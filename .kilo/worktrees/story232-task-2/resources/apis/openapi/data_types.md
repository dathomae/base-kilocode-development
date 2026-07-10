# Data Types

The OpenAPI Specification uses data types based on JSON Schema. This document describes the primitive data types and their formats.

## Primitive Data Types

### string

A sequence of characters.

```yaml
type: string
```

### number

Any numeric value (integer or floating-point).

```yaml
type: number
```

### integer

A whole number (no fractional part).

```yaml
type: integer
```

### boolean

A true or false value.

```yaml
type: boolean
```

### array

An ordered list of values.

```yaml
type: array
items:
  type: string
```

### object

A collection of key-value pairs.

```yaml
type: object
properties:
  name:
    type: string
```

## String Formats

Formats provide additional semantic meaning to string values.

### Date and Time

| Format      | Description          | Example                |
| ----------- | -------------------- | ---------------------- |
| `date`      | Full-date (RFC 3339) | `2024-01-15`           |
| `date-time` | Date-time (RFC 3339) | `2024-01-15T10:30:00Z` |
| `time`      | Time (RFC 3339)      | `10:30:00Z`            |
| `duration`  | Duration (ISO 8601)  | `P1Y2M3DT4H5M6S`       |

### Network Identifiers

| Format          | Description             | Example                    |
| --------------- | ----------------------- | -------------------------- |
| `uri`           | URI (RFC 3986)          | `https://example.com/path` |
| `uri-reference` | URI reference           | `/path/to/resource`        |
| `uri-template`  | URI template (RFC 6570) | `/users/{userId}`          |
| `hostname`      | Hostname (RFC 1123)     | `api.example.com`          |
| `ipv4`          | IPv4 address            | `192.168.1.1`              |
| `ipv6`          | IPv6 address            | `2001:db8::1`              |

### Identifiers

| Format | Description     | Example                                |
| ------ | --------------- | -------------------------------------- |
| `uuid` | UUID (RFC 4122) | `550e8400-e29b-41d4-a716-446655440000` |
| `ulid` | ULID            | `01ARZ3NDEKTSV4RRFFQ69G5FAV`           |

### Strings

| Format                  | Description              | Example            |
| ----------------------- | ------------------------ | ------------------ |
| `email`                 | Email address (RFC 5322) | `user@example.com` |
| `password`              | Password (UI hint)       | `********`         |
| `regex`                 | Regular expression       | `^[a-z]+$`         |
| `json-pointer`          | JSON Pointer (RFC 6901)  | `/users/0/name`    |
| `relative-json-pointer` | Relative JSON Pointer    | `1/name`           |

### Binary Data

| Format   | Description    | Usage                   |
| -------- | -------------- | ----------------------- |
| `binary` | Binary data    | File uploads, responses |
| `byte`   | Base64-encoded | Binary data in JSON     |

## Numeric Formats

### Integer Formats

| Format  | Description           | Range                                                   |
| ------- | --------------------- | ------------------------------------------------------- |
| `int32` | 32-bit signed integer | -2,147,483,648 to 2,147,483,647                         |
| `int64` | 64-bit signed integer | -9,223,372,036,854,775,808 to 9,223,372,036,854,775,807 |

### Number Formats

| Format   | Description                     |
| -------- | ------------------------------- |
| `float`  | Single-precision floating-point |
| `double` | Double-precision floating-point |

## Type Examples

### String Types

```yaml
# Basic string
name:
  type: string

# Email
email:
  type: string
  format: email

# Date
birthDate:
  type: string
  format: date

# Enum
status:
  type: string
  enum:
    - active
    - inactive
    - pending

# Pattern
phoneNumber:
  type: string
  pattern: '^\\+?[1-9]\\d{1,14}$'

# Length constraints
username:
  type: string
  minLength: 3
  maxLength: 20
```

### Numeric Types

```yaml
# Integer
age:
  type: integer
  minimum: 0
  maximum: 150

# 64-bit integer ID
id:
  type: integer
  format: int64

# Floating-point
price:
  type: number
  format: double
  minimum: 0

# Percentage
percentage:
  type: number
  minimum: 0
  maximum: 100
```

### Boolean Type

```yaml
isActive:
  type: boolean
  default: false
```

### Array Types

```yaml
# Simple array
tags:
  type: array
  items:
    type: string

# Array of objects
users:
  type: array
  items:
    type: object
    properties:
      id:
        type: integer
      name:
        type: string

# Array constraints
limitedList:
  type: array
  items:
    type: string
  minItems: 1
  maxItems: 10
  uniqueItems: true
```

### Object Types

```yaml
# Basic object
address:
  type: object
  properties:
    street:
      type: string
    city:
      type: string
    country:
      type: string
  required:
    - street
    - city

# Object with additional properties
metadata:
  type: object
  additionalProperties:
    type: string

# Strict object (no additional properties)
strictObject:
  type: object
  properties:
    name:
      type: string
  additionalProperties: false
```

## Nullable Types

In OpenAPI 3.1+, use an array of types to allow null:

```yaml
# Can be string or null
middleName:
  type:
    - string
    - "null"
```

## Multiple Types

A value can be one of several types:

```yaml
# Can be string or number
id:
  type:
    - string
    - integer
```

## Type Composition

### allOf

Must match all schemas:

```yaml
allOf:
  - type: object
    properties:
      name:
        type: string
  - type: object
    properties:
      email:
        type: string
```

### anyOf

Must match at least one schema:

```yaml
anyOf:
  - type: string
  - type: number
```

### oneOf

Must match exactly one schema:

```yaml
oneOf:
  - $ref: "#/components/schemas/Cat"
  - $ref: "#/components/schemas/Dog"
```

### not

Must not match the schema:

```yaml
not:
  type: null
```

## Default Values

```yaml
role:
  type: string
  default: user
  enum:
    - user
    - admin
```

## Constants

```yaml
version:
  const: "1.0"
```

## See Also

- [Schema Object](schema_object.md) - Complete schema definition
- [Parameter Object](parameter_object.md) - Parameter data types
- [Media Type Object](media_type_object.md) - Content type definitions
