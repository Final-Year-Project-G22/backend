from __future__ import annotations

from jinja2 import Environment, FileSystemLoader, TemplateNotFound, select_autoescape


class PromptLoader:
    def __init__(self, template_dir: str) -> None:
        self._env = Environment(
            loader=FileSystemLoader(template_dir),
            autoescape=select_autoescape(default=False),
            trim_blocks=True,
            lstrip_blocks=True,
        )

    def render_simple(
        self,
        locale: str,
        kb_context: str = "",
    ) -> str:
        return self._render("simple_system.j2", locale=locale, kb_context=kb_context)

    def render_agentic(
        self,
        locale: str,
        tools: list[dict[str, str]] | None = None,
    ) -> str:
        return self._render("agentic_system.j2", locale=locale, tools=tools or [])

    def render_tool_history(
        self,
        tool_history: list[dict[str, str]] | None = None,
    ) -> str:
        if not tool_history:
            return ""
        return self._render("tool_history.j2", tool_history=tool_history)

    def _render(self, template_name: str, **kwargs: object) -> str:
        try:
            template = self._env.get_template(template_name)
        except TemplateNotFound:
            msg = f"Prompt template '{template_name}' not found in '{self._env.loader}'"
            raise FileNotFoundError(msg) from None
        return template.render(**kwargs)
