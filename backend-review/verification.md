# Review verification

The review covers backend commit `216fc0a94ea86bcd712028c2647b36f2e11e0c77`. Tracked repository files were not modified.

The Markdown documents are stored in this folder. Supporting scripts, logs, coverage output, the dependency environment, source snapshot, and build caches are under `/tmp/adisu-backend-review`; the commands below reference those paths.

## Go

Run from `core-backend`:

```bash
GOCACHE=/tmp/adisu-backend-review/go-cache GOTOOLCHAIN=local go test ./...
```

Result: passed. Installed toolchain: Go 1.27.1. The repository and CI specify Go 1.25.3. The sandbox initially blocked an existing test's local TCP listener, so the successful run used approved execution outside that restriction. No live database or broker was provisioned for these tests.

The targeted review probes use Go's overlay support to add test files without editing the application packages:

```bash
GOCACHE=/tmp/adisu-backend-review/go-cache GOTOOLCHAIN=local go test \
  -overlay=/tmp/adisu-backend-review/go-overlay.json \
  ./internal/ws ./internal/shared/repository -run TestReview -v
```

Both probes passed by demonstrating the faulty behavior. They are diagnostic reproductions, not fixes. The WebSocket probe exercises actual hub methods with transport closure suppressed. The repository probe injects an error through GORM's query callback rather than using a database.

## Python

Dependencies came from the checked-in lockfile using:

```bash
UV_PROJECT_ENVIRONMENT=/tmp/adisu-backend-review/python-venv \
UV_CACHE_DIR=/tmp/adisu-backend-review/uv-cache \
uv sync --frozen --all-groups --extra dev
```

Python is 3.11.9. A committed-source snapshot is under `/tmp/adisu-backend-review/snapshot`. Python stubs were generated from the snapshot's proto files using the repository's pinned Python generator versions through a Python-only Buf template. Source tests and settings are unchanged.

Run from the snapshot's `ai-service` directory:

```bash
PYTHONPATH=/tmp/adisu-backend-review/snapshot/ai-service:/tmp/adisu-backend-review/snapshot/ai-service/grpc_stubs \
PYTHONDONTWRITEBYTECODE=1 \
COVERAGE_FILE=/tmp/adisu-backend-review/python.coverage \
/tmp/adisu-backend-review/python-venv/bin/python -m pytest \
  -n 2 \
  --cov-report=term-missing \
  --cov-report=html:/tmp/adisu-backend-review/htmlcov \
  -o cache_dir=/tmp/adisu-backend-review/pytest-cache

/tmp/adisu-backend-review/python-venv/bin/ruff check --no-cache

/tmp/adisu-backend-review/python-venv/bin/bandit \
  -c pyproject.toml -r app core infrastructure workers

/tmp/adisu-backend-review/python-venv/bin/basedpyright \
  --project /tmp/adisu-backend-review/pyrightconfig.json
```

Results:

- Pytest: 434 passed in 14.24 seconds; coverage 83.75%, exceeding the configured 80% gate.
- Ruff: all checks passed.
- Bandit: no issues identified under the repository's configuration. One potential issue is disabled by configured checks; this is not proof of security.
- BasedPyright 1.39.0: 114 source files checked, zero errors, zero warnings.

The first BasedPyright invocation with only `--pythonpath` did not include the isolated environment's site-packages. The final `pyrightconfig.json` inherits the project's complete configuration and sets only `venvPath` and `venv`. Those environment-resolution errors were excluded from findings.

The source-level diagnostic probes run without dependency installation:

```bash
python3 /tmp/adisu-backend-review/python_source_probes.py \
  /home/johna/Projects/adisu/backend
```

They execute selected source functions with fake gRPC handlers and RabbitMQ deliveries. They confirmed the streaming authentication branch and retry-count behavior. They do not establish wire-level behavior or database concurrency under load.

## Scope limits

No live PostgreSQL, RabbitMQ, SeaweedFS, MinIO, external model endpoint, or production proxy was used. No production migration or deployment was performed. Compose, migration, cross-service transaction, and connection-lifetime findings come from code and configuration tracing unless the review explicitly identifies a probe.

Passing the existing tests does not clear the reproduced defects. Several missing cases concern dependency construction, resource lifetime, and failure paths outside the current test coverage.
