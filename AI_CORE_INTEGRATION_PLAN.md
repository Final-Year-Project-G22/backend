# AI Core Integration Plan

## Architecture Overview
- **core-backend** is the ONLY client-facing service.
- **ai-service** is an internal service that provides AI inference.
- Communication between `core-backend` and `ai-service` is strictly **gRPC**.
- Internal service authentication uses a pre-shared API token.

## Roadmap

### Step 1: Clean Up gRPC Code Generation & Fix Package Shadowing
- **Problem**: Python's `grpc` standard library was being shadowed by `ai-service/grpc/` output directory, causing runtime errors and requiring dangerous hacky bootstrap scripts.
- **Action**: 
  - Delete `ai-service/grpc/` completely.
  - Update `buf.gen.yaml` to output Python stubs to `ai-service/grpc_stubs/` to prevent shadowing.
  - Update all imports in `ai-service` (`infrastructure/grpc/core_service.py`) to use `grpc_stubs...`.
  - Update `pyproject.toml`, pre-commit hooks, and GitHub actions to ignore `grpc_stubs/` instead of `grpc/generated/`.
  - Regenerate stubs using `buf generate`.

### Step 2: Define Inference gRPC Contract
- **Action**:
  - Create `proto/ai/inference/v1/service.proto` defining the `AIInferenceService`.
  - Include methods for inference (e.g., `StreamInference` or `GenerateResponse`).
  - Run `buf generate` to create both Go and Python stubs.

### Step 3: Implement Inference Server in `ai-service`
- **Action**:
  - Implement `AIInferenceServiceServicer` in `ai-service/infrastructure/grpc/inference_server.py`.
  - Wire up the server in the FastAPI lifespan or a dedicated runner (`ai-service/main.py`).
  - Add Token-based Auth interceptor/metadata checking to secure the gRPC endpoint.

### Step 4: Implement Inference Client in `core-backend`
- **Action**:
  - Create `core-backend/internal/modules/ai/infrastructure/client/grpc_client.go`.
  - Implement the `AIInferencePort` defined in `core-backend/internal/modules/ai/domain/port/inference.go`.
  - Inject this client into the `core-backend`'s AI module.

### Step 5: End-to-End Testing & Review
- **Action**:
  - Write unit tests for the Python gRPC server.
  - Write unit tests for the Go gRPC client.
  - Verify seamless communication without import or shadowing issues.