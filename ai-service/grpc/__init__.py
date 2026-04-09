from __future__ import annotations

import importlib.util
import sys
from pathlib import Path


def _bootstrap_grpc_runtime() -> None:
    current_init = Path(__file__).resolve()
    current_dir = current_init.parent

    for entry in sys.path:
        try:
            candidate = (Path(entry) / "grpc" / "__init__.py").resolve()
        except OSError:
            continue

        if not candidate.exists() or candidate == current_init:
            continue

        spec = importlib.util.spec_from_file_location(
            __name__,
            candidate,
            submodule_search_locations=[str(candidate.parent)],
        )
        if spec is None or spec.loader is None:
            continue

        module = importlib.util.module_from_spec(spec)
        sys.modules[__name__] = module
        spec.loader.exec_module(module)

        module_paths = list(getattr(module, "__path__", []))
        local_path = str(current_dir)
        if local_path not in module_paths:
            module_paths.append(local_path)
            module.__path__ = module_paths
        return

    raise ModuleNotFoundError("grpcio runtime package is not installed")


_bootstrap_grpc_runtime()
