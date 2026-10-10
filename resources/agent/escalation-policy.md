# Escalation Policy

This is the single escalation policy for every agent persona. Coordinators own implement/review/critic
loops and apply this policy; workers, reviewers, and critics report findings and defer the disposition
to their coordinator. No persona restates or overrides this policy — persona files and PDD scripts
reference it.

Related: [`change-control.md`](change-control.md) (what happens after an escalation),
[`artifact-standards.md`](artifact-standards.md) (the completeness gates that prevent most
escalations), [`report-formats.md`](report-formats.md) (escalation and finding shapes),
[`persona-registry.md`](persona-registry.md).

## Three dispositions

Anything found during plan writing, story writing, or implementation gets exactly one disposition.
The owning coordinator decides it; reviewers and critics may *propose* it.

| Disposition | When | Where it is recorded |
|---|---|---|
| **Fix on the spot** | A local defect at the artifact's own level that changes no upstream decision and is not already covered by a scheduled story. Lean toward fixing small issues rather than deferring. | The implement→review loop; noted in the walkthrough/report. |
| **Record** | A divergence or misalignment that changes no outcome and needs no decision, **or** an issue already covered by a future story. Do not act; do not escalate every misalignment. | The milestone file's **Divergences** section. |
| **Block and escalate** | Any of the five categories below. | Escalation report + resolution log (see below). |

### Never pull work forward

When a finding concerns work that a scheduled future story will do, **record** it and confirm the
future story covers it. Do not fix it now: doing so duplicates work, creates avoidable merge
conflicts, and invalidates the schedule. The coverage check is against the plan's story list and the
milestone file. If no scheduled story covers it, it is not a "record" — reassess as fix-on-the-spot
or escalate.

## The five escalation categories

Universal across plan writing, story writing, and implementation. Each level differs only in its
detectors.

1. **A "we should figure out X" where X is a requirements, architecture, or design issue.** Ideally
   caught at plan generation; if discovered during story execution, stop and consider options with the
   human.
2. **Any open question in a plan or story.** Must be surfaced and closed before plan generation
   finishes; if one slips through, later implementation or review stages block on it.
3. **Any gap in requirements, specification, or design.**
4. **Any missing architecture-level contract or shared spec vocabulary/terminology.**
5. **Any missing design that forces non-trivial coding decisions on the fly during story
   implementation** — including an "easy" task that turns out not to be easy.

**Category 4 disambiguation (mandatory).** Category 4 means the **shared specification vocabulary and
contract definitions** — the terms in `memory-bank/terms.md`, `memory-bank/specifications/glossary.md`,
and the per-subsystem `contract.md` files — and any other *required element of a particular kind of
spec* as listed in [`artifact-standards.md`](artifact-standards.md). It does **not** mean the consuming
product's domain language. State the intended sense explicitly in any escalation report; shared
specification vocabulary and product-domain vocabulary are easily confused.

## Detection and routing

- **Who detects.** `spec-reviewer`/`design-critic` (specs), `plan-reviewer`/`plan-critic`/`design-critic`
  (plans), `story-reviewer`/`story-critic`/`design-critic` (stories), `code-reviewer`/`code-critic` and
  `write-reviewer`/`write-critic` (implementations), and any implementor via the `Blocked` status.
- **Who decides the disposition.** The owning coordinator. A worker, reviewer, or critic reports
  findings and a proposed disposition; it never escalates to the human directly.
- **Routing.** Worker → its coordinator. Coordinator → its invoker. The top-level coordinators
  (`plan-generation-coordinator`, `plan-to-story-coordinator`, `milestone-generation-coordinator`,
  `milestone-implementation-coordinator`) escalate to the human. During an active milestone, schedule-affecting
  issues route through `milestone-generation-coordinator` (see [`change-control.md`](change-control.md)).

## Retry limits

Two independent ladders. Neither applies to a `Blocked` result.

- **Review-fix iterations (per implementor tier):** up to **3** implement → review cycles. If blocking
  findings persist after the third cycle, move up the implementor ladder.
- **Implementor ladder:** `small-code-implementor` / `small-write-implementor` → `code-implementor` /
  `write-implementor` → `big-code-implementor` / `big-write-implementor`. After the big tier fails,
  escalate to the coordinator's invoker.
- **Blocked results:** an implementor that returns `Blocked — needs decision` (categories 1 or 5) does
  **not** consume either ladder. Escalate immediately. A bigger implementor cannot make an
  undecided design decision.

## The `Blocked — needs decision` status

Implementors return this status, rather than `Failed`, when proceeding would require a
requirements, architecture, or design decision (category 1 or 5), or when the task specification is
insufficient to proceed. The report must name the decision needed and the options the implementor can
see. The coordinator treats it as an escalation trigger and must not retry it up the ladder.

## Resolution

A coordinator resolves a finding itself when the change is internal to the artifact it owns and
changes no upstream artifact. It escalates when the change would require modifying a plan, a story, a
spec, the milestone schedule, or another coordinator's artifact.

- **Blocking findings** from a reviewer or critic are mandatory and are **never waivable**.
- **Minor** findings are fixed or recorded at the coordinator's discretion, with a recorded rationale.
- **Advisory** findings are informational; the coordinator may record them without action.

A reviewer passing while a critic raises a blocking finding is **not** a conflict: they judge
different things (intent vs standards). Both sets of blocking findings must be resolved.

## Recording

- **Human resolutions** land in **§4 Resolution log** of
  `memory-bank/planning/tracking/system-definition-tracking-status.md` **and** in the originating
  artifact (the plan, story, or spec is updated so the decision is durable).
- **Non-blocking divergences** land in the milestone file's Divergences section.
- **Escalations** are recorded as an escalation report (see [`report-formats.md`](report-formats.md))
  with the category, the artifact and revision, the evidence, what is blocked, the options, and the
  coordinator's recommendation.
- If the escalation requires changes above the story level, the resolution is executed through
  [`change-control.md`](change-control.md).

## Pause and hold semantics

When a category is escalated, the coordinator **pauses the affected work only** — not the whole arc:

- Stop launching new waves or tasks that depend on the unresolved item.
- Mark affected stories blocked in the milestone file.
- In-flight story branches are **held** (finished but not merged), never discarded silently.
- Work that is independent of the issue may continue, unless the coordinator judges the issue
  contaminates it.

When the human responds, the coordinator records the resolution, resumes, and — if the resolution
changed plans or stories — runs the change-control cascade before rescheduling.

## Escalation report contents

Every escalation report includes: category (1–5); the disposition being escalated; the artifact(s) and
revision(s) involved; the evidence (findings, logs, citations); what work is blocked; the rung
proposed if remediation is required (see [`change-control.md`](change-control.md)); the options the
coordinator sees; and its recommendation. Keep it factual — the human decides.
