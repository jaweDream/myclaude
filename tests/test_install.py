import json
import os
import shutil
import sys
from pathlib import Path

import pytest

import install


ROOT = Path(__file__).resolve().parents[1]
SCHEMA_PATH = ROOT / "config.schema.json"


def write_config(tmp_path: Path, config: dict) -> Path:
    cfg_path = tmp_path / "config.json"
    cfg_path.write_text(json.dumps(config), encoding="utf-8")
    shutil.copy(SCHEMA_PATH, tmp_path / "config.schema.json")
    return cfg_path


@pytest.fixture()
def valid_config(tmp_path):
    sample_file = tmp_path / "sample.txt"
    sample_file.write_text("hello", encoding="utf-8")

    sample_dir = tmp_path / "sample_dir"
    sample_dir.mkdir()
    (sample_dir / "f.txt").write_text("dir", encoding="utf-8")

    config = {
        "version": "1.0",
        "install_dir": "~/.fromconfig",
        "log_file": "install.log",
        "modules": {
            "dev": {
                "enabled": True,
                "description": "dev module",
                "operations": [
                    {"type": "copy_dir", "source": "sample_dir", "target": "devcopy"}
                ],
            },
            "bmad": {
                "enabled": False,
                "description": "bmad",
                "operations": [
                    {"type": "copy_file", "source": "sample.txt", "target": "bmad.txt"}
                ],
            },
            "requirements": {
                "enabled": False,
                "description": "reqs",
                "operations": [
                    {"type": "copy_file", "source": "sample.txt", "target": "req.txt"}
                ],
            },
            "essentials": {
                "enabled": True,
                "description": "ess",
                "operations": [
                    {"type": "copy_file", "source": "sample.txt", "target": "ess.txt"}
                ],
            },
            "advanced": {
                "enabled": False,
                "description": "adv",
                "operations": [
                    {"type": "copy_file", "source": "sample.txt", "target": "adv.txt"}
                ],
            },
        },
    }

    cfg_path = write_config(tmp_path, config)
    return cfg_path, config


def make_ctx(tmp_path: Path) -> dict:
    install_dir = tmp_path / "install"
    return {
        "install_dir": install_dir,
        "log_file": install_dir / "install.log",
        "status_file": install_dir / "installed_modules.json",
        "config_dir": tmp_path,
        "force": False,
    }


def test_parse_args_defaults():
    args = install.parse_args([])
    assert args.install_dir == install.DEFAULT_INSTALL_DIR
    assert args.config == "config.json"
    assert args.module is None
    assert args.list_modules is False
    assert args.force is False


def test_parse_args_custom():
    args = install.parse_args(
        [
            "--install-dir",
            "/tmp/custom",
            "--module",
            "dev,bmad",
            "--config",
            "/tmp/cfg.json",
            "--list-modules",
            "--force",
        ]
    )

    assert args.install_dir == "/tmp/custom"
    assert args.module == "dev,bmad"
    assert args.config == "/tmp/cfg.json"
    assert args.list_modules is True
    assert args.force is True


def test_load_config_success(valid_config):
    cfg_path, config_data = valid_config
    loaded = install.load_config(str(cfg_path))
    assert loaded["modules"]["dev"]["description"] == config_data["modules"]["dev"]["description"]


def test_load_config_invalid_json(tmp_path):
    bad = tmp_path / "bad.json"
    bad.write_text("{broken", encoding="utf-8")
    shutil.copy(SCHEMA_PATH, tmp_path / "config.schema.json")
    with pytest.raises(ValueError):
        install.load_config(str(bad))


def test_load_config_schema_error(tmp_path):
    cfg = tmp_path / "cfg.json"
    cfg.write_text(json.dumps({"version": "1.0"}), encoding="utf-8")
    shutil.copy(SCHEMA_PATH, tmp_path / "config.schema.json")
    with pytest.raises(ValueError):
        install.load_config(str(cfg))


def test_resolve_paths_respects_priority(tmp_path):
    config = {
        "install_dir": str(tmp_path / "from_config"),
        "log_file": "logs/install.log",
        "modules": {},
        "version": "1.0",
    }
    cfg_path = write_config(tmp_path, config)
    args = install.parse_args(["--config", str(cfg_path)])

    ctx = install.resolve_paths(config, args)
    assert ctx["install_dir"] == (tmp_path / "from_config").resolve()
    assert ctx["log_file"] == (tmp_path / "from_config" / "logs" / "install.log").resolve()
    assert ctx["config_dir"] == tmp_path.resolve()

    cli_args = install.parse_args(
        ["--install-dir", str(tmp_path / "cli_dir"), "--config", str(cfg_path)]
    )
    ctx_cli = install.resolve_paths(config, cli_args)
    assert ctx_cli["install_dir"] == (tmp_path / "cli_dir").resolve()


def test_list_modules_output(valid_config, capsys):
    _, config_data = valid_config
    install.list_modules(config_data)
    captured = capsys.readouterr().out
    assert "dev" in captured
    assert "essentials" in captured
    assert "✓" in captured


def test_select_modules_behaviour(valid_config):
    _, config_data = valid_config

    selected_default = install.select_modules(config_data, None)
    assert set(selected_default.keys()) == {"dev", "essentials"}

    selected_specific = install.select_modules(config_data, "bmad")
    assert set(selected_specific.keys()) == {"bmad"}

    with pytest.raises(ValueError):
        install.select_modules(config_data, "missing")


def test_ensure_install_dir(tmp_path, monkeypatch):
    target = tmp_path / "install_here"
    install.ensure_install_dir(target)
    assert target.is_dir()

    file_path = tmp_path / "conflict"
    file_path.write_text("x", encoding="utf-8")
    with pytest.raises(NotADirectoryError):
        install.ensure_install_dir(file_path)

    blocked = tmp_path / "blocked"
    real_access = os.access

    def fake_access(path, mode):
        if Path(path) == blocked:
            return False
        return real_access(path, mode)

    monkeypatch.setattr(os, "access", fake_access)
    with pytest.raises(PermissionError):
        install.ensure_install_dir(blocked)


def test_op_copy_dir_respects_force(tmp_path):
    ctx = make_ctx(tmp_path)
    install.ensure_install_dir(ctx["install_dir"])

    src = tmp_path / "src"
    src.mkdir()
    (src / "a.txt").write_text("one", encoding="utf-8")

    op = {"type": "copy_dir", "source": "src", "target": "dest"}
    install.op_copy_dir(op, ctx)
    target_file = ctx["install_dir"] / "dest" / "a.txt"
    assert target_file.read_text(encoding="utf-8") == "one"

    (src / "a.txt").write_text("two", encoding="utf-8")
    install.op_copy_dir(op, ctx)
    assert target_file.read_text(encoding="utf-8") == "one"

    ctx["force"] = True
    install.op_copy_dir(op, ctx)
    assert target_file.read_text(encoding="utf-8") == "two"


def test_op_copy_file_behaviour(tmp_path):
    ctx = make_ctx(tmp_path)
    install.ensure_install_dir(ctx["install_dir"])

    src = tmp_path / "file.txt"
    src.write_text("first", encoding="utf-8")

    op = {"type": "copy_file", "source": "file.txt", "target": "out/file.txt"}
    install.op_copy_file(op, ctx)
    dst = ctx["install_dir"] / "out" / "file.txt"
    assert dst.read_text(encoding="utf-8") == "first"

    src.write_text("second", encoding="utf-8")
    install.op_copy_file(op, ctx)
    assert dst.read_text(encoding="utf-8") == "first"

    ctx["force"] = True
    install.op_copy_file(op, ctx)
    assert dst.read_text(encoding="utf-8") == "second"


def test_op_run_command_success(tmp_path):
    ctx = make_ctx(tmp_path)
    install.ensure_install_dir(ctx["install_dir"])
    install.op_run_command({"type": "run_command", "command": "echo hello"}, ctx)
    log_content = ctx["log_file"].read_text(encoding="utf-8")
    assert "hello" in log_content


def test_op_run_command_failure(tmp_path):
    ctx = make_ctx(tmp_path)
    install.ensure_install_dir(ctx["install_dir"])
    with pytest.raises(RuntimeError):
        install.op_run_command(
            {"type": "run_command", "command": f"{sys.executable} -c 'import sys; sys.exit(2)'"},
            ctx,
        )
    log_content = ctx["log_file"].read_text(encoding="utf-8")
    assert "returncode: 2" in log_content


def test_execute_module_success(tmp_path):
    ctx = make_ctx(tmp_path)
    install.ensure_install_dir(ctx["install_dir"])
    src = tmp_path / "src.txt"
    src.write_text("data", encoding="utf-8")

    cfg = {"operations": [{"type": "copy_file", "source": "src.txt", "target": "out.txt"}]}
    result = install.execute_module("demo", cfg, ctx)
    assert result["status"] == "success"
    assert (ctx["install_dir"] / "out.txt").read_text(encoding="utf-8") == "data"


def test_execute_module_failure_logs_and_stops(tmp_path):
    ctx = make_ctx(tmp_path)
    install.ensure_install_dir(ctx["install_dir"])
    cfg = {"operations": [{"type": "unknown", "source": "", "target": ""}]}

    with pytest.raises(ValueError):
        install.execute_module("demo", cfg, ctx)

    log_content = ctx["log_file"].read_text(encoding="utf-8")
    assert "failed on unknown" in log_content


def test_write_log_and_status(tmp_path):
    ctx = make_ctx(tmp_path)
    install.ensure_install_dir(ctx["install_dir"])

    install.write_log({"level": "INFO", "message": "hello"}, ctx)
    content = ctx["log_file"].read_text(encoding="utf-8")
    assert "hello" in content

    results = [
        {"module": "dev", "status": "success", "operations": [], "installed_at": "ts"}
    ]
    install.write_status(results, ctx)
    status_data = json.loads(ctx["status_file"].read_text(encoding="utf-8"))
    assert status_data["modules"]["dev"]["status"] == "success"


def test_main_success(valid_config, tmp_path):
    cfg_path, _ = valid_config
    install_dir = tmp_path / "install_final"
    rc = install.main(
        [
            "--config",
            str(cfg_path),
            "--install-dir",
            str(install_dir),
            "--module",
            "dev",
        ]
    )

    assert rc == 0
    assert (install_dir / "devcopy" / "f.txt").exists()
    assert (install_dir / "installed_modules.json").exists()


def test_main_failure_without_force(tmp_path):
    cfg = {
        "version": "1.0",
        "install_dir": "~/.claude",
        "log_file": "install.log",
        "modules": {
            "dev": {
                "enabled": True,
                "description": "dev",
                "operations": [
                    {
                        "type": "run_command",
                        "command": f"{sys.executable} -c 'import sys; sys.exit(3)'",
                    }
                ],
            },
            "bmad": {
                "enabled": False,
                "description": "bmad",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "t.txt"}
                ],
            },
            "requirements": {
                "enabled": False,
                "description": "reqs",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "r.txt"}
                ],
            },
            "essentials": {
                "enabled": False,
                "description": "ess",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "e.txt"}
                ],
            },
            "advanced": {
                "enabled": False,
                "description": "adv",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "a.txt"}
                ],
            },
        },
    }

    cfg_path = write_config(tmp_path, cfg)
    install_dir = tmp_path / "fail_install"
    rc = install.main(
        [
            "--config",
            str(cfg_path),
            "--install-dir",
            str(install_dir),
            "--module",
            "dev",
        ]
    )

    assert rc == 1
    assert not (install_dir / "installed_modules.json").exists()


def test_main_force_records_failure(tmp_path):
    cfg = {
        "version": "1.0",
        "install_dir": "~/.claude",
        "log_file": "install.log",
        "modules": {
            "dev": {
                "enabled": True,
                "description": "dev",
                "operations": [
                    {
                        "type": "run_command",
                        "command": f"{sys.executable} -c 'import sys; sys.exit(4)'",
                    }
                ],
            },
            "bmad": {
                "enabled": False,
                "description": "bmad",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "t.txt"}
                ],
            },
            "requirements": {
                "enabled": False,
                "description": "reqs",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "r.txt"}
                ],
            },
            "essentials": {
                "enabled": False,
                "description": "ess",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "e.txt"}
                ],
            },
            "advanced": {
                "enabled": False,
                "description": "adv",
                "operations": [
                    {"type": "copy_file", "source": "s.txt", "target": "a.txt"}
                ],
            },
        },
    }

    cfg_path = write_config(tmp_path, cfg)
    install_dir = tmp_path / "force_install"
    rc = install.main(
        [
            "--config",
            str(cfg_path),
            "--install-dir",
            str(install_dir),
            "--module",
            "dev",
            "--force",
        ]
    )

    assert rc == 0
    status = json.loads((install_dir / "installed_modules.json").read_text(encoding="utf-8"))
    assert status["modules"]["dev"]["status"] == "failed"
