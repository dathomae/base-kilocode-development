# Report Formats

Standard shapes for the reports the hierarchy exchanges. Consistent shapes let coordinators parse
results, classify findings, and record dispositions without re-reading prose. Every report names the
**artifact revision** it judged, so an escalation can cite evidence rather than memory.

Related: [`escalation-policy.md`](escalation-policy.md), [`change-control.md`](change-control.md),
[`work-context.md`](work-context.md), [`standards-manifests.md`](standards-manifests.md).

## Severity vocabulary

Used by both reviewers and critics.

| Severity | Meaning | Disposition |
|---|---|---|
| `blocking` | The artifact does not meet its assignment (reviewer) or violates a standard (critic). **Never waivable.** | Must be fixed before the artifact advances. |
| `minor` | A real but non-blocking defect or deviation. | Fix or record at the coordinator's discretion, with a recorded rationale. |
| `advisory` | Informational; a suggestion or observation. | Record without action. |

Findings also carry a **proposed disposition**: `fix` | `record` | `escalate` (see
[`escalation-policy.md`](escalation-policy.md)). The coordinator makes the final call.

## Implementation result (implementor → coordinator)

- `work_context` (returned unchanged)
- Status: `Completed` | `Blocked — needs decision` | `Failed`
- Summary of what was implemented
- Files created or updated
- Test results (pass/fail, command)
- Issues encountered
- If `Blocked — needs decision`: the decision needed and the options the implementor can see (do not
  attempt the decision)

## Flow-verification result (flow-verifier → coordinator)

- `work_context` (returned unchanged)
- `flow`: the flow id (`FG<n>/F<group>.<n>`)
- `status`: `Passed` | `Failed` | `Blocked` | `Not Verified`
- `verification`: `automated` | `agent` | `manual`
- `test` (when automated); `command`: the exact command run
- `evidence`: result path / URL / log path / transcript / human report
- `run_by`: `<flow-verifier | human | session id>`
- `notes`: observed vs expected difference on `Failed`; for `manual`, the human's overall result and, on
  failure, the exact symptoms; for `Blocked`, what is needed to run the flow

The verifier does not judge the milestone; it reports the flow's result with reproducible evidence.

## Review report (reviewer → coordinator)

- `work_context` (returned unchanged)
- Artifact and revision judged
- Stage scope applied (for a milestone, its `Stage`); elements belonging to a later stage, listed under
  `not yet applicable`
- Verdict: `pass` | `fail`
- Findings: each with id, severity, the assignment clause or acceptance criterion it violates, the
  evidence, and a proposed disposition
- If the required context was missing: `terminated — insufficient context` with the omission named

Reviewers judge **conformance to intent** only. They do not report standards violations (that is the
critic's role) and do not propose design alternatives.

## Critic report (critic → coordinator)

- `work_context` (returned unchanged)
- Artifact and revision judged
- Stage scope applied (for a milestone, its `Stage`)
- Standards manifest used (the default plus any extensions)
- Findings: each with id, severity, the **standard cited** (document + section), the evidence, and a
  proposed disposition
- Adherence summary: which standards were checked, which **elements were not yet applicable** at this stage
  (belonging to a later stage), and which could not be assessed
- For `code-critic`/`write-critic` on a completed story: the walkthrough anomaly findings, listed
  separately

Critics judge **conformance to standards** only. They do not re-litigate whether the artifact
satisfies its assignment.

## Escalation report (coordinator → invoker / human)

- Category: 1–5 (see [`escalation-policy.md`](escalation-policy.md)); for category 4, state that it is
  the shared spec vocabulary/contract sense
- Disposition being escalated
- Artifact(s) and revision(s)
- Evidence: findings, logs, citations
- What work is blocked (stories/tasks) and what is still safe to continue
- Proposed rung if remediation is required (R0–R4; see [`change-control.md`](change-control.md))
- Options the coordinator sees
- The coordinator's recommendation

Factual, not persuasive: the human decides.

## Walkthrough additions (story completion)

The story walkthrough (see `resources/templates/walkthrough.template.md`) records:

- **What changed** — files created, modified, deleted.
- **Deviations from the story** and why.
- **Verification** — tests run and results.
- **Decisions made during implementation** — anything decided at the coordinator level, with its
  authority.
- **Divergences** — non-blocking misalignments recorded rather than acted on, with the covering story
  if any.
- **Remediations** — any R0/R1 revision applied, with the triggering issue.
- **Lessons learned** — candidates for `memory-bank/lessons-learned.md`.

The critic reads the walkthrough specifically for anomalies: unexplained deviations, silent scope
changes, skipped verification, and decisions made without authority.

## Milestone-close report (milestone-implementation-coordinator → human)

Produced by `scripts/milestone-close.pdd.script.md` at the end of a milestone:

- Milestone name and revision
- Full test suite: the command and its result
- Flow verification: the **per-flow results** from the milestone's Flow verification results section —
  each flow with its status, verification mode, and evidence; a `Failed`, `Blocked`, or `Not Verified`
  flow is blocking
- Evaluation criteria: each criterion with pass/fail and evidence
- Reviewer/critic verdicts: `milestone-reviewer`, `milestone-critic`, `story-reviewer`, `plan-reviewer`,
  `design-critic`, and `code-critic`/`write-critic` — each verdict and its findings with severity and
  disposition
- Recorded divergences
- Any unresolved blocking findings
- **Recommendation:** `ready to merge` or `blocked`
- The proposed merge (`agent-main` → `main`) and the human's decision

## Milestone tracking entries

**Divergence entry** (in the milestone file's Divergences section):

- id, date, story/plan, artifact, description, disposition (`fixed in place` | `recorded` |
  `escalated`), covering story or resolution reference

**Remediation log entry** (in the milestone file's Remediation log):

- id, date, rung (R0–R4), triggering issue and category, artifacts touched (revised / cancelled /
  created), authority obtained, resulting schedule change, resolution reference (§4 log)

**Milestone row** (in `system-definition-tracking-status.md`):

- milestone name, `Stage` | `Status`, preconditions, evaluation result, flow verification result, merge
  reference

## Resolution log entry

Human resolutions recorded in **§4 Resolution log** of
`memory-bank/planning/tracking/system-definition-tracking-status.md`:

- date, issue, category, decision, artifacts updated, affected stories/plans/milestones, schedule impact
