"""Load gRPC stubs from grpc_stubs without polluting the global import namespace.

The generated protobuf stubs live under grpc_stubs/core/ which collides with
the project's own core/ package.  This module isolates the load by
manipulating sys.path and sys.modules temporarily.

Because the core stubs are only reachable at runtime, their static types are
described by the Protocol classes below; the message shapes mirror the
mypy-protobuf ``*_pb2.pyi`` stubs next to the generated modules.
"""

from __future__ import annotations

import importlib
import sys
from collections.abc import Awaitable, Iterable
from pathlib import Path
from types import ModuleType
from typing import Protocol, cast

import grpc.aio

_GRPC_STUBS_DIR = Path(__file__).parent.parent.parent / "grpc_stubs"


class GetUserProfileRequestProto(Protocol):
    user_id: str


class GetUserProfileResponseProto(Protocol):
    user_id: str
    account_id: str
    tier: str
    preferred_language: str


class CoreUserServiceStub(Protocol):
    def GetUserProfile(  # noqa: N802
        self,
        request: GetUserProfileRequestProto,
    ) -> Awaitable[GetUserProfileResponseProto]: ...


class GetSignedUrlRequestProto(Protocol):
    document_id: str
    account_id: str
    expires_in_seconds: int


class GetSignedUrlResponseProto(Protocol):
    signed_url: str
    expires_at: int
    content_type: str
    content_length: int


class DocumentFetchServiceStub(Protocol):
    def GetSignedUrl(  # noqa: N802
        self,
        request: GetSignedUrlRequestProto,
    ) -> Awaitable[GetSignedUrlResponseProto]: ...


class AIToolProto(Protocol):
    name: str
    description: str
    parameter_schema_json: str


class ListToolsRequestProto(Protocol): ...


class ListToolsResponseProto(Protocol):
    tools: Iterable[AIToolProto]


class ExecuteToolRequestProto(Protocol):
    tool: str
    arguments_json: str
    account_id: str
    user_id: str


class ExecuteToolResponseProto(Protocol):
    success: bool
    result_json: str
    error_message: str


class AIToolServiceStub(Protocol):
    def ListTools(  # noqa: N802
        self,
        request: ListToolsRequestProto,
    ) -> Awaitable[ListToolsResponseProto]: ...

    def ExecuteTool(  # noqa: N802
        self,
        request: ExecuteToolRequestProto,
    ) -> Awaitable[ExecuteToolResponseProto]: ...


class _GrpcStubLoader:
    """Lazily loads protobuf stubs from grpc_stubs in isolation."""

    _document_pb2: ModuleType | None = None
    _document_pb2_grpc: ModuleType | None = None
    _user_pb2: ModuleType | None = None
    _user_pb2_grpc: ModuleType | None = None
    _ai_tool_pb2: ModuleType | None = None
    _ai_tool_pb2_grpc: ModuleType | None = None

    @classmethod
    def _load(cls, module_name: str) -> ModuleType | None:
        file_path = _GRPC_STUBS_DIR / module_name.replace(".", "/")
        py_file = file_path.with_suffix(".py")
        if not py_file.exists():
            return None

        # Import the real core module and patch its __path__ so Python
        # can discover grpc-stub subpackages (core.user, core.document,
        # core.ai_tool).
        # Without this, core has no way to find core.user.v1.service_pb2
        # because the project's core/ is a regular package that shadows
        # the grpc_stubs/core/ namespace portion.
        import core as _core_mod  # noqa: PLC0415

        _original_path = list(_core_mod.__path__)
        _core_mod.__path__.append(str(_GRPC_STUBS_DIR / "core"))

        # Add grpc_stubs to sys.path so that top-level package imports
        # referenced by the generated stubs (buf.validate, etc.) resolve.
        _stubs_str = str(_GRPC_STUBS_DIR)
        _sys_path_added = _stubs_str not in sys.path
        if _sys_path_added:
            sys.path.insert(0, _stubs_str)

        try:
            return importlib.import_module(module_name)
        except Exception:
            return None
        finally:
            _core_mod.__path__[:] = _original_path
            if _sys_path_added:
                sys.path.remove(_stubs_str)
            # Reset cached child modules so the patched path doesn't leak
            for _key in list(sys.modules):
                if _key.startswith(("core.user.", "core.document.", "core.ai_tool.")):
                    del sys.modules[_key]

    @classmethod
    def document_pb2(cls) -> ModuleType | None:
        if cls._document_pb2 is None:
            cls._document_pb2 = cls._load("core.document.v1.document_pb2")
        return cls._document_pb2

    @classmethod
    def document_pb2_grpc(cls) -> ModuleType | None:
        if cls._document_pb2_grpc is None:
            cls._document_pb2_grpc = cls._load("core.document.v1.document_pb2_grpc")
        return cls._document_pb2_grpc

    @classmethod
    def user_pb2(cls) -> ModuleType | None:
        if cls._user_pb2 is None:
            cls._user_pb2 = cls._load("core.user.v1.service_pb2")
        return cls._user_pb2

    @classmethod
    def user_pb2_grpc(cls) -> ModuleType | None:
        if cls._user_pb2_grpc is None:
            cls._user_pb2_grpc = cls._load("core.user.v1.service_pb2_grpc")
        return cls._user_pb2_grpc

    @classmethod
    def ai_tool_pb2(cls) -> ModuleType | None:
        if cls._ai_tool_pb2 is None:
            cls._ai_tool_pb2 = cls._load("core.ai_tool.v1.tool_service_pb2")
        return cls._ai_tool_pb2

    @classmethod
    def ai_tool_pb2_grpc(cls) -> ModuleType | None:
        if cls._ai_tool_pb2_grpc is None:
            cls._ai_tool_pb2_grpc = cls._load("core.ai_tool.v1.tool_service_pb2_grpc")
        return cls._ai_tool_pb2_grpc


def get_document_fetch_stub(channel: grpc.aio.Channel) -> DocumentFetchServiceStub | None:
    grpc_mod = _GrpcStubLoader.document_pb2_grpc()
    if grpc_mod is None:
        return None
    ctor = getattr(grpc_mod, "DocumentFetchServiceStub", None)
    if not callable(ctor):
        return None
    return cast(DocumentFetchServiceStub, ctor(channel))


def build_get_signed_url_request(
    document_id: str,
    account_id: str,
    expires_in_seconds: int,
) -> GetSignedUrlRequestProto:
    pb2 = _GrpcStubLoader.document_pb2()
    if pb2 is not None:
        ctor = getattr(pb2, "GetSignedUrlRequest", None)
        if callable(ctor):
            return cast(
                GetSignedUrlRequestProto,
                ctor(
                    document_id=document_id,
                    account_id=account_id,
                    expires_in_seconds=expires_in_seconds,
                ),
            )
    # Fallback for when protobuf runtime is unavailable
    return cast(
        GetSignedUrlRequestProto,
        type(
            "GetSignedUrlRequest",
            (),
            {
                "document_id": document_id,
                "account_id": account_id,
                "expires_in_seconds": expires_in_seconds,
            },
        )(),
    )


def get_core_user_stub(channel: grpc.aio.Channel) -> CoreUserServiceStub | None:
    grpc_mod = _GrpcStubLoader.user_pb2_grpc()
    if grpc_mod is None:
        return None
    ctor = getattr(grpc_mod, "CoreUserServiceStub", None)
    if not callable(ctor):
        return None
    return cast(CoreUserServiceStub, ctor(channel))


def build_get_user_request(user_id: str) -> GetUserProfileRequestProto:
    pb2 = _GrpcStubLoader.user_pb2()
    if pb2 is not None:
        ctor = getattr(pb2, "GetUserProfileRequest", None)
        if callable(ctor):
            return cast(GetUserProfileRequestProto, ctor(user_id=user_id))
    return cast(
        GetUserProfileRequestProto, type("GetUserProfileRequest", (), {"user_id": user_id})()
    )


def get_ai_tool_stub(channel: grpc.aio.Channel) -> AIToolServiceStub | None:
    grpc_mod = _GrpcStubLoader.ai_tool_pb2_grpc()
    if grpc_mod is None:
        return None
    ctor = getattr(grpc_mod, "AIToolServiceStub", None)
    if not callable(ctor):
        return None
    return cast(AIToolServiceStub, ctor(channel))


def build_list_tools_request() -> ListToolsRequestProto:
    pb2 = _GrpcStubLoader.ai_tool_pb2()
    if pb2 is not None:
        ctor = getattr(pb2, "ListToolsRequest", None)
        if callable(ctor):
            return cast(ListToolsRequestProto, ctor())
    return cast(ListToolsRequestProto, type("ListToolsRequest", (), {})())


def build_execute_tool_request(
    tool: str,
    arguments_json: str,
    account_id: str,
    user_id: str,
) -> ExecuteToolRequestProto:
    pb2 = _GrpcStubLoader.ai_tool_pb2()
    if pb2 is not None:
        ctor = getattr(pb2, "ExecuteToolRequest", None)
        if callable(ctor):
            return cast(
                ExecuteToolRequestProto,
                ctor(
                    tool=tool,
                    arguments_json=arguments_json,
                    account_id=account_id,
                    user_id=user_id,
                ),
            )
    return cast(
        ExecuteToolRequestProto,
        type(
            "ExecuteToolRequest",
            (),
            {
                "tool": tool,
                "arguments_json": arguments_json,
                "account_id": account_id,
                "user_id": user_id,
            },
        )(),
    )


def build_ai_tool_response(response: object) -> ExecuteToolResponseProto | None:
    if response is not None and hasattr(response, "success"):
        return cast(ExecuteToolResponseProto, response)
    pb2 = _GrpcStubLoader.ai_tool_pb2()
    if pb2 is not None:
        ctor = getattr(pb2, "ExecuteToolResponse", None)
        if callable(ctor):
            return cast(
                ExecuteToolResponseProto,
                ctor(
                    success=getattr(response, "success", False),
                    result_json=getattr(response, "result_json", ""),
                    error_message=getattr(response, "error_message", ""),
                ),
            )
    return cast(ExecuteToolResponseProto | None, response)


__all__ = [
    "AIToolProto",
    "AIToolServiceStub",
    "CoreUserServiceStub",
    "DocumentFetchServiceStub",
    "ExecuteToolRequestProto",
    "ExecuteToolResponseProto",
    "GetSignedUrlRequestProto",
    "GetSignedUrlResponseProto",
    "GetUserProfileRequestProto",
    "GetUserProfileResponseProto",
    "ListToolsRequestProto",
    "ListToolsResponseProto",
    "build_ai_tool_response",
    "build_execute_tool_request",
    "build_get_signed_url_request",
    "build_get_user_request",
    "build_list_tools_request",
    "get_ai_tool_stub",
    "get_core_user_stub",
    "get_document_fetch_stub",
]
