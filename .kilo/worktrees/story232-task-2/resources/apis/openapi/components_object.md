# Components Object

Holds a set of reusable objects for different aspects of the OAS. All objects defined within the components object will have no effect on the API unless they are explicitly referenced from properties outside the components object.

## Fixed Fields

| Field Name      |                                                              Type                                                              | Description                                                                                       |
| --------------- | :----------------------------------------------------------------------------------------------------------------------------: | ------------------------------------------------------------------------------------------------- |
| schemas         |                  Map[`string`, [Schema Object](schema_object.md) \| [Reference Object](reference_object.md)]                   | An object to hold reusable [Schema Objects](schema_object.md).                                    |
| responses       |        Map[`string`, [Response Object](responses_object.md#response-object) \| [Reference Object](reference_object.md)]        | An object to hold reusable [Response Objects](responses_object.md#response-object).               |
| parameters      |               Map[`string`, [Parameter Object](parameter_object.md) \| [Reference Object](reference_object.md)]                | An object to hold reusable [Parameter Objects](parameter_object.md).                              |
| examples        |                 Map[`string`, [Example Object](example_object.md) \| [Reference Object](reference_object.md)]                  | An object to hold reusable [Example Objects](example_object.md).                                  |
| requestBodies   |            Map[`string`, [Request Body Object](request_body_object.md) \| [Reference Object](reference_object.md)]             | An object to hold reusable [Request Body Objects](request_body_object.md).                        |
| headers         |                  Map[`string`, [Header Object](header_object.md) \| [Reference Object](reference_object.md)]                   | An object to hold reusable [Header Objects](header_object.md).                                    |
| securitySchemes | Map[`string`, [Security Scheme Object](security_objects.md#security-scheme-object) \| [Reference Object](reference_object.md)] | An object to hold reusable [Security Scheme Objects](security_objects.md#security-scheme-object). |
| links           |                    Map[`string`, [Link Object](link_object.md) \| [Reference Object](reference_object.md)]                     | An object to hold reusable [Link Objects](link_object.md).                                        |
| callbacks       |                Map[`string`, [Callback Object](callback_object.md) \| [Reference Object](reference_object.md)]                 | An object to hold reusable [Callback Objects](callback_object.md).                                |
| pathItems       |         Map[`string`, [Path Item Object](paths_object.md#path-item-object) \| [Reference Object](reference_object.md)]         | An object to hold reusable [Path Item Objects](paths_object.md#path-item-object).                 |

This object MAY be extended with Specification Extensions.

## Naming Conventions

All the fixed fields declared above are objects that MUST use keys that match the regular expression: `^[a-zA-Z0-9._-]+$`

### Valid Keys

- `User`
- `User_Model`
- `user-model`
- `user.model`
- `User123`

### Invalid Keys

- `User Model` (contains space)
- `User@Model` (contains special character)
- `User/Model` (contains forward slash)

## Components Object Example

```yaml
components:
  schemas:
    GeneralError:
      type: object
      properties:
        code:
          type: integer
          format: int32
        message:
          type: string
    User:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
  parameters:
    skipParam:
      name: skip
      in: query
      description: number of items to skip
      required: true
      schema:
        type: integer
        format: int32
    limitParam:
      name: limit
      in: query
      description: max records to return
      required: true
      schema:
        type: integer
        format: int32
  responses:
    NotFound:
      description: Entity not found.
    IllegalInput:
      description: Illegal input for operation.
    GeneralError:
      description: General Error
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/GeneralError"
  securitySchemes:
    api_key:
      type: apiKey
      name: api_key
      in: header
    petstore_auth:
      type: oauth2
      flows:
        implicit:
          authorizationUrl: https://example.org/api/oauth/dialog
          scopes:
            write:pets: modify pets in your account
            read:pets: read your pets
```

## Referencing Components

Components are referenced using JSON Reference syntax:

### Schema Reference

```yaml
schema:
  $ref: "#/components/schemas/User"
```

### Parameter Reference

```yaml
parameters:
  - $ref: "#/components/parameters/skipParam"
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

### Security Scheme Reference

```yaml
security:
  - $ref: "#/components/securitySchemes/api_key"
```

## External References

Components can also reference external documents:

### External Schema

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
  $ref: "./schemas/common.yaml#/components/schemas/Error"
```

## Best Practices

### Organization

Organize components logically:

```yaml
components:
  schemas:
    # Domain models
    User:
      # ...
    Pet:
      # ...

    # Request/response models
    UserRequest:
      # ...
    UserResponse:
      # ...

    # Common models
    Error:
      # ...
    Pagination:
      # ...
```

### Reusability

Define components for commonly used patterns:

```yaml
components:
  parameters:
    PageOffset:
      name: offset
      in: query
      schema:
        type: integer
        minimum: 0
        default: 0
    PageLimit:
      name: limit
      in: query
      schema:
        type: integer
        minimum: 1
        maximum: 100
        default: 20
  headers:
    X-RateLimit-Limit:
      description: Request limit per hour
      schema:
        type: integer
    X-RateLimit-Remaining:
      description: Remaining requests
      schema:
        type: integer
```

### Schema Composition

Use composition for shared properties:

```yaml
components:
  schemas:
    BaseObject:
      type: object
      properties:
        id:
          type: string
          format: uuid
        createdAt:
          type: string
          format: date-time
        updatedAt:
          type: string
          format: date-time

    User:
      allOf:
        - $ref: "#/components/schemas/BaseObject"
        - type: object
          properties:
            name:
              type: string
            email:
              type: string
              format: email
```

## See Also

- [Schema Object](schema_object.md) - Schema definitions
- [Parameter Object](parameter_object.md) - Parameter definitions
- [Responses Object](responses_object.md) - Response definitions
- [Security Objects](security_objects.md) - Security scheme definitions
- [Reference Object](reference_object.md) - Reference syntax
