from __future__ import annotations

import uuid

import httpx
import pytest

from infrastructure.tools.local.search_trusted_web import SearchTrustedWebTool


@pytest.fixture
def tool() -> SearchTrustedWebTool:
    return SearchTrustedWebTool()


def test_tool_name(tool: SearchTrustedWebTool) -> None:
    assert tool.name == "search_trusted_web"


def test_parameter_schema(tool: SearchTrustedWebTool) -> None:
    schema = tool.parameter_schema
    assert schema["type"] == "object"
    assert "url" in schema["properties"]
    assert schema["required"] == ["url"]


@pytest.mark.asyncio
async def test_execute_empty_url(tool: SearchTrustedWebTool) -> None:
    result = await tool.execute(
        arguments={"url": ""},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is False
    assert result.error_message == "Empty URL"


@pytest.mark.asyncio
async def test_execute_invalid_scheme(tool: SearchTrustedWebTool) -> None:
    result = await tool.execute(
        arguments={"url": "ftp://mint.gov.et/file"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is False
    assert "URL must start with" in result.result_text


@pytest.mark.asyncio
async def test_execute_untrusted_domain(tool: SearchTrustedWebTool) -> None:
    result = await tool.execute(
        arguments={"url": "https://malicious.site/evil"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )
    assert result.success is False
    assert "not in trusted whitelist" in result.result_text


@pytest.mark.asyncio
async def test_execute_http_failure() -> None:
    transport = httpx.MockTransport(lambda _: httpx.Response(500))
    client = httpx.AsyncClient(transport=transport)
    tool = SearchTrustedWebTool(http_client=client)

    result = await tool.execute(
        arguments={"url": "https://www.mint.gov.et/page"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is False
    assert "500" in result.error_message or "Failed" in result.result_text


@pytest.mark.asyncio
async def test_execute_success_with_trusted_domain() -> None:
    html = "<html><body><h1>Trade License Info</h1><p>Apply at your local office.</p></body></html>"

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text=html)

    transport = httpx.MockTransport(handler)
    client = httpx.AsyncClient(transport=transport)
    tool = SearchTrustedWebTool(http_client=client)

    result = await tool.execute(
        arguments={"url": "https://www.mint.gov.et/trade-license"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is True
    assert "Trade License Info" in result.result_text
    assert "mint.gov.et" in result.result_text


@pytest.mark.asyncio
async def test_execute_content_truncation() -> None:
    long_text = "word " * 3000
    html = f"<html><body><p>{long_text}</p></body></html>"

    def handler(request: httpx.Request) -> httpx.Response:
        return httpx.Response(200, text=html)

    transport = httpx.MockTransport(handler)
    client = httpx.AsyncClient(transport=transport)
    tool = SearchTrustedWebTool(http_client=client)

    result = await tool.execute(
        arguments={"url": "http://mint.gov.et/page"},
        account_id=str(uuid.uuid4()),
        user_id=str(uuid.uuid4()),
    )

    assert result.success is True
    assert len(result.result_text) < 5000
