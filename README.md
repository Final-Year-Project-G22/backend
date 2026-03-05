# Backend Services

## Prerequisites

Install in order if not already installed:

1. **Go 1.25.3**
   - Download from: https://go.dev/dl/
   - Or on Linux:
     ```bash
     wget -q https://go.dev/dl/go1.25.3.linux-amd64.tar.gz -O /tmp/go.tar.gz
     sudo tar -C /usr/local -xzf /tmp/go.tar.gz
     rm /tmp/go.tar.gz
     ```
   - Add to PATH: `export PATH=$PATH:/usr/local/go/bin`

2. **Python 3.11.9**
   - Ubuntu/Debian:
     ```bash
     sudo apt update
     sudo apt install python3.11 python3-pip python3.11-venv
     ```
   - macOS: `brew install python@3.11`

3. **uv (Python Package Manager)**
   ```bash
   # Option 1: curl (recommended)
   curl -LsSf https://astral.sh/uv/install.sh | sh

   # Option 2: pip
   pip install uv

   # Option 3: brew
   brew install uv
   ```

4. **Docker & Docker Compose**
   - Download from: https://docs.docker.com/get-docker/

5. **Atlas CLI** (for migrations)
   ```bash
   curl -sSf https://atlasgo.sh | sh
   ```

## Quick Start

### 1. Install All Dependencies

```bash
make setup
```

This will:
- Check for Go, Python, uv installation
- Install pre-commit
- Install Go tools (air, goimports, golangci-lint)
- Install Python tools (ruff)
- Install pre-commit hooks

### 2. Start Infrastructure

```bash
make infra-up
```

Starts: PostgreSQL (:5432), Redis (:6379), SeaweedFS (:8080)

### 3. Run Migrations

```bash
cd core-backend
make migration-generate name=init
make migration-apply
```

### 4. Start Development

```bash
# All services
make dev

# Or individually
make dev-go      # core-backend (Go)
make dev-python  # ai-service (Python)
```

## Commands

| Command | Description |
|---------|-------------|
| `make install-go` | Install Go 1.25.3 (reference) |
| `make install-python` | Install Python 3.11 (reference) |
| `make install-uv` | Install uv package manager (reference) |
| `make setup` | Install all dev tools & hooks |
| `make check-prerequisites` | Check if Go, Python, uv are installed |
| `make lint` | Run pre-commit on affected files |
| `make lint-all` | Run pre-commit on all files |
| `make fmt-all` | Format all code |
| `make check` | Format + Lint |
| `make infra-up` | Start Docker services |
| `make infra-down` | Stop Docker services |
| `make dev` | Start all dev servers |
| `make clean` | Clean build artifacts |

## Pre-commit Hooks

The project uses pre-commit for linting and formatting:

| Hook | Language | Tool | Action |
|------|----------|------|--------|
| go-lint | Go | golangci-lint | Lint code |
| go-fmt | Go | goimports | Format code |
| python-ruff | Python | ruff | Lint code |
| python-format | Python | ruff | Format code |
| commitlint | Commit | commitlint | Validate commit messages |

Hooks run automatically on `git commit`. To run manually:

```bash
# Test on staged files
make lint

# Test on all files
make lint-all

# Format all code
make fmt-all
```

## Services

| Service | Language | Port | Description |
|---------|----------|------|-------------|
| core-backend | Go | 8080 | Main API |
| ai-service | Python | 3000 | AI processing |
| PostgreSQL | - | 5432 | Database |
| Redis | - | 6379 | Cache |
| SeaweedFS | - | 8080 | File storage |

## Project Structure

```
backend/
├── .pre-commit-config.yaml    # Pre-commit hook definitions
├── Makefile                   # Root make commands
├── package.json               # Node.js dependencies (husky, pre-commit)
├── docker-compose.yml         # Infrastructure services
├── README.md                  # This file
│
├── core-backend/             # Go API service
│   ├── .golangci.yml         # Go linting config
│   ├── Makefile              # Go make commands
│   ├── .air.toml             # Hot reload config
│   ├── .env                  # Environment variables
│   └── ...
│
└── ai-service/               # Python AI service
    ├── pyproject.toml        # Python config (ruff)
    ├── .python-version       # Python version
    └── ...
```

## License

ISC
