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
- `resources/extensions/` — project setup, specialization-pack, task-kind, composition, and ownership
  contracts.
- `resources/templates/` — plans, stories, milestones, arcs, decisions, and walkthrough templates.
- `scripts/` — reusable Prompt Driven Development workflows and protected GitHub App setup tooling.
- `memory-bank/` — a deliberately empty project scaffold satisfying the framework's expected paths.

## Using the toolkit

Start with the step-by-step [`project setup guide`](resources/extensions/project-setup.md). It covers
core-only software, hardware/manufacturing, complex-system, and combined projects. Pin a released
version rather than automatically following `main`; the resulting composition lock records exactly
which pack owns each managed file and named agent/task-kind entry, and updates stop rather than
overwrite local changes.

Then customize these project-owned seed files:

1. `memory-bank/brief.md`, `terms.md`, `concepts.md`, `requirements.md`, and `design.md`.
2. `resources/code-conventions.md` and any language-specific review guidance.
3. `memory-bank/specifications/architecture/`, `design/`, `requirements/`, and `patterns/`.
4. The default models and permissions in `kilo.jsonc`.
5. The pre-PR verification command described in `resources/agent/git-workflow.md`.

Specialization packs can add task kinds, personas, standards, templates, and unique files without
patching this repository. A project may compose multiple independent packs—for example hardware
development and complex-system development—when it needs both capabilities.

The GitHub App identity and protected-branch workflow are optional only for projects that do not use
agent milestone implementation. Before an agent implements a milestone, complete
`AGENT_REPOSITORY_SETUP.md`; the agent must work from `agent-main` or its child branches and can never
modify `main`. Do not commit App keys or installation tokens.

## Source boundary

Chronocone was used only as a read-only source. Files in this repository should contain no Chronocone
product plans, source code, credentials, repository identifiers, or active Kilo session data.
