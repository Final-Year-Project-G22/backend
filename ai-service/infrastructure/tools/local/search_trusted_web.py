from __future__ import annotations

import time
from typing import Any, ClassVar

import httpx
from bs4 import BeautifulSoup

from core.domain.tools import ToolResult

TRUSTED_DOMAINS: dict[str, str] = {
    "mint.gov.et": "Ministry of Innovation and Technology",
    "erca.gov.et": "Ethiopian Revenue and Customs Authority",
    "ethiopianbusiness.org": "Ethiopian Business Development",
    "molsa.gov.et": "Ministry of Labor and Social Affairs",
    "ethiocontrol.gov.et": "Ethiopian Standards Agency",
    "ethiopianinvestment.org": "Ethiopian Investment Commission",
}

USER_AGENT = "AdisuSerategnaAI/1.0"
_MAX_CONTENT_LENGTH = 4000


class SearchTrustedWebTool:
    name = "search_trusted_web"

    description: str = (
        "Fetch content from trusted Ethiopian government and regulatory websites. "
        "Use this to get the latest official forms, fee schedules, or announcements."
    )

    parameter_schema: ClassVar[dict[str, Any]] = {
        "type": "object",
        "properties": {
            "url": {
                "type": "string",
                "description": "Full URL to fetch content from. Must be one of the trusted domains.",
            },
        },
        "required": ["url"],
    }

    def __init__(self, http_client: httpx.AsyncClient | None = None) -> None:
        self._client = http_client or httpx.AsyncClient(timeout=15.0)

    async def execute(
        self,
        arguments: dict[str, Any],
        account_id: str = "",
        user_id: str = "",
    ) -> ToolResult:
        url = arguments.get("url", "").strip()

        if not url:
            return ToolResult(
                tool_name=self.name,
                arguments=arguments,
                result_text="No URL provided.",
                success=False,
                error_message="Empty URL",
            )

        scheme_ok = url.startswith(("https://", "http://"))
        if not scheme_ok:
            return ToolResult(
                tool_name=self.name,
                arguments=arguments,
                result_text="URL must start with http:// or https://",
                success=False,
                error_message="Invalid URL scheme",
            )

        domain = url.split("/")[2].lower().removeprefix("www.")
        if not any(
            domain == trusted or domain.endswith("." + trusted) for trusted in TRUSTED_DOMAINS
        ):
            trusted_list = ", ".join(sorted(TRUSTED_DOMAINS))
            return ToolResult(
                tool_name=self.name,
                arguments=arguments,
                result_text=f"Domain not in trusted whitelist. Trusted domains: {trusted_list}",
                success=False,
                error_message=f"Untrusted domain: {domain}",
            )

        start = time.perf_counter()
        try:
            response = await self._client.get(
                url,
                headers={"User-Agent": USER_AGENT},
                follow_redirects=True,
            )
            response.raise_for_status()
        except httpx.HTTPError as e:
            elapsed = int((time.perf_counter() - start) * 1000)
            return ToolResult(
                tool_name=self.name,
                arguments=arguments,
                result_text=f"Failed to fetch {url}: {e}",
                success=False,
                error_message=str(e),
                execution_ms=elapsed,
            )

        elapsed = int((time.perf_counter() - start) * 1000)

        soup = BeautifulSoup(response.text, "html.parser")
        for tag in soup(["script", "style", "nav", "footer", "header"]):
            tag.decompose()
        text = soup.get_text(separator="\n", strip=True)

        lines = [line.strip() for line in text.split("\n") if line.strip()]
        content = "\n".join(lines[:100])

        if len(content) > _MAX_CONTENT_LENGTH:
            content = content[:_MAX_CONTENT_LENGTH] + "..."

        result_text = (
            f"Content from {url}\nSource: {TRUSTED_DOMAINS.get(domain, domain)}\n\n{content}"
        )

        return ToolResult(
            tool_name=self.name,
            arguments=arguments,
            result_text=result_text,
            success=True,
            execution_ms=elapsed,
        )
