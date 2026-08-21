"""CLI entrypoint: ``uv run python -m evals.cli run [options]``."""

from __future__ import annotations

import argparse
import asyncio
import json
import logging
import shutil
import subprocess  # nosec B404 - only invokes a resolved git binary with fixed argv
import sys
import uuid
from datetime import UTC, datetime
from pathlib import Path
from typing import Any

import httpx

from app.config import Settings
from core.ports.llm import LLMPort
from evals.dataset import DatasetError, build_fixture_index, load_golden_set
from evals.models import (
    CellKey,
    CellScore,
    GoldenItem,
    ItemScores,
    ItemTrace,
    PromptVariant,
    RunManifest,
)
from evals.runner import RunnerError, build_environment, preflight, run_item
from evals.scoring import EvalJudge, ItemScorer, aggregate_cell
from infrastructure.database.connection import engine
from infrastructure.llm import create_llm_adapter
from infrastructure.llm.gemini import GeminiLLMAdapter

logger = logging.getLogger(__name__)

EVALS_DIR = Path(__file__).resolve().parent
SPOT_CHECK_COUNT = 5


def build_judge(settings: Settings, http_client: httpx.AsyncClient) -> EvalJudge:
    """Gemini 2.5 Flash at temperature 0 by default; configured provider as fallback."""
    llm: LLMPort
    if settings.GEMINI_API_KEY.strip():
        llm = GeminiLLMAdapter(
            api_key=settings.GEMINI_API_KEY,
            model=settings.GEMINI_LLM_MODEL,
            http_client=http_client,
            use_vertex=False,
        )
    elif settings.GEMINI_USE_VERTEX and settings.GEMINI_VERTEX_PROJECT.strip():
        llm = GeminiLLMAdapter(
            api_key="",
            model=settings.GEMINI_LLM_MODEL,
            http_client=http_client,
            use_vertex=True,
            vertex_project=settings.GEMINI_VERTEX_PROJECT,
            vertex_location=settings.GEMINI_VERTEX_LOCATION,
        )
    else:
        llm = create_llm_adapter(settings, http_client=http_client)
        logger.warning(
            "no Gemini judge configured; falling back to %s/%s — record this in calibration",
            llm.provider,
            llm.model,
        )
    cache_dir = EVALS_DIR / ".judge_cache"
    return EvalJudge(llm, cache_dir=cache_dir)


def git_sha() -> str:
    git = shutil.which("git")
    if git is None:
        return ""
    try:
        return subprocess.run(  # noqa: S603 # nosec B603 - fixed, resolved executable
            [git, "rev-parse", "--short", "HEAD"],
            capture_output=True,
            text=True,
            cwd=EVALS_DIR.parent,
            check=True,
        ).stdout.strip()
    except (OSError, subprocess.CalledProcessError):
        return ""


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    parser = argparse.ArgumentParser(prog="evals", description="Agentic Ask evaluation runner")
    sub = parser.add_subparsers(dest="command", required=True)

    run = sub.add_parser("run", help="run the golden set end to end")
    run.add_argument("--variant", choices=[v.value for v in PromptVariant], default="clean")
    run.add_argument("--cell", action="append", default=[], help="intent:locale filter, repeatable")
    run.add_argument("--limit", type=int, default=None, help="cap items (after filtering)")
    run.add_argument("--skip-judge", action="store_true", help="deterministic gates only")
    run.add_argument("--concurrency", type=int, default=1)
    run.add_argument(
        "--resume-from",
        type=Path,
        default=None,
        help="items.jsonl from a prior run: reuse successful traces, re-run failed items",
    )
    run.add_argument("--output-dir", type=Path, default=None)

    verify = sub.add_parser("verify-dataset", help="validate golden.jsonl against the fixture")
    verify.add_argument("--golden", type=Path, default=None)
    return parser.parse_args(argv)


def select_items(items: list[GoldenItem], args: argparse.Namespace) -> list[GoldenItem]:
    selected = items
    if args.cell:
        wanted: set[tuple[str, str]] = set()
        for cell in args.cell:
            intent, _, locale = cell.partition(":")
            if not locale:
                msg = f"--cell expects intent:locale, got {cell!r}"
                raise ValueError(msg)
            wanted.add((intent, locale))
        selected = [i for i in selected if (i.intent.value, i.locale.value) in wanted]
    if args.limit is not None:
        selected = selected[: args.limit]
    return selected


async def command_run(args: argparse.Namespace) -> int:
    variant = PromptVariant(args.variant)
    started_at = datetime.now(UTC)

    golden_items = load_golden_set()
    fixture_index = build_fixture_index()
    selected = select_items(golden_items, args)
    if not selected:
        logger.error("no items selected")
        return 2

    environment = build_environment(variant)
    await preflight(environment.container, fixture_index)

    judge: EvalJudge | None = None
    if not args.skip_judge:
        judge = build_judge(environment.settings, environment.http_client)

    scorer = ItemScorer(fixture_index, judge)
    semaphore = asyncio.Semaphore(max(1, args.concurrency))
    prior_traces = _load_prior_traces(args.resume_from)

    async def bounded(item: GoldenItem) -> tuple[ItemTrace, ItemScores]:
        return await _run_and_score(
            item,
            scorer=scorer,
            run=run_item,
            container=environment.container,
            fixture_index=fixture_index,
            semaphore=semaphore,
            prior_traces=prior_traces,
        )

    results = list(await asyncio.gather(*(bounded(item) for item in selected)))
    traces = [trace for trace, _ in results]
    scores = [score for _, score in results]

    finished_at = datetime.now(UTC)
    output_dir = args.output_dir or (
        EVALS_DIR / "results" / f"{started_at.strftime('%Y%m%d-%H%M%S')}-{variant.value}"
    )
    output_dir.mkdir(parents=True, exist_ok=True)

    manifest = _build_manifest(
        output_dir=output_dir.name,
        started_at=started_at,
        finished_at=finished_at,
        variant=variant,
        fixture_version=fixture_index.fixture_version,
        environment=environment,
        judge=judge,
        concurrency=args.concurrency,
        skip_judge=args.skip_judge,
    )

    _write_jsonl(output_dir / "items.jsonl", [trace.model_dump(mode="json") for trace in traces])
    _write_jsonl(output_dir / "scores.jsonl", [score.model_dump(mode="json") for score in scores])

    cells: list[CellScore] = []
    for cell_key in sorted({(item.intent, item.locale) for item in selected}):
        cell_items = [i for i in selected if (i.intent, i.locale) == cell_key]
        cell_scores = [s for s in scores if s.item_id in {i.id for i in cell_items}]
        cells.append(
            aggregate_cell(CellKey(intent=cell_key[0], locale=cell_key[1]), cell_items, cell_scores)
        )
    dumped_cells: list[dict[str, Any]] = [cell.model_dump(mode="json") for cell in cells]
    _write_json(output_dir / "cells.json", dumped_cells)
    _write_json(output_dir / "manifest.json", manifest.model_dump(mode="json"))

    _write_spot_check(output_dir / "human_spot_check.jsonl", selected, traces, scores)

    _print_summary(manifest, dumped_cells, output_dir)
    await environment.http_client.aclose()
    await engine.dispose()
    return 0


async def _run_and_score(
    item: GoldenItem,
    *,
    scorer: ItemScorer,
    run: Any,
    container: Any,
    fixture_index: Any,
    semaphore: asyncio.Semaphore,
    prior_traces: dict[str, ItemTrace],
) -> tuple[ItemTrace, ItemScores]:
    prior = prior_traces.get(item.id)
    if prior is not None and not prior.error:
        logger.info("item %s reused from previous run", item.id)
        trace = prior
    else:
        async with semaphore:
            try:
                trace = await run(item, container, fixture_index)
            except Exception as exc:
                logger.exception("item %s failed", item.id)
                trace = ItemTrace(
                    item_id=item.id, conversation_id=_zero_uuid(), answer="", error=str(exc)
                )
    try:
        scores = await scorer.score(item, trace)
    except Exception as exc:
        logger.exception("scoring item %s failed", item.id)
        scores = ItemScores(
            item_id=item.id,
            hard_fail=True,
            hard_fail_reasons=[f"scoring error: {exc}"],
        )
    return trace, scores


def _build_manifest(
    *,
    output_dir: str,
    started_at: datetime,
    finished_at: datetime,
    variant: PromptVariant,
    fixture_version: str,
    environment: Any,
    judge: EvalJudge | None,
    concurrency: int,
    skip_judge: bool,
) -> RunManifest:
    settings: Settings = environment.settings
    return RunManifest(
        run_id=output_dir,
        started_at=started_at,
        finished_at=finished_at,
        variant=variant,
        dataset_version="1.0.0",
        fixture_version=fixture_version,
        llm_provider=_provider_of(environment),
        llm_model=_model_of(environment),
        embedding_provider=settings.EMBEDDING_PROVIDER,
        embedding_model=settings.COHERE_EMBEDDING_MODEL,
        judge_provider=judge.provider if judge else "none",
        judge_model=judge.model if judge else "none",
        am_faithfulness_trusted=False,
        git_sha=git_sha(),
        config={
            "agentic_max_iterations": settings.AI_AGENTIC_MAX_ITERATIONS,
            "llm_provider_setting": settings.LLM_PROVIDER,
            "concurrency": concurrency,
            "skip_judge": skip_judge,
        },
    )


def _load_prior_traces(resume_path: Path | None) -> dict[str, ItemTrace]:
    if resume_path is None:
        return {}
    traces: dict[str, ItemTrace] = {}
    with resume_path.open(encoding="utf-8") as handle:
        for line in handle:
            stripped = line.strip()
            if not stripped:
                continue
            trace = ItemTrace.model_validate(json.loads(stripped))
            traces[trace.item_id] = trace
    logger.info("resume: %d prior traces loaded from %s", len(traces), resume_path)
    return traces


def _zero_uuid() -> Any:
    return uuid.UUID(int=0)


def _provider_of(environment: Any) -> str:
    try:
        return environment.container.llm_port().provider
    except Exception:
        return environment.settings.LLM_PROVIDER


def _model_of(environment: Any) -> str:
    try:
        return environment.container.llm_port().model
    except Exception:
        return ""


def _write_jsonl(path: Path, rows: list[dict[str, Any]]) -> None:
    with path.open("w", encoding="utf-8") as handle:
        for row in rows:
            handle.write(json.dumps(row, ensure_ascii=False) + "\n")


def _write_json(path: Path, payload: Any) -> None:
    path.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")


def _write_spot_check(
    path: Path,
    items: list[GoldenItem],
    traces: list[ItemTrace],
    scores: list[ItemScores],
) -> None:
    am_rows = [
        {
            "item_id": item.id,
            "query": item.query,
            "answer": next(t.answer for t in traces if t.item_id == item.id),
            "faithfulness": next(
                s.faithfulness.model_dump(mode="json") for s in scores if s.item_id == item.id
            ),
            "review_verdict": "pending-human-review",
        }
        for item in items
        if item.locale.value == "am"
    ]
    _write_jsonl(path, am_rows[:SPOT_CHECK_COUNT])


def _print_summary(manifest: RunManifest, cells: list[dict[str, Any]], output_dir: Path) -> None:
    metric_names = (
        "citation_coverage",
        "grounding_precision",
        "tool_adherence",
        "context_recall",
        "faithfulness",
        "answer_relevancy",
    )
    print(
        f"\nrun {manifest.run_id} ({manifest.variant.value}) — judge: "
        f"{manifest.judge_provider}/{manifest.judge_model}"
    )
    header = f"{'cell':22s} " + " ".join(f"{name[:14]:>16s}" for name in metric_names)
    header += f" {'unknown':>8s} {'hard_fails':>10s}"
    print(header)
    for cell in cells:
        metrics: dict[str, float] = cell["metrics"]
        key = f"{cell['cell']['intent']}:{cell['cell']['locale']}"
        row = f"{key:22s} "
        row += " ".join(f"{metrics.get(name, float('nan')):>16.3f}" for name in metric_names)
        unknown = cell["unknown_handling_rate"]
        row += f" {unknown:>8.3f}" if unknown is not None else f" {'n/a':>8s}"
        row += f" {len(cell['hard_fail_items']):>10d}"
        print(row)
    print(f"\noutputs: {output_dir}")


async def command_verify(args: argparse.Namespace) -> int:
    golden_path = args.golden or (EVALS_DIR / "golden.jsonl")
    load_golden_set(golden_path)
    fixture_index = build_fixture_index()
    print(
        f"OK: golden set valid; fixture v{fixture_index.fixture_version} indexes "
        f"{len(fixture_index.id_to_ref)} chunks"
    )
    return 0


def main(argv: list[str] | None = None) -> int:
    logging.basicConfig(level=logging.INFO, format="%(levelname)s %(name)s: %(message)s")
    args = parse_args(argv)
    try:
        if args.command == "run":
            return asyncio.run(command_run(args))
        if args.command == "verify-dataset":
            return asyncio.run(command_verify(args))
    except (DatasetError, RunnerError, ValueError) as exc:
        logger.error("%s", exc)
        return 2
    return 1


if __name__ == "__main__":
    sys.exit(main())
