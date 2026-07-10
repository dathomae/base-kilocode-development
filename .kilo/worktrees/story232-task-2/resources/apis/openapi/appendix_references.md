# Appendix: Reference Resolution

This appendix provides guidance on parsing and resolving references in OpenAPI Descriptions.

## Base URI Determination

The base URI of an OpenAPI document is determined in the following order:

### 1. Base URI Within Content (Highest Priority)

The `$self` field in the OpenAPI Object or `$id` in Schema Objects:

```yaml
openapi: 3.2.0
$self: https://example.com/api/openapi
info:
  title: Example API
  version: 1.0.0
```

### 2. Base URI From Encapsulating Entity

For documents within a larger container (e.g., multipart/related):

```
--boundary
Content-Type: application/openapi+yaml
Content-Location: https://example.com/api/openapi.yaml

openapi: 3.2.0
...
--boundary
```

### 3. Base URI From Retrieval URI

The URI used to retrieve the document:

```
GET /api/openapi.yaml HTTP/1.1
Host: example.com
```

Base URI: `https://example.com/api/openapi.yaml`

### 4. Application-Specific Default

When no other source is available, applications may define a default base URI.

## Reference Resolution Examples

### Internal References

```yaml
schema:
  $ref: "#/components/schemas/User"
```

Resolves to the User schema within the current document.

### Relative File References

```yaml
schema:
  $ref: "./schemas/User.yaml"
```

Resolved relative to the current document's base URI.

### Absolute URL References

```yaml
schema:
  $ref: "https://example.com/schemas/User.json"
```

Resolved directly to the absolute URL.

### Fragment References

```yaml
schema:
  $ref: "common.yaml#/components/schemas/Error"
```

Loads `common.yaml` and navigates to the specified fragment.

## Multi-Document Resolution

### Entry Document

```yaml
# openapi.yaml
openapi: 3.2.0
$self: https://example.com/api/openapi
paths:
  /users:
    $ref: "paths/users.yaml"
components:
  schemas:
    User:
      $ref: "schemas/user.yaml"
```

### Referenced Path Item

```yaml
# paths/users.yaml
$self: https://example.com/api/paths/users
get:
  operationId: listUsers
  responses:
    "200":
      description: List of users
```

### Referenced Schema

```yaml
# schemas/user.yaml
$self: https://example.com/api/schemas/user
type: object
properties:
  id:
    type: integer
  name:
    type: string
```

## Relative $self and $id

Using relative URIs allows documents to work across different hosts:

```yaml
# Retrieved from https://staging.example.com/api/openapi
openapi: 3.2.0
$self: /api/openapi # Resolves to https://staging.example.com/api/openapi
paths:
  /users:
    $ref: "paths/users" # Resolves to https://staging.example.com/api/paths/users
```

This same document deployed to production:

```yaml
# Retrieved from https://api.example.com/openapi
openapi: 3.2.0
$self: /api/openapi # Resolves to https://api.example.com/api/openapi
paths:
  /users:
    $ref: "paths/users" # Resolves to https://api.example.com/api/paths/users
```

## Implicit Connections

Some objects reference others by name rather than URI:

### Security Requirements

```yaml
security:
  - api_key: [] # References Security Scheme named "api_key"
```

### Discriminator Mapping

```yaml
discriminator:
  propertyName: petType
  mapping:
    dog: "#/components/schemas/Dog" # Explicit reference
    cat: Cat # Implicit reference to schema named "Cat"
```

### Operation Tags

```yaml
get:
  tags:
    - users # References Tag Object named "users"
```

### Link operationId

```yaml
links:
  userOrders:
    operationId: getOrders # References operation with this ID
```

## Resolution Best Practices

### Use $self for Multi-Document APIs

```yaml
openapi: 3.2.0
$self: https://api.example.com/v1/openapi
```

### Use Relative References for Portability

```yaml
# Good - works in any environment
$ref: './schemas/User.yaml'

# Avoid - breaks when moved
$ref: 'https://staging.example.com/schemas/User.yaml'
```

### Resolve Implicit Connections from Entry Document

When resolving security requirements and other implicit connections, use the entry document's components, not the current document's.

## Parsing Guidance

### Document Detection

Implementations may detect document types by:

1. Media type (`application/openapi+json`, `application/openapi+yaml`)
2. Root `openapi` field (OpenAPI document)
3. JSON Schema keywords (Schema document)

### Fragment Parsing

Parsing fragments in isolation is NOT RECOMMENDED:

- Base URI may be different
- Security risks from incorrect reference resolution
- Behavior is undefined

### Context Conflicts

If the same object is parsed in different contexts (e.g., as both a Schema Object and a Path Item Object), behavior is implementation-defined.

## Common Issues

### Circular References

```yaml
components:
  schemas:
    User:
      properties:
        manager:
          $ref: "#/components/schemas/User"
```

Handle with depth limits or detection algorithms.

### Missing References

```yaml
schema:
  $ref: "#/components/schemas/NonExistent"
```

Should produce an error during validation.

### Invalid JSON Pointers

```yaml
$ref: "#/invalid/pointer"
```

JSON Pointer must reference an existing location.

## See Also

- [Reference Object](reference_object.md) - Reference syntax
- [OpenAPI Object](openapi_object.md) - $self field
- [Schema Object](schema_object.md) - $id field
- [Components Object](components_object.md) - Reusable components
