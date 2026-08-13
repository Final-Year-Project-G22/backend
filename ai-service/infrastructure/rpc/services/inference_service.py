from __future__ import annotations

import json
import logging
import time
import uuid
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
            try:
                request_id = uuid.UUID(request.request_id)
                user_id = uuid.UUID(request.user_id)
                account_id = uuid.UUID(request.account_id)

                session_id = None
                if request.session_id:
                    session_id = uuid.UUID(request.session_id)

                language = Language.ENGLISH
                if request.language:
                    language = Language(request.language)
            except ValueError as e:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID or enum: {e}")  # type: ignore
                return service_pb2.AskResponse()  # type: ignore

            strategy = getattr(request, "strategy", "simple") or "simple"
            debug_mode = getattr(request, "debug_mode", False)

            domain_req = AskAICommand(
                user_id=user_id,
                account_id=account_id,
                prompt=request.query,
                language=language,
                conversation_id=session_id,
                title=request.title if getattr(request, "title", "") else None,
                vector_top_k=request.top_k if request.top_k > 0 else 5,
                strategy=strategy,
                debug_mode=debug_mode,
            )

            response = await self._ask_ai_usecase.execute(domain_req)

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

            usage = None
            if response.ai_message.token_usage:
                usage = service_pb2.Usage(  # type: ignore
                    prompt_tokens=response.ai_message.token_usage.prompt_tokens,
                    completion_tokens=response.ai_message.token_usage.completion_tokens,
                    total_tokens=response.ai_message.token_usage.total_tokens,
                )

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

    async def AskStream(self, request: Any, context: Any) -> Any:  # type: ignore[override]  # noqa: N802, PLR0912, PLR0915
        if not await self._ensure_ask_enabled(context):
            return

        start_time = time.perf_counter()

        try:
            user_id = uuid.UUID(request.user_id)
            account_id = uuid.UUID(request.account_id)

            session_id = None
            if request.session_id:
                session_id = uuid.UUID(request.session_id)

            language = Language.ENGLISH
            if request.language:
                language = Language(request.language)

            strategy = getattr(request, "strategy", "simple") or "simple"
            debug_mode = getattr(request, "debug_mode", False)

            command = AskAICommand(
                user_id=user_id,
                account_id=account_id,
                prompt=request.query,
                language=language,
                conversation_id=session_id,
                title=request.title if getattr(request, "title", "") else None,
                vector_top_k=request.top_k if request.top_k > 0 else 5,
                strategy=strategy,
                debug_mode=debug_mode,
            )

            ai_message = None
            merged_hits = None
            conversation = None
            llm_port = self._ask_ai_usecase._llm_port

            async for event in self._ask_ai_usecase.execute_stream_with_tools(
                command,
            ):
                if event.is_text and event.text:
                    yield service_pb2.AskStreamChunk(text=service_pb2.TextChunk(text=event.text))

                if event.is_tool_call and event.tool_name:
                    yield service_pb2.AskStreamChunk(
                        tool_use=service_pb2.ToolUseChunk(
                            tool=event.tool_name,
                            arguments_json=json.dumps(event.tool_arguments or {}),
                        )
                    )

                if event.is_tool_result and event.tool_name:
                    yield service_pb2.AskStreamChunk(
                        tool_result=service_pb2.ToolResultChunk(
                            tool=event.tool_name,
                            result_summary=event.tool_result_summary or "",
                        )
                    )

                if event.is_thinking and event.text and debug_mode:
                    yield service_pb2.AskStreamChunk(
                        thinking=service_pb2.ThinkingChunk(text=event.text),
                    )

                if event.is_tool_suppressed and event.tool_name and debug_mode:
                    yield service_pb2.AskStreamChunk(
                        tool_suppressed=service_pb2.ToolSuppressedChunk(
                            tool=event.tool_name,
                            reason=event.suppression_reason or "",
                            matched_query=event.matched_query or "",
                        )
                    )

                if event.is_done:
                    ai_message = event.ai_message
                    merged_hits = event.merged_hits
                    conversation = event.done

            if ai_message is None or conversation is None:
                logger.error("execute_stream_with_tools finished without done event")
                return

            citations = [
                service_pb2.Citation(  # type: ignore
                    document_id=str(hit.document_id),
                    chunk_id=str(hit.chunk_id),
                    source_type="chunk",
                    score=hit.score,
                    title=getattr(hit, "document_title", None) or f"Chunk {hit.chunk_index}",
                    excerpt=hit.chunk_text[:300],
                )
                for hit in (merged_hits or [])
            ]

            yield service_pb2.AskStreamChunk(
                citations=service_pb2.CitationsChunk(citations=citations)
            )

            usage = None
            if ai_message.token_usage:
                usage = service_pb2.Usage(  # type: ignore
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
                error=service_pb2.ErrorChunk(code="RESOURCE_EXHAUSTED", message=str(e))
            )
        except ValidationError as e:
            detail = _validation_error_detail(e)
            logger.info("AskStream validation failed: %s", detail)
            yield service_pb2.AskStreamChunk(
                error=service_pb2.ErrorChunk(code="INVALID_ARGUMENT", message=detail)
            )
        except Exception:
            logger.exception("Error in AskStream")
            yield service_pb2.AskStreamChunk(
                error=service_pb2.ErrorChunk(code="INTERNAL", message="Internal server error")
            )
