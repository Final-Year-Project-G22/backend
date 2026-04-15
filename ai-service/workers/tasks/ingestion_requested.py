from __future__ import annotations

import logging
from typing import Any, cast

from core.domain.ingestion_events import (
    ACCEPTED_INGESTION_REQUESTED_SCHEMA_VERSIONS,
    DOCUMENT_INGESTION_REQUESTED_V1,
)
from core.security import EnvelopeVerificationError, EnvelopeVerifier
from infrastructure.messagebus import MessageHandlingRejectError

logger = logging.getLogger(__name__)


class IngestionRequestedTaskHandler:
    def __init__(self, *, envelope_verifier: EnvelopeVerifier) -> None:
        self._envelope_verifier = envelope_verifier

    async def handle(self, payload: dict[str, Any]) -> None:
        try:
            self._envelope_verifier.verify(payload)
        except EnvelopeVerificationError as exc:
            raise MessageHandlingRejectError("invalid ingestion envelope") from exc

        event_type = str(payload.get("event_type", ""))
        if event_type != DOCUMENT_INGESTION_REQUESTED_V1:
            raise MessageHandlingRejectError("unsupported event type")

        payload_message = payload.get("payload")
        if not isinstance(payload_message, dict):
            raise MessageHandlingRejectError("missing envelope payload")
        payload_message_dict = cast(dict[str, Any], payload_message)

        ingestion_requested_payload = payload_message_dict.get("ingestion_requested")
        if not isinstance(ingestion_requested_payload, dict):
            raise MessageHandlingRejectError("missing ingestion_requested payload")
        ingestion_requested_payload_dict = cast(dict[str, Any], ingestion_requested_payload)

        payload_schema_version = str(
            ingestion_requested_payload_dict.get("payload_schema_version", "")
        )
        if payload_schema_version not in ACCEPTED_INGESTION_REQUESTED_SCHEMA_VERSIONS:
            raise MessageHandlingRejectError("unsupported ingestion payload schema")

        event_id = str(payload.get("event_id", ""))
        logger.info(
            "ingestion event received event_id=%s event_type=%s",
            event_id,
            event_type,
        )


__all__ = ["IngestionRequestedTaskHandler"]
