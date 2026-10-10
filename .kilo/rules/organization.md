# Organization

There are three key top-level directories:

1. `memory-bank` — captured information referenced by each task.
2. `scripts` — step-by-step procedures; every script file ends in `.script.md`, and a Prompt Driven
   Development script ends in `.pdd.script.md`.
3. `resources` — templates, API notes, writing guidance, and the shared agent resources.

# The agent configuration

- `kilo.jsonc` (project root) — the **registration source**. One `agent` entry per persona with
  `description`, `mode` (`all` = human-selectable and coordinator-dispatchable, or `subagent` =
  internal only), `permission`, and a `prompt` that reads the persona's rules file. It also holds the
  top-level settings (`model`, `small_model`, `default_agent`, `instructions`, `provider`,
  `permission`, `indexing`, `compaction`, `experimental`). No per-persona models are set; a human
  selects the model per session.
- `.kilo/personas/<persona>.md` — the persona rules (the full role definition). One file per persona.
  Deliberately **outside** the `.kilo/agent/` discovery path, because any `.md` under `.kilo/agent/`
  auto-registers a second agent with its own `mode`/`description`. The `.kilo/agent/` directory does not
  exist.
- `.kilo/rules/` — system-level rules that apply to every persona:
  - `system-rules.md` — persona selection, boundary rules, and file discovery. Injected into every
    session via the `instructions` config field.
  - `organization.md` — this file.
  - `end_of_task_activities.md` — end-of-story and end-of-task activities.
  - `toc.md` — the index of the above.
- `resources/agent/` — the shared agent resources: `persona-registry.md` (the authoritative roster),
  `escalation-policy.md`, `change-control.md`, `artifact-standards.md`, `standards-manifests.md`,
  `work-context.md`, `report-formats.md`, `git-workflow.md`, and `toc.md`.

The registry is the single source of truth for the persona set: nothing may define a persona absent
from it, and every entry must have a matching persona file and `kilo.jsonc` registration.

# The memory-bank directory

Standard files:

1. `terms.md` — a table of terms and their definitions.
2. `concepts.md` — a table of concepts (more general or abstract than terms) with closely related
   concepts/terms.
3. `lessons-learned.md` — longer-term learnings so mistakes are not repeated.
4. `bugs.md` — coding-specific learnings, so coding bugs are not repeated.
5. `brief.md` — a high-level overview of the project.
6. `context.md` — the task(s) currently in progress, with pointers to the relevant plan/story files.
   Not a historical log: when a task or story finishes, its information is removed (at most a brief
   statement that it is complete). Updated at the start and end of every task (and story).
7. `requirements.md` — the requirements for the project.
8. `design.md` — the design of the project.

# The planning directory

All current planning documents live under `memory-bank/planning/`:

- `plans/` — targeted objectives, with `active-plans/`, `finished-plans/`, and `canceled-plans/`.
- `stories/` — larger work broken into independently executable stories, with `active-stories/`,
  `finished-stories/`, and `canceled-stories/`.
- `arcs/` — per-component/subsystem **arc files**: an **arc** is the plan and set of stories for one
  component/subsystem **within one milestone**, with `active-arcs/`, `finished-arcs/`, and
  `canceled-arcs/`. The arc file (`Arc_<ABBR>.md`) is durable and per-component: it accumulates the
  component's milestones and carries design, history, and coherence notes; it grows as more milestones
  are planned and implemented, and moves to `finished-arcs/` only when the component has no remaining
  planned work.
- `milestones/` — **delivery increments**: a milestone delivers one user-level capability, is verified by
  its **milestone acceptance flows**, and is evaluated and merged `agent-main → main`, with
  `active-milestones/`, `finished-milestones/`, and `canceled-milestones/`. Milestones are named
  `Milestone_M<n>_<short-name>.md` and form a linear sequence (one predecessor, at most one successor),
  planned in batches. Each milestone file records its order, components, flow definitions and flow
  verification results, gap register, arcs and plans, story set, preconditions, wave schedule, evaluation
  criteria and results, merge record, divergences, and remediations. Its header carries **`Stage`**
  (`definition` → `arc planning` → `story planning` → `story implementation` → `complete`, plus
  `canceled`) and **`Status`** (`Not Started` | `In Progress` | `Done`, within the current stage); passing
  a human gate advances `Stage` and resets `Status` to `Not Started`. A milestone is reviewed at several
  stages, so the `milestone-reviewer`/`milestone-critic` scope their checks to the current `Stage`.
- `tracking/` — live registers: `system-definition-tracking-status.md` (requirements, architecture and
  design, deferred items, and the §4 Resolution log) and
  `implementation-gap-tracking.md` (the source-code implementation gap, by component).

Every directory under `memory-bank/planning/` contains a `toc.md` in the list format defined by
`memory-bank/specifications/toc-format-spec.md`. The `toc.md` files in the `active-`, `finished-`, and
`canceled-` directories list the individual plans, stories, arcs, or milestones and capture each one's
state (`Not Started` | `In Progress` | `Done`).

**Plan files** capture shorter, targeted objectives: a description of the objective and a numbered,
step-by-step plan, usually executed across multiple tasks with the user checking the result between
steps. A plan is generated for one **arc within one milestone** and is named `Plan_M<n>_<ABBR>.md`; it
records the arc and the milestone it serves.

**Story files** capture larger work: one story per markdown file. Stories carry a structured
footprint/dependency section (hard `blocked-by`, soft preferred-ordering dependencies, expected
footprint with shared-infra flags, and computed `conflicts-with`) so milestone scheduling can be computed
from data rather than prose, plus an optional `flows` coverage field.

**Arc files** record the per-component work: the component's scope, the milestones it serves (with their
milestone-qualified plans and stories), and design, history, and coherence notes. **Milestone files**
record delivery: the milestone's order, components, flow definitions, gap register, arcs and plans, story
set, preconditions, wave schedule, evaluation criteria and results, flow verification results, merge
record, divergences, and remediations. `milestone-generation-coordinator` owns both; they are revised as
issues arise or specifications change.

# The historical archive

`memory-bank/historical-archive/` is a frozen record of the planning documents from before the
2026-09-16 planning reorganization. `memory-bank/specifications/design/historical-archive/` is a frozen
record of the design Epic Restraining Orders. Do not edit files in either archive and do not add new
work there.

# The scripts directory

Scripts encode step-by-step procedures. A script file ends in `.script.md`; a Prompt Driven Development
script ends in `.pdd.script.md`. The coordinator PDD scripts are:

- `story-planning.pdd.script.md` — `story-planning-coordinator`.
- `story-implementation.pdd.script.md` — `story-implementation-coordinator`.
- `plan-generation.pdd.script.md` — `plan-generation-coordinator`.
- `plan-to-story.pdd.script.md` — `plan-to-story-coordinator`.
- `milestone-generation.pdd.script.md` — `milestone-generation-coordinator`.
- `milestone-implementation.pdd.script.md` — `milestone-implementation-coordinator`.
- `change-control.pdd.script.md` — any coordinator, for remediation.

# The resources directory

The `resources` directory captures information of many types, two of which are:

1. `templates/` — template files, ending in `.template.md`.
2. `apis/` — subdirectories describing APIs or libraries.

It also holds `agent/` (the shared agent resources described above), `code-conventions.md`,
`code_review/`, and `writing_resources/`. The `resources/toc.md` and `templates/toc.md` files index
this directory.

# Mode discipline and transitions

Movement between personas and lifecycle stages follows a rigorous, non-linear path that prevents
premature action and implementation:

1. **Investigation-to-action boundary.** An inquiry in `ask` must never lead directly to a filesystem
   change; findings go to the user, who decides whether to proceed.
2. **Strategy before execution.** No implementation work may begin without an approved plan or story.
3. **Generation reviews.** A plan or story is not complete while it has an open question or a
   figure-it-out item. See `resources/agent/artifact-standards.md`.
4. **Human gates.** Milestone definition has three human gates — **A** (milestone), **B** (arc plans),
   **C** (story set and waves) — and no stage begins until the human answers the preceding gate. No
   implementation begins until gate C is approved and the human directs work to start. Escalations stop
   and ask the human. Merging `agent-main` into `main` is always the human's action. The gates advance the
   milestone's `Stage` (`definition` → `arc planning` → `story planning` → `story implementation`); the
   merge is not a stage, it is recorded in the milestone's merge record. See
   `resources/agent/git-workflow.md`.
5. **Change control.** An issue requiring changes above the story level follows the remediation ladder
   in `resources/agent/change-control.md`.
