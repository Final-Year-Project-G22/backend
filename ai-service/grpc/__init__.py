from __future__ import annotations

import importlib.util
import sys
import sysconfig
from pathlib import Path

_site_packages = Path(sysconfig.get_paths()["purelib"])
_real_grpc_dir = _site_packages / "grpc"
_real_grpc_init = _real_grpc_dir / "__init__.py"

if not _real_grpc_init.exists():
    msg = "grpcio package is not installed in site-packages"
    raise ImportError(msg)

_spec = importlib.util.spec_from_file_location(
    __name__,
    _real_grpc_init,
    submodule_search_locations=[str(_real_grpc_dir)],
)
if _spec is None or _spec.loader is None:
    msg = "failed to load grpc package spec"
    raise ImportError(msg)

_module = importlib.util.module_from_spec(_spec)
sys.modules[__name__] = _module
_spec.loader.exec_module(_module)

globals().update(_module.__dict__)
