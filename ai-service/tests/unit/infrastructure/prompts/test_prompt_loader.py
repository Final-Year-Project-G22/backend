from __future__ import annotations

from pathlib import Path

import pytest

from infrastructure.prompts import PromptLoader


@pytest.fixture
def prompt_loader() -> PromptLoader:
    template_dir = str(Path(__file__).resolve().parents[4] / "prompts")
    return PromptLoader(template_dir)


def test_loader_loads_simple_system(prompt_loader: PromptLoader) -> None:
    result = prompt_loader.render_simple(locale="en", kb_context="Some context")
    assert "Adisu Serategna AI Advisor" in result
    assert "locale is en" in result
    assert "Some context" in result
    assert "GROUNDED TRUTH" in result
    assert "CITATION MANDATE" in result


def test_loader_loads_agentic_system(prompt_loader: PromptLoader) -> None:
    tools = [
        {"name": "search_knowledge_base", "description": "Search KB"},
        {"name": "get_user_profile", "description": "Get profile"},
    ]
    result = prompt_loader.render_agentic(locale="am", tools=tools)

    assert "Adisu Serategna AI Advisor" in result
    assert "locale is am" in result
    assert "search_knowledge_base" in result
    assert "Search KB" in result
    assert "get_user_profile" in result
    assert "Tool usage rules" in result
    assert "GROUNDED TRUTH" in result


def test_loader_agentic_with_no_tools(prompt_loader: PromptLoader) -> None:
    result = prompt_loader.render_agentic(locale="en")
    assert "have access to the following tools" in result
    assert "No tools" not in result


def test_loader_render_tool_history(prompt_loader: PromptLoader) -> None:
    history = [
        {"tool_name": "search_knowledge_base", "result_summary": "Found 3 docs"},
        {"tool_name": "get_user_profile", "result_summary": "Restaurant, Addis"},
    ]
    result = prompt_loader.render_tool_history(tool_history=history)
    assert "Previous tool use" in result
    assert "search_knowledge_base" in result
    assert "Found 3 docs" in result
    assert "get_user_profile" in result
    assert "Restaurant, Addis" in result


def test_loader_render_tool_history_empty(prompt_loader: PromptLoader) -> None:
    result = prompt_loader.render_tool_history(tool_history=[])
    assert result == ""


def test_loader_render_tool_history_none(prompt_loader: PromptLoader) -> None:
    result = prompt_loader.render_tool_history()
    assert result == ""


def test_loader_raises_on_missing_template(tmp_path: Path) -> None:
    loader = PromptLoader(str(tmp_path / "nonexistent_prompts"))
    with pytest.raises(FileNotFoundError):
        loader.render_simple(locale="en")
