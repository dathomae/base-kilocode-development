# Request Body Object

Describes a single request body.

## Fixed Fields

| Field Name  |                           Type                           | Description                                                                                                                                                                                                                                                                                    |
| ----------- | :------------------------------------------------------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| description |                         `string`                         | A brief description of the request body. This could contain examples of use. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                       |
| content     | Map[`string`, [Media Type Object](media_type_object.md)] | **REQUIRED**. The content of the request body. The key is a media type or [media type range](https://tools.ietf.org/html/rfc7231#appendix-D) and the value describes it. For requests that match multiple keys, only the most specific key is applicable. e.g. `text/plain` overrides `text/*` |
| required    |                        `boolean`                         | Determines if the request body is required in the request. Defaults to `false`.                                                                                                                                                                                                                |

This object MAY be extended with Specification Extensions.

## Request Body Example

```yaml
description: user to add to the system
content:
  application/json:
    schema:
      $ref: "#/components/schemas/User"
    examples:
      user:
        summary: User Example
        externalValue: "http://foo.bar/examples/user-example.json"
  application/xml:
    schema:
      $ref: "#/components/schemas/User"
    examples:
      user:
        summary: User Example in XML
        externalValue: "http://foo.bar/examples/user-example.xml"
  text/plain:
    examples:
      user:
        summary: User Example in Plain Text
        externalValue: "http://foo.bar/examples/user-example.txt"
  "*/*":
    examples:
      user:
        summary: User Example
        externalValue: "http://foo.bar/examples/user-example.json"
```

## Media Type Selection

When multiple media types are specified, the client and server negotiate the content type. The `content` map uses media type keys that follow these rules:

### Media Type Matching

Media types are matched from most specific to least specific:

1. Exact match: `application/json`
2. Wildcard subtype: `application/*`
3. Wildcard type: `*/*`

### Example

```yaml
requestBody:
  content:
    application/json:
      schema:
        $ref: "#/components/schemas/Pet"
    application/xml:
      schema:
        $ref: "#/components/schemas/Pet"
    text/plain:
      schema:
        type: string
    "*/*":
      schema:
        type: string
```

In this example:

- A request with `Content-Type: application/json` uses the JSON schema
- A request with `Content-Type: application/xml` uses the XML schema
- A request with `Content-Type: text/plain` uses the string schema
- Any other content type falls back to `*/*`

## Form Data

Request bodies can describe form data submissions:

### URL-Encoded Form

```yaml
requestBody:
  content:
    application/x-www-form-urlencoded:
      schema:
        type: object
        properties:
          name:
            type: string
          email:
            type: string
            format: email
        required:
          - name
          - email
```

### Multipart Form Data

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
          metadata:
            type: object
            properties:
              name:
                type: string
              description:
                type: string
      encoding:
        file:
          contentType: application/octet-stream
        metadata:
          contentType: application/json
```

## File Upload

For file uploads, use the `binary` format:

```yaml
requestBody:
  content:
    application/octet-stream:
      schema:
        type: string
        format: binary
```

### Multiple File Upload

```yaml
requestBody:
  content:
    multipart/form-data:
      schema:
        type: object
        properties:
          files:
            type: array
            items:
              type: string
              format: binary
```

## Reusable Request Bodies

Request bodies can be defined in the [Components Object](components_object.md) and referenced:

```yaml
components:
  requestBodies:
    UserBody:
      description: A user object
      required: true
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/User"

paths:
  /users:
    post:
      requestBody:
        $ref: "#/components/requestBodies/UserBody"
```

## HTTP Methods and Request Bodies

The HTTP 1.1 specification [RFC7231](https://tools.ietf.org/html/rfc7231#section-4.3.1) explicitly defines semantics for request bodies for:

- POST
- PUT
- PATCH

For other methods (GET, HEAD, DELETE, OPTIONS), request bodies are permitted but have undefined semantics. It is RECOMMENDED to avoid request bodies with these methods.

### Example with GET (Not Recommended)

```yaml
paths:
  /search:
    get:
      requestBody:
        description: Search parameters (not recommended for GET)
        content:
          application/json:
            schema:
              type: object
              properties:
                query:
                  type: string
                filters:
                  type: array
                  items:
                    type: string
```

## See Also

- [Media Type Object](media_type_object.md) - Media type definitions
- [Schema Object](schema_object.md) - Schema definitions
- [Components Object](components_object.md) - Reusable components
- [Operation Object](operation_object.md) - Operation definitions
