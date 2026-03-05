# ==============================================================================
# Prerequisites Installation (Reference Commands)
# ==============================================================================

# Install Go 1.25.3
install-go:
	@echo "Installing Go 1.25.3..."
	wget -q https://go.dev/dl/go1.25.3.linux-amd64.tar.gz -O /tmp/go.tar.gz
	sudo tar -C /usr/local -xzf /tmp/go.tar.gz
	rm /tmp/go.tar.gz
	@echo "Go installed. Add to PATH: export PATH=\$$PATH:/usr/local/go/bin"

# Install Python 3.11
install-python:
	@echo "Installing Python 3.11..."
	sudo apt update
	sudo apt install -y python3.11 python3-pip python3.11-venv
	@echo "Python installed."

# Install uv (Python package manager)
install-uv:
	@echo "Installing uv..."
	curl -LsSf https://astral.sh/uv/install.sh | sh
	@echo "uv installed. Restart shell or source profile."

# ==============================================================================
# Setup (Install all dev tools)
# ==============================================================================

setup: check-prerequisites
	@echo "Setting up development environment..."
	@echo "Installing pre-commit..."
	uv pip install pre-commit
	@echo "Installing Go tools..."
	cd core-backend && go install github.com/cosmtrek/air@latest
	cd core-backend && go install golang.org/x/tools/cmd/goimports@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin
	@echo "Installing Python tools..."
	cd ai-service && uv pip install ruff
	@echo "Installing pre-commit hooks..."
	pre-commit install
	pre-commit install --hook-type commit-msg
	@echo "Setup complete!"

# Check prerequisites
check-prerequisites:
	@echo "Checking prerequisites..."
	@which go >/dev/null 2>&1 || (echo "Go not found. Run: make install-go" && exit 1)
	@which python3 >/dev/null 2>&1 || (echo "Python not found. Run: make install-python" && exit 1)
	@which uv >/dev/null 2>&1 || (echo "uv not found. Run: make install-uv" && exit 1)
	@echo "All prerequisites satisfied."

# ==============================================================================
# Linting & Formatting
# ==============================================================================

# Run pre-commit on affected files
lint:
	pre-commit run

# Run pre-commit on all files
lint-all:
	pre-commit run --all-files

# Format all code
fmt-all:
	cd core-backend && goimports -l -w .
	cd ai-service && uv run ruff format .

# Format + Lint
check: fmt-all lint-all

# ==============================================================================
# Development
# ==============================================================================

# Start infrastructure services
infra-up:
	docker compose up -d

# Stop infrastructure services
infra-down:
	docker compose down

# Start all dev servers
dev:
	cd core-backend && air
	cd ai-service && uv run python main.py

# Start core-backend only
dev-go:
	cd core-backend && air

# Start ai-service only
dev-python:
	cd ai-service && uv run python main.py

# ==============================================================================
# Clean
# ==============================================================================

clean:
	cd core-backend && rm -rf tmp
	rm -rf .ruff_cache
