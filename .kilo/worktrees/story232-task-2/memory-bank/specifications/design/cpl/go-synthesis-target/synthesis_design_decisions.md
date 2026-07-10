# Design Decisions: Go Synthesis Target

This file captures design decisions related to the Go synthesis target design document. Design decisions explain _why_ a particular approach was chosen, what alternatives were considered, and the rationale — providing context that the design document alone doesn't capture.

## Placement Convention

- This file is co-located with the design document it references
- Naming: if the design doc is `<name>.md`, the decisions file is `<name>_design_decisions.md`
- One design decisions file per design document (not per directory)
- References must be to design documents and/or EROs, never to stories (which are transient)

## Format

Each entry follows the lessons-learned.md style: a **bold header phrase** that is easily searchable, followed by a colon (`:`) and detailed explanation on the following lines. Sub-sections are indented with a leading colon (`:`).

---

## Design Decisions

**CILContext is a separate struct from CILAbstractTransformer**
: `CILContext` was defined as a standalone struct rather than embedding its fields directly into `CILAbstractTransformer`. This separation distinguishes temporal tree-walking state from persistent synthesis configuration.

: **Alternatives considered:**
: - Embedding context fields directly into the transformer struct: Rejected because it would bloat the transformer with ephemeral state and prevent reentrancy.

: **Rationale:**
: The transformer holds persistent type mappings and connector metadata that live for the entire synthesis pass. The context holds ephemeral tree-walking state that changes as the visitor descends into different subtrees. A fresh context can be created for each top-level metric computation while reusing the same transformer instance and its registered type mappings. Additional per-walk state (e.g., symbol scopes, parent node references) can be added to `CILContext` without bloating the transformer struct.

: **References:**
: - [`synthesis.md`](synthesis.md)
: - [ERO005: Metrics Synthesis Phase](eros/ERO005_metrics-synthesis-phase.md)

---

**Use strconv.FormatFloat for literal value formatting**
: `strconv.FormatFloat(node.Value, 'f', -1, 64)` was chosen to convert `float64` literal values into string snippets for generated Go code.

: **Alternatives considered:**
: - `fmt.Sprintf("%v", node.Value)`: Rejected because it can produce trailing `.0` for whole numbers, resulting in less readable generated code.
: - `fmt.Sprintf("%f", node.Value)`: Rejected because it forces a fixed precision that produces trailing zeros.
: - `fmt.Sprintf("%g", node.Value)`: Rejected because it switches to exponential notation for large or small values, producing less predictable generated code.

: **Rationale:**
: The format verb `'f'` with precision `-1` produces the smallest number of digits necessary such that `ParseFloat` will return the exact same value. This yields clean output without trailing `.0` and avoids unpredictable exponential notation, ensuring deterministic and readable generated Go code.

: **References:**
: - [`synthesis.md`](synthesis.md)
: - [ERO005: Metrics Synthesis Phase](eros/ERO005_metrics-synthesis-phase.md)

---

**Tiered nil safety for CIL visitor methods**
: Three layers of nil checking were implemented: nil receiver returns an explicit error, nil context returns `nil` (no-op), and nil computation returns `nil` (no-op).

: **Alternatives considered:**
: - Uniform error on all nil conditions: Rejected because a missing context or computation is often a legitimate "nothing to do" state rather than a programming error.
: - Panic on nil conditions: Rejected because it would make the synthesis pipeline fragile against partially initialized call sites.

: **Rationale:**
: This tiered approach distinguishes between "the transformer itself is unusable" (a programming error that should be reported) and "there is nowhere to emit code right now" (a graceful degradation case where the method should complete without effect). It ensures robustness against partially initialized call sites without being overly defensive.

: **References:**
: - [`synthesis.md`](synthesis.md)
: - [ERO005: Metrics Synthesis Phase](eros/ERO005_metrics-synthesis-phase.md)

---
