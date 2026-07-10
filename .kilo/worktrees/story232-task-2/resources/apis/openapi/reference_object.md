# Reference Object

A simple object to allow referencing other components in the OpenAPI Description, internally and externally.

The Reference Object is defined by [JSON Reference](https://tools.ietf.org/html/draft-pbryan-zyp-json-ref-03) and follows the same structure.

## Fixed Fields

| Field Name |   Type   | Description                         |
| ---------- | :------: | ----------------------------------- |
| $ref       | `string` | **REQUIRED**. The reference string. |

This object cannot be extended with additional properties and any properties added SHALL be ignored.

## Reference Syntax

### Internal References

References to components within the same document use JSON Pointer syntax:

```yaml
$ref: '#/components/schemas/User'
$ref: '#/components/parameters/userId'
$ref: '#/components/responses/NotFound'
$ref: '#/paths/~1users~1{id}/get'
```

### External References

References to external documents:

```yaml
# Absolute URL
$ref: 'https://example.com/schemas/user.json'

# Relative file path
$ref: './schemas/User.yaml'

# Relative file path with fragment
$ref: './common.yaml#/components/schemas/Error'
```

## Reference Resolution

### JSON Pointer Escaping

When using JSON Pointer syntax, certain characters must be escaped:

| Character | Escaped Form |
| --------- | ------------ |
| `~`       | `~0`         |
| `/`       | `~1`         |

Example:

```yaml
# Path: /users/{id}
$ref: '#/paths/~1users~1{id}/get'

# Path: /api/v1/users
$ref: '#/paths/~1api~1v1~1users/get'
```

### Base URI Resolution

References are resolved relative to the base URI:

1. If `$self` is present in the document, it sets the base URI
2. Otherwise, the retrieval URI is used
3. Relative references are resolved against the base URI

## Reference Examples

### Schema Reference

```yaml
schema:
  $ref: "#/components/schemas/User"
```

### Parameter Reference

```yaml
parameters:
  - $ref: "#/components/parameters/userId"
```

### Response Reference

```yaml
responses:
  "404":
    $ref: "#/components/responses/NotFound"
```

### Request Body Reference

```yaml
requestBody:
  $ref: "#/components/requestBodies/UserBody"
```

### Path Item Reference

```yaml
paths:
  /users:
    $ref: "#/components/pathItems/UserPath"
```

### External Schema Reference

```yaml
schema:
  $ref: "https://example.com/schemas/user.json"
```

### Relative File Reference

```yaml
schema:
  $ref: "./schemas/User.yaml"
```

### Reference with Fragment

```yaml
schema:
  $ref: "./common.yaml#/components/schemas/Error"
```

## Circular References

Circular references are allowed but should be used carefully:

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        name:
          type: string
        manager:
          $ref: "#/components/schemas/User" # Circular reference
        reports:
          type: array
          items:
            $ref: "#/components/schemas/User" # Circular reference
```

## Reference vs Direct Definition

### Using References (Recommended for Reuse)

```yaml
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string

paths:
  /users:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
  /users/{id}:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
```

### Direct Definition (For Single Use)

```yaml
paths:
  /status:
    get:
      responses:
        "200":
          content:
            application/json:
              schema:
                type: object
                properties:
                  status:
                    type: string
```

## Multi-Document References

References can span multiple documents:

### Entry Document

```yaml
openapi: 3.2.0
$self: https://api.example.com/openapi.yaml
info:
  title: Example API
  version: 1.0.0
paths:
  /users:
    $ref: "paths/users.yaml" # External path item
components:
  schemas:
    User:
      $ref: "schemas/user.yaml" # External schema
```

### External Schema Document

```yaml
# schemas/user.yaml
$self: https://api.example.com/schemas/user.yaml
type: object
properties:
  id:
    type: integer
  name:
    type: string
  email:
    type: string
    format: email
```

## Reference Resolution Order

When resolving references:

1. Check if the reference is internal (starts with `#`)
2. If external, resolve against the base URI
3. Fetch the external document if not already loaded
4. Apply the fragment (if any) to locate the target

## Best Practices

### Use Components for Reuse

```yaml
# Good - Define once, reference many times
components:
  schemas:
    Error:
      type: object
      properties:
        code:
          type: integer
        message:
          type: string

paths:
  /users:
    get:
      responses:
        "404":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
  /products:
    get:
      responses:
        "404":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Error"
```

### Use Descriptive Names

```yaml
# Good
$ref: '#/components/schemas/UserResponse'
$ref: '#/components/schemas/UserCreateRequest'

# Avoid
$ref: '#/components/schemas/User1'
$ref: '#/components/schemas/User2'
```

### Organize External Files

```
api/
├── openapi.yaml
├── paths/
│   ├── users.yaml
│   └── products.yaml
├── schemas/
│   ├── user.yaml
│   ├── product.yaml
│   └── common.yaml
└── parameters/
    └── pagination.yaml
```

## See Also

- [Components Object](components_object.md) - Defining reusable components
- [Schema Object](schema_object.md) - Schema definitions
- [OpenAPI Object](openapi_object.md) - Base URI and document structure
- [Appendix F: Base URI Examples](appendix_f.md) - Reference resolution examples
