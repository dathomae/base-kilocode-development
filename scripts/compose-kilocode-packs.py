#!/usr/bin/env python3
"""Compose version-pinned KiloCode packs into a consuming project.

The composer intentionally acquires no remote content. Composition files point at trusted local
checkouts or unpacked release artifacts, and each expected name/version is verified before use.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path, PurePosixPath
import re
import stat
import sys
import tempfile
from typing import Any


LOCK_NAME = ".kilocode-pack.lock.json"
MANIFEST_NAME = "kilocode-pack.json"
CONTRIBUTION_KINDS = ("agents", "task_kinds")
IGNORED_EXPORT_NAMES = {".DS_Store", "__pycache__"}
IGNORED_EXPORT_SUFFIXES = {".pyc", ".pyo"}
NAME_PATTERN = re.compile(r"^[a-z][a-z0-9-]*$")
VERSION_PATTERN = re.compile(r"^[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$")


class CompositionError(RuntimeError):
    """A safe, user-actionable composition failure."""


def sha256_bytes(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def canonical_sha(value: Any) -> str:
    data = json.dumps(value, sort_keys=True, separators=(",", ":")).encode("utf-8")
    return sha256_bytes(data)


def strip_jsonc(text: str) -> str:
    """Remove JavaScript comments and trailing commas without changing string contents."""
    out: list[str] = []
    index = 0
    in_string = False
    escaped = False
    while index < len(text):
        char = text[index]
        if in_string:
            out.append(char)
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == '"':
                in_string = False
            index += 1
            continue
        if char == '"':
            in_string = True
            out.append(char)
            index += 1
            continue
        if char == "/" and index + 1 < len(text) and text[index + 1] == "/":
            index += 2
            while index < len(text) and text[index] not in "\r\n":
                index += 1
            continue
        if char == "/" and index + 1 < len(text) and text[index + 1] == "*":
            end = text.find("*/", index + 2)
            if end < 0:
                raise CompositionError("unterminated block comment in JSONC")
            index = end + 2
            continue
        out.append(char)
        index += 1

    uncommented = "".join(out)
    out = []
    index = 0
    in_string = False
    escaped = False
    while index < len(uncommented):
        char = uncommented[index]
        if in_string:
            out.append(char)
            if escaped:
                escaped = False
            elif char == "\\":
                escaped = True
            elif char == '"':
                in_string = False
            index += 1
            continue
        if char == '"':
            in_string = True
            out.append(char)
            index += 1
            continue
        if char == ",":
            lookahead = index + 1
            while lookahead < len(uncommented) and uncommented[lookahead].isspace():
                lookahead += 1
            if lookahead < len(uncommented) and uncommented[lookahead] in "}]":
                index += 1
                continue
        out.append(char)
        index += 1
    return "".join(out)


def load_json(path: Path, *, jsonc: bool = False) -> dict[str, Any]:
    try:
        text = path.read_text(encoding="utf-8")
        value = json.loads(strip_jsonc(text) if jsonc else text)
    except (OSError, UnicodeError, json.JSONDecodeError, CompositionError) as error:
        raise CompositionError(f"cannot read {path}: {error}") from error
    if not isinstance(value, dict):
        raise CompositionError(f"{path} must contain a JSON object")
    return value


def safe_relative(value: str, label: str) -> PurePosixPath:
    if not isinstance(value, str) or "\\" in value or ":" in value or "\x00" in value:
        raise CompositionError(f"{label} must use a portable relative POSIX path: {value!r}")
    path = PurePosixPath(value)
    if not value or path.is_absolute() or ".." in path.parts:
        raise CompositionError(f"{label} must be a non-empty relative path without '..': {value!r}")
    return path


def output_path(root: Path, value: str) -> Path:
    relative = safe_relative(value, "output target")
    candidate = root.joinpath(*relative.parts)
    if not inside(candidate, root):
        raise CompositionError(f"output target escapes the project root: {value!r}")
    cursor = candidate
    while cursor != root:
        if cursor.is_symlink():
            raise CompositionError(f"symlinked output targets are not supported: {candidate}")
        cursor = cursor.parent
    return candidate


def inside(path: Path, root: Path) -> bool:
    try:
        path.resolve().relative_to(root.resolve())
        return True
    except ValueError:
        return False


def source_path(root: Path, value: str, label: str) -> Path:
    relative = safe_relative(value, label)
    candidate = root.joinpath(*relative.parts)
    if not inside(candidate, root):
        raise CompositionError(f"{label} escapes its pack root: {value!r}")
    if not candidate.exists():
        raise CompositionError(f"{label} does not exist: {candidate}")
    if candidate.is_symlink():
        raise CompositionError(f"symlinked pack sources are not supported: {candidate}")
    return candidate


def validate_manifest(manifest: dict[str, Any], path: Path) -> None:
    required = ("schema_version", "name", "version", "kind", "requires", "exports")
    missing = [key for key in required if key not in manifest]
    if missing:
        raise CompositionError(f"{path} is missing: {', '.join(missing)}")
    if manifest["schema_version"] != 1:
        raise CompositionError(f"unsupported schema_version in {path}: {manifest['schema_version']!r}")
    if not isinstance(manifest["name"], str) or not NAME_PATTERN.fullmatch(manifest["name"]):
        raise CompositionError(f"invalid pack name in {path}: {manifest['name']!r}")
    if not isinstance(manifest["version"], str) or not VERSION_PATTERN.fullmatch(manifest["version"]):
        raise CompositionError(f"invalid pack version in {path}: {manifest['version']!r}")
    if manifest["kind"] not in ("core", "specialization"):
        raise CompositionError(f"invalid pack kind in {path}: {manifest['kind']!r}")
    if not isinstance(manifest["requires"], list) or not isinstance(manifest["exports"], list):
        raise CompositionError(f"requires and exports must be arrays in {path}")


def read_packs(composition_path: Path) -> tuple[dict[str, Any], list[dict[str, Any]]]:
    composition = load_json(composition_path)
    if (
        composition.get("schema_version") != 1
        or not isinstance(composition.get("output"), str)
        or not composition["output"]
        or not isinstance(composition.get("packs"), list)
        or not composition["packs"]
    ):
        raise CompositionError("composition must have schema_version 1, an output path, and a non-empty packs array")
    packs: list[dict[str, Any]] = []
    names: set[str] = set()
    for entry in composition["packs"]:
        if (
            not isinstance(entry, dict)
            or not all(isinstance(entry.get(key), str) and entry[key] for key in ("path", "name", "version"))
        ):
            raise CompositionError("each composition pack needs path, name, and version")
        root = (composition_path.parent / entry["path"]).resolve()
        manifest_path = root / MANIFEST_NAME
        manifest = load_json(manifest_path)
        validate_manifest(manifest, manifest_path)
        if manifest["name"] != entry["name"] or manifest["version"] != entry["version"]:
            raise CompositionError(
                f"pack identity mismatch for {root}: expected {entry['name']}@{entry['version']}, "
                f"found {manifest['name']}@{manifest['version']}"
            )
        if manifest["name"] in names:
            raise CompositionError(f"duplicate pack name: {manifest['name']}")
        names.add(manifest["name"])
        packs.append({"root": root, "manifest_path": manifest_path, "manifest": manifest})
    return composition, packs


def order_packs(packs: list[dict[str, Any]]) -> list[dict[str, Any]]:
    by_name = {pack["manifest"]["name"]: pack for pack in packs}
    cores = [pack for pack in packs if pack["manifest"]["kind"] == "core"]
    if len(cores) != 1:
        raise CompositionError(f"composition requires exactly one core pack; found {len(cores)}")
    dependencies: dict[str, set[str]] = {}
    for pack in packs:
        manifest = pack["manifest"]
        needed: set[str] = set()
        for requirement in manifest["requires"]:
            if (
                not isinstance(requirement, dict)
                or not all(
                    isinstance(requirement.get(key), str) and requirement[key]
                    for key in ("name", "version")
                )
            ):
                raise CompositionError(f"invalid dependency in {pack['manifest_path']}")
            if requirement["name"] in needed:
                raise CompositionError(
                    f"duplicate dependency {requirement['name']!r} in {pack['manifest_path']}"
                )
            dependency = by_name.get(requirement["name"])
            if dependency is None:
                raise CompositionError(
                    f"{manifest['name']} requires missing pack {requirement['name']}@{requirement['version']}"
                )
            actual = dependency["manifest"]["version"]
            if actual != requirement["version"]:
                raise CompositionError(
                    f"{manifest['name']} requires {requirement['name']}@{requirement['version']}; found {actual}"
                )
            needed.add(requirement["name"])
        dependencies[manifest["name"]] = needed

    ordered: list[dict[str, Any]] = []
    remaining = set(by_name)
    while remaining:
        ready = sorted(name for name in remaining if not (dependencies[name] & remaining))
        if not ready:
            raise CompositionError(f"pack dependency cycle: {', '.join(sorted(remaining))}")
        for name in ready:
            ordered.append(by_name[name])
            remaining.remove(name)

    core_name = cores[0]["manifest"]["name"]
    reachable = {core_name}
    changed = True
    while changed:
        changed = False
        for name, needed in dependencies.items():
            if name not in reachable and needed & reachable:
                reachable.add(name)
                changed = True
    disconnected = sorted(set(by_name) - reachable)
    if disconnected:
        raise CompositionError(f"packs do not depend on the core: {', '.join(disconnected)}")
    return ordered


def expand_exports(ordered: list[dict[str, Any]]) -> dict[str, dict[str, Any]]:
    exports: dict[str, dict[str, Any]] = {}
    for pack in ordered:
        manifest = pack["manifest"]
        for number, entry in enumerate(manifest["exports"], 1):
            if (
                not isinstance(entry, dict)
                or not all(
                    isinstance(entry.get(key), str) and entry[key]
                    for key in ("source", "target", "strategy")
                )
            ):
                raise CompositionError(f"invalid export #{number} in {pack['manifest_path']}")
            if entry["strategy"] not in ("managed", "seed"):
                raise CompositionError(f"invalid export strategy in {pack['manifest_path']}: {entry['strategy']!r}")
            source = source_path(pack["root"], entry["source"], "export source")
            target_root = safe_relative(entry["target"], "export target")
            files = [source] if source.is_file() else sorted(
                path
                for path in source.rglob("*")
                if path.is_file()
                and not any(part in IGNORED_EXPORT_NAMES for part in path.relative_to(source).parts)
                and path.suffix not in IGNORED_EXPORT_SUFFIXES
            )
            for file_path in files:
                if file_path.is_symlink():
                    raise CompositionError(f"symlinked pack sources are not supported: {file_path}")
                relative = PurePosixPath() if source.is_file() else PurePosixPath(file_path.relative_to(source).as_posix())
                target = target_root / relative
                target_name = target.as_posix()
                if target_name == LOCK_NAME:
                    raise CompositionError(f"packs may not export the reserved lock file: {LOCK_NAME}")
                if target_name in exports:
                    other = exports[target_name]
                    raise CompositionError(
                        f"export collision at {target_name}: {other['owner']} and {manifest['name']}"
                    )
                data = file_path.read_bytes()
                exports[target_name] = {
                    "owner": manifest["name"],
                    "strategy": entry["strategy"],
                    "data": data,
                    "sha256": sha256_bytes(data),
                    "mode": stat.S_IMODE(file_path.stat().st_mode),
                }
    for target_name in exports:
        for parent in PurePosixPath(target_name).parents:
            parent_name = parent.as_posix()
            if parent_name != "." and parent_name in exports:
                raise CompositionError(f"export file/directory collision at {parent_name} and {target_name}")
    return exports


def load_old_lock(output: Path) -> dict[str, Any]:
    lock_path = output / LOCK_NAME
    if not lock_path.exists():
        return {"managed_files": {}, "seed_files": {}, "contributions": {}}
    lock = load_json(lock_path)
    if lock.get("schema_version") != 1:
        raise CompositionError(f"unsupported lock schema in {lock_path}")
    return lock


def current_bytes(output: Path, pending: dict[str, tuple[bytes | None, int]], target: str) -> bytes | None:
    if target in pending:
        return pending[target][0]
    path = output_path(output, target)
    if not path.exists():
        return None
    if not path.is_file() or path.is_symlink():
        raise CompositionError(f"output target is not a regular file: {path}")
    return path.read_bytes()


def plan_exports(
    output: Path,
    exports: dict[str, dict[str, Any]],
    old_lock: dict[str, Any],
) -> tuple[dict[str, tuple[bytes | None, int]], dict[str, Any], dict[str, Any], list[str]]:
    pending: dict[str, tuple[bytes | None, int]] = {}
    actions: list[str] = []
    old_managed = old_lock.get("managed_files", {})
    new_managed: dict[str, Any] = {}
    new_seed: dict[str, Any] = {}
    for target, entry in sorted(exports.items()):
        existing = current_bytes(output, pending, target)
        if entry["strategy"] == "seed":
            if existing is None:
                pending[target] = (entry["data"], entry["mode"])
                actions.append(f"seed {target} from {entry['owner']}")
                installed_sha = entry["sha256"]
            else:
                installed_sha = sha256_bytes(existing)
            new_seed[target] = {"owner": entry["owner"], "installed_sha256": installed_sha}
            continue

        previous = old_managed.get(target)
        if existing is not None:
            existing_sha = sha256_bytes(existing)
            if previous is None and existing_sha != entry["sha256"]:
                raise CompositionError(f"refusing to replace unowned file: {target}")
            if previous is not None and existing_sha != previous.get("sha256"):
                raise CompositionError(f"managed file was locally modified: {target}")
        if existing != entry["data"]:
            pending[target] = (entry["data"], entry["mode"])
            actions.append(f"write {target} from {entry['owner']}")
        new_managed[target] = {
            "owner": entry["owner"],
            "sha256": entry["sha256"],
            "mode": entry["mode"],
        }

    for target, previous in sorted(old_managed.items()):
        if target in new_managed:
            continue
        existing = current_bytes(output, pending, target)
        if existing is None:
            continue
        if sha256_bytes(existing) != previous.get("sha256"):
            raise CompositionError(f"removed managed file was locally modified: {target}")
        pending[target] = (None, 0)
        actions.append(f"remove obsolete managed file {target}")
    return pending, new_managed, new_seed, actions


def contribution_maps(ordered: list[dict[str, Any]], kind: str) -> tuple[str, str, dict[str, Any]] | None:
    target: str | None = None
    target_key: str | None = None
    entries: dict[str, Any] = {}
    for pack in ordered:
        contribution = pack["manifest"].get("contributions", {}).get(kind)
        if contribution is None:
            continue
        for key in ("source", "source_key", "target", "target_key"):
            if not isinstance(contribution.get(key), str) or not contribution[key]:
                raise CompositionError(f"{kind} contribution in {pack['manifest_path']} is missing {key}")
        contribution_target = safe_relative(contribution["target"], f"{kind} target").as_posix()
        if target is None:
            target, target_key = contribution_target, contribution["target_key"]
        elif target != contribution_target or target_key != contribution["target_key"]:
            raise CompositionError(f"all {kind} contributions must use the same target and target_key")
        source = source_path(pack["root"], contribution["source"], f"{kind} contribution source")
        document = load_json(source, jsonc=source.suffix == ".jsonc")
        source_entries = document.get(contribution["source_key"])
        if not isinstance(source_entries, dict):
            raise CompositionError(f"{source} field {contribution['source_key']!r} must be an object")
        for name, value in source_entries.items():
            if name in entries:
                raise CompositionError(f"{kind} entry collision for {name!r}")
            entries[name] = {"owner": pack["manifest"]["name"], "value": value}
    if target is None or target_key is None:
        return None
    return target, target_key, entries


def plan_contributions(
    output: Path,
    pending: dict[str, tuple[bytes | None, int]],
    ordered: list[dict[str, Any]],
    old_lock: dict[str, Any],
    actions: list[str],
) -> dict[str, Any]:
    new_lock: dict[str, Any] = {}
    old_groups = old_lock.get("contributions", {})
    for kind in CONTRIBUTION_KINDS:
        resolved = contribution_maps(ordered, kind)
        if resolved is None:
            continue
        target, target_key, entries = resolved
        data = current_bytes(output, pending, target)
        if data is None:
            document: dict[str, Any] = {}
        else:
            try:
                document = json.loads(strip_jsonc(data.decode("utf-8")))
            except (UnicodeError, json.JSONDecodeError, CompositionError) as error:
                raise CompositionError(f"cannot merge {kind} into {target}: {error}") from error
        current = document.setdefault(target_key, {})
        if not isinstance(current, dict):
            raise CompositionError(f"target field {target_key!r} in {target} must be an object")
        previous_entries = old_groups.get(kind, {}).get("entries", {})

        for name, previous in list(previous_entries.items()):
            if name in entries:
                continue
            if name in current and canonical_sha(current[name]) != previous.get("sha256"):
                raise CompositionError(f"removed pack-owned {kind} entry was locally modified: {name}")
            current.pop(name, None)
            actions.append(f"remove obsolete {kind} entry {name}")

        locked_entries: dict[str, Any] = {}
        for name, entry in sorted(entries.items()):
            previous = previous_entries.get(name)
            if previous is not None and name in current and canonical_sha(current[name]) != previous.get("sha256"):
                raise CompositionError(f"pack-owned {kind} entry was locally modified: {name}")
            if previous is None and name in current and canonical_sha(current[name]) != canonical_sha(entry["value"]):
                raise CompositionError(f"unowned {kind} entry collides with pack entry: {name}")
            if current.get(name) != entry["value"]:
                actions.append(f"merge {kind} entry {name} from {entry['owner']}")
            current[name] = entry["value"]
            locked_entries[name] = {"owner": entry["owner"], "sha256": canonical_sha(entry["value"])}

        rendered = (json.dumps(document, indent=2, ensure_ascii=False) + "\n").encode("utf-8")
        existing_mode = pending.get(target, (None, 0o644))[1]
        pending[target] = (rendered, existing_mode or 0o644)
        new_lock[kind] = {"target": target, "target_key": target_key, "entries": locked_entries}
    return new_lock


def parse_pending_json(
    output: Path,
    pending: dict[str, tuple[bytes | None, int]],
    target: str,
) -> dict[str, Any]:
    data = current_bytes(output, pending, target)
    if data is None:
        raise CompositionError(f"required composed document is missing: {target}")
    try:
        value = json.loads(strip_jsonc(data.decode("utf-8")))
    except (UnicodeError, json.JSONDecodeError, CompositionError) as error:
        raise CompositionError(f"cannot validate composed document {target}: {error}") from error
    if not isinstance(value, dict):
        raise CompositionError(f"composed document must be an object: {target}")
    return value


def validate_resolved_task_kinds(
    output: Path,
    pending: dict[str, tuple[bytes | None, int]],
    contributions: dict[str, Any],
) -> None:
    if "task_kinds" not in contributions or "agents" not in contributions:
        raise CompositionError("composition must resolve both agents and task_kinds contributions")
    task_group = contributions["task_kinds"]
    agent_group = contributions["agents"]
    task_document = parse_pending_json(output, pending, task_group["target"])
    agent_document = parse_pending_json(output, pending, agent_group["target"])
    task_kinds = task_document.get(task_group["target_key"])
    agents = agent_document.get(agent_group["target_key"])
    if not isinstance(task_kinds, dict) or not isinstance(agents, dict):
        raise CompositionError("resolved task-kind and agent registries must be objects")
    required = {
        "description",
        "artifact",
        "explicit_annotation_required",
        "implementors",
        "default_implementor",
        "failure_retries_per_implementor",
        "reviewer",
        "verifier",
        "critic",
        "contract_schema",
    }
    for kind_name, entry in task_kinds.items():
        if not isinstance(entry, dict):
            raise CompositionError(f"task kind {kind_name!r} must be an object")
        missing = sorted(required - set(entry))
        if missing:
            raise CompositionError(f"task kind {kind_name!r} is missing: {', '.join(missing)}")
        if not isinstance(entry["description"], str) or not entry["description"]:
            raise CompositionError(f"task kind {kind_name!r} needs a description")
        if not isinstance(entry["artifact"], str) or not entry["artifact"]:
            raise CompositionError(f"task kind {kind_name!r} needs an artifact type")
        if not isinstance(entry["explicit_annotation_required"], bool):
            raise CompositionError(f"task kind {kind_name!r} has an invalid annotation requirement")
        implementors = entry["implementors"]
        default = entry["default_implementor"]
        retries = entry["failure_retries_per_implementor"]
        if not isinstance(implementors, list) or not implementors or not all(isinstance(x, str) for x in implementors):
            raise CompositionError(f"task kind {kind_name!r} needs a non-empty implementors list")
        if not isinstance(default, int) or isinstance(default, bool) or not 0 <= default < len(implementors):
            raise CompositionError(f"task kind {kind_name!r} has an invalid default_implementor")
        if len(implementors) > 1 and default == len(implementors) - 1:
            raise CompositionError(
                f"task kind {kind_name!r} reserves its final implementor for escalation"
            )
        if not isinstance(retries, int) or isinstance(retries, bool) or retries < 0:
            raise CompositionError(f"task kind {kind_name!r} has invalid failure retries")
        reviewer = entry["reviewer"]
        critic = entry["critic"]
        verifier = entry["verifier"]
        if not isinstance(reviewer, str) or not reviewer or not isinstance(critic, str) or not critic:
            raise CompositionError(f"task kind {kind_name!r} needs reviewer and critic agent names")
        if verifier is not None and (not isinstance(verifier, str) or not verifier):
            raise CompositionError(f"task kind {kind_name!r} has an invalid verifier")
        roles = [*implementors, reviewer, critic]
        if verifier is not None:
            roles.append(verifier)
        missing_roles = sorted({role for role in roles if role not in agents})
        if missing_roles:
            raise CompositionError(
                f"task kind {kind_name!r} references unregistered agents: {', '.join(missing_roles)}"
            )
        contract = entry["contract_schema"]
        if contract is not None and (not isinstance(contract, str) or not contract):
            raise CompositionError(f"task kind {kind_name!r} has an invalid contract_schema")
        if contract is not None:
            contract_target = safe_relative(contract, f"task kind {kind_name} contract_schema").as_posix()
            if current_bytes(output, pending, contract_target) is None:
                raise CompositionError(f"task kind {kind_name!r} contract schema is missing: {contract_target}")


def render_lock(
    ordered: list[dict[str, Any]],
    managed: dict[str, Any],
    seed: dict[str, Any],
    contributions: dict[str, Any],
) -> bytes:
    packs = []
    for pack in ordered:
        manifest_bytes = pack["manifest_path"].read_bytes()
        packs.append(
            {
                "name": pack["manifest"]["name"],
                "version": pack["manifest"]["version"],
                "kind": pack["manifest"]["kind"],
                "manifest_sha256": sha256_bytes(manifest_bytes),
            }
        )
    lock = {
        "schema_version": 1,
        "packs": packs,
        "managed_files": managed,
        "seed_files": seed,
        "contributions": contributions,
    }
    return (json.dumps(lock, indent=2, ensure_ascii=False) + "\n").encode("utf-8")


def apply_pending(output: Path, pending: dict[str, tuple[bytes | None, int]]) -> None:
    for target, (data, _mode) in sorted(pending.items(), key=lambda item: len(PurePosixPath(item[0]).parts), reverse=True):
        if data is not None:
            continue
        path = output_path(output, target)
        if path.exists():
            path.unlink()
    for target, (data, mode) in sorted(pending.items()):
        if data is None:
            continue
        path = output_path(output, target)
        path.parent.mkdir(parents=True, exist_ok=True)
        descriptor, temporary_name = tempfile.mkstemp(prefix=f".{path.name}.", dir=path.parent)
        try:
            with os.fdopen(descriptor, "wb") as temporary:
                temporary.write(data)
                temporary.flush()
                os.fsync(temporary.fileno())
            os.chmod(temporary_name, mode)
            os.replace(temporary_name, path)
        finally:
            if os.path.exists(temporary_name):
                os.unlink(temporary_name)


def compose(composition_path: Path, output_override: Path | None, check: bool) -> list[str]:
    composition_path = composition_path.resolve()
    composition, packs = read_packs(composition_path)
    ordered = order_packs(packs)
    output = output_override.resolve() if output_override else (composition_path.parent / composition["output"]).resolve()
    for pack in ordered:
        if inside(output, pack["root"]):
            raise CompositionError(f"output may not be inside a pack source: {output}")
    if not check:
        output.mkdir(parents=True, exist_ok=True)
    old_lock = load_old_lock(output)
    exports = expand_exports(ordered)
    pending, managed, seed, actions = plan_exports(output, exports, old_lock)
    contributions = plan_contributions(output, pending, ordered, old_lock, actions)
    validate_resolved_task_kinds(output, pending, contributions)
    for group in contributions.values():
        target = group["target"]
        if target in managed and target in pending and pending[target][0] is not None:
            managed[target]["sha256"] = sha256_bytes(pending[target][0])
    pending[LOCK_NAME] = (render_lock(ordered, managed, seed, contributions), 0o644)
    if not check:
        apply_pending(output, pending)
    pack_order = " -> ".join(f"{pack['manifest']['name']}@{pack['manifest']['version']}" for pack in ordered)
    return [f"pack order: {pack_order}", *actions, "check passed" if check else f"wrote {output / LOCK_NAME}"]


def parse_args(argv: list[str]) -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--composition", type=Path, required=True, help="project composition JSON")
    parser.add_argument("--output", type=Path, help="override the composition output directory")
    parser.add_argument("--check", action="store_true", help="validate and report without writing")
    return parser.parse_args(argv)


def main(argv: list[str] | None = None) -> int:
    args = parse_args(sys.argv[1:] if argv is None else argv)
    try:
        for line in compose(args.composition, args.output, args.check):
            print(line)
    except CompositionError as error:
        print(f"error: {error}", file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
