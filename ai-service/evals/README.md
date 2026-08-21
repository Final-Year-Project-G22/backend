# AI Evaluation Harness (FIN-69 / FIN-87)

Canonical runner for the 36-item Agentic Ask golden set. Drives the in-process
`AskAIUseCase.execute` path, captures full traces, scores deterministic gates
plus LLM-judge metrics, and aggregates per-cell scores for pass-bar
calibration (FIN-77).

## Layout

| Component | Location | Purpose |
|---|---|---|
| Golden set | `golden.jsonl` | 36 items: 3 intents × 2 locales × 6 (2 easy / 2 medium / 2 hard per cell); 2 unknown-handling items per locale |
| Dataset loader | `dataset.py` | Matrix validation + fixture chunk-ref index (uuid5 reverse map) |
| Runner | `runner.py` | Container bootstrap, fixture preflight, in-process ask with fresh conversations and cache-bypass assertion |
| Deterministic gates | `scoring/deterministic.py` | Required-citation coverage, grounding precision, tool adherence, context recall, unknown handling |
| Judge | `scoring/judge.py` | Faithfulness (claim extraction + verification) and answer relevancy; Gemini 2.5 Flash temp-0 by default, configured provider as fallback; input-hash cache, position-swapped re-judge on disagreement |
| Aggregation | `scoring/aggregate.py` | Per intent×locale cell means; not-applicable metrics excluded from denominators |
| CLI | `cli.py` | `run` and `verify-dataset` commands |

## Running

Prerequisites: fixture provisioned **with embeddings** (`evals/fixture/scripts/provision_corpus.py`,
no `--skip-embeddings`), core-backend gRPC up (remote tools), provider keys configured.

```sh
# from backend/ai-service/
uv run python -m evals.cli verify-dataset          # validate golden set vs fixture
uv run python -m evals.cli run                     # clean variant, full set, judge on
uv run python -m evals.cli run --variant broken    # deliberately degraded prompt (calibration contrast)
uv run python -m evals.cli run --cell knowledge:en --skip-judge   # quick deterministic pass
```

Or from `backend/`: `make eval-ai [EVAL_ARGS="..."]`, `make eval-verify`.

Outputs land in `evals/results/<timestamp>-<variant>/`: `items.jsonl` (raw traces),
`scores.jsonl` (per item), `cells.json` (per-cell aggregates), `manifest.json`
(versions, providers, git sha), `human_spot_check.jsonl` (5 AM items for the
manual faithfulness spot check — Amharic faithfulness stays an untrusted gate
until reviewed).

## Calibration contract (FIN-77)

Provisional bars live in `evals/scoring/scorer.py::PROVISIONAL_BARS`. Compare a
clean run against a broken run (`--variant broken` swaps the agentic system
prompt for one that drops grounding/citation/unknown-handling directives);
bars sit between the two distributions and require human approval. Judge
inconclusiveness, wrong-family tool use, and fabricated answers on unknown
items are hard failures — never silent passes.
