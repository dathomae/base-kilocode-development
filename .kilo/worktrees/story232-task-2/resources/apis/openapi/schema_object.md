# Schema Object

The Schema Object allows the definition of input and output data types. The following properties are taken directly from the [JSON Schema definition](https://json-schema.org/draft/2020-12/json-schema-core.html) and follow the same specifications:

## JSON Schema Properties

The Schema Object is a superset of JSON Schema, with additional OpenAPI-specific properties.

### Core JSON Schema Properties

| Property    |          Type          | Description                                                                                                              |
| ----------- | :--------------------: | ------------------------------------------------------------------------------------------------------------------------ |
| $id         |        `string`        | The canonical URI for this schema.                                                                                       |
| $schema     |        `string`        | The JSON Schema version being used.                                                                                      |
| $ref        |        `string`        | A reference to another schema.                                                                                           |
| $comment    |        `string`        | A comment for the schema.                                                                                                |
| $defs       | Map[`string`, Schema]  | A collection of schema definitions for reuse.                                                                            |
| type        | `string` \| [`string`] | The type of the value. Can be a single type or array of types.                                                           |
| enum        |         [any]          | An array of allowed values.                                                                                              |
| const       |          any           | A single allowed value.                                                                                                  |
| default     |          any           | The default value for the property.                                                                                      |
| title       |        `string`        | A title for the schema.                                                                                                  |
| description |        `string`        | A description of the schema. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation. |
| examples    |         [any]          | An array of example values.                                                                                              |
| deprecated  |       `boolean`        | Indicates the schema is deprecated.                                                                                      |

### String Constraints

| Property  |   Type    | Description                                                |
| --------- | :-------: | ---------------------------------------------------------- |
| minLength | `integer` | Minimum string length.                                     |
| maxLength | `integer` | Maximum string length.                                     |
| pattern   | `string`  | A regex pattern the string must match.                     |
| format    | `string`  | A semantic format (see [String Formats](#string-formats)). |

### Numeric Constraints

| Property         |   Type   | Description                              |
| ---------------- | :------: | ---------------------------------------- |
| minimum          | `number` | Minimum value (inclusive).               |
| maximum          | `number` | Maximum value (inclusive).               |
| exclusiveMinimum | `number` | Minimum value (exclusive).               |
| exclusiveMaximum | `number` | Maximum value (exclusive).               |
| multipleOf       | `number` | Value must be a multiple of this number. |

### Array Constraints

| Property    |        Type        | Description                                  |
| ----------- | :----------------: | -------------------------------------------- |
| items       | Schema \| [Schema] | Schema for array items.                      |
| prefixItems |      [Schema]      | Schema for prefix items (tuple validation).  |
| minItems    |     `integer`      | Minimum number of items.                     |
| maxItems    |     `integer`      | Maximum number of items.                     |
| uniqueItems |     `boolean`      | All items must be unique.                    |
| contains    |       Schema       | At least one item must match this schema.    |
| minContains |     `integer`      | Minimum number of items matching `contains`. |
| maxContains |     `integer`      | Maximum number of items matching `contains`. |

### Object Constraints

| Property             |         Type          | Description                                |
| -------------------- | :-------------------: | ------------------------------------------ |
| properties           | Map[`string`, Schema] | Property name to schema mapping.           |
| patternProperties    | Map[`string`, Schema] | Regex pattern to schema mapping.           |
| additionalProperties |  Schema \| `boolean`  | Schema for properties not in `properties`. |
| propertyNames        |        Schema         | Schema for property names.                 |
| required             |      [`string`]       | Array of required property names.          |
| minProperties        |       `integer`       | Minimum number of properties.              |
| maxProperties        |       `integer`       | Maximum number of properties.              |

### Composition

| Property |   Type   | Description                            |
| -------- | :------: | -------------------------------------- |
| allOf    | [Schema] | Must match all schemas.                |
| anyOf    | [Schema] | Must match at least one schema.        |
| oneOf    | [Schema] | Must match exactly one schema.         |
| not      |  Schema  | Must not match this schema.            |
| if       |  Schema  | Conditional schema.                    |
| then     |  Schema  | Schema to apply if `if` matches.       |
| else     |  Schema  | Schema to apply if `if` doesn't match. |

## OpenAPI-Specific Properties

The following properties are specific to OpenAPI and not part of JSON Schema:

| Property      |                               Type                                | Description                                                                                                                                                                                                                                                                                                                                                                                                |
| ------------- | :---------------------------------------------------------------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| discriminator |           [Discriminator Object](#discriminator-object)           | Adds support for polymorphism.                                                                                                                                                                                                                                                                                                                                                                             |
| xml           |                     [XML Object](#xml-object)                     | This MAY be used only on properties schemas. It has no effect on root schemas. Adds additional metadata to describe the XML representation of this property.                                                                                                                                                                                                                                               |
| externalDocs  | [External Documentation Object](external_documentation_object.md) | Additional external documentation for this schema.                                                                                                                                                                                                                                                                                                                                                         |
| example       |                                any                                | A free-form property to include an example of an instance for this schema. To represent examples that cannot be naturally represented in JSON or YAML, a string value can be used to contain the example with escaping where necessary. **Deprecated:** The `example` property has been deprecated in favor of the JSON Schema `examples` keyword. Use of `example` is still permitted but is discouraged. |

## String Formats

Primitives have an optional modifier property: `format`. OAS uses several known formats to define in fine detail the data type being used. However, to support documentation needs, the `format` property is an open `string`-valued property, and can have any value.

### Common Formats

| Type      | Format          | Description                     |
| --------- | --------------- | ------------------------------- |
| `integer` | `int32`         | Signed 32-bit integer           |
| `integer` | `int64`         | Signed 64-bit integer           |
| `number`  | `float`         | Floating-point number           |
| `number`  | `double`        | Double-precision floating-point |
| `string`  | `password`      | Hint to UI to obscure input     |
| `string`  | `binary`        | Binary data (for file uploads)  |
| `string`  | `byte`          | Base64-encoded characters       |
| `string`  | `date`          | Full-date (RFC 3339)            |
| `string`  | `date-time`     | Date-time (RFC 3339)            |
| `string`  | `email`         | Email address (RFC 5322)        |
| `string`  | `hostname`      | Hostname (RFC 1123)             |
| `string`  | `ipv4`          | IPv4 address                    |
| `string`  | `ipv6`          | IPv6 address                    |
| `string`  | `uri`           | URI (RFC 3986)                  |
| `string`  | `uri-reference` | URI reference                   |
| `string`  | `uri-template`  | URI template (RFC 6570)         |
| `string`  | `uuid`          | UUID (RFC 4122)                 |
| `string`  | `json-pointer`  | JSON Pointer (RFC 6901)         |
| `string`  | `regex`         | Regular expression              |

## Discriminator Object

When request bodies or response payloads may be one of a number of different schemas, a `discriminator` object can be used to aid in serialization, deserialization, and validation. The discriminator is a specific object in a schema which is used to inform the consumer of the specification of an alternative schema based on the value associated with it.

### Fixed Fields

| Field Name   |          Type           | Description                                                                                   |
| ------------ | :---------------------: | --------------------------------------------------------------------------------------------- |
| propertyName |        `string`         | **REQUIRED**. The name of the property in the payload that will hold the discriminator value. |
| mapping      | Map[`string`, `string`] | An object to hold mappings between payload values and schema names or references.             |

### Discriminator Example

```yaml
components:
  schemas:
    Pet:
      type: object
      discriminator:
        propertyName: petType
        mapping:
          dog: "#/components/schemas/Dog"
          cat: "#/components/schemas/Cat"
      properties:
        name:
          type: string
        petType:
          type: string
      required:
        - name
        - petType

    Cat:
      allOf:
        - $ref: "#/components/schemas/Pet"
        - type: object
          properties:
            huntingSkill:
              type: string
              description: The measured skill for hunting
              enum:
                - clueless
                - lazy
                - adventurous
                - aggressive
          required:
            - huntingSkill

    Dog:
      allOf:
        - $ref: "#/components/schemas/Pet"
        - type: object
          properties:
            packSize:
              type: integer
              format: int32
              description: The size of the pack the dog is from
              default: 0
              minimum: 0
          required:
            - packSize
```

## XML Object

A metadata object that allows for more fine-tuned XML model definitions.

### Fixed Fields

| Field Name |   Type    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| ---------- | :-------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| name       | `string`  | Replaces the name of the element/attribute used for the described schema property. When defined within `items`, it will affect the name of the individual XML elements within the list. When defined alongside `type` being `array` (outside the `items`), it will affect the wrapping element. If `xml` is specified at both `items` and `array` levels, the `name` at `items` level will be used for the individual elements, and the `name` at `array` level will be used for the wrapping element. |
| namespace  | `string`  | The URI of the namespace definition. This MUST be in the form of an absolute URI.                                                                                                                                                                                                                                                                                                                                                                                                                      |
| prefix     | `string`  | The prefix to be used for the name.                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| attribute  | `boolean` | Declares whether the property definition translates to an attribute instead of an element. Default value is `false`.                                                                                                                                                                                                                                                                                                                                                                                   |
| wrapped    | `boolean` | MAY be used only for an array definition. Signifies whether the array is wrapped (for example, `<books><book/><book/></books>`) or unwrapped (`<book/><book/>`). Default value is `false`. The definition takes effect only when defined alongside `type` being `array` (outside the `items`).                                                                                                                                                                                                         |

### XML Object Examples

#### Basic XML Configuration

```yaml
xml:
  name: Pet
  namespace: http://example.com/schema
  prefix: pet
```

#### Attribute

```yaml
type: object
properties:
  id:
    type: integer
    xml:
      attribute: true
```

#### Wrapped Array

```yaml
type: array
items:
  type: string
  xml:
    name: item
xml:
  name: items
  wrapped: true
```

## Schema Object Examples

### Primitive Types

```yaml
type: string
minLength: 1
maxLength: 100
pattern: "^[a-zA-Z]+$"
```

```yaml
type: integer
format: int64
minimum: 0
maximum: 100
```

### Object Type

```yaml
type: object
required:
  - name
properties:
  name:
    type: string
  address:
    $ref: "#/components/schemas/Address"
  age:
    type: integer
    minimum: 0
additionalProperties: false
```

### Array Type

```yaml
type: array
items:
  $ref: "#/components/schemas/Pet"
minItems: 1
maxItems: 100
uniqueItems: false
```

### Enum

```yaml
type: string
enum:
  - active
  - inactive
  - pending
```

### Composition

```yaml
oneOf:
  - $ref: "#/components/schemas/Cat"
  - $ref: "#/components/schemas/Dog"
discriminator:
  propertyName: petType
```

### Nullable

```yaml
type: string
nullable: true # Deprecated in 3.1+, use type: ["string", "null"] instead
```

### Reference

```yaml
$ref: "#/components/schemas/User"
```

## See Also

- [Components Object](components_object.md) - Reusable schema definitions
- [Data Types](data_types.md) - Primitive data types
- [Parameter Object](parameter_object.md) - Parameter schemas
- [Media Type Object](media_type_object.md) - Media type schemas
