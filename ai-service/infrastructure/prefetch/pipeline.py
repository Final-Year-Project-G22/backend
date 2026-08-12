from __future__ import annotations

import asyncio
import logging

from core.ports.intent_classifier import IntentClass
from core.ports.tool_registry import ToolRegistryPort

logger = logging.getLogger(__name__)


class PreFetchPipeline:
    def __init__(self, tool_registry: ToolRegistryPort) -> None:
        self._tool_registry = tool_registry

    async def pre_fetch(
        self,
        intent: IntentClass,
        query: str,
        account_id: str,
        user_id: str,
        *,
        include_kb: bool = True,
    ) -> dict[str, str]:
        results: dict[str, str] = {}

        if intent == IntentClass.KNOWLEDGE:
            if include_kb:
                _, value = await self._fetch_kb(query, account_id, user_id)
                if value:
                    results["kb"] = value

        elif intent == IntentClass.PERSONAL:
            tasks = [
                asyncio.create_task(self._fetch_profile(account_id, user_id)),
                asyncio.create_task(self._fetch_guide_progress(account_id, user_id)),
                asyncio.create_task(self._fetch_compliance(account_id, user_id)),
            ]
            done, _ = await asyncio.wait(tasks, return_when=asyncio.ALL_COMPLETED)
            for task in done:
                key, value = task.result()
                if value:
                    results[key] = value

        else:
            tasks = [asyncio.create_task(self._fetch_profile(account_id, user_id))]
            if include_kb:
                tasks.append(asyncio.create_task(self._fetch_kb(query, account_id, user_id)))
            tasks.extend(
                [
                    asyncio.create_task(self._fetch_guide_progress(account_id, user_id)),
                    asyncio.create_task(self._fetch_compliance(account_id, user_id)),
                ]
            )
            done, _ = await asyncio.wait(tasks, return_when=asyncio.ALL_COMPLETED)
            for task in done:
                key, value = task.result()
                if value:
                    results[key] = value

        return results

    async def _fetch_kb(self, query: str, account_id: str, user_id: str) -> tuple[str, str]:
        try:
            result = await self._tool_registry.execute_tool(
                "search_knowledge_base",
                {"query": query, "top_k": 5},
                account_id,
                user_id,
            )
            return "kb", result.result_text
        except Exception:
            logger.exception("Pre-fetch KB failed")
            return "kb", ""

    async def _fetch_profile(self, account_id: str, user_id: str) -> tuple[str, str]:
        try:
            result = await self._tool_registry.execute_tool(
                "get_user_profile",
                {},
                account_id,
                user_id,
            )
            return "profile", result.result_text
        except Exception:
            logger.exception("Pre-fetch profile failed")
            return "profile", ""

    async def _fetch_guide_progress(self, account_id: str, user_id: str) -> tuple[str, str]:
        try:
            result = await self._tool_registry.execute_tool(
                "get_guide_progress",
                {},
                account_id,
                user_id,
            )
            return "progress", result.result_text
        except Exception:
            logger.exception("Pre-fetch guide progress failed")
            return "progress", ""

    async def _fetch_compliance(self, account_id: str, user_id: str) -> tuple[str, str]:
        try:
            result = await self._tool_registry.execute_tool(
                "check_compliance_status",
                {},
                account_id,
                user_id,
            )
            return "compliance", result.result_text
        except Exception:
            logger.exception("Pre-fetch compliance failed")
            return "compliance", ""
