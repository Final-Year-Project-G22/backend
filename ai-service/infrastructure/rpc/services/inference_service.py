from __future__ import annotations

import json
import logging
import time
import uuid
from collections.abc import AsyncIterator

import grpc
import grpc.aio
from ai.inference.v1 import service_pb2, service_pb2_grpc
from pydantic import ValidationError

from core.domain.enums import Language
from core.domain.exceptions import AIServiceError, QuotaExceededError, RepositoryError
from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.value_objects import SearchHit
from core.usecases.ask_ai import AskAIUseCase
from core.usecases.contracts import AskAICommand
from infrastructure.rpc.services._mappers import map_citation_from_hit, map_usage

logger = logging.getLogger(__name__)

_MAX_VALIDATION_MSG = 512

_AskContext = grpc.aio.ServicerContext[service_pb2.AskRequest, service_pb2.AskResponse]
_AskStreamContext = grpc.aio.ServicerContext[service_pb2.AskRequest, service_pb2.AskStreamChunk]


def _validation_error_detail(exc: ValidationError) -> str:
    parts: list[str] = []
    for err in exc.errors():
        loc = ".".join(str(x) for x in err["loc"])
        parts.append(f"{loc}: {err['msg']}")
    msg = "; ".join(parts) if parts else str(exc)
    if len(msg) > _MAX_VALIDATION_MSG:
        return msg[: _MAX_VALIDATION_MSG - 3] + "..."
    return msg


class AIInferenceService(service_pb2_grpc.AIInferenceServiceServicer):
    def __init__(self, ask_ai_usecase: AskAIUseCase, *, ask_enabled: bool = True) -> None:
        self._ask_ai_usecase = ask_ai_usecase
        self._ask_enabled = ask_enabled

    async def Ask(  # noqa: N802
        self,
        request: service_pb2.AskRequest,
        context: _AskContext,
    ) -> service_pb2.AskResponse:
        if not self._ask_enabled:
            await context.abort(grpc.StatusCode.UNAVAILABLE, "Ask API is disabled")

        try:
            try:
                request_id = uuid.UUID(request.request_id)
                user_id = uuid.UUID(request.user_id)
                account_id = uuid.UUID(request.account_id)

                session_id: uuid.UUID | None = None
                if request.session_id:
                    session_id = uuid.UUID(request.session_id)

                language = Language.ENGLISH
                if request.language:
                    language = Language(request.language)
            except ValueError as e:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID or enum: {e}")

            strategy = request.strategy or "simple"
            debug_mode = request.debug_mode

            domain_req = AskAICommand(
                user_id=user_id,
                account_id=account_id,
                prompt=request.query,
                language=language,
                conversation_id=session_id,
                title=request.title or None,
                vector_top_k=request.top_k if request.top_k > 0 else 5,
                strategy=strategy,
                debug_mode=debug_mode,
            )

            response = await self._ask_ai_usecase.execute(domain_req)

            citations: list[service_pb2.Citation] = [
                map_citation_from_hit(hit) for hit in response.retrieved_hits
            ]

            usage: service_pb2.Usage | None = None
            if response.ai_message.token_usage:
                usage = map_usage(response.ai_message.token_usage)

            return service_pb2.AskResponse(
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

        except grpc.aio.AbortError:
            # context.abort raises AbortError to terminate the RPC; it must
            # not be swallowed by the generic handler below.
            raise
        except QuotaExceededError as e:
            logger.warning("Quota exceeded: %s", e)
            await context.abort(grpc.StatusCode.RESOURCE_EXHAUSTED, str(e))
        except ValidationError as e:
            detail = _validation_error_detail(e)
            logger.info("Ask validation failed: %s", detail)
            await context.abort(grpc.StatusCode.INVALID_ARGUMENT, detail)
        except RepositoryError:
            logger.exception("Repository error in Ask endpoint")
            await context.abort(grpc.StatusCode.INTERNAL, "Internal data error")
        except AIServiceError:
            logger.exception("AI provider error in Ask endpoint")
            await context.abort(grpc.StatusCode.INTERNAL, "AI provider error")
        except Exception:
            logger.exception("Unexpected error in Ask endpoint")
            await context.abort(grpc.StatusCode.INTERNAL, "Internal server error")

    async def AskStream(  # noqa: N802, PLR0912, PLR0915
        self,
        request: service_pb2.AskRequest,
        context: _AskStreamContext,
    ) -> AsyncIterator[service_pb2.AskStreamChunk]:
        if not self._ask_enabled:
            await context.abort(grpc.StatusCode.UNAVAILABLE, "Ask API is disabled")

        start_time = time.perf_counter()

        try:
            user_id = uuid.UUID(request.user_id)
            account_id = uuid.UUID(request.account_id)

            session_id: uuid.UUID | None = None
            if request.session_id:
                session_id = uuid.UUID(request.session_id)

            language = Language.ENGLISH
            if request.language:
                language = Language(request.language)

            strategy = request.strategy or "simple"
            debug_mode = request.debug_mode

            command = AskAICommand(
                user_id=user_id,
                account_id=account_id,
                prompt=request.query,
                language=language,
                conversation_id=session_id,
                title=request.title or None,
                vector_top_k=request.top_k if request.top_k > 0 else 5,
                strategy=strategy,
                debug_mode=debug_mode,
            )

            ai_message: AIChatMessage | None = None
            merged_hits: list[SearchHit] | None = None
            conversation: AIConversationSession | None = None
            llm_model = self._ask_ai_usecase.llm_model

            async for event in self._ask_ai_usecase.execute_stream_with_tools(command):
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

            citations: list[service_pb2.Citation] = [
                map_citation_from_hit(hit) for hit in (merged_hits or [])
            ]

            yield service_pb2.AskStreamChunk(
                citations=service_pb2.CitationsChunk(citations=citations)
            )

            usage: service_pb2.Usage | None = None
            if ai_message.token_usage:
                usage = map_usage(ai_message.token_usage)

            latency_ms = int((time.perf_counter() - start_time) * 1000)
            yield service_pb2.AskStreamChunk(
                done=service_pb2.DoneChunk(
                    latency_ms=latency_ms,
                    model=llm_model or "default",
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
