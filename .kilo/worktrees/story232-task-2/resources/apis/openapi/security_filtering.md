# Security Filtering

OpenAPI Descriptions can contain sensitive information that should be excluded from certain audiences. The specification provides mechanisms to control what information is visible.

## Overview

Security filtering allows API descriptions to be tailored for different audiences while maintaining a single source of truth. This is useful when:

- Sharing API documentation with external partners
- Providing different views for internal vs external developers
- Hiding deprecated or internal operations
- Protecting sensitive implementation details

## Filtering Mechanisms

### Path Item Filtering

A [Path Item Object](paths_object.md#path-item-object) MAY be empty due to ACL constraints. The path itself is still exposed, but operations and parameters are hidden.

```yaml
# Full API description
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: Success
  /admin:
    get:
      summary: Admin operations
      responses:
        "200":
          description: Success
```

After filtering for external users:

```yaml
# Filtered API description
paths:
  /users:
    get:
      summary: List users
      responses:
        "200":
          description: Success
  /admin: {} # Empty path item - operations hidden
```

### Operation Filtering

Individual operations can be hidden while keeping the path visible:

```yaml
paths:
  /users:
    get:
      summary: List users
    post:
      summary: Create user
    delete:
      summary: Delete user # May be hidden for certain audiences
```

### Schema Filtering

Schemas can be filtered to hide sensitive properties:

```yaml
# Full schema
components:
  schemas:
    User:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
        email:
          type: string
        internalNotes:
          type: string
        salary:
          type: number
```

Filtered for external use:

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
        email:
          type: string
```

## Implementation Approaches

### Server-Side Filtering

Filter the API description before serving:

1. **Role-based**: Different descriptions for different user roles
2. **Partner-based**: Custom descriptions for integration partners
3. **Environment-based**: Different descriptions for dev/staging/prod

### Documentation Generation

Generate multiple documentation sets from a single source:

```yaml
# Source API description with annotations
paths:
  /internal/metrics:
    get:
      summary: System metrics
      x-visibility: internal
      responses:
        "200":
          description: Metrics data

  /public/status:
    get:
      summary: System status
      x-visibility: public
      responses:
        "200":
          description: Status information
```

### Extension-Based Filtering

Use extensions to mark visibility:

```yaml
paths:
  /admin/users:
    get:
      summary: Admin user list
      x-audience:
        - internal
        - admin
      responses:
        "200":
          description: List of all users

  /users:
    get:
      summary: Public user list
      x-audience:
        - public
      responses:
        "200":
          description: List of public users
```

## Best Practices

### Maintain Single Source

Keep one complete API description and filter for distribution:

```
source/
├── openapi-full.yaml      # Complete API description
└── generate-docs.sh       # Script to generate filtered versions

output/
├── openapi-public.yaml    # Public API documentation
├── openapi-partner.yaml   # Partner API documentation
└── openapi-internal.yaml  # Internal API documentation
```

### Use Extensions for Metadata

Add metadata to control filtering:

```yaml
x-visibility: internal
x-roles:
  - admin
  - support
x-department: finance
```

### Document Filtering Rules

Clearly document what is filtered and why:

```yaml
info:
  title: Example API
  x-filtering:
    public:
      description: Public API for external developers
      exclude:
        - x-visibility: internal
        - x-roles: admin
    partner:
      description: Partner API with extended access
      exclude:
        - x-visibility: internal
```

### Preserve Structure

When filtering, preserve the overall structure:

- Keep paths even if operations are hidden
- Maintain component references
- Preserve security scheme names

## Security Considerations

### What to Filter

Consider filtering:

- Internal endpoints
- Admin operations
- Sensitive data fields
- Implementation details
- Deprecated features

### What Not to Filter

Avoid filtering:

- Security requirements (unless the operation is completely hidden)
- Required parameters
- Error responses that help debugging
- Rate limiting information

### Authentication and Authorization

Security filtering is NOT a substitute for proper authentication:

```yaml
# Security filtering hides this from documentation
/admin:
  get:
    x-visibility: internal
    responses:
      '200':
        description: Admin data

# But the endpoint still requires authentication
/admin:
  get:
    security:
      - oauth2:
          - admin:read
    responses:
      '200':
        description: Admin data
      '401':
        description: Unauthorized
      '403':
        description: Forbidden
```

## Example: Multi-Audience API

```yaml
openapi: 3.2.0
info:
  title: Multi-Audience API
  version: 1.0.0

# Public endpoints
paths:
  /products:
    get:
      x-audience: public
      summary: List products
      responses:
        "200":
          description: Product list

  /orders:
    post:
      x-audience: public
      summary: Create order
      responses:
        "201":
          description: Order created

  # Partner endpoints
  /partner/analytics:
    get:
      x-audience: partner
      summary: Sales analytics
      responses:
        "200":
          description: Analytics data

  # Internal endpoints
  /internal/metrics:
    get:
      x-audience: internal
      summary: System metrics
      responses:
        "200":
          description: Metrics data

components:
  securitySchemes:
    oauth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://auth.example.com/authorize
          tokenUrl: https://auth.example.com/token
          scopes:
            public:read: Read public data
            partner:read: Read partner data
            internal:read: Read internal data
```

## See Also

- [Paths Object](paths_object.md) - Path item structure
- [Operation Object](operation_object.md) - Operation definitions
- [Security Objects](security_objects.md) - Security requirements
