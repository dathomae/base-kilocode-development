# External Documentation Object

Allows referencing an external resource for extended documentation.

## Fixed Fields

| Field Name  |   Type   | Description                                                                                                                            |
| ----------- | :------: | -------------------------------------------------------------------------------------------------------------------------------------- |
| description | `string` | A description of the target documentation. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation. |
| url         | `string` | **REQUIRED**. The URL for the target documentation. Value MUST be in the format of a URL.                                              |

This object MAY be extended with Specification Extensions.

## External Documentation Object Example

```yaml
description: Find out more about our API
url: https://example.com/docs/api
```

## Using External Documentation

### At the Root Level

```yaml
openapi: 3.2.0
info:
  title: Pet Store API
  version: 1.0.0
externalDocs:
  description: Find out more about the Pet Store
  url: https://example.com/docs/pet-store
```

### In an Operation

```yaml
paths:
  /pets:
    get:
      summary: List all pets
      externalDocs:
        description: Learn more about pet listing
        url: https://example.com/docs/pets#list
```

### In a Tag

```yaml
tags:
  - name: pets
    description: Pet operations
    externalDocs:
      description: Pet API documentation
      url: https://example.com/docs/pets
```

### In a Schema

```yaml
components:
  schemas:
    Pet:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
      externalDocs:
        description: Pet schema documentation
        url: https://example.com/docs/schemas/pet
```

## Common Use Cases

### API Documentation Portal

```yaml
externalDocs:
  description: Full API documentation
  url: https://docs.example.com/api
```

### Developer Portal

```yaml
externalDocs:
  description: Developer portal with guides and tutorials
  url: https://developers.example.com
```

### GitHub Repository

```yaml
externalDocs:
  description: Source code and issue tracker
  url: https://github.com/example/api
```

### Standards Documentation

```yaml
externalDocs:
  description: OAuth 2.0 specification
  url: https://oauth.net/2/
```

## Best Practices

### Descriptive Text

Provide clear descriptions:

```yaml
# Good
externalDocs:
  description: Comprehensive guide to authentication flows
  url: https://docs.example.com/auth

# Avoid
externalDocs:
  description: More info
  url: https://docs.example.com
```

### Specific URLs

Link to specific sections when possible:

```yaml
# Good - Links to specific section
externalDocs:
  description: Rate limiting documentation
  url: https://docs.example.com/api#rate-limiting

# Less specific - Links to general page
externalDocs:
  description: API documentation
  url: https://docs.example.com
```

### Multiple Documentation Links

Use external docs at different levels for comprehensive coverage:

```yaml
openapi: 3.2.0
externalDocs:
  description: General API documentation
  url: https://docs.example.com/api

tags:
  - name: authentication
    description: Authentication operations
    externalDocs:
      description: Authentication guide
      url: https://docs.example.com/auth

paths:
  /auth/login:
    post:
      summary: User login
      externalDocs:
        description: Login flow documentation
        url: https://docs.example.com/auth#login
```

## See Also

- [OpenAPI Object](openapi_object.md) - Root level external docs
- [Operation Object](operation_object.md) - Operation level external docs
- [Tag Object](tag_object.md) - Tag level external docs
- [Schema Object](schema_object.md) - Schema level external docs
