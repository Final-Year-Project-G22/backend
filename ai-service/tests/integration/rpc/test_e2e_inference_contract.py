from __future__ import annotations

import uuid
from unittest.mock import AsyncMock

import pytest

from infrastructure.rpc.services.inference_service import AIInferenceService

PROMPT_TOKENS = 5
COMPLETION_TOKENS = 9
TOTAL_TOKENS = 14


@pytest.mark.asyncio
async def test_inference_service_contract_shape_matches_proto_expectations() -> None:
    ask_usecase = AsyncMock()
    conversation_id = uuid.uuid4()
    doc_id = uuid.uuid4()
    chunk_id = uuid.uuid4()

    ask_usecase.execute.return_value = type(
        "Response",
        (),
        {
            "conversation": type("Conversation", (), {"id": conversation_id})(),
            "ai_message": type(
                "AIMessage",
                (),
                {
                    "llm_response": "Use the registration portal.",
                    "token_usage": type(
                        "Usage",
                        (),
                        {
                            "prompt_tokens": PROMPT_TOKENS,
                            "completion_tokens": COMPLETION_TOKENS,
                            "total_tokens": TOTAL_TOKENS,
                        },
                    )(),
                },
            )(),
            "retrieved_hits": [
                type(
                    "Hit",
                    (),
                    {
                        "document_id": doc_id,
                        "chunk_id": chunk_id,
                        "score": 0.76,
                    },
                )(),
            ],
        },
    )()

    request = type(
        "AskRequest",
        (),
        {
            "request_id": str(uuid.uuid4()),
            "user_id": str(uuid.uuid4()),
            "account_id": str(uuid.uuid4()),
            "query": "How do I register?",
            "language": "en",
            "session_id": "",
            "top_k": 3,
        },
    )()

    context = AsyncMock()
    service = AIInferenceService(ask_usecase)
    response = await service.Ask(request, context)

    assert response.request_id == request.request_id
    assert response.session_id == str(conversation_id)
    assert response.answer == "Use the registration portal."
    assert response.model == "default"
    assert response.latency_ms == 0
    assert len(response.citations) == 1
    assert response.citations[0].document_id == str(doc_id)
    assert response.citations[0].chunk_id == str(chunk_id)
    assert response.usage.prompt_tokens == PROMPT_TOKENS
    assert response.usage.completion_tokens == COMPLETION_TOKENS
    assert response.usage.total_tokens == TOTAL_TOKENS
