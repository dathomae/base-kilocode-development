# Standards Manifests

Each critic loads an explicit, enumerable set of standards documents. Without a manifest, "compare to
patterns, principles, and standards" degrades into generic opinion. The manifest below is the
**default** for each critic; a coordinator may **extend** it through the work context (see
[`work-context.md`](work-context.md)) but may not remove entries.

Related: [`artifact-standards.md`](artifact-standards.md) (the per-level criteria every critic also
applies), [`persona-registry.md`](persona-registry.md).

## `design-critic` — specifications (requirements / architecture / design)

- `memory-bank/specifications/architecture/introduction.md` — the Definition of Done for architecture.
- `memory-bank/specifications/architecture/**` — the component and subsystem documents, including the
  `README.md` + `contract.md` pairs and `subsystems.md`.
- `memory-bank/specifications/design/**` — version-specific design decisions.
- `memory-bank/specifications/patterns/toc.md` and the pattern files it lists.
- `memory-bank/specifications/glossary.md`, `memory-bank/terms.md` — shared vocabulary.
- `resources/templates/subsystem-architecture.template.md`,
  `resources/templates/subsystem-contract.template.md`, and
  `resources/templates/design_decision.md`.
- `resources/agent/artifact-standards.md`.

Checks: non-overlapping responsibilities; contracts with non-happy paths; interface vs implementation
separation; access modality; the three testing levels; pattern conformance; defined shared vocabulary
(category 4); no technique-level detail in architecture.

## `plan-critic` — plans

- `resources/templates/plan.template.md`.
- `resources/agent/artifact-standards.md` (the Plan section).
- `.kilo/rules/organization.md` — where plans live and how the toc chain works.
- `memory-bank/specifications/toc-format-spec.md`.
- `memory-bank/planning/tracking/system-definition-tracking-status.md` — for status and open issues.

Checks: the plan is for one **arc within one milestone** and records the **Arc** and **Milestone served**
(and proposes no milestones); objective and completion criteria present; settled decisions recorded; scope
in/out; sequence overview; story breakdown; deferred items referenced not open; open-questions section
empty; no figure-it-out story; toc and naming conventions.

## `story-critic` — stories

- `resources/templates/stories.md` — including the footprint/dependency section, flow coverage, task
  sizing, parallel execution, and reintegration rules.
- `resources/agent/artifact-standards.md` (the Story section).
- `memory-bank/specifications/toc-format-spec.md`.
- `memory-bank/concepts.md` — the Logical TDD lifecycle.

Checks: task sizing and `[Small]` routing; parallel groups are genuinely file-disjoint; execution order;
reintegration entries with merge order and integration tests; footprint/dependency metadata complete and
schedule-ready; **flow coverage** (`flows:`) present where the story covers a flow, and a flow test
present only where the flow's functionality is expected to exist; acceptance criteria verifiable;
open-questions section empty; no figure-it-out task; toc and naming conventions.

## `code-critic` — code changes and walkthroughs

- `resources/code-conventions.md` — the canonical project coding conventions.
- `resources/code_review/general_code_review_practices.md`,
  `resources/code_review/python_code_practices.md`.
- `memory-bank/specifications/patterns/toc.md` and the applicable pattern files (e.g.
  `domain-layer-patterns.md`, `database-patterns.md`, `handler-patterns.md`,
  `json-marshal-error-handling.md`, `component-organization.md`).
- `memory-bank/specifications/architecture/testing.md` — the three test levels, including the **flows
  test suite** as the end-to-end level.

Checks: convention adherence; pattern adherence; error handling; test level placement (including flow
tests placed where their functionality exists); naming and file organization. It also reads the story's
**walkthrough** for anomalies — unexplained deviations, silent scope changes, skipped verification, and
decisions made without authority.

## `write-critic` — documents and walkthroughs

- `resources/writing_resources/writing_advice.md`,
  `resources/writing_resources/writing_review.md`, and the rest of
  `resources/writing_resources/**` that applies.
- `memory-bank/terms.md`, `memory-bank/specifications/glossary.md` — terminology.
- `memory-bank/specifications/toc-format-spec.md` — toc format.
- `resources/agent/report-formats.md` — required report shapes.

Checks: clarity and correctness for the stated audience; consistent terminology; document structure
and toc format; no unexplained jargon; the walkthrough's completeness and absence of anomalies.

## `milestone-critic` — milestones

- `resources/templates/milestone.template.md`.
- `resources/agent/artifact-standards.md` (the Milestone and Arc sections).
- `.kilo/rules/organization.md` — where milestones, arcs, and plans live.
- `memory-bank/specifications/toc-format-spec.md`.
- The milestone-generation decisions: the staged dialog and human gates; the arc definition (the plan and
  stories for one component/subsystem within one milestone); the flow rules (≥1 flow, ≤5 flow groups,
  actors at group level, prose `automation.notes` for non-automated flows, flow → story coverage).

Checks are **stage-scoped** to the milestone's `Stage`:
- *all stages:* the current stage's required elements are present; the flow definition block is
  schema-conformant; no flow step is a figure-it-out item; each modified component has an arc; the gap
  register records each gap's resolution and authority; `Stage:`/`Status:` are consistent; naming and toc
  conventions.
- *arc planning:* a plan per modified component and arc-file plan links.
- *story planning:* story set with `flows covered`; every flow covered by ≥1 story; wave schedule present.
- *complete:* evaluation results and flow verification results populated; every flow `Passed`.
A later-stage element is **not yet applicable**, not a finding.

## Reviewers' inputs

Reviewers do not carry a standards manifest; they judge **conformance to intent** against the
assignment and its inputs:

| Reviewer | Inputs |
|---|---|
| `spec-reviewer` | the spec under review + the requirements it must satisfy |
| `plan-reviewer` | the plan + the specs it implements + the settled decisions |
| `story-reviewer` | the story + its plan + acceptance criteria; and at implementation time, the completed story result + its walkthrough |
| `code-reviewer` | the task specification + the diff + the story/plan context |
| `write-reviewer` | the task specification + the document diff + the story/plan context |
| `milestone-reviewer` | the milestone + the human-approved stage decisions (capability, sequence, components, flows, arcs, gap resolutions) + the cited requirements; scope = the milestone's `Stage` |

A reviewer that lacks the context it needs must terminate and report the omission rather than guess.

## Verifiers' inputs

`flow-verifier` does not carry a standards manifest; it executes a defined verification. It is given the
flow definition (its group, actors, preconditions, steps, expected positive/negative outcomes, and
`verification`), the requirement(s) the flow cites, the flow's `test`, the flows-suite contract, and the
revision under test.
