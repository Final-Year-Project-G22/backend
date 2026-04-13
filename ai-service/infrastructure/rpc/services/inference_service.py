from __future__ import annotations

import logging
import uuid
from typing import Any

from ai.inference.v1 import service_pb2, service_pb2_grpc  # type: ignore
from typing_extensions import override

import grpc
from core.domain.enums import Language
from core.domain.exceptions import AIServiceError, QuotaExceededError, RepositoryError
from core.usecases.ask_ai import AskAIUseCase
from core.usecases.contracts import AskAICommand

logger = logging.getLogger(__name__)


class AIInferenceService(service_pb2_grpc.AIInferenceServiceServicer):
    def __init__(self, ask_ai_usecase: AskAIUseCase):
        self._ask_ai_usecase = ask_ai_usecase

    @override
    async def Ask(self, request: Any, context: Any) -> Any:
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
                answer=response.ai_message.llm_response or "",
                citations=citations,
                usage=usage,
                model="default",
                latency_ms=0,
            )

        except QuotaExceededError as e:
            logger.warning("Quota exceeded: %s", e)
            await context.abort(grpc.StatusCode.RESOURCE_EXHAUSTED, str(e))  # type: ignore
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
