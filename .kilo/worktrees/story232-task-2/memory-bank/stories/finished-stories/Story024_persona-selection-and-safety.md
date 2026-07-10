# Story024: Persona Selection and Safety Boundaries

## Goal
To prevent premature implementation and "Completion Bias" by codifying the mapping of request types to Kilocode personas and establishing hard "read-only" boundaries for discovery tasks (Ask mode).

## References
- [.kilocode/rules/system-rules/system-rules.md](../../.kilocode/rules/system-rules/system-rules.md)
- [.kilocode/rules/rules-ask/ask.md](../../.kilocode/rules/rules-ask/ask.md)
- [.kilocode/rules/organization.md](../../.kilocode/rules/organization.md)
- [.cursorrules](../../.cursorrules)
- [GitHub Issue #47](https://github.com/horatiar/chronocone/issues/47)

## Dependencies
- Issue #29: Incomplete Kilocode persona capture (Completed)

## Dependent Stories
- None

## Tasks
1. [x] Update `.cursorrules` to mandate explicit persona identification for every request.
2. [x] Update `.kilocode/rules/system-rules/system-rules.md` to add a Persona Selection Guide based on intent (e.g., Inquiry, Planning, Execution) using examples to guide agent interpretation.
3. [x] Update `.kilocode/rules/rules-ask/ask.md` to establish a strict Tool Allow-List (Safe Tools) for research and discovery, explicitly forbidding any tool not on the list.
4. [x] Update `.kilocode/rules/organization.md` to add the "Mode Discipline and Transitions" section requiring approved plans for transitions.
5. [x] Update `memory-bank/context.md` to reflect the completion of this story.

## Notes
- The Persona Selection Guide should favor conceptual intent (Investigation vs. Strategy vs. Action) over specific keyword matching.
- The Tool Allow-List in Ask mode must be exhaustive for research (e.g., ls, grep, view, list_resources) and strictly exclude any that create or modify files or state.
