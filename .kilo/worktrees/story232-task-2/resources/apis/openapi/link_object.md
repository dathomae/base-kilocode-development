# Link Object

The Link Object represents a possible design-time link for a response. The presence of a link does not guarantee the caller's ability to successfully invoke it, rather it provides a known relationship and traversal mechanism between responses and other operations.

Unlike dynamic links (i.e. links provided in the response payload), the OAS linking mechanism does not require link information in the runtime response.

For computing links, and providing instructions to execute them, a [runtime expression](#runtime-expression) is used for accessing values in an operation and using them as parameters while invoking the linked operation.

## Fixed Fields

| Field Name   |                              Type                               | Description                                                                                                                                                                                                                                                                                                                                                                                                                                        |
| ------------ | :-------------------------------------------------------------: | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| operationRef |                            `string`                             | A relative or absolute URI reference to an OAS operation. This field is mutually exclusive of the `operationId` field, and MUST point to an [Operation Object](operation_object.md). Relative `operationRef` values MAY be used to locate an existing [Operation Object](operation_object.md) in the OAS. See the rules for resolving Relative References.                                                                                         |
| operationId  |                            `string`                             | The name of an existing, resolvable OAS operation, as defined with a unique `operationId`. This field is mutually exclusive of the `operationRef` field.                                                                                                                                                                                                                                                                                           |
| parameters   | Map[`string`, Any \| [Runtime Expression](#runtime-expression)] | A map representing parameters to pass to an operation as specified with `operationId` or identified via `operationRef`. The key is the parameter name to be used, whereas the value can be a constant or an expression to be evaluated and passed to the linked operation. The parameter name can be qualified using the parameter location `[{in}.]{name}` for operations that use the same parameter name in different locations (e.g. path.id). |
| requestBody  |        Any \| [Runtime Expression](#runtime-expression)         | A literal value or `{expression}` to use as a request body when calling the target operation.                                                                                                                                                                                                                                                                                                                                                      |
| description  |                            `string`                             | A description of the link. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                                                                                                                                                                                                             |
| server       |                [Server Object](server_object.md)                | A server object to be used by the target operation.                                                                                                                                                                                                                                                                                                                                                                                                |

This object MAY be extended with Specification Extensions.

## Runtime Expression

Runtime expressions are used to access values from the request/response context.

### Expression Syntax

```
expression = ( "$url" | "$method" | "$statusCode" | "$request." reference | "$response." reference )
reference = header-reference | query-reference | path-reference | body-reference
header-reference = "header." token
query-reference = "query." token
path-reference = "path." token
body-reference = "body" ["#" json-pointer]
```

### Expression Variables

| Expression                 | Description                                   |
| -------------------------- | --------------------------------------------- |
| `$url`                     | The full request URL                          |
| `$method`                  | The HTTP method of the request                |
| `$statusCode`              | The status code of the response               |
| `$request.header.{name}`   | A request header value                        |
| `$request.query.{name}`    | A query parameter value                       |
| `$request.path.{name}`     | A path parameter value                        |
| `$request.body`            | The entire request body                       |
| `$request.body#{pointer}`  | A value from the request body (JSON Pointer)  |
| `$response.header.{name}`  | A response header value                       |
| `$response.body`           | The entire response body                      |
| `$response.body#{pointer}` | A value from the response body (JSON Pointer) |

## Link Object Examples

### Using operationId

```yaml
paths:
  /users/{id}:
    get:
      operationId: getUserById
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: integer
      responses:
        "200":
          description: A user
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
          links:
            userRepositories:
              operationId: getRepositoriesByUser
              parameters:
                username: $response.body#/username
```

### Using operationRef

```yaml
links:
  userRepositories:
    operationRef: "#/paths/~1repositories~1{username}/get"
    parameters:
      username: $response.body#/username
```

### With Request Body

```yaml
links:
  updateUser:
    operationId: updateUser
    requestBody: $response.body#/id
```

### With Description

```yaml
links:
  address:
    operationId: getAddress
    parameters:
      userId: $response.body#/id
    description: Get the user's address
```

### With Server Override

```yaml
links:
  externalService:
    operationRef: "https://api.example.com/operations#getStatus"
    parameters:
      id: $response.body#/transactionId
    server:
      url: https://api.example.com
      description: External API server
```

## Complete Example

```yaml
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: User found
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
          links:
            orders:
              operationId: getOrders
              parameters:
                userId: $request.path.userId
            address:
              operationId: getAddress
              parameters:
                userId: $response.body#/id
              description: Get user's address
            manager:
              operationRef: "#/paths/~1users~1{managerId}/get"
              parameters:
                managerId: $response.body#/managerId

  /users/{userId}/orders:
    get:
      operationId: getOrders
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: Orders for user

  /users/{userId}/address:
    get:
      operationId: getAddress
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: User's address
```

## Parameter Passing

Links can pass parameters in several ways:

### Constant Values

```yaml
parameters:
  status: active
```

### From Request Path

```yaml
parameters:
  userId: $request.path.userId
```

### From Request Body

```yaml
parameters:
  parentId: $request.body#/parentId
```

### From Response Body

```yaml
parameters:
  userId: $response.body#/id
```

### From Response Headers

```yaml
parameters:
  requestId: $response.header.X-Request-ID
```

## Resolving operationRef

The `operationRef` uses JSON Pointer syntax to reference operations:

### Path Escaping

Special characters in paths must be escaped:

| Character | Escaped Form |
| --------- | ------------ |
| `~`       | `~0`         |
| `/`       | `~1`         |

### Examples

```yaml
# /users/{id}
operationRef: '#/paths/~1users~1{id}/get'

# /users/{id}/orders
operationRef: '#/paths/~1users~1{id}~1orders/get'

# External reference
operationRef: 'https://api.example.com/openapi.yaml#/paths/~1users/get'
```

## Links in Responses

Links are defined in the [Response Object](responses_object.md#response-object):

```yaml
responses:
  "200":
    description: Success
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/User"
    links:
      self:
        operationId: getUser
        parameters:
          userId: $response.body#/id
      orders:
        operationId: getOrders
        parameters:
          userId: $response.body#/id
```

## Reusable Links

Links can be defined in the [Components Object](components_object.md):

```yaml
components:
  links:
    UserOrders:
      operationId: getOrders
      parameters:
        userId: $response.body#/id
      description: Get orders for this user

paths:
  /users/{userId}:
    get:
      responses:
        "200":
          links:
            orders:
              $ref: "#/components/links/UserOrders"
```

## See Also

- [Responses Object](responses_object.md) - Defining links in responses
- [Operation Object](operation_object.md) - Target operations
- [Components Object](components_object.md) - Reusable links
