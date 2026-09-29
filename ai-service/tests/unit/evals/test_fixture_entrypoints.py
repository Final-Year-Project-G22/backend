from __future__ import annotations

import importlib

import pytest


@pytest.mark.parametrize("script", ["provision_corpus", "verify_fixture"])
def test_fixture_script_imports_without_database_side_effects(script: str) -> None:
    importlib.import_module(f"evals.fixture.scripts.{script}")
