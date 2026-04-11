import sys
from pathlib import Path

# Add grpc_stubs to sys.path so generated protobuf modules (like core.user, buf.validate) can be imported directly
stubs_path = Path(__file__).resolve().parent.parent.parent / "grpc_stubs"
if str(stubs_path) not in sys.path:
    sys.path.insert(0, str(stubs_path))

from infrastructure.rpc.core_service import CoreServiceGrpcAdapter, CoreUserGrpcClient  # noqa: E402

__all__ = ["CoreServiceGrpcAdapter", "CoreUserGrpcClient"]
