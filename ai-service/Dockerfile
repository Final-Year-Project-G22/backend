FROM python:3.11-slim

ENV UV_SYSTEM_PYTHON=1 \
    UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy

RUN pip install --no-cache-dir uv

WORKDIR /app

COPY ai-service/pyproject.toml ai-service/uv.lock ./
RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev --no-install-project

COPY ai-service/ ./

RUN --mount=type=cache,target=/root/.cache/uv \
    uv sync --frozen --no-dev

RUN addgroup -S app && adduser -S app -G app
USER app

EXPOSE 8000

CMD ["python", "main.py"]
