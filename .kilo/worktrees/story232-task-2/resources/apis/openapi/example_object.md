# Example Object

The Example Object shows an example of a schema or property.

## Fixed Fields

| Field Name    |   Type   | Description                                                                                                                                                                                                                                                                 |
| ------------- | :------: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| summary       | `string` | Short description for the example.                                                                                                                                                                                                                                          |
| description   | `string` | Long description for the example. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                               |
| value         |   Any    | Embedded literal example. The `value` field and `externalValue` field are mutually exclusive. To represent examples of media types that cannot naturally be represented in JSON or YAML, use a string value to contain the example, escaping where necessary.               |
| externalValue | `string` | A URI that points to the literal example. This provides the capability to reference examples that cannot easily be included in JSON or YAML documents. The `value` field and `externalValue` field are mutually exclusive. See the rules for resolving Relative References. |

This object MAY be extended with Specification Extensions.

## Example Object Examples

### Inline Value

```yaml
summary: A simple user example
value:
  id: 1
  name: John Doe
  email: john@example.com
```

### String Value

```yaml
summary: An XML example
value: |
  <user>
    <id>1</id>
    <name>John Doe</name>
    <email>john@example.com</email>
  </user>
```

### External Value

```yaml
summary: A large dataset example
externalValue: "https://example.com/examples/large-dataset.json"
```

### With Description

```yaml
summary: Admin user example
description: |
  This example shows an admin user with full permissions.
  Note that the `role` field is set to `admin`.
value:
  id: 1
  name: Admin User
  email: admin@example.com
  role: admin
  permissions:
    - read
    - write
    - delete
```

## Using Examples

### In Media Type Object

```yaml
content:
  application/json:
    schema:
      $ref: "#/components/schemas/User"
    examples:
      basic:
        summary: Basic user
        value:
          id: 1
          name: John Doe
      admin:
        summary: Admin user
        value:
          id: 2
          name: Admin
          role: admin
```

### In Parameter Object

```yaml
parameters:
  - name: userId
    in: path
    required: true
    schema:
      type: integer
    examples:
      valid:
        summary: A valid user ID
        value: 1
      invalid:
        summary: An invalid user ID
        value: -1
```

### In Header Object

```yaml
headers:
  X-Request-ID:
    schema:
      type: string
      format: uuid
    examples:
      example1:
        summary: Standard UUID
        value: 550e8400-e29b-41d4-a716-446655440000
```

## Multiple Examples

When multiple examples are provided, each must have a unique key:

```yaml
content:
  application/json:
    schema:
      $ref: "#/components/schemas/Pet"
    examples:
      dog:
        summary: A dog
        value:
          name: Fluffy
          petType: dog
      cat:
        summary: A cat
        value:
          name: Mittens
          petType: cat
      bird:
        summary: A bird
        externalValue: "https://example.com/examples/bird.json"
```

## Example vs examples

### Single Example

Use `example` for a single, simple example:

```yaml
schema:
  type: string
example: "Hello, World!"
```

### Multiple Examples

Use `examples` for multiple examples with descriptions:

```yaml
schema:
  type: string
examples:
  greeting:
    summary: A greeting
    value: "Hello, World!"
  farewell:
    summary: A farewell
    value: "Goodbye!"
```

## Reusable Examples

Examples can be defined in the [Components Object](components_object.md):

```yaml
components:
  examples:
    UserExample:
      summary: A typical user
      value:
        id: 1
        name: John Doe
        email: john@example.com
    AdminExample:
      summary: An admin user
      value:
        id: 2
        name: Admin
        email: admin@example.com
        role: admin

paths:
  /users/{id}:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
              examples:
                user:
                  $ref: "#/components/examples/UserExample"
                admin:
                  $ref: "#/components/examples/AdminExample"
```

## External Examples

For large or complex examples, use `externalValue`:

```yaml
examples:
  largeDataset:
    summary: A large dataset
    externalValue: "https://example.com/examples/large-dataset.json"
  imageExample:
    summary: An image file
    externalValue: "https://example.com/examples/image.png"
```

### Relative External Examples

```yaml
examples:
  localExample:
    summary: Local example file
    externalValue: "./examples/user.json"
```

## Example Values and Schema

Examples SHOULD match the schema they are associated with:

```yaml
schema:
  type: object
  required:
    - id
    - name
  properties:
    id:
      type: integer
    name:
      type: string
    email:
      type: string
      format: email
examples:
  valid:
    summary: Valid user
    value:
      id: 1
      name: John Doe
      email: john@example.com
  minimal:
    summary: Minimal valid user
    value:
      id: 2
      name: Jane Doe
```

## Binary Examples

For binary data, use string representation:

```yaml
schema:
  type: string
  format: binary
examples:
  image:
    summary: Base64 encoded image
    value: iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==
```

## See Also

- [Media Type Object](media_type_object.md) - Using examples in media types
- [Parameter Object](parameter_object.md) - Using examples in parameters
- [Header Object](header_object.md) - Using examples in headers
- [Components Object](components_object.md) - Reusable examples
