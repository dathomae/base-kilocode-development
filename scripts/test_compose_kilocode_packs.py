"""Tests for compose-kilocode-packs.py."""

from __future__ import annotations

import importlib.util
import json
from pathlib import Path
import tempfile
import unittest


MODULE_PATH = Path(__file__).with_name("compose-kilocode-packs.py")
SPEC = importlib.util.spec_from_file_location("compose_kilocode_packs", MODULE_PATH)
assert SPEC and SPEC.loader
composer = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(composer)


def write_json(path: Path, value: object) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(value, indent=2) + "\n", encoding="utf-8")


def make_pack(
    root: Path,
    name: str,
    version: str,
    kind: str,
    requires: list[dict[str, str]],
    *,
    agent_name: str,
    task_kind: str,
    managed_file: str,
) -> None:
    root.mkdir(parents=True)
    write_json(
        root / "config.jsonc",
        {
            "model": "project-choice",
            "agent": {
                agent_name: {
                    "description": f"{name} agent",
                    "mode": "all",
                }
            },
        },
    )
    write_json(
        root / "task-kinds.json",
        {
            "schema_version": 1,
            "task_kinds": {
                task_kind: {
                    "description": f"{name} task",
                    "artifact": task_kind,
                    "explicit_annotation_required": kind != "core",
                    "implementors": [agent_name],
                    "default_implementor": 0,
                    "failure_retries_per_implementor": 0,
                    "reviewer": agent_name,
                    "verifier": None,
                    "critic": agent_name,
                    "contract_schema": None,
                }
            },
        },
    )
    managed = root / managed_file
    managed.parent.mkdir(parents=True, exist_ok=True)
    managed.write_text(f"managed by {name}\n", encoding="utf-8")
    write_json(
        root / composer.MANIFEST_NAME,
        {
            "schema_version": 1,
            "name": name,
            "version": version,
            "kind": kind,
            "requires": requires,
            "exports": [
                {
                    "source": managed_file,
                    "target": managed_file,
                    "strategy": "managed",
                }
            ],
            "contributions": {
                "agents": {
                    "source": "config.jsonc",
                    "source_key": "agent",
                    "target": "kilo.jsonc",
                    "target_key": "agent",
                },
                "task_kinds": {
                    "source": "task-kinds.json",
                    "source_key": "task_kinds",
                    "target": "resources/agent/task-kinds.json",
                    "target_key": "task_kinds",
                },
            },
        },
    )


class ComposePacksTests(unittest.TestCase):
    def setUp(self) -> None:
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.core = self.root / "core"
        self.hardware = self.root / "hardware"
        self.systems = self.root / "systems"
        self.project = self.root / "project"
        make_pack(
            self.core,
            "core",
            "1.0.0",
            "core",
            [],
            agent_name="code-agent",
            task_kind="code",
            managed_file="core/rules.md",
        )
        make_pack(
            self.hardware,
            "hardware",
            "2.0.0",
            "specialization",
            [{"name": "core", "version": "1.0.0"}],
            agent_name="cad-agent",
            task_kind="mechanical-cad",
            managed_file="hardware/rules.md",
        )
        make_pack(
            self.systems,
            "systems",
            "3.0.0",
            "specialization",
            [{"name": "core", "version": "1.0.0"}],
            agent_name="service-agent",
            task_kind="go-service",
            managed_file="systems/rules.md",
        )
        core_manifest = json.loads((self.core / composer.MANIFEST_NAME).read_text())
        core_manifest["exports"].extend(
            [
                {"source": "config.jsonc", "target": "kilo.jsonc", "strategy": "seed"},
                {
                    "source": "task-kinds.json",
                    "target": "resources/agent/task-kinds.json",
                    "strategy": "managed",
                },
            ]
        )
        write_json(self.core / composer.MANIFEST_NAME, core_manifest)
        self.composition = self.root / "composition.json"
        self.write_composition()

    def tearDown(self) -> None:
        self.temp.cleanup()

    def write_composition(self, packs: list[tuple[Path, str, str]] | None = None) -> None:
        packs = packs or [
            (self.systems, "systems", "3.0.0"),
            (self.core, "core", "1.0.0"),
            (self.hardware, "hardware", "2.0.0"),
        ]
        write_json(
            self.composition,
            {
                "schema_version": 1,
                "output": str(self.project),
                "packs": [
                    {"path": str(path), "name": name, "version": version}
                    for path, name, version in packs
                ],
            },
        )

    def compose(self, *, check: bool = False) -> list[str]:
        return composer.compose(self.composition, None, check)

    def test_composes_two_independent_specializations(self) -> None:
        messages = self.compose()
        self.assertIn("core@1.0.0 -> hardware@2.0.0 -> systems@3.0.0", messages[0])
        config = json.loads((self.project / "kilo.jsonc").read_text())
        self.assertEqual({"code-agent", "cad-agent", "service-agent"}, set(config["agent"]))
        task_kinds = json.loads((self.project / "resources/agent/task-kinds.json").read_text())
        self.assertEqual({"code", "mechanical-cad", "go-service"}, set(task_kinds["task_kinds"]))
        lock = json.loads((self.project / composer.LOCK_NAME).read_text())
        self.assertEqual(["core", "hardware", "systems"], [pack["name"] for pack in lock["packs"]])

    def test_safe_managed_update_and_seed_preservation(self) -> None:
        self.compose()
        config = json.loads((self.project / "kilo.jsonc").read_text())
        config["model"] = "user-selected-model"
        write_json(self.project / "kilo.jsonc", config)
        (self.hardware / "hardware/rules.md").write_text("updated hardware rules\n", encoding="utf-8")
        self.compose()
        self.assertEqual(
            "updated hardware rules\n",
            (self.project / "hardware/rules.md").read_text(encoding="utf-8"),
        )
        config = json.loads((self.project / "kilo.jsonc").read_text())
        self.assertEqual("user-selected-model", config["model"])

    def test_rejects_locally_modified_managed_file(self) -> None:
        self.compose()
        (self.project / "hardware/rules.md").write_text("local project edit\n", encoding="utf-8")
        with self.assertRaisesRegex(composer.CompositionError, "locally modified"):
            self.compose()

    def test_rejects_locally_modified_pack_owned_entry(self) -> None:
        self.compose()
        config = json.loads((self.project / "kilo.jsonc").read_text())
        config["agent"]["cad-agent"]["description"] = "local project edit"
        write_json(self.project / "kilo.jsonc", config)
        with self.assertRaisesRegex(composer.CompositionError, "pack-owned agents entry was locally modified"):
            self.compose()

    def test_rejects_named_contribution_collision(self) -> None:
        config = json.loads((self.systems / "config.jsonc").read_text())
        config["agent"] = {"cad-agent": {"description": "collision", "mode": "all"}}
        write_json(self.systems / "config.jsonc", config)
        with self.assertRaisesRegex(composer.CompositionError, "entry collision"):
            self.compose()

    def test_rejects_dependency_version_mismatch(self) -> None:
        manifest = json.loads((self.hardware / composer.MANIFEST_NAME).read_text())
        manifest["requires"][0]["version"] = "9.9.9"
        write_json(self.hardware / composer.MANIFEST_NAME, manifest)
        with self.assertRaisesRegex(composer.CompositionError, "requires core@9.9.9"):
            self.compose()

    def test_rejects_export_path_collision(self) -> None:
        manifest = json.loads((self.systems / composer.MANIFEST_NAME).read_text())
        manifest["exports"][0]["target"] = "hardware/rules.md"
        write_json(self.systems / composer.MANIFEST_NAME, manifest)
        with self.assertRaisesRegex(composer.CompositionError, "export collision"):
            self.compose()

    def test_rejects_export_file_directory_collision(self) -> None:
        manifest = json.loads((self.hardware / composer.MANIFEST_NAME).read_text())
        manifest["exports"][0]["target"] = "core"
        write_json(self.hardware / composer.MANIFEST_NAME, manifest)
        with self.assertRaisesRegex(composer.CompositionError, "file/directory collision"):
            self.compose()

    def test_rejects_nonportable_export_path(self) -> None:
        manifest = json.loads((self.hardware / composer.MANIFEST_NAME).read_text())
        manifest["exports"][0]["target"] = "..\\escape.md"
        write_json(self.hardware / composer.MANIFEST_NAME, manifest)
        with self.assertRaisesRegex(composer.CompositionError, "portable relative POSIX path"):
            self.compose()

    def test_rejects_unregistered_task_role(self) -> None:
        registry = json.loads((self.hardware / "task-kinds.json").read_text())
        registry["task_kinds"]["mechanical-cad"]["reviewer"] = "missing-reviewer"
        write_json(self.hardware / "task-kinds.json", registry)
        with self.assertRaisesRegex(composer.CompositionError, "unregistered agents"):
            self.compose()

    def test_rejects_missing_contract_schema(self) -> None:
        registry = json.loads((self.hardware / "task-kinds.json").read_text())
        registry["task_kinds"]["mechanical-cad"]["contract_schema"] = "contracts/cad.schema.json"
        write_json(self.hardware / "task-kinds.json", registry)
        with self.assertRaisesRegex(composer.CompositionError, "contract schema is missing"):
            self.compose()

    def test_reserves_final_implementor_for_escalation(self) -> None:
        registry = json.loads((self.systems / "task-kinds.json").read_text())
        kind = registry["task_kinds"]["go-service"]
        kind["implementors"] = ["service-agent", "cad-agent"]
        kind["default_implementor"] = 1
        write_json(self.systems / "task-kinds.json", registry)
        with self.assertRaisesRegex(composer.CompositionError, "reserves its final implementor"):
            self.compose()

    def test_check_does_not_create_output(self) -> None:
        messages = self.compose(check=True)
        self.assertFalse(self.project.exists())
        self.assertEqual("check passed", messages[-1])

    def test_ignores_generated_cache_files_in_directory_exports(self) -> None:
        (self.core / "bundle").mkdir()
        (self.core / "bundle/keep.md").write_text("keep\n", encoding="utf-8")
        (self.core / "bundle/__pycache__").mkdir()
        (self.core / "bundle/__pycache__/generated.pyc").write_bytes(b"generated")
        manifest = json.loads((self.core / composer.MANIFEST_NAME).read_text())
        manifest["exports"].append(
            {"source": "bundle", "target": "bundle", "strategy": "managed"}
        )
        write_json(self.core / composer.MANIFEST_NAME, manifest)

        self.compose()

        self.assertTrue((self.project / "bundle/keep.md").is_file())
        self.assertFalse((self.project / "bundle/__pycache__/generated.pyc").exists())


class ActualCorePackTests(unittest.TestCase):
    def test_core_pack_composes_and_resolves_core_registries(self) -> None:
        repository = Path(__file__).resolve().parents[1]
        with tempfile.TemporaryDirectory() as temporary:
            root = Path(temporary)
            output = root / "project"
            composition = root / "composition.json"
            write_json(
                composition,
                {
                    "schema_version": 1,
                    "output": str(output),
                    "packs": [
                        {
                            "path": str(repository),
                            "name": "base-kilocode-development",
                            "version": "0.1.0",
                        }
                    ],
                },
            )
            composer.compose(composition, None, False)
            config = json.loads((output / "kilo.jsonc").read_text())
            task_kinds = json.loads((output / "resources/agent/task-kinds.json").read_text())
            self.assertEqual({"code", "document"}, set(task_kinds["task_kinds"]))
            for entry in task_kinds["task_kinds"].values():
                roles = [*entry["implementors"], entry["reviewer"], entry["critic"]]
                self.assertTrue(set(roles).issubset(config["agent"]))
            self.assertTrue((output / composer.LOCK_NAME).is_file())


if __name__ == "__main__":
    unittest.main()
