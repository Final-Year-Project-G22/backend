from __future__ import annotations

from typing import TypeGuard


def is_json_object(raw: object) -> TypeGuard[dict[str, object]]:
    """Narrow arbitrary JSON to a string-keyed object.

    Shared by the LLM adapters as the single primitive for validating
    that a payload level is a JSON object before it is inspected further.
    """
    return isinstance(raw, dict)
