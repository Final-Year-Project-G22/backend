import importlib
import sys
from pathlib import Path
from typing import TYPE_CHECKING, Any

if TYPE_CHECKING:
    from infrastructure.rpc.ai_tool_client import AIToolGrpcClient
    from infrastructure.rpc.core_service import CoreServiceGrpcAdapter, CoreUserGrpcClient

# Add grpc_stubs to sys.path so generated protobuf modules (like core.user, buf.validate) can be imported directly
stubs_path = Path(__file__).resolve().parent.parent.parent / "grpc_stubs"
if str(stubs_path) not in sys.path:
    sys.path.insert(0, str(stubs_path))


def __getattr__(name: str) -> Any:
    if name in {"CoreServiceGrpcAdapter", "CoreUserGrpcClient"}:
        core_service = importlib.import_module("infrastructure.rpc.core_service")
        return {
            "CoreServiceGrpcAdapter": core_service.CoreServiceGrpcAdapter,
            "CoreUserGrpcClient": core_service.CoreUserGrpcClient,
        }[name]
    if name == "AIToolGrpcClient":
        ai_tool_module = importlib.import_module("infrastructure.rpc.ai_tool_client")
        return ai_tool_module.AIToolGrpcClient
    msg = f"module {__name__!r} has no attribute {name!r}"
    raise AttributeError(msg)


__all__ = ["AIToolGrpcClient", "CoreServiceGrpcAdapter", "CoreUserGrpcClient"]
