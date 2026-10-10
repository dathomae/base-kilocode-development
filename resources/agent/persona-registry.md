# Agent Persona Registry

This file is the authoritative registry of the toolkit's agent personas. It defines every persona, its
class and level, how it is registered, whom it dispatches, to whom it escalates, and the standards it
is judged against. The persona rules under `.kilo/personas/` implement this registry; nothing may define
a persona that is absent here, and nothing here may lack a persona file.

A **milestone** is a delivery increment that delivers one user-level capability, is verified by its
**milestone acceptance flows**, and is evaluated and merged `agent-main → main` by a human. An **arc** is
the plan and set of stories for one component/subsystem within one milestone; the durable arc file
(`Arc_<ABBR>.md`) grows across milestones. The delivery coordinators are named `milestone-*`; the
artifacts live in `memory-bank/planning/arcs/` (arc files) and `memory-bank/planning/milestones/`
(milestone files).

Related: [`escalation-policy.md`](escalation-policy.md), [`change-control.md`](change-control.md),
[`artifact-standards.md`](artifact-standards.md), [`standards-manifests.md`](standards-manifests.md),
[`work-context.md`](work-context.md), [`report-formats.md`](report-formats.md),
[`git-workflow.md`](git-workflow.md).

## Registration model

The registration model keeps one explicit source of truth and avoids accidental duplicate discovery.

1. **Registration source.** Every persona is registered in the `kilo.jsonc` `agent` block — the
   declaration the user requires. Each entry carries `description`, `mode`, `permission`, and a
   `prompt` that reads the persona's rules file (the pattern the project already uses).
2. **Rules location.** The persona rules live in `.kilo/personas/<name>.md`, deliberately **outside**
   the `{agent,agents}/**/*.md` discovery path, because any `.md` under `.kilo/agent/` auto-registers
   a second agent with its own `mode`/`description` — the drift this revision removes. `.kilo/agent/`
   is retired.
3. **Mode.** `mode: all` registers a persona that is **both** selectable as a main agent by a human
   and dispatchable via the Task tool by a coordinator; internal-only personas use `mode: subagent`.
   Dual-surface personas therefore need a single `all` entry, not paired primary/subagent
   declarations.
4. **Models.** No model is named for any persona (plan decision 13): the top-level `model` is the
   default and a human selects the model per session. A model *can* be named in agent frontmatter or
   in `kilo.jsonc`, but because rules files are not agent definitions, `kilo.jsonc` is the single
   place to do so if per-persona models are ever wanted.
5. **Runtime confirmation.** Whether a `mode: all` persona is Task-dispatchable from a *primary*
   session (top-level agents cannot invoke other top-level agents) is confirmed in the fresh-session
   verification (plan step 13). Fallback if it is not: register dispatchable personas as
   `mode: subagent` plus a separate `mode: primary` alias that reads the same rules file.

## Naming conventions

- Role-qualified, hyphenated names: `<artifact-or-domain>-<role>`, e.g. `story-reviewer`,
  `code-critic`, `write-implementor`.
- Coordinators are `<level>-<phase>-coordinator`, e.g. `story-planning-coordinator`.
- The invoker-oriented `-for-story-implementor` suffix is retired; workers are reusable by any
  coordinator.
- The canonical rules file is `.kilo/personas/<name>.md`, registered by a `kilo.jsonc` `agent` entry
  of the same name.

## Persona inventory

31 personas in eight classes. `Mode` is the frontmatter `mode`; `all` means human-selectable and
coordinator-dispatchable.

### Coordinators (6)

| Persona | Mode | Level | Dispatches | Escalates to | Input → Output |
|---|---|---|---|---|---|
| `story-planning-coordinator` | all | story (formulate) | `plan-author`, `story-reviewer`, `story-critic`, `design-critic` | `plan-to-story-coordinator`, else human | plan + settled decisions → story file, toc entry |
| `story-implementation-coordinator` | all | story (execute) | `small`/`code`/`big-code-implementor`, `small`/`write`/`big-write-implementor`, `code-reviewer`, `write-reviewer`, `story-reviewer`, `code-critic`, `write-critic`, `integrator` | `milestone-implementation-coordinator`, else human | story → implemented story, walkthrough, PR, status |
| `plan-generation-coordinator` | all | plan (formulate) | `plan-author`, `plan-reviewer`, `plan-critic`, `design-critic` | `milestone-generation-coordinator`, else human | milestone + arc → arc plan |
| `plan-to-story-coordinator` | all | plan (formulate) | `story-planning-coordinator`, `story-reviewer`, `story-critic`, `design-critic` | `milestone-generation-coordinator`, else human | plan (+ milestone flows) → story set, tocs, conflict resolution |
| `milestone-generation-coordinator` | all | milestone (formulate) | `architect`; `plan-generation-coordinator`, `plan-to-story-coordinator`; `milestone-reviewer`, `milestone-critic`, `plan-reviewer`, `story-reviewer`, `plan-critic`, `story-critic` | human | milestone intent → milestone + flows + arcs + plans + stories + wave schedule |
| `milestone-implementation-coordinator` | all | milestone (execute) | `story-implementation-coordinator`, `integrator`, `milestone-generation-coordinator` (consult); close evaluation: `flow-verifier`, `milestone-reviewer`, `milestone-critic`, `design-critic`, `code-critic`/`write-critic`, `story-reviewer`/`plan-reviewer` | human | milestone file → merged stories on `agent-main`, close evaluation, human merge |

The milestone-generation coordinator is the **change-control owner** during delivery and owns the
milestone definition, its acceptance flows, the arcs, and the schedule; milestone-implementation pauses,
holds, and resumes. See [`change-control.md`](change-control.md).

### Implementors (6)

| Persona | Mode | Level | Dispatches | Escalates to | Input → Output |
|---|---|---|---|---|---|
| `code-implementor` | all | task (code) | — | its coordinator | work context → code + tests, or `Blocked — needs decision` |
| `write-implementor` | all | task (writing) | — | its coordinator | work context → document changes |
| `small-code-implementor` | subagent | task (code, narrow) | — | its coordinator | work context → code + tests |
| `big-code-implementor` | subagent | task (code, large) | — | its coordinator | work context → code + tests |
| `small-write-implementor` | subagent | task (writing, narrow) | — | its coordinator | work context → document changes |
| `big-write-implementor` | subagent | task (writing, large) | — | its coordinator | work context → document changes |

### Reviewers (6)

Reviewers judge **conformance to intent**: does the artifact do the assigned activity, completely and
correctly? They may block.

| Persona | Mode | Reviews | Escalates to | Input → Output |
|---|---|---|---|---|
| `spec-reviewer` | all | requirements/architecture/design vs their assigned purpose | its invoker | spec + requirements → review report |
| `plan-reviewer` | all | plans vs the specs they implement | its invoker | plan + specs → review report |
| `story-reviewer` | all | stories (authoring) and completed story results (implementation) | its invoker | story (+ plan) → review report |
| `code-reviewer` | all | code changes vs the task specification | its invoker | task + diff → review report |
| `write-reviewer` | all | document changes vs the task specification | its invoker | task + diff → review report |
| `milestone-reviewer` | all | milestones vs the agreed capability, flows, components, and arcs (stage-scoped to the milestone's `Stage`) | its invoker | milestone (+ stage decisions) → review report |

### Critics (6)

Critics judge **conformance to standards**: does the artifact follow the project's patterns,
principles, and standards, independent of whether it satisfies the assignment? They raise findings;
they do not re-litigate the assignment. See [`standards-manifests.md`](standards-manifests.md).

| Persona | Mode | Critiques | Standards manifest | Input → Output |
|---|---|---|---|---|
| `design-critic` | all | specs (requirements/architecture/design) | architecture/design specs, patterns toc, architecture Definition of Done | artifact → critic report |
| `plan-critic` | all | plans | plan template, `artifact-standards.md`, planning organization, toc-format spec | artifact → critic report |
| `story-critic` | all | stories | story template, sizing/parallelism rules, `artifact-standards.md`, toc-format spec | artifact → critic report |
| `code-critic` | all | code changes and walkthroughs | `resources/code-conventions.md`, `resources/code_review/*`, patterns toc | artifact → critic report |
| `write-critic` | all | documents and walkthroughs | `resources/writing_resources/*`, `terms.md`, toc-format spec | artifact → critic report |
| `milestone-critic` | all | milestones (stage-scoped to the milestone's `Stage`) | milestone template, `artifact-standards.md` (Milestone/Arc), flow rules, planning organization, toc-format spec | artifact → critic report |

### Verifiers (1)

Verifiers execute a defined verification and return evidence; they do not judge intent or standards, and
they do not modify the repository.

| Persona | Mode | Verifies | Escalates to | Input → Output |
|---|---|---|---|---|
| `flow-verifier` | subagent | one milestone acceptance flow (automated test, agent-driven, or manual report) | `milestone-implementation-coordinator` | flow definition + revision → flow-verification result |

### Authors (2)

| Persona | Mode | Level | Dispatches | Escalates to | Input → Output |
|---|---|---|---|---|---|
| `architect` | all | spec | — | human | requirements + discussion → architecture/design documents |
| `plan-author` | all | plan | — | its coordinator | specs + objective → plan or story prose |

`plan-author` replaces `work-planner` as a *worker*: it authors plan and story prose on behalf of the
generation coordinators rather than being a top-level planning mode.

### Integrator (1)

| Persona | Mode | Level | Dispatches | Escalates to | Input → Output |
|---|---|---|---|---|---|
| `integrator` | subagent | milestone (execute) | — | `milestone-implementation-coordinator` | branch/PR/CI instruction → rebased branch, PR, CI status |

The integrator never merges and never pushes `main`.

### Retained (3)

| Persona | Mode | Notes |
|---|---|---|
| `ask` | all | Read-only inquiry; no filesystem changes. |
| `debug` | all | Diagnosis and repair of defects outside the story flow. |
| `mock` | all | Non-functional UI mockups; only when directed. |

## Standards manifests (critic reference sets)

- **`design-critic`:** `memory-bank/specifications/architecture/**`, `memory-bank/specifications/design/**`,
  `memory-bank/specifications/patterns/toc.md`, the architecture Definition of Done,
  `resources/agent/artifact-standards.md`.
- **`plan-critic`:** `resources/templates/plan.template.md`, `resources/agent/artifact-standards.md`,
  `.kilo/rules/organization.md`, `memory-bank/specifications/toc-format-spec.md`.
- **`story-critic`:** `resources/templates/stories.md`, `resources/agent/artifact-standards.md`,
  `memory-bank/specifications/toc-format-spec.md`, the task-sizing and parallel-execution rules.
- **`code-critic`:** `resources/code-conventions.md`, `resources/code_review/**`,
  `memory-bank/specifications/patterns/toc.md`.
- **`write-critic`:** `resources/writing_resources/**`, `memory-bank/terms.md`,
  `memory-bank/specifications/toc-format-spec.md`.
- **`milestone-critic`:** `resources/templates/milestone.template.md`,
  `resources/agent/artifact-standards.md` (Milestone and Arc sections), the milestone-generation
  decisions (staged dialog and gates; the arc definition; the flow rules), `.kilo/rules/organization.md`,
  `memory-bank/specifications/toc-format-spec.md`.

The manifest is a *default*; a coordinator may extend it via the work context (see
[`work-context.md`](work-context.md)), but may not remove entries.

## PDD script map

| Persona | Script |
|---|---|
| `story-planning-coordinator` | `scripts/story-planning.pdd.script.md` |
| `story-implementation-coordinator` | `scripts/story-implementation.pdd.script.md` |
| `plan-generation-coordinator` | `scripts/plan-generation.pdd.script.md` |
| `plan-to-story-coordinator` | `scripts/plan-to-story.pdd.script.md` |
| `milestone-generation-coordinator` | `scripts/milestone-generation.pdd.script.md` |
| `milestone-implementation-coordinator` | `scripts/milestone-implementation.pdd.script.md` |
| `milestone-implementation-coordinator` (milestone close) | `scripts/milestone-close.pdd.script.md` |
| any coordinator (remediation) | `scripts/change-control.pdd.script.md` |

## Retired personas and config

| Retired | Replaced by |
|---|---|
| `code` | `code-implementor` |
| `work-planner` | `plan-author` |
| `story-implementor` | `story-implementation-coordinator` |
| `documentation-implementor` | folded into `story-implementation-coordinator` (story type = documentation) |
| `technical-writer` / `technical-writer-for-story-implementor` | `write-implementor` |
| `technical-editor` / `technical-editor-for-story-implementor` | `write-reviewer` |
| `code-for-story-implementor` | `code-implementor` |
| `small-code-for-story-implementor` | `small-code-implementor` |
| `big-code-for-story-implementor` | `big-code-implementor` |
| `code-reviewer-for-story-implementor` | `code-reviewer` |
| `architect-for-story-planning` | split into `story-reviewer` + `design-critic` |
| `orchestrator` | retired (coordinators own routing) |
| `small-code` (config-only duplicate) | `small-code-implementor` |
| `test`, `thomae-architect`, hidden `plan`, `*-custom` duplicates | deleted (dead config) |
