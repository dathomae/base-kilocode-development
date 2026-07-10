# OpenAPI Specification Overview

## Version 3.2.0

The key words "MUST", "MUST NOT", "REQUIRED", "SHALL", "SHALL NOT", "SHOULD", "SHOULD NOT", "RECOMMENDED", "NOT RECOMMENDED", "MAY", and "OPTIONAL" in this document are to be interpreted as described in [BCP 14](https://tools.ietf.org/html/bcp14) [RFC2119](https://tools.ietf.org/html/rfc2119) [RFC8174](https://tools.ietf.org/html/rfc8174) when, and only when, they appear in all capitals, as shown here.

This document is licensed under [The Apache License, Version 2.0](https://www.apache.org/licenses/LICENSE-2.0.html).

## Introduction

The OpenAPI Specification (OAS) defines a standard, language-agnostic interface to HTTP APIs which allows both humans and computers to discover and understand the capabilities of the service without access to source code, documentation, or through network traffic inspection. When properly defined, a consumer can understand and interact with the remote service by parsing and serializing HTTP messages to and from a data model with a minimal amount of implementation logic.

An OpenAPI Description (OAD) can then be used by documentation generation tools to display the API, code generation tools to generate servers and clients in various programming languages, testing tools, and many other use cases.

For examples of OpenAPI usage and additional documentation, please visit [OpenAPI-Learn](https://learn.openapis.org/).

For extension registries and other specifications published by the OpenAPI Initiative, as well as the authoritative rendering of this specification, please visit [spec.openapis.org](https://spec.openapis.org/).

## Versions and Deprecation

The OpenAPI Specification is versioned using a `major`.`minor`.`patch` versioning scheme. The `major`.`minor` portion of the version string (for example `3.1`) SHALL designate the OAS feature set. _`.patch`_ versions address errors in, or provide clarifications to, this document, not the feature set. Tooling which supports OAS 3.1 SHOULD be compatible with all OAS 3.1.\* versions. The patch version SHOULD NOT be considered by tooling, making no distinction between `3.1.0` and `3.1.1` for example.

Certain fields or features may be marked **Deprecated**.
These fields and features remain part of the specification and can be used like any other field or feature.
However, OpenAPI Description authors should use newer fields and features documented to replace the deprecated ones whenever possible.

At this time, such elements are expected to remain part of the OAS until the next major version, although a future minor version of this specification may define a policy for later removal of deprecated elements.

Occasionally, non-backwards compatible changes may be made in `minor` versions of the OAS where impact is believed to be low relative to the benefit provided.

## Undefined and Implementation-Defined Behavior

This specification deems certain situations to have either _undefined_ or _implementation-defined_ behavior.

Behavior described as _undefined_ is likely, at least in some circumstances, to result in outcomes that contradict the specification.
This description is used when detecting the contradiction is impossible or impractical.
Implementations MAY support undefined scenarios for historical reasons, including ambiguous text in prior versions of the specification.
This support might produce correct outcomes in many cases, but relying on it is NOT RECOMMENDED as there is no guarantee that it will work across all tools or with future specification versions, even if those versions are otherwise strictly compatible with this one.

Behavior described as _implementation-defined_ allows implementations to choose which of several different-but-compliant approaches to a requirement to implement.
This documents ambiguous requirements that API description authors are RECOMMENDED to avoid in order to maximize interoperability.
Unlike undefined behavior, it is safe to rely on implementation-defined behavior if _and only if_ it can be guaranteed that all relevant tools support the same behavior.

## Format

An OpenAPI document that conforms to the OpenAPI Specification is itself a JSON object, which may be represented either in JSON or YAML format.
Examples in this specification will be shown in YAML for brevity.

All field names in the specification are **case-sensitive**.
This includes all fields that are used as keys in a map, except where explicitly noted that keys are **case-insensitive**.

OAS Objects expose two types of fields: _fixed fields_, which have a declared name, and _patterned fields_, which have a declared pattern for the field name.

Patterned fields MUST have unique names within the containing object.

**Note:** While APIs may be described by OpenAPI Descriptions in either YAML or JSON format, the API request and response bodies and other content are not required to be JSON or YAML.

### JSON and YAML Compatibility

In order to preserve the ability to round-trip between YAML and JSON formats, YAML version [1.2](https://yaml.org/spec/1.2/spec.html) is RECOMMENDED along with the additional constraints listed in [RFC9512] [Section 3.4](https://www.rfc-editor.org/rfc/rfc9512.html#name-yaml-and-json).

The recommendation in previous versions of this specification to restrict YAML to its "JSON" [schema ruleset](https://yaml.org/spec/1.2/spec.html#id2803231) allowed for the inclusion of certain values that (despite the name) cannot be represented in JSON.
OAD authors SHOULD NOT rely on any such JSON-incompatible YAML values.

### Case Sensitivity

As most field names and values in the OpenAPI Specification are case-sensitive, this document endeavors to call out any case-insensitive names and values.
However, the case sensitivity of field names and values that map directly to HTTP concepts follow the case sensitivity rules of HTTP, even if this document does not make a note of every concept.

### Rich Text Formatting

Throughout the specification `description` fields are noted as supporting CommonMark markdown formatting.
Where OpenAPI tooling renders rich text it MUST support, at a minimum, markdown syntax as described by [CommonMark 0.27](https://spec.commonmark.org/0.27/). Tooling MAY choose to ignore some CommonMark or extension features to address security concerns.

While the framing of CommonMark 0.27 as a minimum requirement means that tooling MAY choose to implement extensions on top of it, note that any such extensions are by definition implementation-defined and will not be interoperable.
OpenAPI Description authors SHOULD consider how text using such extensions will be rendered by tools that offer only the minimum support.

## Objects and Fields

This section describes the structure of the OpenAPI Description format.
This text is the only normative description of the format.
A JSON Schema is hosted on [spec.openapis.org](https://spec.openapis.org) for informational purposes.
If the JSON Schema differs from this section, then this section MUST be considered authoritative.

In the following description, if a field is not explicitly **REQUIRED** or described with a MUST or SHALL, it can be considered OPTIONAL.

## See Also

- [OpenAPI Object](openapi_object.md) - The root object of the OpenAPI Description
- [Data Types](data_types.md) - Primitive data types and formats
- [Security Filtering](security_filtering.md) - Excluding sensitive information from API descriptions
