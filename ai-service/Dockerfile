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

COPY proto/ ./proto/

RUN apt-get update -qq && apt-get install -y -qq curl ca-certificates && \
    curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-Linux-x86_64" -o /usr/local/bin/buf && \
    chmod +x /usr/local/bin/buf && \
    cd proto && buf generate && cd .. && \
    mkdir -p grpc_stubs && cp -r ai-service/grpc_stubs/* grpc_stubs/ && \
    test -f grpc_stubs/buf/validate/validate_pb2.py && \
    rm -rf proto ai-service/grpc_stubs /usr/local/bin/buf && \
    apt-get remove -y curl ca-certificates && apt-get autoremove -y && \
    rm -rf /var/lib/apt/lists/*

ENV PATH="/app/.venv/bin:$PATH"

RUN groupadd --system app && useradd --system --no-create-home -g app app
USER app

EXPOSE 8000

CMD ["python", "main.py"]
