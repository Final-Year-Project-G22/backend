# AGENTS.md — operating rules for the Adisu Serategna backend

> Monorepo: `core-backend/` (Go 1.25, Gin + huma v2), `ai-service/` (Python 3.11, uv), `proto/` (Buf). Feature work is tracked in GitHub issues (ported from Linear when implementation starts); issues labeled `ready-for-dev` are grabbable by agents.

## Repo layout

- `core-backend/` — Go HTTP API (Gin). Routes registered per module under `internal/modules/<module>/delivery/routes`; DTOs and request/response structs are huma-annotated. **The OpenAPI spec is generated from these structs — never edited by hand.**
- `ai-service/` — Python AI service (ask strategies, embeddings, LLM adapters, tool registry). Managed with `uv` (`uv sync`, `uv run`).
- `proto/` — gRPC contracts between core-backend and ai-service; `buf generate` produces `ai-service/grpc_stubs/` (CI regenerates when protos change).

## Commands

- `make lint` / `make lint-all` — pre-commit hooks (commitlint scopes, Go lint, ruff).
- `make fmt-all` — goimports + ruff format across both services.
- `make check` — fmt + lint.
- `make dev` / `make dev-go` — local dev (see Makefile).
- `make spec` — **regenerate `core-backend/docs/openapi.json` from the huma router** (see API contract below).
- Go: `cd core-backend && go test ./...`
- Python: `cd ai-service && uv sync --frozen --all-groups --extra dev && uv run basedpyright && uv run ruff check && uv run bandit -c pyproject.toml -r app core infrastructure workers`

## Commit conventions

- Conventional commits (`type(scope): summary`). **commitlint scopes are strict: `core` | `ai` | `cross`** — any other scope (e.g. `docs`) fails the pre-commit hook. Use `chore(cross): …` for repo-wide chores.
- Integration branches are `main` + `dev`. Ticket work must use a feature branch based on `origin/dev` and land through a PR targeting `dev`. Direct commits to `dev` are only for cleanup/chores; admins may bypass branch protection for cleanup.
- No doc branches, no doc PRs — the docs home is the `planning` repo (`Final-Year-Project-G22/planning`), not this repo.

## API contract (NON-NEGOTIABLE)

Any change to huma handler structs, routes, DTOs, or validation rules **must** be accompanied by the regenerated spec and downstream typegen, in the same change:

1. `make spec` → regenerates `core-backend/docs/openapi.json` from the router.
2. Copy the spec to consumers and regenerate their types:
   - **web** (admin): `cd ../web && pnpm sync:api` — copies the spec to `src/openapi/openapi.json` and runs orval typegen (`src/lib/api/types`, `src/lib/api/services`).
   - **mobile** (Flutter): copy to `mobile/tool/openapi/openapi.json`, then `cd ../mobile && dart run build_runner build --delete-conflicting-outputs` (generates `lib/core/api/generated/`).
3. Commit the regenerated spec + generated types **in the same commit/PR** as the API change.
4. **Never hand-write** DTO types that the spec defines — web types come from orval, mobile from swagger_dart_code_generator. Hand-written models are for domain objects that are not API-transported.

CI enforces step 1: the `openapi-fresh` job regenerates the spec on every PR touching Go code and fails if `core-backend/docs/openapi.json` is stale. If a PR fails this check: run `make spec`, commit the spec diff, push.

## Agent workflow

- Pick up issues labeled `ready-for-dev` (GitHub) — these are triaged, sized, and unblocked.
- Implement with /tdd where seams allow; run typechecking regularly, single test files regularly, full suite once at the end.
- Use /code-review on the finished change, then commit to the current branch.
- If an implementation touches API shape: the API contract above applies — spec regen + web/mobile typegen propagate **before** the change is considered done.

### Mandatory Ticket Branch/PR Gate

For every implementation requested from a GitHub issue or Linear ticket:

1. Before editing, fetch the latest `origin/dev` and create a feature branch from it: `git fetch origin dev && git switch -c <type>/<short-name> origin/dev`.
2. Never implement or commit ticket work directly on `dev` or `main`. If the current branch is `dev` or `main`, stop and create the feature branch first.
3. After verification, push the feature branch with `git push -u origin <branch>`.
4. Create a GitHub PR with `gh pr create --base dev --head <branch>` and include the issue reference in the PR body.
5. Do not resolve or close the source ticket until the PR exists, then record the PR URL on the source ticket.

## Linear access

Design decisions live in Linear (team *Final Year Project G22*, key `FIN`). Read them before implementing:

- Auth: `LINEAR_TOKEN` in the root `.env` (personal API key from https://linear.app/settings/api).
- Helper: `../scripts/linear.sh` (from repo root). `view <identifier-or-id>` shows a ticket's description, relations, and comments (resolutions are usually in comments).
- A GitHub issue ported from Linear should name its source ticket (e.g. FIN-75); check that ticket's resolution for design decisions before coding.
