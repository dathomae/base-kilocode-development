# Design Decision Template

This file captures design decisions related to a specific design document. Design decisions explain _why_ a particular approach was chosen, what alternatives were considered, and the rationale — providing context that the design document alone doesn't capture.

## Placement Convention

- This file is co-located with the design document it references
- Naming: if the design doc is `<name>.md`, the decisions file is `<name>_design_decisions.md`
- One design decisions file per design document (not per directory)
- References must be to design documents and/or EROs, never to stories (which are transient)
- Pre-2026-09-16 EROs are frozen in `memory-bank/specifications/design/historical-archive/`; reference them at their archive path. New EROs are created under the live ERO directory.

## Format

Each entry follows the lessons-learned.md style: a **bold header phrase** that is easily searchable, followed by a colon (`:`) and detailed explanation on the following lines. Sub-sections are indented with a leading colon (`:`).

---

## Design Decisions

<!-- Add entries below. Each entry format:
**<decision title>**
: What was decided, in 1-2 sentences.

: **Alternatives considered:**
: - Alternative 1: Brief description and why rejected
: - Alternative 2: Brief description and why rejected

: **Rationale:**
: Why the chosen option was selected. What constraints or principles drove the decision.

: **References:**
: - [`<design_doc_name>.md`](relative/path/to/design_doc.md)
: - [ERO<NNN>: <ero_title>](../relative/path/to/ero.md)

---

-->

**<decision title>**  
: What was decided, in 1-2 sentences.

: **Alternatives considered:**
: - Alternative 1: Brief description and why rejected
: - Alternative 2: Brief description and why rejected

: **Rationale:**
: Why the chosen option was selected. What constraints or principles drove the decision.

: **References:**
: - [`<design_doc_name>.md`](relative/path/to/design_doc.md)
: - [ERO<NNN>: <ero_title>](../relative/path/to/ero.md)

---
