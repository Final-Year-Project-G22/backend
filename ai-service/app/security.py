from __future__ import annotations

import json
from typing import Any, cast

from app.config import Settings
from core.security import EnvelopeVerifier


def build_ingestion_envelope_verifier(settings: Settings) -> EnvelopeVerifier:
    previous_keys: dict[str, str] = {}
    raw = settings.INGESTION_SIGNING_PREVIOUS_KEYS_JSON.strip()
    if raw:
        parsed = json.loads(raw)
        if isinstance(parsed, dict):
            parsed_dict = cast(dict[Any, Any], parsed)
            typed_previous_keys: dict[str, str] = {}
            for key, value in parsed_dict.items():
                typed_previous_keys[str(key)] = str(value)
            previous_keys = typed_previous_keys

    return EnvelopeVerifier(
        active_key_id=settings.INGESTION_SIGNING_ACTIVE_KEY_ID,
        active_secret=settings.INGESTION_SIGNING_ACTIVE_KEY_SECRET,
        previous_keys=previous_keys,
    )
