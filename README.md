# Base KiloCode Development Toolkit

This repository contains the reusable agent framework extracted from the Chronocone project. It is
project-neutral: product code, product specifications, active plans, generated worktrees, session
state, demos, and Chronocone-specific operational scripts are intentionally excluded.

## Contents

- `kilo.jsonc` — KiloCode agent registration and permissions.
- `.kilo/personas/` — role definitions for coordinators, implementors, reviewers, critics, authors,
  verifiers, and integration helpers.
- `.kilo/rules/` — system-wide operating rules.
- `resources/agent/` — lifecycle, escalation, artifact, reporting, and Git workflow contracts.
- `resources/extensions/` — specialization-pack, task-kind, composition, and ownership contracts.
- `resources/templates/` — plans, stories, milestones, arcs, decisions, and walkthrough templates.
- `scripts/` — reusable Prompt Driven Development workflows and optional GitHub App setup tooling.
- `memory-bank/` — a deliberately empty project scaffold satisfying the framework's expected paths.

## Using the toolkit

Pin a released version rather than automatically following `main`. Create a composition file from
`resources/templates/kilocode-composition.template.json`, add any version-pinned specialization
packs, and run `scripts/compose-kilocode-packs.py`. The resulting lock records exactly which pack owns
each managed file and named agent/task-kind entry; updates stop rather than overwrite local changes.

Then customize these project-owned seed files:

1. `memory-bank/brief.md`, `terms.md`, `concepts.md`, `requirements.md`, and `design.md`.
2. `resources/code-conventions.md` and any language-specific review guidance.
3. `memory-bank/specifications/architecture/`, `design/`, `requirements/`, and `patterns/`.
4. The default models and permissions in `kilo.jsonc`.
5. The pre-PR verification command described in `resources/agent/git-workflow.md`.

Specialization packs can add task kinds, personas, standards, templates, and unique files without
patching this repository. A project may compose multiple independent packs—for example hardware
development and complex-system development—when it needs both capabilities.

The GitHub App identity and protected-branch workflow are optional. If used, start with
`AGENT_REPOSITORY_SETUP.md`; do not commit App keys or installation tokens.

## Source boundary

Chronocone was used only as a read-only source. Files in this repository should contain no Chronocone
product plans, source code, credentials, repository identifiers, or active Kilo session data.
