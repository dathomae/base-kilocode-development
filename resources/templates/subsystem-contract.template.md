# Subsystem Contract Template

Use this template for the `contract.md` of a **component or subsystem** directory under
[`architecture/`](../../memory-bank/specifications/architecture/). It captures the **static
interface** (synchronous endpoints with request/response shapes, asynchronous events with payload
shapes) and the **behavior** (happy and non-happy paths) for every operation.

Its companion is
[`subsystem-architecture.template.md`](subsystem-architecture.template.md), which captures
responsibilities, owned data, dependencies, seams, and access modality. Implementations belong in
[`design/`](../../memory-bank/specifications/design/), never here.

The contract is what callers depend on. It is:

- **Implementation-agnostic.** One or more implementations sit behind it; the implementation map at
  the end points into `design/`.
- **Caller-facing.** It describes the operations a caller invokes and the payloads exchanged, not
  the internal structure that serves them.
- **Complete on failure.** The happy path alone is not a contract; the non-happy paths are part of
  it.

## Notation and representation

A contract is an **abstract interface**: the operations and the *semantics* of the payloads exchanged,
independent of any encoding or protocol. The JSON-like blocks in a contract are **shape notation** for
readability; they are not a wire format and not the interface. Access modality determines what the
concrete representation is:

- **In-process interface** — the representation is the implementation language's native types; the
  implementation design records them.
- **HTTPS-reachable interface** — the wire representation (JSON and its OpenAPI schema) is part of the
  **transport adapter's** design; the abstract contract states the payload shapes and behavior, and the
  gateway owns the wire mapping.

A contract must never be read as mandating JSON, HTTP, or a particular type.

Before writing, read
[`architecture/introduction.md`](../../memory-bank/specifications/architecture/introduction.md)
(Definition of Done §2 and §4) and the project's documented seam or interface pattern, if one
exists.

Instantiation rules:

- Copy the skeleton below into the component or subsystem directory as `contract.md`.
- Replace every `<...>` placeholder and delete the instructional angle-bracket notes.
- Adjust relative link targets to match the instantiated directory depth; the skeleton is written
  for a `README.md`/`contract.md` pair in the same directory.
- Omit a section that genuinely does not apply (for example, a subsystem with no HTTPS surface has
  no synchronous endpoint table), but say so explicitly rather than leaving it empty.
- Endpoint paths follow the
  [`/<group>/<version>/<endpoint>` convention](../../memory-bank/specifications/architecture/introduction.md#https-endpoint-convention);
  use `v1` unless a later version is genuinely required.
- Payload blocks are **shape notation**, not a wire format (see [Notation and
  representation](#notation-and-representation)).

The full document skeleton to copy:

````markdown
# <Component or Subsystem> — Contract

<One or two sentences: which component or subsystem this contract governs, and a link to its
[`README.md`](README.md) for responsibilities and boundaries.>

## Interface vs. Implementation

This document captures the **interface**: the operations and the payload/behavior contract around
them. It is implementation-agnostic and says nothing about how the operations are served.
Implementations and technique-level detail live in `design/`; see the Implementation Map at the end.
The payload blocks below are **shape notation**, not a wire format; concrete native types for an
in-process interface and wire/schema mappings for a network interface are fixed in `design/`.

## Access Modality

- **HTTPS-reachable:** <yes/no>. <If yes: the group family `/<group>/<version>/...` and a
  pointer to the Synchronous Endpoints table below. If no: in-process/library, with the reason
  (normally a foundational component); document any exception.>
- **Asynchronous surface:** <yes/no>. <If yes: reached by putting a message on the message bus; see
  the Asynchronous Events table below.>
- **Transport adapter note:** the HTTP handlers are a transport adapter, not this interface. The API
  gateway may call the interface directly, in-process; HTTPS endpoints exist only when the
  subsystem must be reached across a distribution boundary.

## Static Interface

### Synchronous Endpoints

<If the subsystem is not HTTPS-reachable, replace this section's contents with "None — the
component is reached in-process/library." Otherwise list every endpoint:>

| Operation | Method + path | Request | Response |
| --- | --- | --- | --- |
| `<operation>` | `GET /<group>/v1/<function>` | `<RequestType>` | `<ResponseType>` |

#### `<operation>`

**Request shape**

```json
{
  "<field>": "<type — description>"
}
```

**Response shape**

```json
{
  "<field>": "<type — description>"
}
```

**Payload notes**

<Natural keys vs internal ids, nullability, versioning fields, and any format rules
that apply to this operation's payloads.>

### Asynchronous Events

<If the subsystem has no asynchronous surface, replace this section's contents with "None." Otherwise
list every event this subsystem emits or consumes:>

| Event | Direction | Trigger | Payload | Ordering key | Idempotency key |
| --- | --- | --- | --- | --- | --- |
| `<event.name>` | emits / consumes | <when it is produced> | `<EventType>` | <field> | <field> |

#### `<event.name>`

**Payload shape**

```json
{
  "<field>": "<type — description>"
}
```

**Contract notes**

<Thin vs. fat payload (reference + correlation data vs. full body), schema version, ordering,
at-least-once delivery, and dedupe expectation.>

## Behavior

<One subsection per operation and per event-handling obligation. Cover the happy path and every
non-happy path that applies: invalid input; missing entity; foreign entity/scope; conflict
(staleness or concurrent mutation); authorization failure; partial failure; idempotency/retry;
precondition failure. Where a category does not apply, say so explicitly (for example, "no partial
failure: the operation is atomic").>

### `<operation>` — Behavior

**Happy path**

1. <Precondition and input.>
2. <Effect on owned data.>
3. <Result returned to the caller, and any event emitted.>

**Non-happy paths**

| Condition | Result | Notes |
| --- | --- | --- |
| Invalid input | <error / status, response shape> | <which fields, why> |
| Missing entity | <...> | <...> |
| Foreign entity / scope | <...> | <...> |
| Conflict (staleness / concurrent mutation) | <...> | <...> |
| Authorization failure | <...> | <...> |
| Partial failure | <...> | <...> |
| Idempotency / retry | <...> | <...> |

## Implementation Map

<Where each interface element is implemented today. Implementations are design, not architecture:
link into `design/`. Where no implementation design exists yet, write `TODO` in the design-location
column so the gap is visible.>

| Interface element | Implementation | Design location |
| --- | --- | --- |
| `<operation>` | <Go package / provider> | `TODO` or [`design/...`](relative/design.md) |
| `<event.name>` | <...> | `TODO` or [`design/...`](relative/design.md) |

## Related

- [`README.md`](README.md) — responsibilities, owned data, dependencies, access modality.
- <Subsystem spine entry, requirements, patterns the payloads follow, source tree briefs.>
````
