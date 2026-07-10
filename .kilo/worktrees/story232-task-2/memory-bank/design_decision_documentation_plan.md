# Plan: Design Decision Documentation Practice

## Objective

Establish a standardized practice for capturing coding design decisions alongside design documents, including a template, a PDD script for adding decisions, and placement conventions.

## Background

The user wants to capture the reasons for coding design decisions in design docs for future reference. Design decisions explain _why_ a particular approach was chosen, what alternatives were considered, and the rationale — providing context that design documents and EROs alone don't capture.

## Plan Steps

1. Create the design decision template at [`resources/templates/design_decision.md`](resources/templates/design_decision.md) — Completed
   - Use a list-like format with bold header phrases (matching the [`lessons-learned.md`](memory-bank/lessons-learned.md) style) for searchability
   - Each entry must include: **what was decided**, **alternatives considered**, **rationale**, and **references** (to design docs and/or EROs, never stories)
   - Include a brief header explaining the file's purpose and placement conventions
   - Reference the existing template format conventions from [`resources/templates/`](resources/templates/)

2. Create the PDD script at [`scripts/add-design-decision.pdd.script.md`](scripts/add-design-decision.pdd.script.md) — Completed
   - Follow the PDD script template format from [`resources/templates/prompt-driven-development-script.template.md`](resources/templates/prompt-driven-development-script.template.md)
   - Parameters: `design_doc_path` (path to the design doc the decision relates to), `decision_title` (bold header phrase), `decision_description` (what was decided), `alternatives` (alternatives considered), `rationale` (why chosen), `references` (design/ERO doc references)
   - Operations: validate parameters, construct `*_design_decisions.md` filename from design doc path, create file if new or append entry if existing, update relevant `toc.md`
   - Pattern reference: [`scripts/add-new-story.pdd.script.md`](scripts/add-new-story.pdd.script.md)

3. Document placement conventions — Completed
   - Convention: `*_design_decisions.md` files are co-located with the design doc they reference (e.g., [`synthesis.md`](memory-bank/specifications/design/cpl/go-synthesis-target/synthesis.md) → `synthesis_design_decisions.md` in same directory)
   - Convention: One design decisions file per design document (not per directory)
   - Convention: Design decisions reference design docs and EROs, never stories
   - These conventions should be documented in the template header and in the PDD script

4. Update relevant toc.md files — Completed
   - Add the new template to [`resources/templates/toc.md`](resources/templates/toc.md)
   - Add the new PDD script to [`scripts/toc.md`](scripts/toc.md) if one exists, or note it in the project's main toc

## Completion Criteria

- Template file exists at `resources/templates/design_decision.md` with proper format ✓
- PDD script exists at `scripts/add-design-decision.pdd.script.md` following PDD conventions ✓
- Placement conventions are documented in both the template and script ✓
- Relevant toc.md files are updated ✓
