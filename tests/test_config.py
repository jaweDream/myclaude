import copy
import json
import unittest
from pathlib import Path

import jsonschema


CONFIG_PATH = Path(__file__).resolve().parents[1] / "config.json"
SCHEMA_PATH = Path(__file__).resolve().parents[1] / "config.schema.json"
ROOT = CONFIG_PATH.parent


def load_config():
    with CONFIG_PATH.open(encoding="utf-8") as f:
        return json.load(f)


def load_schema():
    with SCHEMA_PATH.open(encoding="utf-8") as f:
        return json.load(f)


class ConfigSchemaTest(unittest.TestCase):
    def test_config_matches_schema(self):
        config = load_config()
        schema = load_schema()
        jsonschema.validate(config, schema)

    def test_required_modules_present(self):
        modules = load_config()["modules"]
        self.assertEqual(set(modules.keys()), {"dev", "bmad", "requirements", "essentials", "advanced"})

    def test_enabled_defaults_and_flags(self):
        modules = load_config()["modules"]
        self.assertTrue(modules["dev"]["enabled"])
        self.assertTrue(modules["essentials"]["enabled"])
        self.assertFalse(modules["bmad"]["enabled"])
        self.assertFalse(modules["requirements"]["enabled"])
        self.assertFalse(modules["advanced"]["enabled"])

    def test_operations_have_expected_shape(self):
        config = load_config()
        for name, module in config["modules"].items():
            self.assertTrue(module["operations"], f"{name} should declare at least one operation")
            for op in module["operations"]:
                self.assertIn("type", op)
                if op["type"] in {"copy_dir", "copy_file"}:
                    self.assertTrue(op.get("source"), f"{name} operation missing source")
                    self.assertTrue(op.get("target"), f"{name} operation missing target")
                elif op["type"] == "run_command":
                    self.assertTrue(op.get("command"), f"{name} run_command missing command")
                    if "env" in op:
                        self.assertIsInstance(op["env"], dict)
                else:
                    self.fail(f"Unsupported operation type: {op['type']}")

    def test_operation_sources_exist_on_disk(self):
        config = load_config()
        for module in config["modules"].values():
            for op in module["operations"]:
                if op["type"] in {"copy_dir", "copy_file"}:
                    path = (ROOT / op["source"]).expanduser()
                    self.assertTrue(path.exists(), f"Source path not found: {path}")

    def test_schema_rejects_invalid_operation_type(self):
        config = load_config()
        invalid = copy.deepcopy(config)
        invalid["modules"]["dev"]["operations"][0]["type"] = "unknown_op"
        schema = load_schema()
        with self.assertRaises(jsonschema.exceptions.ValidationError):
            jsonschema.validate(invalid, schema)


if __name__ == "__main__":
    unittest.main()
