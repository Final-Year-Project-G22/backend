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

# Install pnpm (Node package manager)
install-pnpm:
	@echo "Installing pnpm..."
	@if command -v corepack >/dev/null 2>&1; then \
		corepack enable && corepack prepare pnpm@latest --activate; \
	elif command -v npm >/dev/null 2>&1; then \
		npm install -g pnpm; \
	else \
		echo "Node.js/npm not found. Install Node.js LTS first."; \
		exit 1; \
	fi
	@echo "pnpm installed."

# ==============================================================================
# Setup (Install all dev tools)
# ==============================================================================

setup: check-prerequisites
	@echo "Setting up development environment..."
	@$(MAKE) ensure-pnpm
	@echo "Installing pre-commit (system Python)..."
	uv pip install --system pre-commit
	@echo "Installing Node.js dependencies (pnpm)..."
	pnpm install
	@echo "Installing Go tools..."
	cd core-backend && go install github.com/air-verse/air@latest
	cd core-backend && go install golang.org/x/tools/cmd/goimports@latest
	cd core-backend && go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.2
	@echo "Installing Python tools (system Python)..."
	uv pip install --system ruff
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
	@which node >/dev/null 2>&1 || (echo "Node.js not found. Install Node.js LTS first." && exit 1)
	@echo "All prerequisites satisfied."

# Ensure pnpm is installed before setup
ensure-pnpm:
	@which pnpm >/dev/null 2>&1 || $(MAKE) install-pnpm

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
