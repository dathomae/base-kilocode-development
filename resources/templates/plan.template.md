# Plan Template

A plan captures a targeted objective and the settled implementation decisions that achieve it, plus an
approximate story breakdown. A plan is generated **for one arc within one milestone** (see
[`arc.template.md`](arc.template.md)): its inputs are the milestone and the arc. Plans record
**implementation-design** decisions inside settled architecture and design; a decision that would change
requirements, architecture, or design is not a plan task — it is an escalation.

Save the plan under `memory-bank/planning/plans/active-plans/` as `Plan_M<n>_<ABBR>.md` and list it in
that directory's `toc.md` grouped by milestone; move it to `finished-plans/` or `canceled-plans/` when
its milestone ends. Author it with `plan-author` under `plan-generation-coordinator`, and check it with
`plan-reviewer`, `plan-critic`, and `design-critic`. A plan does **not** propose milestones: milestones
are defined first and drive plan generation.

A plan is **not complete** while its `Open Questions` section is non-empty or while any story in the
breakdown is a figure-it-out item. See `resources/agent/artifact-standards.md`.

---

# Plan: <title>

> **Source:** <where this plan came from>
> **Type:** plan (targeted, multi-task).
> **Milestone served:** [`Milestone_M<n>_<short-name>.md`](../../milestones/active-milestones/Milestone_M<n>_<short-name>.md)
> **Arc:** [`Arc_<ABBR>.md`](../../arcs/active-arcs/Arc_<ABBR>.md) — <component/subsystem>
> **Status of steps:** track each step's status below.

## Objective

<What this plan achieves and why.>

## Context

<The current state, constraints, and references that shape the plan.>

## Settled decisions

<Every implementation-design decision this plan settles, numbered and explicit. Each states the
decision and its rationale. Nothing may be left implicit for a story to discover.>

1. <decision> — <rationale>

## Scope

In scope:

- <...>

Out of scope:

- <...>

## Sequence overview

<The order in which the work proceeds, and the dependency relationships between parts.>

```
1 <part>
    └─> 2 <part>
            └─> 3 <part>
```

## Story breakdown

<The approximate story set this plan decomposes into, in implementation order. Stories are authored by
`plan-to-story-coordinator`; this is the plan's intent, not the final story files.>

## Test levels

<Which of the three test levels (unit / subsystem / integration) apply, and any build or verification
commands.>

## Deferred items

<Only references to a requirements or architecture document section, each with human sign-off. Most
apparent deferrals are open questions under another label.>

## Open Questions

<Must be **empty** when the plan is complete. A non-empty section blocks completion. A figure-it-out
item is an escalation, not a plan step.>

## Tracking

<How progress is tracked: step statuses, the tracking register row(s), and the milestone file when
one exists.>
