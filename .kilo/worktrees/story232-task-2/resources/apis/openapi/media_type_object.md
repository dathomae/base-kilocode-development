# Media Type Object

Each Media Type Object provides schema and examples for the media type identified by its key.

## Fixed Fields

| Field Name |                                             Type                                              | Description                                                                                                                                                                                                                                                                                                              |
| ---------- | :-------------------------------------------------------------------------------------------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| schema     |         [Schema Object](schema_object.md) \| [Reference Object](reference_object.md)          | The schema defining the content of the request, response, or parameter.                                                                                                                                                                                                                                                  |
| example    |                                              Any                                              | Example of the media type. The example object SHOULD be in the correct format as specified by the media type. The `example` field is mutually exclusive of the `examples` field. Furthermore, if referencing a schema that contains an example, the `example` value SHALL _override_ the example provided by the schema. |
| examples   | Map[`string`, [Example Object](example_object.md) \| [Reference Object](reference_object.md)] | Examples of the media type. Each example object SHOULD match the media type and specified schema if present. The `examples` field is mutually exclusive of the `example` field. Furthermore, if referencing a schema that contains an example, the `examples` value SHALL _override_ the example provided by the schema. |
| encoding   |                      Map[`string`, [Encoding Object](#encoding-object)]                       | A map between a property name and its encoding information. The key, being the property name, MUST exist in the schema as a property. The `encoding` object SHALL only apply to `requestBody` objects when the media type is `multipart` or `application/x-www-form-urlencoded`.                                         |

This object MAY be extended with Specification Extensions.

## Media Type Object Examples

### JSON Response

```yaml
application/json:
  schema:
    $ref: "#/components/schemas/Pet"
  examples:
    dog:
      summary: An example of a dog
      value:
        name: Fluffy
        petType: Dog
    cat:
      summary: An example of a cat
      value:
        name: Mittens
        petType: Cat
```

### XML Response

```yaml
application/xml:
  schema:
    $ref: "#/components/schemas/Pet"
  example: |
    <Pet>
      <name>Fluffy</name>
      <petType>Dog</petType>
    </Pet>
```

### Binary File

```yaml
application/octet-stream:
  schema:
    type: string
    format: binary
```

### Plain Text

```yaml
text/plain:
  schema:
    type: string
  example: "Hello, World!"
```

## Encoding Object

Single encoding definition applied to a single schema property.

### Fixed Fields

| Field Name    |                                            Type                                             | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| ------------- | :-----------------------------------------------------------------------------------------: | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| contentType   |                                          `string`                                           | The Content-Type for encoding a specific property. Default value depends on the property type: for `string` with `format` being `binary` – `application/octet-stream`; for other primitive types – `text/plain`; for `object` - `application/json`; for `array` – the default is defined based on the inner type. The value can be a specific media type (e.g. `application/json`), a wildcard media type (e.g. `image/*`), or a comma-separated list of the two types. |
| headers       | Map[`string`, [Header Object](header_object.md) \| [Reference Object](reference_object.md)] | A map allowing additional information to be provided as headers, for example `Content-Disposition`. `Content-Type` is described separately and SHALL be ignored in this section. This property SHALL be ignored if the request body media type is not a `multipart`.                                                                                                                                                                                                    |
| style         |                                          `string`                                           | Describes how a specific property value will be serialized depending on its type. See [Parameter Object](parameter_object.md) for details on the `style` property. The behavior follows the same values as `query` parameters, including default values. This property SHALL be ignored if the request body media type is not `application/x-www-form-urlencoded`.                                                                                                      |
| explode       |                                          `boolean`                                          | When this is true, property values of type `array` or `object` generate separate parameters for each value of the array, or key-value-pair of the map. For other types of properties this property has no effect. When `style` is `form`, the default value is `true`. For all other styles, the default value is `false`. This property SHALL be ignored if the request body media type is not `application/x-www-form-urlencoded`.                                    |
| allowReserved |                                          `boolean`                                          | Determines whether the parameter value SHOULD allow reserved characters, as defined by [RFC3986](https://tools.ietf.org/html/rfc3986#section-2.2) `:/?#[]@!$&'()*+,;=` to be included without percent-encoding. The default value is `false`. This property SHALL be ignored if the request body media type is not `application/x-www-form-urlencoded`.                                                                                                                 |

This object MAY be extended with Specification Extensions.

### Encoding Object Example

```yaml
requestBody:
  content:
    multipart/form-data:
      schema:
        type: object
        properties:
          id:
            type: string
            format: uuid
          address:
            $ref: "#/components/schemas/Address"
          historyMetadata:
            description: metadata in XML format
            type: object
          profileImage:
            type: string
            format: binary
      encoding:
        historyMetadata:
          contentType: application/xml
        profileImage:
          contentType: image/png, image/jpeg
          headers:
            X-Custom-Header:
              description: Custom header for profile image
              schema:
                type: string
```

## Form URL-Encoded Data

For `application/x-www-form-urlencoded` request bodies:

```yaml
requestBody:
  content:
    application/x-www-form-urlencoded:
      schema:
        type: object
        properties:
          name:
            type: string
          colors:
            type: array
            items:
              type: string
      encoding:
        colors:
          style: form
          explode: false
          # Serializes as: colors=red,green,blue
```

## Multipart Form Data

For `multipart/form-data` request bodies, each property can have its own content type:

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
          contentType: application/pdf
        metadata:
          contentType: application/json
```

## Content Type Selection

When multiple content types are specified, clients and servers negotiate the format:

### Media Type Ranges

Media type ranges allow matching multiple content types:

```yaml
content:
  image/*:
    schema:
      type: string
      format: binary
  text/*:
    schema:
      type: string
  application/*:
    schema:
      type: object
```

### Specificity Rules

1. Exact match takes precedence over wildcard
2. More specific wildcards take precedence over less specific

```yaml
content:
  application/json:
    schema:
      $ref: "#/components/schemas/Pet"
  application/*:
    schema:
      type: string
      format: binary
  "*/*":
    schema:
      type: string
```

In this example:

- `application/json` matches JSON specifically
- `application/*` matches any other application type
- `*/*` matches any other content type

## See Also

- [Schema Object](schema_object.md) - Schema definitions
- [Example Object](example_object.md) - Example definitions
- [Header Object](header_object.md) - Header definitions for multipart
- [Parameter Object](parameter_object.md) - Serialization style reference
- [Request Body Object](request_body_object.md) - Request body definitions
- [Responses Object](responses_object.md) - Response definitions
