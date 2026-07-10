# Header Object

The Header Object follows the structure of the [Parameter Object](parameter_object.md) with the following changes:

1. `name` MUST NOT be specified, it is given in the corresponding `headers` map
2. `in` MUST NOT be specified, it is implicitly in `header`
3. All traits that are affected by the location MUST be applicable to a location of `header` (for example, `style`)

## Fixed Fields

| Field Name  |                                             Type                                              | Description                                                                                                                                                                                                                                                                                                                 |
| ----------- | :-------------------------------------------------------------------------------------------: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| description |                                           `string`                                            | A brief description of the parameter. This could contain examples of use. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                                       |
| required    |                                           `boolean`                                           | Determines whether this parameter is mandatory. Default value is `false`.                                                                                                                                                                                                                                                   |
| deprecated  |                                           `boolean`                                           | Specifies that a parameter is deprecated and SHOULD be transitioned out of usage. Default value is `false`.                                                                                                                                                                                                                 |
| style       |                                           `string`                                            | Describes how the parameter value will be serialized. For `header`, the default value is `simple`.                                                                                                                                                                                                                          |
| explode     |                                           `boolean`                                           | When this is true, parameter values of type `array` or `object` generate separate parameters for each value of the array, or key-value-pair of the map. For other types of parameters this property has no effect. When `style` is `form`, the default value is `true`. For all other styles, the default value is `false`. |
| schema      |         [Schema Object](schema_object.md) \| [Reference Object](reference_object.md)          | The schema defining the type used for the parameter.                                                                                                                                                                                                                                                                        |
| example     |                                              Any                                              | Example of the parameter's potential value.                                                                                                                                                                                                                                                                                 |
| examples    | Map[`string`, [Example Object](example_object.md) \| [Reference Object](reference_object.md)] | Examples of the parameter's potential value.                                                                                                                                                                                                                                                                                |
| content     |                   Map[`string`, [Media Type Object](media_type_object.md)]                    | A map containing the representations for the parameter.                                                                                                                                                                                                                                                                     |

This object MAY be extended with Specification Extensions.

## Header Object Examples

### Simple Header

```yaml
description: The number of allowed requests in the current period
schema:
  type: integer
```

### Header with Format

```yaml
description: The rate limit reset time
schema:
  type: string
  format: date-time
```

### Header with Example

```yaml
description: A unique request identifier
schema:
  type: string
  format: uuid
example: 550e8400-e29b-41d4-a716-446655440000
```

### Header with Multiple Examples

```yaml
description: Content type of the response
schema:
  type: string
examples:
  json:
    value: application/json
  xml:
    value: application/xml
```

### Deprecated Header

```yaml
description: Legacy header, use X-Request-ID instead
deprecated: true
schema:
  type: string
```

### Array Header

```yaml
description: List of allowed methods
style: simple
explode: false
schema:
  type: array
  items:
    type: string
example: GET, POST, PUT
```

## Common Headers

### Rate Limiting Headers

```yaml
headers:
  X-RateLimit-Limit:
    description: Request limit per hour
    schema:
      type: integer
  X-RateLimit-Remaining:
    description: Remaining requests in the current period
    schema:
      type: integer
  X-RateLimit-Reset:
    description: Seconds until rate limit resets
    schema:
      type: integer
```

### Pagination Headers

```yaml
headers:
  X-Total-Count:
    description: Total number of items
    schema:
      type: integer
  X-Page-Count:
    description: Total number of pages
    schema:
      type: integer
  Link:
    description: Pagination links
    schema:
      type: string
    example: '<https://api.example.com/users?page=2>; rel="next", <https://api.example.com/users?page=5>; rel="last"'
```

### Content Headers

```yaml
headers:
  Content-Type:
    description: Media type of the response body
    schema:
      type: string
  Content-Length:
    description: Size of the response body in bytes
    schema:
      type: integer
  Content-Disposition:
    description: Suggested filename for download
    schema:
      type: string
    example: 'attachment; filename="report.pdf"'
```

### Caching Headers

```yaml
headers:
  ETag:
    description: Entity tag for cache validation
    schema:
      type: string
  Cache-Control:
    description: Caching directives
    schema:
      type: string
    example: "max-age=3600, public"
  Last-Modified:
    description: Last modification time
    schema:
      type: string
      format: date-time
```

## Headers in Responses

Headers are defined in the [Response Object](responses_object.md#response-object):

```yaml
responses:
  "200":
    description: Success
    headers:
      X-Request-ID:
        description: Unique request identifier
        schema:
          type: string
          format: uuid
      X-RateLimit-Remaining:
        description: Remaining requests
        schema:
          type: integer
    content:
      application/json:
        schema:
          $ref: "#/components/schemas/User"
```

## Headers in Multipart Encoding

Headers can be specified for multipart form data parts:

```yaml
requestBody:
  content:
    multipart/form-data:
      schema:
        type: object
        properties:
          file:
            type: string
            format: binary
      encoding:
        file:
          contentType: application/octet-stream
          headers:
            Content-Disposition:
              description: File disposition
              schema:
                type: string
```

## Reusable Headers

Headers can be defined in the [Components Object](components_object.md):

```yaml
components:
  headers:
    X-Request-ID:
      description: Unique request identifier
      schema:
        type: string
        format: uuid
    X-RateLimit-Limit:
      description: Request limit per hour
      schema:
        type: integer

paths:
  /users:
    get:
      responses:
        "200":
          headers:
            X-Request-ID:
              $ref: "#/components/headers/X-Request-ID"
            X-RateLimit-Limit:
              $ref: "#/components/headers/X-RateLimit-Limit"
```

## Header Serialization

Headers use the `simple` style by default:

### Simple Style (Default)

```yaml
# Single value
X-Custom-Header: value

# Array
X-Custom-Header: value1,value2,value3

# Object
X-Custom-Header: name,John,age,30
```

### With Explode

```yaml
style: simple
explode: true
# Object with explode
X-Custom-Header: name=John,age=30
```

## Ignored Headers

The following headers are ignored when defined in `headers`:

- `Content-Type` - Defined in the `content` field instead
- `Accept` - Defined in the request
- `Authorization` - Defined in [Security Scheme Object](security_objects.md#security-scheme-object)

## See Also

- [Parameter Object](parameter_object.md) - Similar structure for parameters
- [Responses Object](responses_object.md) - Defining response headers
- [Media Type Object](media_type_object.md) - Headers in multipart encoding
- [Components Object](components_object.md) - Reusable headers
