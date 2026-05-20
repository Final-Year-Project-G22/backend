FROM python:3.11-slim

ENV UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy \
    UV_NO_MANAGED_PYTHON=1 \
    UV_PYTHON=/usr/local/bin/python3

RUN pip install --no-cache-dir uv

WORKDIR /app

COPY ai-service/pyproject.toml ai-service/uv.lock ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev --no-install-project

COPY ai-service/ ./

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

ENV PATH="/app/.venv/bin:$PATH"

RUN groupadd --system app && useradd --system --no-create-home -g app app
USER app

EXPOSE 8000

CMD ["python", "main.py"]
