"""Compute deterministic chunk refs for the eval fixture corpus (offline, no DB/API).

Runs the real ParserRegistry + ChunkingRegistry with the production chunking
strategy (structural, max 512 tokens, overlap 50) exactly as the ingestion
orchestrator does, and prints per-document chunk counts and token counts.
"""

from __future__ import annotations

import asyncio
import json
import sys
from pathlib import Path
from typing import Any

sys.path.insert(0, str(Path(__file__).resolve().parents[3]))

from core.ports.chunking import ChunkingStrategy
from infrastructure.chunking.registry import ChunkingRegistry
from infrastructure.parsers.registry import ParserRegistry

CORPUS = Path(__file__).resolve().parents[3] / "evals" / "fixture" / "corpus"
CONTENT_TYPES = {
    ".md": "text/markdown",
}
CHUNKING_STRATEGY = ChunkingStrategy(max_tokens=512, overlap_tokens=50)


# Stable mapping from corpus file stem -> knowledge_documents.source value.
# Frozen with the fixture: changing a source value changes the document_key.
SOURCE_BY_STEM = {
    "vat-registration": "tax_code",
    "tin-registration": "tax_code",
    "business-registration": "legal",
    "trade-license": "government",
    "compliance-deadlines": "guide",
    "formalization-overview": "guide",
}


def document_key_for(path: Path) -> str:
    parts = path.parts
    locale = parts[-2]
    stem = path.stem
    source = SOURCE_BY_STEM[stem]
    return f"{locale}:{source}:{stem}"


async def main() -> None:
    parser_registry = ParserRegistry()
    chunking_registry = ChunkingRegistry()
    results: dict[str, dict[str, Any]] = {}
    for path in sorted(CORPUS.glob("*/*.md")):
        content_type = CONTENT_TYPES[path.suffix]
        parser = parser_registry.get_parser(content_type)
        assert parser is not None, f"no parser for {content_type}"  # nosec B101 - operator script guard
        parsed = await parser.parse(path.read_bytes(), metadata={"filename": path.name})
        chunks = await chunking_registry.chunk("structural", parsed, CHUNKING_STRATEGY)
        results[document_key_for(path)] = {
            "path": str(path.relative_to(CORPUS)),
            "chunk_count": len(chunks),
            "token_counts": [c.token_count for c in chunks],
            "section_count": len(parsed.sections),
        }
        print(
            f"{document_key_for(path):40s} chunks={len(chunks):2d} tokens={sum(c.token_count for c in chunks)}"
        )

    out = Path(__file__).resolve().parents[3] / "evals" / "fixture" / "chunk_refs.json"
    out.write_text(json.dumps(results, indent=2) + "\n")
    print(f"\nwrote {out}")


if __name__ == "__main__":
    asyncio.run(main())
