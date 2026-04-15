from __future__ import annotations

import logging
from typing import Any

logger = logging.getLogger(__name__)


class IngestionRequestedTaskHandler:
    async def handle(self, payload: dict[str, Any]) -> None:
        event_id = str(payload.get("event_id", ""))
        event_type = str(payload.get("event_type", ""))
        logger.info(
            "ingestion event received event_id=%s event_type=%s",
            event_id,
            event_type,
        )


__all__ = ["IngestionRequestedTaskHandler"]
