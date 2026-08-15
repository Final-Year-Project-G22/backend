from __future__ import annotations

import json
import logging
import time
import uuid
from collections.abc import AsyncIterator
from typing import NamedTuple

import grpc
import grpc.aio
from ai.inference.v1 import service_pb2, service_pb2_grpc
from pydantic import ValidationError

from core.domain.enums import Language
from core.domain.exceptions import AIServiceError, QuotaExceededError, RepositoryError
from core.domain.models import AIChatMessage, AIConversationSession
from core.domain.stream_events import AskStreamEvent
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


class _ParsedRequest(NamedTuple):
    user_id: uuid.UUID
    account_id: uuid.UUID
    session_id: uuid.UUID | None
    language: Language


class AIInferenceService(service_pb2_grpc.AIInferenceServiceServicer):
    def __init__(self, ask_ai_usecase: AskAIUseCase, *, ask_enabled: bool = True) -> None:
        self._ask_ai_usecase = ask_ai_usecase
        self._ask_enabled = ask_enabled

    def _parse_request_ids(self, request: service_pb2.AskRequest) -> _ParsedRequest:
        user_id = uuid.UUID(request.user_id)
        account_id = uuid.UUID(request.account_id)

        session_id: uuid.UUID | None = None
        if request.session_id:
            session_id = uuid.UUID(request.session_id)

        language = Language.ENGLISH
        if request.language:
            language = Language(request.language)

        return _ParsedRequest(
            user_id=user_id,
            account_id=account_id,
            session_id=session_id,
            language=language,
        )

    def _build_command(
        self,
        request: service_pb2.AskRequest,
        parsed: _ParsedRequest,
    ) -> AskAICommand:
        return AskAICommand(
            user_id=parsed.user_id,
            account_id=parsed.account_id,
            prompt=request.query,
            language=parsed.language,
            conversation_id=parsed.session_id,
            title=request.title or None,
            vector_top_k=request.top_k if request.top_k > 0 else 5,
            strategy=request.strategy or "simple",
            debug_mode=request.debug_mode,
        )

    def _map_event_chunk(
        self, event: AskStreamEvent, *, debug_mode: bool
    ) -> service_pb2.AskStreamChunk | None:
        if event.is_text and event.text:
            return self._text_chunk(event.text)
        if event.is_tool_call and event.tool_name:
            return self._tool_use_chunk(event.tool_name, event.tool_arguments or {})
        if event.is_tool_result and event.tool_name:
            return self._tool_result_chunk(event.tool_name, event.tool_result_summary or "")
        if event.is_thinking and event.text and debug_mode:
            return self._thinking_chunk(event.text)
        if event.is_tool_suppressed and event.tool_name and debug_mode:
            return self._tool_suppressed_chunk(
                event.tool_name,
                event.suppression_reason or "",
                event.matched_query or "",
            )
        return None

    def _text_chunk(self, text: str) -> service_pb2.AskStreamChunk:
        return service_pb2.AskStreamChunk(text=service_pb2.TextChunk(text=text))

    def _tool_use_chunk(
        self, tool_name: str, arguments: dict[str, object]
    ) -> service_pb2.AskStreamChunk:
        return service_pb2.AskStreamChunk(
            tool_use=service_pb2.ToolUseChunk(
                tool=tool_name,
                arguments_json=json.dumps(arguments),
            )
        )

    def _tool_result_chunk(self, tool_name: str, result_summary: str) -> service_pb2.AskStreamChunk:
        return service_pb2.AskStreamChunk(
            tool_result=service_pb2.ToolResultChunk(
                tool=tool_name,
                result_summary=result_summary,
            )
        )

    def _thinking_chunk(self, text: str) -> service_pb2.AskStreamChunk:
        return service_pb2.AskStreamChunk(thinking=service_pb2.ThinkingChunk(text=text))

    def _tool_suppressed_chunk(
        self, tool_name: str, reason: str, matched_query: str
    ) -> service_pb2.AskStreamChunk:
        return service_pb2.AskStreamChunk(
            tool_suppressed=service_pb2.ToolSuppressedChunk(
                tool=tool_name,
                reason=reason,
                matched_query=matched_query,
            )
        )

    def _citations_chunk(self, merged_hits: list[SearchHit]) -> service_pb2.AskStreamChunk:
        citations = [map_citation_from_hit(hit) for hit in merged_hits]
        return service_pb2.AskStreamChunk(citations=service_pb2.CitationsChunk(citations=citations))

    def _done_chunk(
        self,
        conversation: AIConversationSession,
        ai_message: AIChatMessage,
        llm_model: str | None,
        start_time: float,
    ) -> service_pb2.AskStreamChunk:
        usage: service_pb2.Usage | None = None
        if ai_message.token_usage:
            usage = map_usage(ai_message.token_usage)
        latency_ms = int((time.perf_counter() - start_time) * 1000)
        return service_pb2.AskStreamChunk(
            done=service_pb2.DoneChunk(
                latency_ms=latency_ms,
                model=llm_model or "default",
                usage=usage,
                session_id=str(conversation.id),
                session_created_at=conversation.created_at.isoformat(),
                session_updated_at=conversation.updated_at.isoformat(),
            )
        )

    def _error_chunk(self, code: str, message: str) -> service_pb2.AskStreamChunk:
        return service_pb2.AskStreamChunk(error=service_pb2.ErrorChunk(code=code, message=message))

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
                parsed = self._parse_request_ids(request)
            except ValueError as e:
                await context.abort(grpc.StatusCode.INVALID_ARGUMENT, f"Invalid UUID or enum: {e}")

            command = self._build_command(request, parsed)

            response = await self._ask_ai_usecase.execute(command)

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

    async def AskStream(  # noqa: N802
        self,
        request: service_pb2.AskRequest,
        context: _AskStreamContext,
    ) -> AsyncIterator[service_pb2.AskStreamChunk]:
        if not self._ask_enabled:
            await context.abort(grpc.StatusCode.UNAVAILABLE, "Ask API is disabled")

        start_time = time.perf_counter()

        try:
            parsed = self._parse_request_ids(request)
            command = self._build_command(request, parsed)

            ai_message: AIChatMessage | None = None
            merged_hits: list[SearchHit] | None = None
            conversation: AIConversationSession | None = None
            llm_model = self._ask_ai_usecase.llm_model

            async for event in self._ask_ai_usecase.execute_stream_with_tools(command):
                if event.is_done:
                    ai_message = event.ai_message
                    merged_hits = event.merged_hits
                    conversation = event.done

                chunk = self._map_event_chunk(event, debug_mode=command.debug_mode)
                if chunk is not None:
                    yield chunk

            if ai_message is None or conversation is None:
                logger.error("execute_stream_with_tools finished without done event")
                return

            yield self._citations_chunk(merged_hits or [])
            yield self._done_chunk(conversation, ai_message, llm_model, start_time)

        except QuotaExceededError as e:
            logger.warning("Quota exceeded in AskStream: %s", e)
            yield self._error_chunk("RESOURCE_EXHAUSTED", str(e))
        except ValidationError as e:
            detail = _validation_error_detail(e)
            logger.info("AskStream validation failed: %s", detail)
            yield self._error_chunk("INVALID_ARGUMENT", detail)
        except Exception:
            logger.exception("Error in AskStream")
            yield self._error_chunk("INTERNAL", "Internal server error")
