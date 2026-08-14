"""
Type stub for the generated ``ai/inference/v1/service_pb2_grpc`` module.

The buf ``grpc/python`` plugin emits ``service_pb2_grpc.py`` without type
hints, and no buf plugin emits gRPC-level stubs, so this hand-written .pyi
mirrors the generated runtime surface. Request/response message types come
from the mypy-protobuf ``service_pb2.pyi`` so the servicer layer
typechecks in strict mode.
"""

import collections.abc
import typing

import grpc
import grpc.aio

from ai.inference.v1 import service_pb2 as _service_pb2

_AskRequest = _service_pb2.AskRequest
_AskResponse = _service_pb2.AskResponse
_AskStreamChunk = _service_pb2.AskStreamChunk


class AIInferenceServiceStub:
    def __init__(self, channel: grpc.aio.Channel) -> None: ...
    Ask: grpc.aio.UnaryUnaryMultiCallable[_AskRequest, _AskResponse]
    AskStream: grpc.aio.UnaryStreamMultiCallable[_AskRequest, _AskStreamChunk]


class AIInferenceServiceServicer:
    def Ask(
        self,
        request: _AskRequest,
        context: grpc.aio.ServicerContext[_AskRequest, _AskResponse],
    ) -> typing.Awaitable[_AskResponse] | _AskResponse: ...
    def AskStream(
        self,
        request: _AskRequest,
        context: grpc.aio.ServicerContext[_AskRequest, _AskStreamChunk],
    ) -> collections.abc.AsyncIterable[_AskStreamChunk]: ...


def add_AIInferenceServiceServicer_to_server(
    servicer: AIInferenceServiceServicer,
    server: grpc.aio.Server,
) -> None: ...
