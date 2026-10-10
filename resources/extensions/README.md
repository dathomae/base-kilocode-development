# KiloCode Specialization Packs

The toolkit has one domain-neutral core and any number of independently versioned specialization
packs. A consuming project may compose no specialization, one specialization, or several. Packs are
capability layers, not mutually exclusive project types.

## Pack contract

Every pack contains `kilocode-pack.json` at its repository root. The manifest declares:

- an immutable pack name and semantic version;
- whether it is the `core` or a `specialization`;
- exact-version dependencies on other packs;
- files and directories exported with either `managed` or `seed` ownership;
- optional named contributions to the KiloCode agent registry and task-kind registry.

Exact dependency versions are deliberate. A specialization release is tested against one core
release; changing that relationship requires a new specialization release.

## Ownership strategies

- **Managed:** owned by the supplying pack. An update replaces it only when the installed file still
  matches the prior lock. A local edit causes a conflict instead of being overwritten.
- **Seed:** copied only when absent and then owned by the consuming project. Updates never replace it.

Agent and task-kind contributions are merged by named entry. Two packs may not claim the same name.
Project-defined entries are preserved; a pack may not silently take them over. Pack-owned entries are
updated only when the installed value still matches the prior lock.

## Composition

Create a project composition file from
`resources/templates/kilocode-composition.template.json`, pointing at local checkouts or unpacked
release artifacts. Then run:

```bash
python scripts/compose-kilocode-packs.py --composition kilocode-composition.json
```

The composer validates dependencies, orders packs topologically, detects path and named-entry
collisions, applies safe updates, and writes `.kilocode-pack.lock.json`. Use `--check` to validate the
same operation without writing. Commit both `kilocode-composition.json` and
`.kilocode-pack.lock.json` in the consuming project so its declared inputs and resolved installation
remain reviewable and reproducible.

The first version intentionally separates artifact acquisition from composition. Git checkouts,
release automation, or another trusted downloader supplies the pack directories; the composer never
selects an unpinned remote revision.

## Task-kind extension point

The resolved `resources/agent/task-kinds.json` drives task routing. Core supplies `code` and
`document`. A specialization contributes explicitly annotated kinds such as mechanical CAD, PCB
design, embedded firmware, or a domain-specific verifier without patching the core coordinator.

Each contributed kind names its implementor ladder, reviewer, optional independent verifier, critic,
artifact type, and optional contract schema. Specialized tasks use `[Kind: <id>]` and must satisfy
the referenced contract before dispatch.

Implementors are ordered from the narrowest/least costly role to the strongest fallback. The
zero-based `default_implementor` selects normal non-small work; when a kind has more than one
implementor, its final entry is reserved for escalation. `failure_retries_per_implementor` controls
fresh-worktree retries for execution failures and is separate from the review-fix limit.

## Authoring a specialization

Copy `resources/templates/kilocode-specialization-pack.template.json` to `kilocode-pack.json` in the
new repository. Keep domain-owned files
under paths unique to the specialization, and use named contributions for agents and task kinds that
must join the shared registries. A typical repository is:

```text
kilocode-pack.json
agents.json
task-kinds.json
.kilo/personas/<domain-persona>.md
resources/<domain>/...
```

The specialization manifest must pin the exact core release it was tested with. Its contributed
agent names and task-kind ids are public identifiers: prefix them with a stable domain term, do not
rename them within a release line, and never reuse an identifier for different semantics.

`agents.json` contains only the entries that the specialization adds:

```json
{
  "agent": {
    "example-domain-implementor": {
      "description": "Implement one example-domain task",
      "mode": "subagent",
      "permission": { "read": "allow", "edit": "allow", "bash": "allow", "task": "deny", "skill": "allow" },
      "prompt": "Read .kilo/personas/example-domain-implementor.md for instructions."
    },
    "example-domain-reviewer": {
      "description": "Review an example-domain artifact against its task intent",
      "mode": "subagent",
      "permission": { "read": "allow", "edit": "deny", "bash": "allow", "task": "deny", "skill": "allow" },
      "prompt": "Read .kilo/personas/example-domain-reviewer.md for instructions."
    },
    "example-domain-verifier": {
      "description": "Verify an example-domain artifact and report evidence",
      "mode": "subagent",
      "permission": { "read": "allow", "edit": "deny", "bash": "allow", "task": "deny", "skill": "allow" },
      "prompt": "Read .kilo/personas/example-domain-verifier.md for instructions."
    },
    "example-domain-critic": {
      "description": "Critique an example-domain artifact against domain standards",
      "mode": "subagent",
      "permission": { "read": "allow", "edit": "deny", "bash": "allow", "task": "deny", "skill": "allow" },
      "prompt": "Read .kilo/personas/example-domain-critic.md for instructions."
    }
  }
}
```

`task-kinds.json` maps an explicit story annotation to that pack's execution roles:

```json
{
  "schema_version": 1,
  "task_kinds": {
    "example-domain": {
      "description": "Produce and verify an example-domain artifact",
      "artifact": "example-domain-artifact",
      "explicit_annotation_required": true,
      "implementors": ["example-domain-implementor"],
      "default_implementor": 0,
      "failure_retries_per_implementor": 0,
      "reviewer": "example-domain-reviewer",
      "verifier": "example-domain-verifier",
      "critic": "example-domain-critic",
      "contract_schema": "resources/example-domain/task-contract.schema.json"
    }
  }
}
```

If a task requires domain-specific planning fields or verification evidence, export a JSON Schema
at a unique project-relative path and set the task kind's `contract_schema` to that installed path.
The composer verifies the schema exists and that every referenced implementor, reviewer, verifier,
and critic is registered.

