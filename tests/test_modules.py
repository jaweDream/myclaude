import json
import shutil
import sys
from pathlib import Path

import pytest

import install


ROOT = Path(__file__).resolve().parents[1]
SCHEMA_PATH = ROOT / "config.schema.json"


def _write_schema(target_dir: Path) -> None:
    shutil.copy(SCHEMA_PATH, target_dir / "config.schema.json")


def _base_config(install_dir: Path, modules: dict) -> dict:
    return {
        "version": "1.0",
        "install_dir": str(install_dir),
        "log_file": "install.log",
        "modules": modules,
    }


def _prepare_env(tmp_path: Path, modules: dict) -> tuple[Path, Path, Path]:
    """Create a temp config directory with schema and config.json."""

    config_dir = tmp_path / "config"
    install_dir = tmp_path / "install"
    config_dir.mkdir()
    _write_schema(config_dir)

    cfg_path = config_dir / "config.json"
    cfg_path.write_text(
        json.dumps(_base_config(install_dir, modules)), encoding="utf-8"
    )
    return cfg_path, install_dir, config_dir


def _sample_sources(config_dir: Path) -> dict:
    sample_dir = config_dir / "sample_dir"
    sample_dir.mkdir()
    (sample_dir / "nested.txt").write_text("dir-content", encoding="utf-8")

    sample_file = config_dir / "sample.txt"
    sample_file.write_text("file-content", encoding="utf-8")

    return {"dir": sample_dir, "file": sample_file}


def _read_status(install_dir: Path) -> dict:
    return json.loads((install_dir / "installed_modules.json").read_text("utf-8"))


def test_single_module_full_flow(tmp_path):
    cfg_path, install_dir, config_dir = _prepare_env(
        tmp_path,
        {
            "solo": {
                "enabled": True,
                "description": "single module",
                "operations": [
                    {"type": "copy_dir", "source": "sample_dir", "target": "payload"},
                    {
                        "type": "copy_file",
                        "source": "sample.txt",
                        "target": "payload/sample.txt",
                    },
                    {
                        "type": "run_command",
                        "command": f"{sys.executable} -c \"from pathlib import Path; Path('run.txt').write_text('ok', encoding='utf-8')\"",
                    },
                ],
            }
        },
    )

    _sample_sources(config_dir)
    rc = install.main(["--config", str(cfg_path), "--module", "solo"])

    assert rc == 0
    assert (install_dir / "payload" / "nested.txt").read_text(encoding="utf-8") == "dir-content"
    assert (install_dir / "payload" / "sample.txt").read_text(encoding="utf-8") == "file-content"
    assert (install_dir / "run.txt").read_text(encoding="utf-8") == "ok"

    status = _read_status(install_dir)
    assert status["modules"]["solo"]["status"] == "success"
    assert len(status["modules"]["solo"]["operations"]) == 3


def test_multi_module_install_and_status(tmp_path):
    modules = {
        "alpha": {
            "enabled": True,
            "description": "alpha",
            "operations": [
                {
                    "type": "copy_file",
                    "source": "sample.txt",
                    "target": "alpha.txt",
                }
            ],
        },
        "beta": {
            "enabled": True,
            "description": "beta",
            "operations": [
                {
                    "type": "copy_dir",
                    "source": "sample_dir",
                    "target": "beta_dir",
                }
            ],
        },
    }

    cfg_path, install_dir, config_dir = _prepare_env(tmp_path, modules)
    _sample_sources(config_dir)

    rc = install.main(["--config", str(cfg_path)])
    assert rc == 0

    assert (install_dir / "alpha.txt").read_text(encoding="utf-8") == "file-content"
    assert (install_dir / "beta_dir" / "nested.txt").exists()

    status = _read_status(install_dir)
    assert set(status["modules"].keys()) == {"alpha", "beta"}
    assert all(mod["status"] == "success" for mod in status["modules"].values())


def test_force_overwrites_existing_files(tmp_path):
    modules = {
        "forcey": {
            "enabled": True,
            "description": "force copy",
            "operations": [
                {
                    "type": "copy_file",
                    "source": "sample.txt",
                    "target": "target.txt",
                }
            ],
        }
    }

    cfg_path, install_dir, config_dir = _prepare_env(tmp_path, modules)
    sources = _sample_sources(config_dir)

    install.main(["--config", str(cfg_path), "--module", "forcey"])
    assert (install_dir / "target.txt").read_text(encoding="utf-8") == "file-content"

    sources["file"].write_text("new-content", encoding="utf-8")

    rc = install.main(["--config", str(cfg_path), "--module", "forcey", "--force"])
    assert rc == 0
    assert (install_dir / "target.txt").read_text(encoding="utf-8") == "new-content"

    status = _read_status(install_dir)
    assert status["modules"]["forcey"]["status"] == "success"


def test_failure_triggers_rollback_and_restores_status(tmp_path):
    # First successful run to create a known-good status file.
    ok_modules = {
        "stable": {
            "enabled": True,
            "description": "stable",
            "operations": [
                {
                    "type": "copy_file",
                    "source": "sample.txt",
                    "target": "stable.txt",
                }
            ],
        }
    }

    cfg_path, install_dir, config_dir = _prepare_env(tmp_path, ok_modules)
    _sample_sources(config_dir)
    assert install.main(["--config", str(cfg_path)]) == 0
    pre_status = _read_status(install_dir)
    assert "stable" in pre_status["modules"]

    # Rewrite config to introduce a failing module.
    failing_modules = {
        **ok_modules,
        "broken": {
            "enabled": True,
            "description": "will fail",
            "operations": [
                {
                    "type": "copy_file",
                    "source": "sample.txt",
                    "target": "broken.txt",
                },
                {
                    "type": "run_command",
                    "command": f"{sys.executable} -c 'import sys; sys.exit(5)'",
                },
            ],
        },
    }

    cfg_path.write_text(
        json.dumps(_base_config(install_dir, failing_modules)), encoding="utf-8"
    )

    rc = install.main(["--config", str(cfg_path)])
    assert rc == 1

    # The failed module's file should have been removed by rollback.
    assert not (install_dir / "broken.txt").exists()
    # Previously installed files remain.
    assert (install_dir / "stable.txt").exists()

    restored_status = _read_status(install_dir)
    assert restored_status == pre_status

    log_content = (install_dir / "install.log").read_text(encoding="utf-8")
    assert "Rolling back" in log_content

