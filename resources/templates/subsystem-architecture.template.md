# Subsystem Architecture Template

Use this template for the `README.md` of a **component or subsystem** directory under
[`architecture/`](../../memory-bank/specifications/architecture/). It captures the component's
**placement and boundaries** — responsibilities, owned data, dependencies, seams, and access
modality — and nothing about a particular implementation.

Its companion is
[`subsystem-contract.template.md`](subsystem-contract.template.md), which captures the **interface
and behavior**. Implementations belong in
[`design/`](../../memory-bank/specifications/design/), never in `architecture/`.

Before writing, read:

- [`architecture/introduction.md`](../../memory-bank/specifications/architecture/introduction.md) —
  the Definition of Done these documents must satisfy.
- The project's architecture index — the subsystem spine and naming.
- The project's documented seam or interface pattern, if one exists.

Instantiation rules:

- Copy the skeleton below into the component or subsystem directory as `README.md`.
- Replace every `<...>` placeholder and delete the instructional angle-bracket notes.
- Adjust relative link targets to match the instantiated directory depth; the skeleton is written
  for a `README.md`/`contract.md` pair in the same directory.
- **Reference, do not repeat** content from `introduction.md`, `subsystems.md`, or the
  concept/requirements levels.
- Keep responsibilities non-overlapping; name the owner of every adjacent concern you exclude.

The full document skeleton to copy:

```markdown
# <Component or Subsystem> — Architecture

## Purpose

<One to three sentences: what this component or subsystem is for and where it sits in the system.>

## Responsibilities

<A non-overlapping, bulleted list of what this component or subsystem is responsible for. Each
responsibility should map to something a reader could assign work against.>

## What It Does Not Own

<Explicit non-responsibilities, naming the owner of each adjacent concern. This is what makes the
responsibilities disjoint (Definition of Done §1); a concern named here must appear in exactly one
other component's Responsibilities, never nowhere.>

## Owned Data

<The tables, registries, or persistent state this component or subsystem owns.

- For each item: its lifecycle and which write path mutates it.
- State only ownership; physical storage belongs in the implementation design.
- Record cross-subsystem references: data this component reads but must never write, and which
  subsystem owns it (call the owner's interface instead).>

## Dependencies

<What this component or subsystem depends on, and in which direction.

- Distinguish synchronous and asynchronous dependencies.
- Identify dependencies that are seams (interfaces) rather than concrete implementations.
- Do not duplicate the full dependency graph in `subsystems.md`; state this component's slice of it
  and reference the spine.>

## Seams

<The interfaces this component or subsystem provides or consumes. The full static interface and
behavior live in [`contract.md`](contract.md). Name each interface and the implementations known to
sit behind it, and cite `patterns/extensible-component-interfaces.md` where the seam exists to allow
future implementations.>

## Access Modality

<How this component or subsystem is reachable:

- **HTTPS-reachable: yes/no.** If yes, give the endpoint family under the
  `/<version>/<subsystem>/...` convention and point at the per-operation endpoint table in
  [`contract.md`](contract.md). If no, state that it is in-process/library (a foundational
  component) and, if there is an exception, document it here.
- Note that HTTP handling is a **transport adapter**, not the interface: the API gateway may call
  this component or subsystem directly, and an asynchronous interface is called through the
  message bus.>

## Testing

<Which of the three levels (unit, subsystem, integration) cover this component or subsystem, and
what specifically each level covers. Reference `architecture/testing.md` for the level definitions
rather than restating them.>

## Related

- [`contract.md`](contract.md) — the static interface and behavior.
- <Spine entry, requirements, design docs, patterns, source tree briefs.>
```
