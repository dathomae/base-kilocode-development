# Operation Object

Describes a single API operation on a path.

## Fixed Fields

| Field Name   |                                              Type                                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ------------ | :---------------------------------------------------------------------------------------------: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| tags         |                                           [`string`]                                            | A list of tags for API documentation control. Tags can be used for logical grouping of operations by resources or any other qualifier.                                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| summary      |                                            `string`                                             | A short summary of what the operation does.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| description  |                                            `string`                                             | A verbose explanation of the operation behavior. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                                                                                                                                                                                                                                                                                                                                              |
| externalDocs |                [External Documentation Object](external_documentation_object.md)                | Additional external documentation for this operation.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                     |
| operationId  |                                            `string`                                             | Unique string used to identify the operation. The id MUST be unique among all operations described in the API. The operationId value is **case-sensitive**. Tools and libraries MAY use the operationId to uniquely identify an operation, therefore, it is RECOMMENDED to follow common programming naming conventions.                                                                                                                                                                                                                                                                                  |
| parameters   |      [[Parameter Object](parameter_object.md) \| [Reference Object](reference_object.md)]       | A list of parameters that are applicable for this operation. If a parameter is already defined at the [Path Item](paths_object.md#path-item-object), the new definition will override it but can never remove it. The list MUST NOT include duplicated parameters. A unique parameter is defined by a combination of a name and location. The list can use the [Reference Object](reference_object.md) to link to parameters that are defined at the [OpenAPI Object's components/parameters](components_object.md#parameters).                                                                           |
| requestBody  |    [Request Body Object](request_body_object.md) \| [Reference Object](reference_object.md)     | The request body applicable for this operation. The `requestBody` is fully supported in HTTP methods where the HTTP 1.1 specification [RFC7231](https://tools.ietf.org/html/rfc7231#section-4.3.1) has explicitly defined semantics for request bodies. In other cases where the HTTP spec is vague (such as [GET](https://tools.ietf.org/html/rfc7231#section-4.3.2), [HEAD](https://tools.ietf.org/html/rfc7231#section-4.3.3) and [DELETE](https://tools.ietf.org/html/rfc7231#section-4.3.5)), `requestBody` is permitted but does not have well-defined semantics and SHOULD be avoided if possible. |
| responses    |                             [Responses Object](responses_object.md)                             | The list of possible responses as they are returned from executing this operation.                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| callbacks    | Map[`string`, [Callback Object](callback_object.md) \| [Reference Object](reference_object.md)] | A map of possible out-of band callbacks related to the parent operation. The key is a unique identifier for the Callback Object. Each value in the map is a [Callback Object](callback_object.md) that describes a request that may be initiated by the API provider and the expected responses.                                                                                                                                                                                                                                                                                                          |
| deprecated   |                                            `boolean`                                            | Declares this operation to be deprecated. Consumers SHOULD refrain from usage of the declared operation. Default value is `false`.                                                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| security     |        [[Security Requirement Object](security_objects.md#security-requirement-object)]         | A declaration of which security mechanisms can be used for this operation. The list of values includes alternative security requirement objects that can be used. Only one of the security requirement objects need to be satisfied to authorize a request. To make security optional, an empty security requirement (`{}`) can be included in the array. This definition overrides any declared top-level `security`. To remove a top-level security declaration, an empty array can be used.                                                                                                            |
| servers      |                               [[Server Object](server_object.md)]                               | An alternative `servers` array to service this operation. If an alternative `servers` array is specified at the Path Item Object or root level, it will be overridden by this value.                                                                                                                                                                                                                                                                                                                                                                                                                      |

This object MAY be extended with Specification Extensions.

## Operation Object Example

```yaml
tags:
  - pet
summary: Updates a pet in the store with form data
operationId: updatePetWithForm
parameters:
  - name: petId
    in: path
    description: ID of pet that needs to be updated
    required: true
    schema:
      type: string
requestBody:
  content:
    application/x-www-form-urlencoded:
      schema:
        type: object
        properties:
          name:
            description: Updated name of the pet
            type: string
          status:
            description: Updated status of the pet
            type: string
        required:
          - status
responses:
  "200":
    description: Pet updated.
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/Pet"
      application/xml:
        schema:
          $ref: "#/components/schemas/Pet"
  "405":
    description: Method Not Allowed
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/Error"
      application/xml:
        schema:
          $ref: "#/components/schemas/Error"
security:
  - petstore_auth:
      - write:pets
      - read:pets
deprecated: true
```

## Operation ID Best Practices

The `operationId` should:

1. Be unique across all operations in the API
2. Follow common programming naming conventions (e.g., camelCase, snake_case)
3. Be descriptive of the operation's purpose
4. Remain stable across API versions when the operation semantics haven't changed

### Common Naming Patterns

```yaml
# Resource-based naming
operationId: getUsers
operationId: getUserById
operationId: createUser
operationId: updateUser
operationId: deleteUser

# Action-based naming
operationId: uploadFile
operationId: processPayment
operationId: calculateTax
```

## Parameter Inheritance

Parameters can be defined at multiple levels:

1. **Path Item level**: Apply to all operations on that path
2. **Operation level**: Apply only to that specific operation

Operation-level parameters override Path Item-level parameters with the same name and location.

### Example

```yaml
paths:
  /users/{id}:
    parameters:
      - name: id
        in: path
        required: true
        schema:
          type: integer
    get:
      operationId: getUserById
      # Inherits the 'id' parameter
    put:
      operationId: updateUser
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string # Overrides the integer type
      # Uses the overridden 'id' parameter
```

## Security Override

The `security` field at the operation level can:

1. **Override** top-level security requirements
2. **Add** additional security requirements
3. **Remove** security requirements (using an empty array)

### Examples

```yaml
# Top-level security
security:
  - api_key: []

paths:
  /public:
    get:
      security: [] # No security required

  /admin:
    get:
      security:
        - oauth2:
            - admin:read
            - admin:write
        # Overrides top-level, requires OAuth

  /mixed:
    get:
      security:
        - api_key: []
        - oauth2:
            - read
        # Either API key OR OAuth is acceptable
```

## See Also

- [Paths Object](paths_object.md) - Container for path items
- [Parameter Object](parameter_object.md) - Parameter definitions
- [Request Body Object](request_body_object.md) - Request body definitions
- [Responses Object](responses_object.md) - Response definitions
- [Security Objects](security_objects.md) - Security scheme definitions
