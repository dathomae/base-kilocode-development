# Artifact Standards

Per-level completeness criteria. These are the hard checks the generation-stage reviewers and critics
apply: a plan or story is **not finished** while it contains an open question or a figure-it-out item, a
spec is not complete while a required element is missing, and a milestone is not complete until the
close evaluation passes and the human merges it. They are the primary prevention for escalation
categories 1–5 (see [`escalation-policy.md`](escalation-policy.md)).

Related: [`standards-manifests.md`](standards-manifests.md) (which persona loads which standard). This file
is loaded by `plan-critic`, `story-critic`, `design-critic`, and `milestone-critic`, and informs
`milestone-reviewer`; see also [`change-control.md`](change-control.md).

## Cross-cutting gates

These apply to every artifact at every level.

1. **Settled-decision rule.** Every artifact may rely only on decisions already settled and recorded at
   its level or above. An artifact must not require a downstream actor to make an upstream decision.
2. **No open questions at completion.** An `Open Questions` section must be **empty** when the
   artifact is declared complete. A non-empty section blocks completion.
3. **No figure-it-out items.** No task or requirement may say "investigate/evaluate/choose/design/
   determine X" where X is a requirements, architecture, or design matter. That is category 1.
4. **Deferred items.** An artifact may reference a deliberately deferred item **only** as a reference
   to a section of a requirements or architecture document. Deferral requires explicit human
   sign-off; the only question is whether it was deliberately deferred. Most apparent deferrals are
   open questions under another label and are therefore blocking.
5. **Reference, do not repeat.** Content settled at a higher level is referenced, not restated, so
   there is one authoritative copy.
6. **Terminology.** Every term is either defined at its level or already in `memory-bank/terms.md` /
   `memory-bank/specifications/glossary.md`. A missing shared term or contract is category 4.

## Requirements

- **Captures:** user-perceivable behavior, grouped by topic; what the system must do, not how.
- **Must decide:** what behavior is required and for whom; what is explicitly out of scope.
- **Required elements:** a stable identifier per requirement; the behavior; the actor; acceptance
  semantics; explicit non-goals where ambiguity is likely.
- **Complete when:** every requirement is testable or otherwise verifiable; no requirement depends on
  an architecture or design choice that is not yet settled (if it does, that choice is a gap —
  category 3); terminology is defined.
- **Common failures:** internal terms pushed up from architecture; requirements that encode an
  implementation; missing non-goals.

## Architecture

- **Captures:** placement, boundaries, and interfaces — *where* functionality belongs and *what*
  agreement it honors. Architecture stops at the interface.
- **Must decide:** the component/subsystem decomposition; each subsystem's responsibilities and
  explicit non-ownership; the data each owns; the contracts between them; access modality.
- **Required elements:** the five Definition-of-Done criteria in
  `memory-bank/specifications/architecture/introduction.md` — (1) non-overlapping responsibilities
  including what is not owned, (2) a contract with the static interface, representation-abstract
  shapes, and a behavior specification covering happy and non-happy paths, (3) interface vs
  implementation separation, (4) access modality, (5) the three testing levels. A component or
  subsystem is documented with the `README.md` + `contract.md` pair per
  `resources/templates/subsystem-architecture.template.md` and
  `resources/templates/subsystem-contract.template.md`.
- **Complete when:** all five criteria hold; the contract's non-happy paths are specified (invalid
  input, missing/foreign entity, conflict, authorization failure, partial failure, idempotency/retry);
  shared vocabulary is defined.
- **Common failures:** technique-level detail (belongs in design); a contract with only the happy
  path; overlapping responsibilities; an undefined shared term (category 4).

## Design

- **Captures:** version-specific and technique-level decisions not general enough for architecture —
  concrete types, package maps, SQL, algorithms, file layout.
- **Must decide:** the technique-level decisions an implementor would otherwise have to invent.
- **Required elements:** the decision and its rationale; the alternatives considered; the consequences
  and constraints it imposes; references to the architecture contract it implements. Design-decision
  entries follow `resources/templates/design_decision.md`.
- **Complete when:** no non-trivial coding decision or flow is left to be figured out on the fly
  during implementation (category 5); each design decision traces to an architecture contract or
  requirement.
- **Common failures:** design that is really architecture restated; decisions with no rationale;
  leaving an algorithm or data shape to the implementor.

## Plan

- **Captures:** a targeted objective and the settled implementation decisions that achieve it, plus an
  approximate story breakdown. Plans are where implementation-design decisions are made and recorded
  **inside settled architecture and design**. A plan is generated **for one arc within one milestone**
  (component/subsystem + milestone as input); it does **not** propose milestones.
- **Must decide:** the objective and completion criteria; the settled implementation decisions;
  scope in/out; the story breakdown with sequencing intent.
- **Required elements:** objective; context and references; the **Arc** and **Milestone served**; settled
  decisions; scope (in/out); sequence overview; the story breakdown; test levels where code is involved;
  deferred items (referenced, not open); tracking. Follow `resources/templates/plan.template.md`.
- **Complete when:** the `Open Questions` section is empty; no story in the breakdown is a
  figure-it-out item; every settled decision is recorded in the plan; no decision requires changing a
  spec (if it does, that is category 1/3 and escalates).
- **Common failures:** an open question left in the plan; a decision left implicit; a story breakdown
  whose stories require new design decisions; a plan decision that actually changes architecture.

## Story

- **Captures:** task-level implementation work to implement a settled plan, spec, or design decision.
- **Must decide:** nothing architectural or design-level. A story task that requires a design or
  architecture decision is out of scope.
- **Required elements:** goal; references; dependencies (hard `blocked-by`) and dependent stories;
  **structured footprint/dependency metadata** — hard `blocked-by`, soft preferred-ordering
  dependencies, expected `footprint` (files/packages with shared-infra flags), and `conflicts-with`
  (computed by `milestone-generation-coordinator`, may be empty); tasks with statuses and `[Small]` routing
  annotations; task sizing; parallel execution groups; execution order; reintegration; constraints;
  intent; acceptance criteria; test-first (Logical TDD) notes for coding tasks. Follow
  `resources/templates/stories.md`.
- **Complete when:** the `Open Questions` section is empty; every task implements an already-settled
  decision; each task is sized and routed; the footprint/dependency metadata is present and specific
  enough for `milestone-generation-coordinator` to compute a schedule without reading prose; dependencies
  are consistent with other stories in the plan.
- **Common failures:** a task that says "figure out X"; missing footprint metadata; dependencies that
  contradict another story; a task whose acceptance criteria are not verifiable.

## Arc

- **Captures:** the plan and set of stories for **one system component or subsystem within one
  milestone** — the unit of focused requirements/architecture/design-level checking. A **durable
  per-component arc file** (`Arc_<ABBR>.md`) is updated and grows across milestones.
- **Must decide:** the component's scope for the milestone; nothing above the milestone's already-settled
  decisions.
- **Required elements:** the component/subsystem and its register abbreviation; the milestones served,
  each with its milestone-qualified plan and story links; design notes; history; coherence notes. Follow
  `resources/templates/arc.template.md`.
- **Complete when:** every component marked **modified** in a milestone has an arc whose plan and stories
  exist, and the arc file reflects the current milestone state.
- **Common failures:** an arc for a **traversed** component; a component marked modified with no arc;
  stale plan or story pointers; an arc that records a milestone decision that belongs in the milestone.

## Milestone

- **Captures:** a **delivery increment** — the arcs, plans, and stories that deliver one user-level
  capability, verified by its **milestone acceptance flows**, and can be evaluated and merged
  `agent-main → main`. A milestone mirrors a user experience and may cross UI areas and subsystems.
- **Must decide:** the delivered capability; the milestone's position in the sequence; the participating
  components; the flow groups and flows; the arcs; the gap resolutions; the story set; preconditions; and
  evaluation criteria.
- **Required elements by stage:**
  - *definition:* delivered functionality; milestone order (order/predecessor/successor/batch); components
    (marked modified or traversed); flow definitions (**≥1 flow**, **≤5 flow groups**, the §4 definition
    block); gap register; preconditions; evaluation criteria; the arcs identified.
  - *arc planning:* a §6 plan per modified component; arc-file plan links.
  - *story planning:* story set with `flows covered`; wave schedule; arc-file story links.
  - *complete:* evaluation results; flow verification results (separate section) fully populated; every
    flow `Passed`; and **finalization** — the pre-merge §13 merge record, the finished-tree move, and the
    toc/register updates, completed **before** the merge request. Follow
    `resources/templates/milestone.template.md`.
- **State:** two header fields — `Stage` (`definition` → `arc planning` → `story planning` →
  `story implementation` → `complete`; plus `canceled`) and `Status` (`Not Started` | `In Progress` |
  `Done`, within the current stage). Passing a gate advances `Stage` and resets `Status` to `Not Started`.
- **Complete when** (applied only at stage `complete`, i.e. at milestone close): every story is `Done` and
  merged to `agent-main`; no story is blocked or failed; no open escalation; the full test suite passes on
  the integrated branch; the evaluation criteria pass; **every flow is `Passed`** and none is `Blocked` or
  `Not Verified`; and the close reviewers and critics have no unresolved blocking findings. The human
  `agent-main → main` merge is **not** part of this clause and is **never** an acceptance flow — it is the
  single post-close human action, its absence is never a reviewer/critic finding, and the coordinator does
  not monitor it. Finalization (the pre-merge merge record, the finished-tree move, and the toc/register
  updates) is complete **before** the merge request, so no agent step requires a second `main` merge.
- **Common failures:** a milestone with no flow; a flow with a figure-it-out step; a flow not covered by
  any story; a non-automated flow with no recorded automation discussion; more than five flow groups; an
  unmet evaluation criterion; a blocking critic finding left unresolved; a story merged without its
  dependency; a reviewer/critic applying a later stage's checks at an earlier stage; a close ordering that
  needs a second `agent-main → main` merge; a milestone that lists the human merge as an acceptance flow.

## Reviewer vs critic at each level

- **Reviewers** check the criteria above as *conformance to intent*: does this artifact do the assigned
  activity, completely and correctly, given its inputs and acceptance criteria?
- **Critics** check the same artifacts against the *standards* in
  [`standards-manifests.md`](standards-manifests.md) — templates, formats, patterns, principles — and
  do not re-litigate whether the artifact satisfies its assignment.

Both sets of blocking findings are mandatory. See [`escalation-policy.md`](escalation-policy.md).

A milestone is reviewed repeatedly as it matures. The `milestone-reviewer` (intent) and `milestone-critic`
(standards) **stage-scope** their checklists to the milestone's `Stage` (`definition`, `arc planning`,
`story planning`, or `complete`); an element belonging to a later stage is not yet applicable and is not a
finding. At **milestone close** (stage `complete`) the reviewers and critics run at milestone scope:
`story-reviewer` and `plan-reviewer` ask whether the delivered increment matches what the milestone claims;
`milestone-reviewer`/`milestone-critic` check the milestone artifact; `design-critic`, `code-critic`, and
`write-critic` check standards conformance and walkthrough anomalies; and the flow verification results are
checked against the §4 flow definitions. See `scripts/milestone-close.pdd.script.md`.
