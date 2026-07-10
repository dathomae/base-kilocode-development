# Parameter Object

Describes a single operation parameter.

A unique parameter is defined by a combination of a [name](#parameter-name) and [location](#parameter-in).

## Parameter Locations

Parameters can be specified in several locations, identified by the `in` field:

| Location | Description                                                                                                                                                                                                                                                          |
| -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `query`  | Parameters that are appended to the URL. For example, in `/items?id=###`, the query parameter is `id`.                                                                                                                                                               |
| `header` | Custom headers that are expected as part of the request. Note that [RFC7230](https://tools.ietf.org/html/rfc7230#page-22) states header names are case insensitive.                                                                                                  |
| `path`   | Used together with [Path Templating](paths_object.md#path-templating), where the parameter value is actually part of the operation's URL. This does not include the host or base path of the API. For example, in `/items/{itemId}`, the path parameter is `itemId`. |
| `cookie` | Used to pass a specific cookie value to the API.                                                                                                                                                                                                                     |

## Fixed Fields

| Field Name      |   Type    | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| --------------- | :-------: | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| name            | `string`  | **REQUIRED**. The name of the parameter. A parameter name is case-sensitive. If `in` is `"path"`, the `name` field MUST correspond to a template expression occurring within the path in the [Paths Object](paths_object.md). If `in` is `"header"` and the `name` field is `"Accept"`, `"Content-Type"` or `"Authorization"`, the parameter definition SHALL be ignored. For all other cases, the `name` corresponds to the parameter name used by the `in` property. |
| in              | `string`  | **REQUIRED**. The location of the parameter. Possible values are `"query"`, `"header"`, `"path"` or `"cookie"`.                                                                                                                                                                                                                                                                                                                                                        |
| description     | `string`  | A brief description of the parameter. This could contain examples of use. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                                                                                                                                                                                  |
| required        | `boolean` | Determines whether this parameter is mandatory. If the parameter location is `"path"`, this property is **REQUIRED** and its value MUST be `true`. Otherwise, the property MAY be included and its default value is `false`.                                                                                                                                                                                                                                           |
| deprecated      | `boolean` | Specifies that a parameter is deprecated and SHOULD be transitioned out of usage. Default value is `false`.                                                                                                                                                                                                                                                                                                                                                            |
| allowEmptyValue | `boolean` | Sets the ability to pass empty-valued parameters. This is valid only for `query` parameters and allows sending a parameter with an empty value. Default value is `false`. If `style` is used, and if behavior is `n/a` (cannot be serialized), the `allowEmptyValue` behavior is `n/a` and the `allowEmptyValue` field SHOULD be ignored. **This field is deprecated.**                                                                                                |

## Serialization Fields

The serialization of parameters is controlled by the following fields:

| Field Name    |   Type    | Description                                                                                                                                                                                                                                                                                                                   |
| ------------- | :-------: | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| style         | `string`  | Describes how the parameter value will be serialized. For simple scenarios, default values SHOULD be used. See [Style Values](#style-values) for the default values and possible values for each `in` value.                                                                                                                  |
| explode       | `boolean` | When this is true, parameter values of type `array` or `object` generate separate parameters for each value of the array, or key-value-pair of the map. For other types of parameters this property has no effect. When `style` is `"form"`, the default value is `true`. For all other styles, the default value is `false`. |
| allowReserved | `boolean` | Determines whether the parameter value SHOULD allow reserved characters, as defined by [RFC3986](https://tools.ietf.org/html/rfc3986#section-2.2) `:/?#[]@!$&'()*+,;=` to be included without percent-encoding. This property only applies to parameters with an `in` value of `query`. The default value is `false`.         |

## Schema Fields

Parameters can define their structure using either a schema or content:

| Field Name |                                             Type                                              | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| ---------- | :-------------------------------------------------------------------------------------------: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| schema     |         [Schema Object](schema_object.md) \| [Reference Object](reference_object.md)          | The schema defining the type used for the parameter.                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| example    |                                              Any                                              | Example of the parameter's potential value. The example SHOULD match the specified schema and encoding properties if present. The `example` field is mutually exclusive of the `examples` field. Furthermore, if referencing a schema that contains an example, the `example` value SHALL _override_ the example provided by the schema. To represent examples of media types that cannot naturally be represented in JSON or YAML, a string value can contain the example with escaping where necessary. |
| examples   | Map[`string`, [Example Object](example_object.md) \| [Reference Object](reference_object.md)] | Examples of the parameter's potential value. Each example SHOULD contain a value in the correct format as specified in the parameter encoding. The `examples` field is mutually exclusive of the `example` field. Furthermore, if referencing a schema that contains an example, the `examples` value SHALL _override_ the example provided by the schema.                                                                                                                                                |

## Content Field

| Field Name |                           Type                           | Description                                                                                                                                                                                                                                                                                            |
| ---------- | :------------------------------------------------------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| content    | Map[`string`, [Media Type Object](media_type_object.md)] | A map containing the representations for the parameter. The key is the media type and the value describes it. The map MUST only contain one entry. For `in` values of `"path"`, `"query"`, `"header"`, and `"cookie"`, the `content` field is mutually exclusive with the `schema` and `style` fields. |

## Style Values

The `style` property determines how parameters are serialized:

### Query Parameters (`in: "query"`)

| Style            | Description                                                                 | Default |
| ---------------- | --------------------------------------------------------------------------- | :-----: |
| `form`           | Ampersand-separated values (e.g., `color=blue&color=red`)                   |    ✓    |
| `spaceDelimited` | Space-separated values (e.g., `color=blue red`)                             |         |
| `pipeDelimited`  | Pipe-separated values (e.g., `color=blue\|red`)                             |         |
| `deepObject`     | Nested objects using prop[value] syntax (e.g., `color[R]=100&color[G]=200`) |         |

### Path Parameters (`in: "path"`)

| Style    | Description                                               | Default |
| -------- | --------------------------------------------------------- | :-----: |
| `simple` | Comma-separated values (e.g., `blue,red`)                 |    ✓    |
| `label`  | Dot-prefixed values (e.g., `.blue.red`)                   |         |
| `matrix` | Semicolon-prefixed values (e.g., `;color=blue;color=red`) |         |

### Header Parameters (`in: "header"`)

| Style    | Description            | Default |
| -------- | ---------------------- | :-----: |
| `simple` | Comma-separated values |    ✓    |

### Cookie Parameters (`in: "cookie"`)

| Style  | Description                | Default |
| ------ | -------------------------- | :-----: |
| `form` | Ampersand-separated values |    ✓    |

## Parameter Object Examples

### Path Parameter

```yaml
name: username
in: path
description: username to fetch
required: true
schema:
  type: string
```

### Query Parameter with Array

```yaml
name: ids
in: query
description: A list of IDs
required: false
schema:
  type: array
  items:
    type: integer
style: form
explode: false
# Serializes as: ids=1,2,3
```

### Query Parameter with Explode

```yaml
name: color
in: query
schema:
  type: array
  items:
    type: string
style: form
explode: true
# Serializes as: color=blue&color=red&color=green
```

### Header Parameter

```yaml
name: X-Request-ID
in: header
description: Unique request identifier
required: false
schema:
  type: string
  format: uuid
```

### Cookie Parameter

```yaml
name: session_id
in: cookie
description: Session identifier
required: true
schema:
  type: string
```

### Content-Based Parameter

```yaml
name: filter
in: query
content:
  application/json:
    schema:
      type: object
      properties:
        status:
          type: string
        created_after:
          type: string
          format: date-time
```

### Deep Object Style

```yaml
name: filter
in: query
style: deepObject
explode: true
schema:
  type: object
  properties:
    status:
      type: string
    tags:
      type: array
      items:
        type: string
# Serializes as: filter[status]=active&filter[tags]=tag1&filter[tags]=tag2
```

## Serialization Examples

### Object Serialization

Given this schema:

```yaml
schema:
  type: object
  properties:
    firstName:
      type: string
    lastName:
      type: string
```

Different serialization styles produce:

| Style        | Explode | Result                                |
| ------------ | :-----: | ------------------------------------- |
| `form`       | `false` | `color=firstName,Alex,lastName,Smith` |
| `form`       | `true`  | `firstName=Alex&lastName=Smith`       |
| `deepObject` | `true`  | `firstName=Alex&lastName=Smith`       |

### Array Serialization

Given this schema:

```yaml
schema:
  type: array
  items:
    type: string
  example: [blue, black, brown]
```

Different serialization styles produce:

| Style            | Explode | Result                               |
| ---------------- | :-----: | ------------------------------------ |
| `form`           | `false` | `color=blue,black,brown`             |
| `form`           | `true`  | `color=blue&color=black&color=brown` |
| `spaceDelimited` | `false` | `color=blue%20black%20brown`         |
| `pipeDelimited`  | `false` | `color=blue\|black\|brown`           |

## See Also

- [Paths Object](paths_object.md) - Path-level parameters
- [Operation Object](operation_object.md) - Operation-level parameters
- [Schema Object](schema_object.md) - Parameter schema definitions
- [Media Type Object](media_type_object.md) - Content-based parameter definitions
- [Appendix C: RFC6570 Serialization](appendix_c.md) - Detailed serialization guidance
