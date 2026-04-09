from __future__ import annotations

import uuid
from types import SimpleNamespace
from unittest.mock import AsyncMock

import grpc
import pytest

from core.domain.exceptions import AIServiceError, QuotaExceededError, RepositoryError
from core.usecases.contracts import AskAICommand
from infrastructure.rpc.services.inference_service import AIInferenceService

TOP_K_INPUT = 7
PROMPT_TOKENS = 11
COMPLETION_TOKENS = 22
TOTAL_TOKENS = 33


def _make_request(**overrides: object) -> SimpleNamespace:
    payload: dict[str, object] = {
        "request_id": str(uuid.uuid4()),
        "user_id": str(uuid.uuid4()),
        "account_id": str(uuid.uuid4()),
        "query": "How do I register a business?",
        "language": "en",
        "session_id": "",
        "top_k": 5,
    }
    payload.update(overrides)
    return SimpleNamespace(**payload)


@pytest.mark.asyncio
async def test_ask_maps_successful_response() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase)

    convo_id = uuid.uuid4()
    hit_document_id = uuid.uuid4()
    hit_chunk_id = uuid.uuid4()
    usecase.execute.return_value = SimpleNamespace(
        conversation=SimpleNamespace(id=convo_id),
        ai_message=SimpleNamespace(
            llm_response="Start by obtaining your trade license.",
            token_usage=SimpleNamespace(
                prompt_tokens=PROMPT_TOKENS,
                completion_tokens=COMPLETION_TOKENS,
                total_tokens=TOTAL_TOKENS,
            ),
        ),
        retrieved_hits=[
            SimpleNamespace(
                document_id=hit_document_id,
                chunk_id=hit_chunk_id,
                score=0.87,
            )
        ],
    )

    request = _make_request(language="am", top_k=TOP_K_INPUT)
    context = AsyncMock()

    response = await service.Ask(request, context)

    usecase.execute.assert_awaited_once()
    command = usecase.execute.await_args.args[0]
    assert isinstance(command, AskAICommand)
    assert command.user_id == uuid.UUID(request.user_id)
    assert command.prompt == request.query
    assert command.vector_top_k == TOP_K_INPUT

    assert response.request_id == request.request_id
    assert response.session_id == str(convo_id)
    assert response.answer == "Start by obtaining your trade license."
    assert len(response.citations) == 1
    assert response.citations[0].document_id == str(hit_document_id)
    assert response.citations[0].chunk_id == str(hit_chunk_id)
    assert response.usage.prompt_tokens == PROMPT_TOKENS
    assert response.usage.total_tokens == TOTAL_TOKENS


@pytest.mark.asyncio
async def test_ask_aborts_for_invalid_uuid() -> None:
    usecase = AsyncMock()
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request(user_id="not-a-uuid")
    context = AsyncMock()

    await service.Ask(request, context)

    context.abort.assert_awaited_once()
    assert context.abort.await_args.args[0] is grpc.StatusCode.INVALID_ARGUMENT
    usecase.execute.assert_not_awaited()


@pytest.mark.asyncio
@pytest.mark.parametrize(
    ("error", "status_code"),
    [
        (QuotaExceededError("quota reached"), grpc.StatusCode.RESOURCE_EXHAUSTED),
        (RepositoryError("db down"), grpc.StatusCode.INTERNAL),
        (AIServiceError("provider failed"), grpc.StatusCode.INTERNAL),
    ],
)
async def test_ask_maps_domain_errors_to_grpc_status(
    error: Exception, status_code: grpc.StatusCode
) -> None:
    usecase = AsyncMock()
    usecase.execute.side_effect = error
    service = AIInferenceService(ask_ai_usecase=usecase)

    request = _make_request()
    context = AsyncMock()

    await service.Ask(request, context)

    context.abort.assert_awaited_once()
    assert context.abort.await_args.args[0] is status_code
