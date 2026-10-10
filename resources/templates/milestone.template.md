# Milestone Template

A milestone is a **delivery increment**: the set of arcs, plans, and stories that delivers one user-level
capability, is verified by its **milestone acceptance flows**, and can be evaluated and merged
`agent-main → main`. It is the unit of delivery and of **demonstrable user value**; an **arc** is the plan
and set of stories for **one system component or subsystem within this milestone**
(see [`arc.template.md`](arc.template.md)).

Save the milestone under `memory-bank/planning/milestones/active-milestones/` as
`Milestone_M<n>_<short-name>.md` and list it in that directory's `toc.md`; move it to
`finished-milestones/` or `canceled-milestones/` when it ends. Owned by
`milestone-generation-coordinator`; maintained by `milestone-implementation-coordinator`. Each stage of
the milestone definition and the final story set require human approval (the **gates**).

Rules:

- Every milestone carries **at least one milestone acceptance flow** and **at most five flow groups**.
  There are no pure-infrastructure milestones: every milestone demonstrates user-level capability.
- Every flow is covered by at least one story in this milestone (flow → story coverage). Stories need not
  map to a flow.
- The human `agent-main → main` merge is **never** a milestone acceptance flow: it is the human's single
  action after finalization, and its absence is never a flow result or a reviewer/critic finding.
- Flow ids are stable once approved; additions get new ids and existing ids are never renumbered.
- Flow **definitions** and flow **verification results** live in separate sections (§4 and §12).

## State model

A milestone has two orthogonal header fields:

- **`Stage`** — the phase: `definition` → `arc planning` → `story planning` → `story implementation` →
  `complete`. `complete` means the milestone is **delivered**: all stories are `Done`, their PRs are merged
  to `agent-main`, and every flow is `Passed`. A canceled milestone carries `Stage: canceled`.
- **`Status`** — progress **within the current Stage**: `Not Started` → `In Progress` → `Done`. Passing a
  human gate advances `Stage` and resets `Status` to `Not Started`. `Stage: complete` + `Status: Done` is
  the terminal delivered state.

The human's `agent-main → main` merge is **not** part of `Stage`/`Status`. Finalization (the pre-merge §13
Merge record, `Stage: complete`, the move to `finished-milestones/`, the tocs and tracking register, and
the close record) is completed **before the merge request**; the human's merge is the single **last**
action, and no agent step follows it on `main`. See
[`artifact-standards.md`](../../../../resources/agent/artifact-standards.md) (Milestone section) and
`.kilo/rules/organization.md`.

The `milestone-reviewer`/`milestone-critic` use the milestone's `Stage` as their review scope (see the
personas). A section that belongs to a later stage is **not yet applicable** at the current stage and must
not be raised as a finding.

## Sections by stage

| Stage | Sections completed in this stage |
|---|---|
| `definition` | §1–§5; §6 **arcs identified**; §8; §10 |
| `arc planning` | §6 plans + arc-file plan links |
| `story planning` | §7; §9; arc-file story links |
| `story implementation` | §14; §15; §11 as waves complete |
| `complete` | §11; §12 populated; every flow `Passed`; **finalization**: §13 pre-merge merge record, the finished-tree move, the toc/register updates |

---

# Milestone: <name>

**Id:** M<n>
**Stage:** definition | arc planning | story planning | story implementation | complete | canceled
**Status:** Not Started | In Progress | Done
**Created:** <YYYY-MM-DD>
**Plan revision:** <n> (last revised <YYYY-MM-DD>)

## 1. Delivered functionality

<The user-visible capability this milestone delivers, and how it can be evaluated.>

## 2. Milestone order

| Field | Value |
|---|---|
| Order | <n> |
| Predecessor | [M<k> <name>](Milestone_M<k>_<short-name>.md) — or "none (first)" |
| Successor | [M<k> <name>](Milestone_M<k>_<short-name>.md) — or "none yet" / "none (last)" |
| Batch | <batch id/name, or "standalone"> |

Milestones form a linear sequence and only one is implemented at a time. A batch's last milestone has an
empty successor with a note; defining a later milestone back-fills this milestone's successor.

## 3. Components

| Component / subsystem | Abbr | Role | Arc | Notes |
|---|---|---|---|---|
| <name> | `CORE` | modified / traversed | [`Arc_CORE.md`](../../arcs/active-arcs/Arc_CORE.md) | |

- **modified** — this milestone changes it; it must have an arc and a plan.
- **traversed** — a flow passes through it but it needs no work; no arc is created.

## 4. Flow definitions

Human-inspection table (the machine source is the YAML block below):

| Group | Flow | Name | Verification | Covered by | Checkable after |
|---|---|---|---|---|---|
| FG1 | F1.1 | <name> | automated | StoryNNN | StoryNNN |

**Flow × story coverage matrix:**

| Flow | Story A | Story B | ... |
|---|---|---|---|
| F1.1 | author | — | |

```yaml
flow_definitions:
  milestone: Milestone_M<n>_<short-name>.md
  flow_groups:
    - id: FG1
      name: <capability bucket>
      capability: >-
        <the one capability this group exercises>
      requirements: [<REQ-ID>]
      actors:
        - { id: <actor-id>, persona: <persona>, grants: [<grant>, ...] }
      flows:
        - id: F1.1
          name: <flow name>
          preconditions:
            - <state that must already hold>
          steps:
            - { actor: <actor-id>, do: "<activity>", via: ui }   # ui | http | cli
          expected:
            positive: [ "<observable positive outcome>" ]
            negative: [ "<observable negative/denied outcome>" ]
          verification: automated        # automated | agent | manual
          test: tests/flows/<flow-test-id>
          automation:
            live_database: true
            mocked_components: none
            notes: null                  # prose; required when verification is agent or manual
          covered_by: [StoryNNN]
          checkable_after: [StoryNNN]
          status: Not Verified
```

For `agent` and `manual` flows, `automation.notes` must be non-empty prose: what makes automation hard,
the alternatives considered, and the decision. A `manual` flow also states that the human is presented the
detailed flow and reports the overall result (with exact symptoms if it fails).

## 5. Gap register

Issues identified while defining the milestone — requirements, architecture, or design gaps that must
close before arc planning and story generation. Record each with its status (`open` until it is
resolved, then `closed`), its resolution, and its authority.

| Id | Type | Description | Status | Resolution | Authority |
|---|---|---|---|---|---|
| GAP-1 | requirements / architecture / design | | open / closed | | |

## 6. Arc(s) and plan(s)

| Arc | Component / subsystem | Plan | Stories |
|---|---|---|---|
| [`Arc_CORE.md`](../../arcs/active-arcs/Arc_CORE.md) | Core capability | [`Plan_M<n>_CORE.md`](../../plans/active-plans/Plan_M<n>_CORE.md) | StoryNNN, StoryNNN |

## 7. Story set

| Story | Arc | Plan | flows covered | blocked-by | Status |
|---|---|---|---|---|---|
| [`StoryNNN_...`](../../stories/active-stories/StoryNNN_....md) | | | | | |

## 8. Preconditions

<What must hold before this milestone starts: e.g. "spec X complete", "milestone M<k> merged".>

## 9. Wave schedule

<The execution order within this milestone: which stories run in parallel in each wave, and why the
grouping is safe. Computed from hard `blocked-by` dependencies, footprint-overlap conflicts, and soft
preferred ordering.>

| Wave | Stories (parallel) | Preconditions | Status |
|---|---|---|---|
| 1 | | | |

## 10. Evaluation criteria

<What "done and correct" means for this milestone — the acceptance check the close evaluation runs.
References the flow-group ids in §4 plus the other gates (full suite, reviewer/critic findings).>

## 11. Evaluation results

<The milestone-close report: full-suite result, evaluation-criteria result, `milestone-reviewer` /
`milestone-critic` / plan / code / write reviewer and critic findings, and their dispositions. See
[`report-formats.md`](../../../../resources/agent/report-formats.md).>

## 12. Flow verification results

The per-flow verification record (machine source), written at milestone close. A `Failed`, `Blocked`, or
`Not Verified` flow blocks the milestone.

```yaml
flow_verification:
  milestone: Milestone_M<n>_<short-name>.md
  revision: <agent-main commit SHA>
  date: <YYYY-MM-DD>
  results:
    - flow: F1.1
      status: Passed                     # Passed | Failed | Blocked | Not Verified
      verification: automated            # automated | agent | manual
      test: tests/flows/<flow-test-id>
      command: <exact command run>
      evidence: <result path / URL / log path / transcript / human report>
      run_by: <flow-verifier | human | session id>
      notes: ""
```

## 13. Merge record

<The **pre-merge** merge record, filled at finalization before the merge is requested: the `main`
**baseline SHA** and the pre-merge `agent-main` **head SHA** (the exact head the human will merge), the
close-report reference, and the readiness recommendation. The human's merge PR and its merged `main` SHA
are authoritative on GitHub and derivable from the PR (for example `gh pr view <n> --json mergeCommit`);
no agent writes them back after the merge and no second `main` merge is required.>

## 14. Divergences

| Id | Date | Story/Plan | Description | Disposition |
|---|---|---|---|---|

## 15. Remediations

| Id | Date | Rung | Issue | Artifacts | Authority | Schedule change |
|---|---|---|---|---|---|---|

## 16. Tracking

<The milestone's row in
[`system-definition-tracking-status.md`](../../tracking/system-definition-tracking-status.md), the arcs it
feeds, and the resolved gaps promoted to the register.>

- **Last review:** `<milestone-critic|milestone-reviewer>` @ stage `<Stage>` — `<pass|fail>`, revision
  `<n>` (<YYYY-MM-DD>). Records the stage scope most recently passed, so a resumed session knows which
  reviews remain.
