# AI Service

Python gRPC service for AI inference, conversations, and ingestion workers.

## Environment

Key variables:

```env
GRPC_PORT=50051
CORE_GRPC_ENDPOINT=localhost:50052
INTERNAL_GRPC_AUTH_TOKEN=
AI_ASK_ENABLED=true
```

## Ask Feature Flag

- `AI_ASK_ENABLED=true` allows `Ask` and `AskStream` gRPC methods.
- `AI_ASK_ENABLED=false` causes `Ask` and `AskStream` to return `UNAVAILABLE` with message `Ask API is disabled`.
