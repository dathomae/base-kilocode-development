# Callback Object

A map of possible out-of band callbacks related to the parent operation. Each value in the map is a Path Item Object that describes a set of requests that may be initiated by the API provider and the expected responses. The key is used to identify the callback object.

## Patterned Fields

| Field Pattern |                         Type                         | Description                                                                                                                               |
| ------------- | :--------------------------------------------------: | ----------------------------------------------------------------------------------------------------------------------------------------- |
| {expression}  | [Path Item Object](paths_object.md#path-item-object) | A Path Item Object used to define a callback request and expected responses. A [complete example](#callback-object-example) is available. |

This object MAY be extended with Specification Extensions.

## Runtime Expression

The key that identifies the Path Item Object is a runtime expression that can be evaluated in the context of a runtime request to identify the URL to use for the callback request.

### Expression Syntax

A runtime expression is defined by the following grammar:

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

## Callback Object Example

```yaml
myCallback:
  "{$request.body#/callbackUrl}":
    post:
      requestBody:
        description: Callback payload
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/SomePayload"
      responses:
        "200":
          description: Callback successfully processed
```

## Using Callbacks in Operations

Callbacks are defined in the [Operation Object](operation_object.md):

```yaml
paths:
  /subscribe:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                callbackUrl:
                  type: string
                  format: uri
      responses:
        "201":
          description: Subscription created
      callbacks:
        onData:
          "{$request.body#/callbackUrl}":
            post:
              requestBody:
                description: Webhook payload
                content:
                  application/json:
                    schema:
                      type: object
                      properties:
                        event:
                          type: string
                        data:
                          type: object
              responses:
                "200":
                  description: Webhook processed
```

## Multiple Callback Events

A single callback can define multiple events:

```yaml
callbacks:
  notifications:
    "{$request.body#/webhookUrl}/events":
      post:
        requestBody:
          content:
            application/json:
              schema:
                oneOf:
                  - $ref: "#/components/schemas/UserCreatedEvent"
                  - $ref: "#/components/schemas/UserUpdatedEvent"
                  - $ref: "#/components/schemas/UserDeletedEvent"
        responses:
          "200":
            description: Event acknowledged
```

## Reusable Callbacks

Callbacks can be defined in the [Components Object](components_object.md) and referenced:

```yaml
components:
  callbacks:
    transactionCallback:
      "{$request.body#/callbackUrl}":
        post:
          requestBody:
            content:
              application/json:
                schema:
                  $ref: "#/components/schemas/TransactionEvent"
          responses:
            "200":
              description: Callback received

paths:
  /transactions:
    post:
      callbacks:
        transactionEvent:
          $ref: "#/components/callbacks/transactionCallback"
```

## Expression Examples

### Using Request Headers

```yaml
callbacks:
  myCallback:
    "https://{$request.header.host}/callback":
      post:
        requestBody:
          content:
            application/json:
              schema:
                type: object
        responses:
          "200":
            description: OK
```

### Using Response Body

```yaml
callbacks:
  webhookCallback:
    "{$response.body#/webhookUrl}":
      post:
        requestBody:
          content:
            application/json:
              schema:
                type: object
        responses:
          "200":
            description: OK
```

### Using Path Parameters

```yaml
paths:
  /users/{id}:
    post:
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
      callbacks:
        userCallback:
          "https://example.com/callbacks/{$request.path.id}":
            post:
              requestBody:
                content:
                  application/json:
                    schema:
                      type: object
              responses:
                "200":
                  description: OK
```

## Callbacks vs Webhooks

While similar, callbacks and webhooks serve different purposes:

### Callbacks

- Defined within an operation
- URL is determined at runtime from request/response data
- Used for asynchronous responses to specific requests

### Webhooks

- Defined at the root level using the `webhooks` field
- URL is typically pre-configured
- Used for event notifications independent of specific requests

```yaml
webhooks:
  userCreated:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/UserCreatedEvent"
      responses:
        "200":
          description: Webhook processed
```

## See Also

- [Operation Object](operation_object.md) - Defining callbacks in operations
- [Path Item Object](paths_object.md#path-item-object) - Callback request definitions
- [Components Object](components_object.md) - Reusable callbacks
