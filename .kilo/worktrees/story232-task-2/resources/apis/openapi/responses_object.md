# Responses Object

A container for the expected responses of an operation. The container maps a HTTP response code to the expected response.

The documentation is not necessarily expected to cover all possible HTTP response codes because they may not be known in advance. However, documentation is expected to cover a successful operation response and any known errors.

The `default` MAY be used as a default response object for all HTTP codes that are not covered individually by the specification.

The `Responses Object` MUST contain at least one response code, and if only one response code is provided it SHOULD be the `default` response.

## Patterned Fields

| Field Pattern             |                                      Type                                      | Description                                                                                                                                                                                                                                                                                                                                    |
| ------------------------- | :----------------------------------------------------------------------------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| default                   | [Response Object](#response-object) \| [Reference Object](reference_object.md) | The documentation of responses other than the ones declared for specific HTTP response codes. Use this field to cover undeclared responses.                                                                                                                                                                                                    |
| {HTTP Status Code}        | [Response Object](#response-object) \| [Reference Object](reference_object.md) | Any [HTTP status code](https://www.iana.org/assignments/http-status-codes/http-status-codes.xhtml) can be used as the property name, but only one property per code, to describe the expected response for that HTTP status code. This field MUST be enclosed in quotation marks (for example, "200") for compatibility between JSON and YAML. |
| {1XX, 2XX, 3XX, 4XX, 5XX} | [Response Object](#response-object) \| [Reference Object](reference_object.md) | Any HTTP status code range can be used as the property name to cover all response codes in that range. Only one range per status code range, to describe the expected response for that HTTP status code range. This field MUST be enclosed in quotation marks (for example, "4XX") for compatibility between JSON and YAML.                   |

This object MAY be extended with Specification Extensions.

## Responses Object Example

```yaml
"200":
  description: a pet to be returned
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/Pet"
default:
  description: Unexpected error
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/ErrorModel"
```

## Response Object

Describes a single response from an API Operation, including design-time, static `links` to operations based on the response.

### Fixed Fields

| Field Name  |                                            Type                                             | Description                                                                                                                                                                                                                                                                                                    |
| ----------- | :-----------------------------------------------------------------------------------------: | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| description |                                          `string`                                           | **REQUIRED**. A description of the response. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                                                       |
| headers     | Map[`string`, [Header Object](header_object.md) \| [Reference Object](reference_object.md)] | Maps a header name to its definition. [RFC7230](https://tools.ietf.org/html/rfc7230#page-22) states header names are case insensitive. If a response header is defined with the name `"Content-Type"`, it SHALL be ignored.                                                                                    |
| content     |                  Map[`string`, [Media Type Object](media_type_object.md)]                   | A map containing descriptions of potential response payloads. The key is a media type or [media type range](https://tools.ietf.org/html/rfc7231#appendix-D) and the value describes it. For responses that match multiple keys, only the most specific key is applicable. e.g. `text/plain` overrides `text/*` |
| links       |   Map[`string`, [Link Object](link_object.md) \| [Reference Object](reference_object.md)]   | A map of operations links that can be followed from the response. The key of the map is a short name for the link, following the naming constraints of the names for [Components Objects](components_object.md).                                                                                               |

This object MAY be extended with Specification Extensions.

### Response Object Examples

#### Simple Response

```yaml
description: A simple string response
```

#### Response with Headers

```yaml
description: A simple string response
headers:
  X-Rate-Limit-Limit:
    description: The number of allowed requests in the current period
    schema:
      type: integer
  X-Rate-Limit-Remaining:
    description: The number of remaining requests in the current period
    schema:
      type: integer
  X-Rate-Limit-Reset:
    description: The number of seconds left in the current period
    schema:
      type: integer
```

#### Response with Content

```yaml
description: A pet to be returned
content:
  application/json:
    schema:
      $ref: "#/components/schemas/Pet"
  application/xml:
    schema:
      $ref: "#/components/schemas/Pet"
```

#### Response with Links

```yaml
description: A user object
content:
  application/json:
    schema:
      $ref: "#/components/schemas/User"
links:
  userRepositories:
    operationId: getRepositoriesByUser
    parameters:
      username: $response.body#/username
  userFollowers:
    operationId: getFollowersByUser
    parameters:
      username: $response.body#/username
```

## Status Code Ranges

Using status code ranges allows covering multiple response codes with a single response definition:

```yaml
responses:
  "2XX":
    description: Successful response
    content:
      application/json:
        schema:
          type: object
  "4XX":
    description: Client error
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/Error"
  "5XX":
    description: Server error
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/Error"
```

### Range Matching Priority

When matching responses:

1. Exact status code match takes precedence
2. Status code range match is used if no exact match
3. `default` is used if no other match

```yaml
responses:
  "200":
    description: Success
  "4XX":
    description: Client error
  "404":
    description: Not found # This overrides 4XX for 404
  default:
    description: Any other response
```

## Media Type Matching

Response content uses media type matching similar to request bodies:

```yaml
responses:
  "200":
    description: Success
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/Pet"
      application/xml:
        schema:
          $ref: "#/components/schemas/Pet"
      text/html:
        schema:
          type: string
      "*/*":
        schema:
          type: string
```

## Reusable Responses

Responses can be defined in the [Components Object](components_object.md) and referenced:

```yaml
components:
  responses:
    NotFound:
      description: The specified resource was not found
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"
    Unauthorized:
      description: Unauthorized
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/Error"

paths:
  /users/{id}:
    get:
      responses:
        "200":
          description: A user
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/User"
        "401":
          $ref: "#/components/responses/Unauthorized"
        "404":
          $ref: "#/components/responses/NotFound"
```

## See Also

- [Media Type Object](media_type_object.md) - Media type definitions
- [Header Object](header_object.md) - Response header definitions
- [Link Object](link_object.md) - Response links
- [Components Object](components_object.md) - Reusable components
