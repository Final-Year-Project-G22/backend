from __future__ import annotations

import logging
import time
import uuid
from datetime import UTC, datetime
from typing import Any

from ai.inference.v1 import service_pb2, service_pb2_grpc  # type: ignore
from pydantic import ValidationError

import grpc
from core.domain.enums import Language
from core.domain.exceptions import AIServiceError, QuotaExceededError, RepositoryError
from core.usecases.ask_ai import AskAIUseCase
from core.usecases.contracts import AskAICommand

logger = logging.getLogger(__name__)

_MAX_VALIDATION_MSG = 512


def _validation_error_detail(exc: ValidationError) -> str:
    parts: list[str] = []
    for err in exc.errors():
        loc = ".".join(str(x) for x in err["loc"])
        parts.append(f"{loc}: {err['msg']}")
    msg = "; ".join(parts) if parts else str(exc)
    if len(msg) > _MAX_VALIDATION_MSG:
        return msg[: _MAX_VALIDATION_MSG - 3] + "..."
    return msg


class AIInferenceService(service_pb2_grpc.AIInferenceServiceServicer):  # type: ignore
    def __init__(self, ask_ai_usecase: AskAIUseCase, *, ask_enabled: bool = True):
        self._ask_ai_usecase = ask_ai_usecase
        self._ask_enabled = ask_enabled

    async def _ensure_ask_enabled(self, context: Any) -> bool:
        if self._ask_enabled:
            return True
        await context.abort(grpc.StatusCode.UNAVAILABLE, "Ask API is disabled")
        return False

    async def Ask(self, request: Any, context: Any) -> Any:  # type: ignore[override]  # noqa: N802
        if not await self._ensure_ask_enabled(context):
            return service_pb2.AskResponse()  # type: ignore

        try:
            # Parse request fields
            try:
                request_id = uuid.UUID(request.request_id)
                user_id = uuid.UUID(request.user_id)
                _account_id = uuid.UUID(request.account_id)  # currently unused in domain

                session_id = None
                if request.session_id:
                    session_id = uuid.UUID(request.session_id)

                language = Language.ENGLISH
                if request.language:
                    language = Language(request.language)
            except ValueError as e:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID or enum: {e}")  # type: ignore
                return service_pb2.AskResponse()  # type: ignore

            domain_req = AskAICommand(
                user_id=user_id,
                prompt=request.query,
                language=language,
                conversation_id=session_id,
                title=request.title if getattr(request, "title", "") else None,
                vector_top_k=request.top_k if request.top_k > 0 else 5,
            )

            # Execute use case
            response = await self._ask_ai_usecase.execute(domain_req)

            # Map citations
            citations: list[Any] = [
                service_pb2.Citation(  # type: ignore
                    document_id=str(hit.document_id),
                    chunk_id=str(hit.chunk_id),
                    source_type="chunk",
                    score=hit.score,
                    title=getattr(hit, "document_title", None) or f"Chunk {hit.chunk_index}",
                    excerpt=hit.chunk_text[:300],
                )
                for hit in response.retrieved_hits
            ]

            # Map usage
            usage = None
            if response.ai_message.token_usage:
                usage = service_pb2.Usage(  # type: ignore
                    prompt_tokens=response.ai_message.token_usage.prompt_tokens,
                    completion_tokens=response.ai_message.token_usage.completion_tokens,
                    total_tokens=response.ai_message.token_usage.total_tokens,
                )

            # Return response
            return service_pb2.AskResponse(  # type: ignore
                request_id=str(request_id),
                session_id=str(response.conversation.id),
                session_created_at=response.conversation.created_at.isoformat(),
                session_updated_at=response.conversation.updated_at.isoformat(),
                answer=response.ai_message.llm_response or "",
                citations=citations,
                usage=usage,
                model="default",
                latency_ms=0,
            )

        except QuotaExceededError as e:
            logger.warning("Quota exceeded: %s", e)
            await context.abort(grpc.StatusCode.RESOURCE_EXHAUSTED, str(e))  # type: ignore
        except ValidationError as e:
            detail = _validation_error_detail(e)
            logger.info("Ask validation failed: %s", detail)
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, detail)  # type: ignore
        except RepositoryError:
            logger.exception("Repository error in Ask endpoint")
            await context.abort(grpc.StatusCode.INTERNAL, "Internal data error")  # type: ignore
        except AIServiceError:
            logger.exception("AI provider error in Ask endpoint")
            await context.abort(grpc.StatusCode.INTERNAL, "AI provider error")  # type: ignore
        except Exception:
            logger.exception("Unexpected error in Ask endpoint")
            await context.abort(grpc.StatusCode.INTERNAL, "Internal server error")  # type: ignore

        return service_pb2.AskResponse()  # type: ignore

    async def AskStream(self, request: Any, context: Any) -> Any:  # type: ignore[override]  # noqa: N802
        # context arg required by gRPC streaming signature
        if not await self._ensure_ask_enabled(context):
            return

        start_time = time.perf_counter()

        try:
            user_id = uuid.UUID(request.user_id)
            _account_id = uuid.UUID(request.account_id)

            session_id = None
            if request.session_id:
                session_id = uuid.UUID(request.session_id)

            language = Language.ENGLISH
            if request.language:
                language = Language(request.language)

            command = AskAICommand(
                user_id=user_id,
                prompt=request.query,
                language=language,
                conversation_id=session_id,
                title=request.title if getattr(request, "title", "") else None,
                vector_top_k=request.top_k if request.top_k > 0 else 5,
            )

            now = datetime.now(UTC)
            conversation = await self._ask_ai_usecase._resolve_conversation(command, now)
            user_message = await self._ask_ai_usecase._persist_user_message(
                command, conversation, now
            )

            query_embedding = await self._ask_ai_usecase._embedding_port.embed_query(command.prompt)
            user_message = await self._ask_ai_usecase._update_user_message_embedding(
                user_message,
                query_embedding,
                now,
            )

            vector_hits, bm25_hits = await self._ask_ai_usecase._retrieve_context(
                command, query_embedding
            )
            merged_hits = self._ask_ai_usecase._merge_and_dedupe_hits(vector_hits, bm25_hits)

            prompt = self._ask_ai_usecase._build_prompt(command.prompt, merged_hits)

            llm_port = self._ask_ai_usecase._llm_port
            full_response_parts = []
            async for chunk in llm_port.generate_stream(prompt):
                full_response_parts.append(chunk)
                yield service_pb2.AskStreamChunk(text=service_pb2.TextChunk(text=chunk))

            full_response = "".join(full_response_parts)
            ai_message = await self._ask_ai_usecase._persist_ai_message(
                command,
                conversation,
                full_response,
                now,
                retrieved_hits=merged_hits,
            )

            cache_key = self._ask_ai_usecase._build_cache_key(command, conversation.id)
            await self._ask_ai_usecase._cache_response(cache_key, full_response)
            await self._ask_ai_usecase._publish_query_event(command, conversation, ai_message)

            citations = [
                service_pb2.Citation(
                    document_id=str(hit.document_id),
                    chunk_id=str(hit.chunk_id),
                    source_type="chunk",
                    score=hit.score,
                    title=getattr(hit, "document_title", None) or f"Chunk {hit.chunk_index}",
                    excerpt=hit.chunk_text[:300],
                )
                for hit in merged_hits
            ]

            yield service_pb2.AskStreamChunk(
                citations=service_pb2.CitationsChunk(citations=citations)
            )

            usage = None
            if ai_message.token_usage:
                usage = service_pb2.Usage(
                    prompt_tokens=ai_message.token_usage.prompt_tokens,
                    completion_tokens=ai_message.token_usage.completion_tokens,
                    total_tokens=ai_message.token_usage.total_tokens,
                )

            latency_ms = int((time.perf_counter() - start_time) * 1000)
            yield service_pb2.AskStreamChunk(
                done=service_pb2.DoneChunk(
                    latency_ms=latency_ms,
                    model=llm_port.model,
                    usage=usage,
                    session_id=str(conversation.id),
                    session_created_at=conversation.created_at.isoformat(),
                    session_updated_at=conversation.updated_at.isoformat(),
                )
            )

        except QuotaExceededError as e:
            logger.warning("Quota exceeded in AskStream: %s", e)
            yield service_pb2.AskStreamChunk(
                error=service_pb2.ErrorChunk(
                    code="RESOURCE_EXHAUSTED",
                    message=str(e),
                )
            )
        except ValidationError as e:
            detail = _validation_error_detail(e)
            logger.info("AskStream validation failed: %s", detail)
            yield service_pb2.AskStreamChunk(
                error=service_pb2.ErrorChunk(
                    code="INVALID_ARGUMENT",
                    message=detail,
                )
            )
        except Exception:
            logger.exception("Error in AskStream")
            yield service_pb2.AskStreamChunk(
                error=service_pb2.ErrorChunk(
                    code="INTERNAL",
                    message="Internal server error",
                )
            )
