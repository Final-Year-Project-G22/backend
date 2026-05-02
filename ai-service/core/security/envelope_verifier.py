from __future__ import annotations

import hashlib
import hmac
import json
import logging
from typing import Any

from core.domain.ingestion_events import ACCEPTED_ENVELOPE_SCHEMA_VERSIONS


class EnvelopeVerificationError(ValueError):
    pass


class EnvelopeVerifier:
    def __init__(
        self, active_key_id: str, active_secret: str, previous_keys: dict[str, str] | None = None
    ) -> None:
        self._keys: dict[str, str] = {}
        if active_key_id and active_secret:
            self._keys[active_key_id] = active_secret
        if previous_keys:
            self._keys.update(previous_keys)

    def verify(self, envelope: dict[str, Any]) -> None:
        logger = logging.getLogger(__name__)
        schema_version = str(envelope.get("schema_version", ""))
        if schema_version not in ACCEPTED_ENVELOPE_SCHEMA_VERSIONS:
            logger.error(
                "EnvelopeVerificationError: unsupported schema_version: %s (accepted: %s)",
                schema_version,
                ACCEPTED_ENVELOPE_SCHEMA_VERSIONS,
            )
            raise EnvelopeVerificationError(f"unsupported schema_version: {schema_version}")

        key_id = str(envelope.get("key_id", ""))
        signature = str(envelope.get("signature", ""))
        if not key_id or not signature:
            logger.error(
                "EnvelopeVerificationError: missing key_id=%r or signature=%r",
                key_id,
                bool(signature),
            )
            raise EnvelopeVerificationError("missing key_id or signature")

        secret = self._keys.get(key_id)
        if not secret:
            logger.error(
                "EnvelopeVerificationError: unknown key_id=%r (available: %s)",
                key_id,
                list(self._keys.keys()),
            )
            raise EnvelopeVerificationError(f"unknown key_id: {key_id}")

        payload = dict(envelope)
        payload.pop("signature", None)
        payload.pop("key_id", None)
        canonical = _canonical_json(payload)
        expected = hmac.new(secret.encode("utf-8"), canonical, hashlib.sha256).hexdigest()
        if not hmac.compare_digest(expected, signature):
            logger.error(
                "EnvelopeVerificationError: signature mismatch for key_id=%r "
                "expected=%s... got=%s... canonical_len=%d",
                key_id,
                expected[:16],
                signature[:16],
                len(canonical),
            )
            raise EnvelopeVerificationError("invalid envelope signature")


def _canonical_json(value: Any) -> bytes:
    return json.dumps(value, separators=(",", ":"), sort_keys=True).encode("utf-8")
