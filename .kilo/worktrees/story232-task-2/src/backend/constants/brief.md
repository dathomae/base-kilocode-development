# Backend Constants

Cross-cutting sentinel values that are shared across every backend layer.

## Purpose

This package exists for one specific reason: to give agents and developers a single grep
result when searching for a sentinel value. Without it, constants that are needed at multiple
layers must either be re-exported through each layer (creating a chain of definitions that
agents may not follow) or be duplicated (causing drift and confusion).

**The "single-grep" principle**: Any constant defined here should have exactly one
authoritative definition. Layers that re-export it for convenience (`domain.InfinityTime`,
`handlers.InfinityTime`) must source it here and carry only a one-line comment pointing
back. The rationale, design choices, and portability notes belong here — not in the
re-export site.

**Agent behavior note**: When an agent searches for a sentinel value and finds three
definitions at three layers, it cannot distinguish which one is authoritative. It may
read only the first result and miss critical design context, or silently define a fourth.
A single definition here prevents all of those failure modes.

## What belongs here

A constant belongs in this package if and only if:

1. It is used by two or more independent backend layers (e.g., both `serde/rdbms` and `handlers`)
2. It has no dependency on any internal package — only the Go standard library
3. It carries design rationale that would otherwise be duplicated or lost

Do not add constants here just because they are used in multiple files within the same layer.
Do not add business-logic constants (validation thresholds, limits, default values). Those
belong in the layer that owns the policy.

## Package Boundaries

This package must remain dependency-free with respect to internal packages. Importing anything
from the backend codebase here would create circular dependency risks and defeat the purpose
of the package.

| Package                      | Status         | Reason                                               |
| ---------------------------- | -------------- | ---------------------------------------------------- |
| Go standard library          | Required       | Time, math, and other primitives as needed           |
| Any `chronocone/backend/...` | **Prohibited** | Would create coupling and potential circular imports |
| Any third-party library      | **Prohibited** | Constants must be portable and zero-dependency       |

## Import Convention

Import this package with the alias `c`:

```go
import c "chronocone/backend/constants"
```

Usage at call sites is then `c.InfinityTime`, `c.SomeOtherConstant`, etc. The short alias
keeps call sites readable without sacrificing clarity about where the value comes from. This
is a project-wide convention — do not import this package under any other name.

## When you work here

- Every exported identifier must include the full design rationale in its doc comment.
  This is the only place that rationale will live; re-export sites carry only a pointer back.
- Do not add constants without clear evidence they are cross-layer. When in doubt, define
  the constant in the layer that owns it and promote it here only when a second layer needs it.
- After adding a constant, update the re-export sites and their briefs to point here.
- This package has no tests of its own. Values here are trivially verifiable by inspection.
  If a constant requires a test, it is probably not a constant.
