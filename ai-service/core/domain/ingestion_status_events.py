from __future__ import annotations

from typing import Any

from core.domain.enums import IngestionStage


def build_status_updated_payload(
    document_id: str,
    from_stage: IngestionStage | None,
    to_stage: IngestionStage,
    *,
    is_terminal: bool = False,
    retry_count: int = 0,
    error_message: str | None = None,
    chunks_processed_count: int | None = None,
    chunks_failed_count: int | None = None,
    metadata: dict[str, Any] | None = None,
) -> dict[str, Any]:
    status_updated: dict[str, Any] = {
        "document_id": document_id,
        "from_stage": from_stage.value if from_stage else None,
        "to_stage": to_stage.value,
        "is_terminal": is_terminal,
        "retry_count": retry_count,
    }

    if error_message is not None:
        status_updated["error_message"] = error_message

    if chunks_processed_count is not None:
        status_updated["chunks_processed_count"] = chunks_processed_count

    if chunks_failed_count is not None:
        status_updated["chunks_failed_count"] = chunks_failed_count

    if metadata is not None:
        status_updated["metadata"] = metadata

    return {
        "status_updated": status_updated,
    }


__all__ = ["build_status_updated_payload"]
