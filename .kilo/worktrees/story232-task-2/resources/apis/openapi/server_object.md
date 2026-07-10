# Server Object

An object representing a Server.

## Fixed Fields

| Field Name  |                               Type                               | Description                                                                                                                                                                                                                                                                                                                      |
| ----------- | :--------------------------------------------------------------: | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| url         |                             `string`                             | **REQUIRED**. A URL to the target host. It MAY be relative, to indicate that the host location is relative to the location where the OpenAPI document is being served. Variables are enclosed in curly brackets, for example `{scheme}://{host}:{port}/{basePath}`. See [Server Variables](#server-variable-object) for details. |
| description |                             `string`                             | An optional string describing the host designated by the URL. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                                                        |
| variables   | Map[`string`, [Server Variable Object](#server-variable-object)] | A map between a variable name and its value. The value is used for substitution in the server's URL template.                                                                                                                                                                                                                    |

This object MAY be extended with Specification Extensions.

## Server Variable Object

An object representing a Server Variable for server URL template substitution.

### Fixed Fields

| Field Name  |    Type    | Description                                                                                                                                                                                                                                                                                |
| ----------- | :--------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| enum        | [`string`] | An enumeration of string values to be used if the substitution options are from a limited set. The array MUST NOT be empty.                                                                                                                                                                |
| default     |  `string`  | **REQUIRED**. The default value to use for substitution, which SHALL be sent if an alternate value is not supplied. Note this behavior is different from the default value of object properties in JSON Schema. If the `enum` field is defined, the value MUST exist in the enum's values. |
| description |  `string`  | An optional description for the server variable. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                               |

This object MAY be extended with Specification Extensions.

## URL Template Substitution

Server URLs can contain variables enclosed in curly braces `{}`. These variables are substituted with values from the `variables` map.

### Variable Syntax

Variables in the URL template follow these rules:

- Variable names are case-sensitive
- Variables MUST be defined in the `variables` map
- The `default` value is used if no override is provided
- If `enum` is specified, the value MUST be one of the enumerated values

### Template Examples

```yaml
url: https://{username}.gigantic-server.com:{port}
variables:
  username:
    default: demo
    description: The username for the server
  port:
    enum:
      - "8443"
      - "443"
    default: "8443"
```

## Server Object Examples

### Single Server

```yaml
servers:
  - url: https://development.gigantic-server.com/v1
    description: Development server
```

### Multiple Servers

```yaml
servers:
  - url: https://development.gigantic-server.com/v1
    description: Development server
  - url: https://staging.gigantic-server.com/v1
    description: Staging server
  - url: https://api.gigantic-server.com/v1
    description: Production server
```

### Server with Variables

```yaml
servers:
  - url: https://{username}.gigantic-server.com:{port}/{basePath}
    description: The production API server
    variables:
      username:
        default: demo
        description: This value is assigned by the service provider
      port:
        enum:
          - "8443"
          - "443"
        default: "8443"
      basePath:
        default: v2
```

### Relative URL

```yaml
servers:
  - url: /v1
    description: Relative to current location
```

## Default Behavior

If the `servers` field is not provided or is an empty array, the default value is a single Server Object with:

```yaml
servers:
  - url: /
```

## Server Override

The `servers` field can be specified at different levels:

1. **Root level** ([OpenAPI Object](openapi_object.md)): Applies to all operations unless overridden
2. **Path Item level** ([Path Item Object](path_item_object.md)): Overrides root servers for all operations in that path
3. **Operation level** ([Operation Object](operation_object.md)): Overrides path item and root servers for that specific operation

### Override Example

```yaml
servers:
  - url: https://api.example.com/v1

paths:
  /users:
    servers:
      - url: https://users.example.com/v1
    get:
      servers:
        - url: https://public-users.example.com/v1
      # This operation uses public-users.example.com
    post:
      # This operation uses users.example.com
  /admin:
    # This path uses api.example.com
```

## See Also

- [OpenAPI Object](openapi_object.md) - Root level servers definition
- [Path Item Object](path_item_object.md) - Path level server override
- [Operation Object](operation_object.md) - Operation level server override
