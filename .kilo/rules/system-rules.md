# System Rules

These rules apply to every persona. Role-specific rules live in `.kilo/personas/<persona>.md`; the
authoritative roster, dispatch map, and escalation targets are in
[`resources/agent/persona-registry.md`](../../resources/agent/persona-registry.md).

## Mandatory reads

Before any task, read:

- `memory-bank/lessons-learned.md` — avoid repeating past mistakes.
- `memory-bank/concepts.md` and `memory-bank/terms.md` — the concepts and terms in use.

Then, depending on the work: the `toc.md` chain for the area (see [File discovery](#file-discovery)),
the patterns index `memory-bank/specifications/patterns/toc.md`, and the shared resources under
[`resources/agent/`](../../resources/agent/toc.md).

## Persona selection

Personas fall into eight classes. Use the registry for the full roster and each persona's inputs,
outputs, dispatches, and escalation target.

| Class | Personas | Purpose |
|---|---|---|
| **Coordinators** | `story-planning-`, `story-implementation-`, `plan-generation-`, `plan-to-story-`, `milestone-generation-`, `milestone-implementation-coordinator` | Own a formulate/implement → review → criticize loop and its escalations. |
| **Implementors** | `small-`/`code`/`big-code-implementor`, `small-`/`write`/`big-write-implementor` | Produce code or documents for one task. |
| **Reviewers** | `spec-`, `plan-`, `story-`, `code-`, `write-`, `milestone-reviewer` | Judge conformance to **intent**; may block. |
| **Critics** | `design-`, `plan-`, `story-`, `code-`, `write-`, `milestone-critic` | Judge conformance to **standards**; raise findings. |
| **Verifiers** | `flow-verifier` | Execute a defined verification and return evidence; do not judge. |
| **Authors** | `architect`, `plan-author` | Author specifications, and plan/story prose. |
| **Integrator** | `integrator` | Perform rebase/push/PR/CI operations on command. |
| **Retained** | `ask`, `debug`, `mock` | Inquiry, defect diagnosis, and UI prototyping. |

Selection:

- **Specifications** (requirements, architecture, design): `architect`, with `spec-reviewer` and
  `design-critic`. Only the human and `architect` author specs.
- **Plans**: `plan-generation-coordinator` (with `plan-author`, `plan-reviewer`, `plan-critic`,
  `design-critic`).
- **Stories**: `plan-to-story-coordinator` → `story-planning-coordinator` (with `story-reviewer`,
  `story-critic`, `design-critic`).
- **Implementation**: `milestone-implementation-coordinator` → `story-implementation-coordinator`.
- **Milestones**: `milestone-generation-coordinator` (staged dialog with human gates A/B/C; with
  `milestone-reviewer`, `milestone-critic`; flow verification by `flow-verifier`).
- **Scheduling and integration**: `milestone-generation-coordinator` (schedule, change control) and
  `milestone-implementation-coordinator` (execution, integration).
- **Inquiry**: `ask` (read-only). **Defects outside the story flow**: `debug`. **UI mocks**: `mock`.

## Review vs critic

- A **reviewer** asks whether the artifact does the assigned activity, completely and correctly, given
  its inputs and acceptance criteria. It may block.
- A **critic** asks whether the artifact follows the project's patterns, principles, and standards,
  independent of whether it satisfies the assignment. It raises findings.
- A reviewer passing while a critic raises a blocking finding is not a conflict. Both sets of blocking
  findings are mandatory. See [`resources/agent/escalation-policy.md`](../../resources/agent/escalation-policy.md).

## Boundary rules

1. **Ask is read-only.** It never modifies the codebase, filesystem, or project state; it reports
   findings and stops.
2. **Coordinators do not implement.** A coordinator delegates all implementation and review, and edits
   only its tracking artifacts (the story/plan file, tocs, walkthrough, milestone file, tracking
   register). It holds `edit` only because subagents inherit permissions from their dispatcher.
3. **Implementors do not decide.** An implementor that needs a requirements, architecture, or design
   decision returns `Blocked — needs decision` rather than choosing.
4. **Reviewers and critics do not modify the artifact they judge.**
5. **Specs are authored only by the human and `architect`.** Coordinators detect and escalate spec
   defects; they never author specs.
6. **Escalation routes upward.** Worker → its coordinator → its invoker → the human. Only the
   top-level coordinators escalate to the human. Apply
   [`resources/agent/escalation-policy.md`](../../resources/agent/escalation-policy.md).

## File discovery

**Always consult the top-level [`toc.md`](../../toc.md) first** when looking for files, scripts,
specifications, or documentation. For source files, consult `src/toc.md` first, then the toc files it
names. Only resort to directory searching when the toc chain is insufficient. Context window is a
precious resource; use it wisely.

## Archive handling

Ignore any path containing an `archive` directory unless the user explicitly directs otherwise. The
`memory-bank/historical-archive/` and `memory-bank/specifications/design/historical-archive/` trees are
frozen records and must not be edited.

## Core organization rules

- [`organization.md`](organization.md) — directory structure and organization.
- [`end_of_task_activities.md`](end_of_task_activities.md) — end-of-task activities.

## Design and implementation patterns

Whenever performing coding, architecture, planning, or documentation work, **read the patterns index
[`memory-bank/specifications/patterns/toc.md`](../../memory-bank/specifications/patterns/toc.md)
first.**

## Above all else

1. Don't assume. Don't hide confusion. Surface tradeoffs.
2. Minimum work that solves the problem. Nothing speculative.
3. Touch only what you must. Clean up only your own mess.
4. Define success criteria. Loop until verified.
