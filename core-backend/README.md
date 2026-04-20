# Core Backend API

Go-based REST API service built with Gin framework.

## Prerequisites

- Go 1.25.3+
- Docker & Docker Compose (for infrastructure)
- Atlas CLI (for migrations)
- air (hot reload)
- goimports (formatting)
- golangci-lint (linting)

## Quick Start

### 1. Install Dependencies

```bash
make deps
```

### 2. Start Infrastructure

```bash
# From root directory
make infra-up
```

### 3. Run Migrations

```bash
# Generate initial migration
make migration-generate name=init

# Apply migrations
make migration-apply
```

### 4. Start Development Server

```bash
# With hot reload (recommended)
make dev

# Or without hot reload
make run
```

## Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application |
| `make run` | Run the application |
| `make dev` | Run with hot reload (air) |
| `make test` | Run tests |
| `make lint` | Run linter |
| `make fmt` | Format code (go fmt) |
| `make fmt-imports` | Format code (goimports) |
| `make check` | Format + Lint |
| `make clean` | Clean build artifacts |
| `make deps` | Download & tidy dependencies |
| `make migration-list` | List registered modules |
| `make migration-generate name=<name>` | Generate migration |
| `make migration-apply` | Apply migrations |

## Environment Variables

Create a `.env` file:

```env
DATABASE_USER=user_adisu
DATABASE_PASSWORD=your_password
DATABASE_DBNAME=adisu_db
AI_INFERENCE_GRPC_ENDPOINT=localhost:50051
AI_INFERENCE_AUTH_TOKEN=
AI_ASK_ENABLED=true
```

### Ask Feature Flag

- `AI_ASK_ENABLED=true` enables Ask REST endpoints (`/api/v1/ai/ask`, `/api/v1/ai/ask/stream`, conversations routes).
- `AI_ASK_ENABLED=false` disables Ask route registration (endpoints return `404` because they are not mounted).

## Features

- Gin web framework
- GORM for database
- Redis for caching
- SeaweedFS for file storage
- Atlas for migrations
- Air for hot reload

## License

ISC
