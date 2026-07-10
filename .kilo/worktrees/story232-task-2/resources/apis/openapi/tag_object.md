# Tag Object

Adds metadata to a single tag that is used by the [Operation Object](operation_object.md). It is not mandatory to have a Tag Object per tag defined in the Operation Object instances.

## Fixed Fields

| Field Name   |                               Type                                | Description                                                                                                            |
| ------------ | :---------------------------------------------------------------: | ---------------------------------------------------------------------------------------------------------------------- |
| name         |                             `string`                              | **REQUIRED**. The name of the tag.                                                                                     |
| description  |                             `string`                              | A description for the tag. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation. |
| externalDocs | [External Documentation Object](external_documentation_object.md) | Additional external documentation for this tag.                                                                        |

This object MAY be extended with Specification Extensions.

## Tag Object Example

```yaml
name: pet
description: Pets operations
externalDocs:
  description: Find out more
  url: https://example.com/docs/pets
```

## Using Tags

### Defining Tags

Tags are defined at the root level in the [OpenAPI Object](openapi_object.md):

```yaml
openapi: 3.2.0
info:
  title: Pet Store API
  version: 1.0.0
tags:
  - name: pets
    description: Pet operations
  - name: store
    description: Access to pet store orders
  - name: user
    description: Operations about user
    externalDocs:
      description: Find out more about users
      url: https://example.com/docs/users
```

### Using Tags in Operations

Operations reference tags by name:

```yaml
paths:
  /pets:
    get:
      tags:
        - pets
      summary: List all pets
      operationId: listPets
      responses:
        "200":
          description: A list of pets
    post:
      tags:
        - pets
      summary: Create a pet
      operationId: createPet
      responses:
        "201":
          description: Pet created
  /store/orders:
    get:
      tags:
        - store
      summary: List orders
      operationId: listOrders
```

## Tag Organization

### Grouping Operations

Tags are used to group operations in documentation:

```yaml
tags:
  - name: users
    description: User management operations
  - name: authentication
    description: Authentication and authorization
  - name: products
    description: Product catalog operations
  - name: orders
    description: Order management

paths:
  /users:
    get:
      tags: [users]
  /auth/login:
    post:
      tags: [authentication]
  /products:
    get:
      tags: [products]
  /orders:
    get:
      tags: [orders]
    post:
      tags: [orders]
```

### Multiple Tags

Operations can have multiple tags:

```yaml
paths:
  /users/{id}/orders:
    get:
      tags:
        - users
        - orders
      summary: Get orders for a user
```

### Tag Order

The order of tags in the `tags` array determines their display order in documentation:

```yaml
tags:
  - name: overview
    description: Overview and general information
  - name: authentication
    description: Authentication endpoints
  - name: users
    description: User management
  - name: products
    description: Product catalog
  - name: orders
    description: Order processing
```

## Undeclared Tags

Tags used in operations but not declared in the `tags` array are still valid:

```yaml
tags:
  - name: pets
    description: Pet operations

paths:
  /pets:
    get:
      tags: [pets] # Declared tag
  /admin:
    get:
      tags: [admin] # Undeclared tag - still valid
```

Undeclared tags:

- Are valid and functional
- May be displayed differently by tools
- Cannot have descriptions or external docs

## External Documentation

Tags can link to external documentation:

```yaml
tags:
  - name: payments
    description: Payment processing
    externalDocs:
      description: Payment API documentation
      url: https://docs.example.com/payments
  - name: webhooks
    description: Webhook configuration
    externalDocs:
      description: Webhook guide
      url: https://docs.example.com/webhooks
```

## Best Practices

### Naming Conventions

Use clear, consistent tag names:

```yaml
# Good
tags:
  - name: Users
  - name: User Authentication
  - name: Order Management

# Avoid
tags:
  - name: user-stuff
  - name: auth
  - name: orders_v2_api
```

### Description Guidelines

Provide helpful descriptions:

```yaml
tags:
  - name: users
    description: |
      User management operations including:
      - User registration
      - Profile management
      - Account settings
  - name: admin
    description: Administrative operations requiring elevated privileges
```

### Logical Grouping

Group related operations:

```yaml
tags:
  # Public API
  - name: public
    description: Public endpoints accessible without authentication
  - name: search
    description: Search and discovery operations

  # User API
  - name: users
    description: User account management
  - name: profile
    description: User profile operations

  # Admin API
  - name: admin
    description: Administrative operations
  - name: reports
    description: Reporting and analytics
```

## Tag Extensions

Tags can be extended with custom properties:

```yaml
tags:
  - name: internal
    description: Internal API endpoints
    x-display-order: 99
    x-visibility: internal
    x-team: platform
```

## See Also

- [Operation Object](operation_object.md) - Using tags in operations
- [OpenAPI Object](openapi_object.md) - Defining tags at the root level
- [External Documentation Object](external_documentation_object.md) - External documentation links
