# Source Code

Root of the application source tree. Code-bearing packages live under
`backend/` and `frontend/`.

## Responsibilities

_(none — this brief is a tree-level marker; see `backend/brief.md` and `frontend/brief.md` for the layers' responsibilities.)_

## Package Boundaries

`chronocone_planning_language/` is **off-limits** to this hierarchical
brief process unless authorized by the user — it is owned by the kilocode process. Do not edit code
under that directory or add `brief.md` files within it as part of work
driven from this tree unless you get that authorization.

Otherwise this is a tree-level marker; see child briefs (`backend/`,
`frontend/`) for inter-package boundaries.

## When you work here

_(none — pick the most specific child brief that matches your task and work from there.)_

## Lessons Learned

**Ask for guidance when you uncover architectural issues**

If a task surfaces problems that look architectural rather than local — wrong abstraction, missing pattern, conflicting conventions, a fundamental design assumption broken (e.g., "this framework doesn't actually support the operation I need") — stop and ask. Don't refactor architecture in flight to make a failing test pass; the right call may be to redesign, to scope the task differently, or to land a smaller fix first. The user should make that call, not you.

**Evaluate layer responsibilities from the interface, not the implementation**

When assessing what a service or handler needs to do, reason exclusively from the **interface contract** of the layers it calls — not from knowledge of what those layers do internally. If the interface accepts `[]InsertStatement` and returns `error`, the caller's job is to construct `InsertStatement`s. Full stop. If your rationale for deferring or expanding work includes implementation details of a dependency, you are violating this rule. Re-evaluate from the interface.

**When architecture seems limiting, ask rather than bypass**

When you hit friction with a pattern (reaching for raw alternatives, skipping validation, adding a special case), pause. The pattern likely covers the case; friction usually signals incomplete understanding. Patterns exist to maintain invariants — bypassing them for convenience breaks invariants silently. Ask for guidance on how to apply the pattern correctly rather than working around it.
