"""
Type stub for the generated ``ai/conversation/v1/service_pb2_grpc`` module.

The buf ``grpc/python`` plugin emits ``service_pb2_grpc.py`` without type
hints, and no buf plugin emits gRPC-level stubs, so this hand-written .pyi
mirrors the generated runtime surface. Request/response message types come
from the mypy-protobuf ``service_pb2.pyi`` so the servicer layer
typechecks in strict mode.
"""

import typing

import grpc
import grpc.aio

from ai.conversation.v1 import service_pb2 as _service_pb2

_ListConversationsRequest = _service_pb2.ListConversationsRequest
_ListConversationsResponse = _service_pb2.ListConversationsResponse
_GetConversationRequest = _service_pb2.GetConversationRequest
_GetConversationResponse = _service_pb2.GetConversationResponse
_ArchiveConversationRequest = _service_pb2.ArchiveConversationRequest
_ArchiveConversationResponse = _service_pb2.ArchiveConversationResponse


class AIConversationServiceStub:
    def __init__(self, channel: grpc.aio.Channel) -> None: ...
    ListConversations: grpc.aio.UnaryUnaryMultiCallable[
        _ListConversationsRequest, _ListConversationsResponse
    ]
    GetConversation: grpc.aio.UnaryUnaryMultiCallable[
        _GetConversationRequest, _GetConversationResponse
    ]
    ArchiveConversation: grpc.aio.UnaryUnaryMultiCallable[
        _ArchiveConversationRequest, _ArchiveConversationResponse
    ]


class AIConversationServiceServicer:
    def ListConversations(
        self,
        request: _ListConversationsRequest,
        context: grpc.aio.ServicerContext[
            _ListConversationsRequest, _ListConversationsResponse
        ],
    ) -> typing.Awaitable[_ListConversationsResponse] | _ListConversationsResponse: ...
    def GetConversation(
        self,
        request: _GetConversationRequest,
        context: grpc.aio.ServicerContext[_GetConversationRequest, _GetConversationResponse],
    ) -> typing.Awaitable[_GetConversationResponse] | _GetConversationResponse: ...
    def ArchiveConversation(
        self,
        request: _ArchiveConversationRequest,
        context: grpc.aio.ServicerContext[
            _ArchiveConversationRequest, _ArchiveConversationResponse
        ],
    ) -> typing.Awaitable[_ArchiveConversationResponse] | _ArchiveConversationResponse: ...


def add_AIConversationServiceServicer_to_server(
    servicer: AIConversationServiceServicer,
    server: grpc.aio.Server,
) -> None: ...
