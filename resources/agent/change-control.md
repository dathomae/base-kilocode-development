# Change Control

Change control is what happens **after** an escalation: deciding how far up the artifact hierarchy a
fix must go, obtaining the required authority, and executing the cascade. It is a shared procedure —
`scripts/change-control.pdd.script.md` — not a separate persona. Reviewers and critics may propose a
rung; the owning coordinator confirms it; the human decides wherever the table below requires.

Related: [`escalation-policy.md`](escalation-policy.md) (detection and the three dispositions),
[`artifact-standards.md`](artifact-standards.md), [`report-formats.md`](report-formats.md),
[`git-workflow.md`](git-workflow.md).

## The remediation ladder

| Rung | Scope | Decision authority | Executor(s) |
|---|---|---|---|
| **R0** | Story-internal fix within settled decisions | story coordinator | the story coordinator (`story-implementation-coordinator`) |
| **R1** | Revise the story — tasks, decomposition, or acceptance criteria — with no changed design or plan decision | `plan-to-story-coordinator` (autonomous) | `plan-to-story-coordinator` → `story-planning-coordinator` |
| **R2** | Revise the plan and its story set | **human sign-off** | `plan-generation-coordinator` → `plan-to-story-coordinator` |
| **R3** | Revise a requirements/architecture/design spec, then the plan, then the story set | **human + `architect`** | `architect` → `plan-generation-coordinator` → `plan-to-story-coordinator` |
| **R4** | Cancel the plan and its stories; author a new plan and new stories | **human** | `plan-generation-coordinator` → `plan-to-story-coordinator` |

Choosing the lowest sufficient rung is required: do not escalate to R2 what R1 resolves, and do not
"fix" at R1 something that actually changes a plan decision.

## Ownership

- **During an active milestone, `milestone-generation-coordinator` is the change-control owner.** It alone has the
  cross-plan view and owns the schedule and tracking plan. It receives the issue, confirms the rung,
  drives the cascade, records it, and reschedules.
- **Before an arc exists** (during plan or story generation), the invoking generation coordinator owns
  it: `plan-generation-coordinator` for R2/R3, `plan-to-story-coordinator` for R1.
- **The human is the top authority** for R2–R4 and for any cancellation.
- `milestone-implementation-coordinator` pauses, holds, and resumes execution; it does not decide the rung.

## Cascade order

Always fix top-down so the downstream artifacts are regenerated from settled upstream decisions:

1. **Spec** (R3): the human and `architect` revise the requirements/architecture/design document, and
   the revision is recorded in the spec.
2. **Plan** (R2/R3): `plan-generation-coordinator` revises the plan in place and marks it revised.
3. **Stories** (R1–R4): `plan-to-story-coordinator` revises, cancels, or regenerates the affected
   stories through `story-planning-coordinator`.
4. **Reschedule** (all rungs): `milestone-generation-coordinator` recomputes the waves, updates the arc
   tracking plan, and reports the new schedule.

Do not regenerate stories before the plan revision is settled, and do not reschedule before the story
set is settled.

## Revision vs replacement

- **R1–R3 revise in place.** The plan and story files keep their identity and gain a **revision note**
  recording what changed, why, and the triggering issue.
- **R4 cancels and replaces.** The plan and its stories move to `canceled-plans/` and
  `canceled-stories/`; their tocs are updated; a new plan is authored and new stories generated.
- **Cancellation always requires human authorization**, at any rung. A coordinator never cancels a
  story or plan on its own.
- **In-flight branches:** a cancelled story's worktree branch is abandoned, never merged. A revised
  story's coordinator decides whether to continue its branch or restart it from the revised story;
  the decision is recorded.

## What R1 may and may not change

R1 may change a story's task decomposition, task ordering, task sizing, parallel groups, acceptance
criteria, or footprint metadata. It may **not** change a plan decision, introduce a new design
decision, or add work that a plan did not authorize. If resolving the issue requires any of those, the
rung is R2 or higher.

## Milestones

A milestone is the **delivery unit**; an arc is the plan and set of stories for one
component/subsystem within one milestone.

- **Re-scoping** — moving stories between milestones, reordering milestones, or adjusting waves within a
  milestone — is a **schedule revision**: `milestone-generation-coordinator`'s autonomous authority
  within already approved scope. Record it in the milestone file.
- Changes that alter **plan scope** go through **R2**.
- **Cancelling a milestone requires human authorization**: move it to `canceled-milestones/`, update the
  tocs and the tracking register, and reassign or cancel its stories (a cancelled story's in-flight
  branch is abandoned, never merged).
- R0–R4 otherwise apply unchanged; the cascade order remains spec → plan → stories → reschedule.

## Recording

- The milestone file gains a **Remediation log** entry per remediation: rung, issue, artifacts
  touched (revised/cancelled/created), the authority obtained, the resulting schedule change, and the
  date.
- Human resolutions also land in **§4 Resolution log** of
  `memory-bank/planning/tracking/system-definition-tracking-status.md` and in the originating artifact.
- Revised plans and stories carry their revision note; the milestone file references it.

## Flow

1. **Receive** the escalation with the proposed rung and evidence.
2. **Confirm the rung** (lowest sufficient) and identify the affected artifact set.
3. **Obtain authority**: autonomous for R0/R1 (unless a category triggered the escalation), human
   sign-off for R2–R4 and all cancellations.
4. **Execute the cascade** top-down via the owning coordinators.
5. **Record** the remediation and the revision.
6. **Reschedule** via `milestone-generation-coordinator` and **resume** via `milestone-implementation-coordinator`.

## Re-scheduling after remediation

The milestone definition, arc plans, and the story set and waves are human-approved (see [`git-workflow.md`](git-workflow.md) and plan
decision 11). A reschedule that follows a human-approved R2–R4 remediation is part of the same
decision and needs no further approval. `milestone-generation-coordinator` may also revise within
already-approved scope (reorder, defer a story, adjust a parallel group) autonomously and record it;
anything that changes cross-plan scope, or that implies a defect, goes through the ladder above.
