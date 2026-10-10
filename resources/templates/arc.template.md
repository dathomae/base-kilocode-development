# Arc Template

An **arc** is the **plan and set of stories for one system component or subsystem within one milestone**
— the unit of focused requirements, architecture, and design-level checking. Planning is **per-arc,
per-milestone**. The arc **file** is **durable and per-component**: it accumulates the component's
milestones over time and carries design, history, and coherence notes, and it grows as more milestones are
planned and implemented.

Save the arc under `memory-bank/planning/arcs/active-arcs/` as `Arc_<ABBR>.md` and list it in that
directory's `toc.md` (the toc lists **components**); move it to `finished-arcs/` only when the component
has **no remaining planned work**. Owned by `milestone-generation-coordinator`.

`<ABBR>` is the component/subsystem's abbreviation from the **Abbreviations** table in
[`system-definition-tracking-status.md`](../../tracking/system-definition-tracking-status.md). Milestone
files are `Milestone_M<n>_<short-name>.md`; plans are `Plan_M<n>_<ABBR>.md`. See
[`milestone.template.md`](milestone.template.md) and [`plan.template.md`](plan.template.md).

---

# Arc: <component/subsystem> (`<ABBR>`)

**Status:** Not Started | In Progress | Done | Canceled
**Created:** <YYYY-MM-DD>
**Last updated:** <YYYY-MM-DD>

## 1. Component scope

<The component/subsystem this arc covers, its responsibility boundary, and what it does not own. This is
the durable scope; each milestone below is one increment of work on it.>

## 2. Milestones served

One row per milestone that includes work on this component, in sequence order.

| Milestone | Plan | Stories | Status |
|---|---|---|---|
| [`Milestone_M1_...`](../../milestones/active-milestones/Milestone_M1_....md) | [`Plan_M1_<ABBR>.md`](../../plans/active-plans/Plan_M1_<ABBR>.md) | StoryNNN | |

## 3. Design notes

<Design decisions, interfaces, and constraints specific to this component's work.>

## 4. History

<What each milestone changed in this component, in order; what was deferred and why; notable divergences
and remediations.>

## 5. Coherence notes

<Interfaces, shared vocabulary, and cross-milestone consistency that must hold for this component —
the reason the arc file persists and grows beyond a single milestone.>

## 6. Tracking

<The component's row and open issues in
[`system-definition-tracking-status.md`](../../tracking/system-definition-tracking-status.md), and the
plans that cover it.>
