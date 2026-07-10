# Appendix: Serialization

This appendix provides detailed guidance on serialization of parameters and request bodies using RFC6570-based serialization.

## RFC6570 URI Templates

OpenAPI uses [RFC6570](https://datatracker.ietf.org/doc/html/rfc6570) URI Templates for parameter serialization. The `style` and `explode` fields control how values are serialized.

### Style to RFC6570 Mapping

| Style            | RFC6570 Equivalent   |
| ---------------- | -------------------- |
| `simple`         | No prefix operator   |
| `matrix`         | `;` prefix operator  |
| `label`          | `.` prefix operator  |
| `form`           | `?` prefix operator  |
| `spaceDelimited` | No direct equivalent |
| `pipeDelimited`  | No direct equivalent |
| `deepObject`     | No direct equivalent |

### Explode Behavior

| Style    | Explode | RFC6570 Equivalent  |
| -------- | :-----: | ------------------- |
| `simple` | `false` | No modifier         |
| `simple` | `true`  | `*` modifier suffix |
| `form`   | `false` | No modifier         |
| `form`   | `true`  | `*` modifier suffix |

## Style Values by Location

### Query Parameters

| Style            | Array Serialization                | Object Serialization                     |
| ---------------- | ---------------------------------- | ---------------------------------------- |
| `form` (default) | `color=blue,red,green`             | `color=R,100,G,200,B,150`                |
| `form` (explode) | `color=blue&color=red&color=green` | `R=100&G=200&B=150`                      |
| `spaceDelimited` | `color=blue%20red%20green`         | Not defined                              |
| `pipeDelimited`  | `color=blue\|red\|green`           | Not defined                              |
| `deepObject`     | Not defined                        | `color[R]=100&color[G]=200&color[B]=150` |

### Path Parameters

| Style              | Array Serialization                 | Object Serialization       |
| ------------------ | ----------------------------------- | -------------------------- |
| `simple` (default) | `blue,red,green`                    | `R,100,G,200,B,150`        |
| `simple` (explode) | `blue,red,green`                    | `R=100,G=200,B=150`        |
| `label`            | `.blue.red.green`                   | `.R.100.G.200.B.150`       |
| `label` (explode)  | `.blue.red.green`                   | `.R=100.G=200.B=150`       |
| `matrix`           | `;color=blue,red,green`             | `;color=R,100,G,200,B,150` |
| `matrix` (explode) | `;color=blue;color=red;color=green` | `;R=100;G=200;B=150`       |

### Header Parameters

| Style              | Array Serialization | Object Serialization |
| ------------------ | ------------------- | -------------------- |
| `simple` (default) | `blue,red,green`    | `R,100,G,200,B,150`  |
| `simple` (explode) | `blue,red,green`    | `R=100,G=200,B=150`  |

### Cookie Parameters

| Style            | Array Serialization                | Object Serialization      |
| ---------------- | ---------------------------------- | ------------------------- |
| `form` (default) | `color=blue,red,green`             | `color=R,100,G,200,B,150` |
| `form` (explode) | `color=blue&color=red&color=green` | `R=100&G=200&B=150`       |

## Serialization Examples

### Form Style (Query)

```yaml
parameters:
  - name: color
    in: query
    schema:
      type: array
      items:
        type: string
      example: [blue, black, brown]
```

With `style: form` and `explode: false`:

```
?color=blue,black,brown
```

With `style: form` and `explode: true`:

```
?color=blue&color=black&color=brown
```

### Space Delimited Style

```yaml
parameters:
  - name: color
    in: query
    style: spaceDelimited
    schema:
      type: array
      items:
        type: string
      example: [blue, black, brown]
```

Result:

```
?color=blue%20black%20brown
```

### Pipe Delimited Style

```yaml
parameters:
  - name: color
    in: query
    style: pipeDelimited
    schema:
      type: array
      items:
        type: string
      example: [blue, black, brown]
```

Result:

```
?color=blue|black|brown
```

### Deep Object Style

```yaml
parameters:
  - name: filter
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
```

With values `{ status: "active", tags: ["urgent", "important"] }`:

```
?filter[status]=active&filter[tags]=urgent&filter[tags]=important
```

### Label Style (Path)

```yaml
parameters:
  - name: color
    in: path
    style: label
    schema:
      type: array
      items:
        type: string
      example: [blue, black, brown]
```

With `explode: false`:

```
/blue.black.brown
```

With `explode: true`:

```
/.blue.black.brown
```

### Matrix Style (Path)

```yaml
parameters:
  - name: color
    in: path
    style: matrix
    schema:
      type: array
      items:
        type: string
      example: [blue, black, brown]
```

With `explode: false`:

```
;color=blue,black,brown
```

With `explode: true`:

```
;color=blue;color=black;color=brown
```

## Percent-Encoding

### Reserved Characters

RFC3986 defines reserved characters that may need encoding:

```
:/?#[]@!$&'()*+,;=
```

### Encoding Behavior

|  `allowReserved`  | Behavior                                |
| :---------------: | --------------------------------------- |
| `false` (default) | Reserved characters are percent-encoded |
|      `true`       | Reserved characters are passed through  |

### Example with allowReserved

```yaml
parameters:
  - name: path
    in: query
    allowReserved: true
    schema:
      type: string
      example: "/users/123"
```

With `allowReserved: false`:

```
?path=%2Fusers%2F123
```

With `allowReserved: true`:

```
?path=/users/123
```

## Delimiter Handling

### Delimiter Characters

Different styles use different delimiters:

| Style            | Delimiter                     |
| ---------------- | ----------------------------- |
| `simple`         | `,` (comma)                   |
| `form`           | `&` (ampersand)               |
| `spaceDelimited` | ` ` (space, encoded as `%20`) |
| `pipeDelimited`  | `\|` (pipe)                   |
| `label`          | `.` (dot)                     |
| `matrix`         | `;` (semicolon)               |

### Delimiter Escaping

When values contain delimiter characters, they must be handled carefully:

1. Values are joined with the delimiter
2. Delimiters in values should be percent-encoded before joining
3. On parsing, split by delimiter first, then decode values

## Form Media Types

### application/x-www-form-urlencoded

For request bodies with form data:

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
```

### multipart/form-data

For file uploads and mixed content:

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
      encoding:
        metadata:
          contentType: application/json
```

## See Also

- [Parameter Object](parameter_object.md) - Parameter definitions
- [Request Body Object](request_body_object.md) - Request body definitions
- [Media Type Object](media_type_object.md) - Encoding configuration
