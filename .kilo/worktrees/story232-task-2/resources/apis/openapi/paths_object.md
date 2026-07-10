# Paths Object

Holds the relative paths to the individual endpoints and their operations. The path is appended to the URL from the [Server Object](server_object.md) in order to construct the full URL. The Paths Object MAY be empty, for example if the API only has webhooks.

## Patterned Fields

| Field Pattern |                  Type                   | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ------------- | :-------------------------------------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| /{path}       | [Path Item Object](path_item_object.md) | A relative path to an individual endpoint. The field name MUST begin with a forward slash (`/`). The path is appended (without a separator) to the expanded URL from the [Server Object](server_object.md)'s `url` field in order to construct the full URL. [Path templating](#path-templating) is allowed. When matching URLs, concrete (non-templated) paths would be matched before their templated counterparts. Templated paths with the same hierarchy but different templated names MUST NOT exist as they are identical. In case of ambiguous matching, it's up to the tooling to decide which one to use. |

This object MAY be extended with Specification Extensions.

## Path Templating

Path templating refers to the usage of template expressions within a path to denote variable parts that are substituted with actual values when the API is called.

### Template Syntax

- Template expressions are enclosed in curly braces: `{variable}`
- The variable name can contain letters, digits, underscores, and hyphens
- Template expressions MUST NOT contain spaces

### Path Matching Rules

When matching URLs:

1. Concrete (non-templated) paths are matched before templated paths
2. More specific templated paths are matched before less specific ones
3. If multiple templated paths have the same level of specificity, the behavior is implementation-defined

### Path Templating Example

```yaml
paths:
  /pets:
    get:
      summary: List all pets
  /pets/{petId}:
    get:
      summary: Get a pet by ID
  /pets/{petId}/owner:
    get:
      summary: Get the owner of a pet
  /pets/mine: # This is matched before /pets/{petId}
    get:
      summary: Get my pets
```

## Paths Object Example

```yaml
/pets:
  get:
    description: Returns all pets from the system that the user has access to
    responses:
      "200":
        description: A list of pets.
        content:
          application/json:
            schema:
              type: array
              items:
                $ref: "#/components/schemas/pet"
  post:
    description: Creates a new pet in the store. Duplicates are allowed
    requestBody:
      description: Pet to add to the store
      required: true
      content:
        application/json:
          schema:
            $ref: "#/components/schemas/newPet"
    responses:
      "201":
        description: Pet response
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/pet"
/pets/{id}:
  get:
    description: Returns a user based on a single ID, if the user does not have access to the pet
    parameters:
      - name: id
        in: path
        description: ID of pet to fetch
        required: true
        schema:
          type: integer
          format: int64
    responses:
      "200":
        description: pet response
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/pet"
      "404":
        description: pet not found
```

## Path Item Object

Describes the operations available on a single path. A Path Item MAY be empty, due to [ACL constraints](security_filtering.md). The path itself is still exposed to the documentation viewer but they will not know which operations and parameters are available.

### Fixed Fields

| Field Name  |                                         Type                                         | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      |
| ----------- | :----------------------------------------------------------------------------------: | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| $ref        |                                       `string`                                       | Allows for an external definition of this path item. The referenced structure MUST be in the format of a [Path Item Object](#path-item-object). In case a Path Item Object field appears both in the defined object and the referenced object, the behavior is undefined. See the rules for resolving [Relative References](openapi_object.md#relative-references-in-api-description-uris).                                                                                                      |
| summary     |                                       `string`                                       | An optional, string summary, intended to apply to all operations in this path.                                                                                                                                                                                                                                                                                                                                                                                                                   |
| description |                                       `string`                                       | An optional, string description, intended to apply to all operations in this path. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                                                                                                                                                                                                                                                                   |
| get         |                       [Operation Object](operation_object.md)                        | A definition of a GET operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| put         |                       [Operation Object](operation_object.md)                        | A definition of a PUT operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| post        |                       [Operation Object](operation_object.md)                        | A definition of a POST operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| delete      |                       [Operation Object](operation_object.md)                        | A definition of a DELETE operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                                 |
| options     |                       [Operation Object](operation_object.md)                        | A definition of an OPTIONS operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| head        |                       [Operation Object](operation_object.md)                        | A definition of a HEAD operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| patch       |                       [Operation Object](operation_object.md)                        | A definition of a PATCH operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| trace       |                       [Operation Object](operation_object.md)                        | A definition of a TRACE operation on this path.                                                                                                                                                                                                                                                                                                                                                                                                                                                  |
| servers     |                         [[Server Object](server_object.md)]                          | An alternative `servers` array to service all operations in this path.                                                                                                                                                                                                                                                                                                                                                                                                                           |
| parameters  | [[Parameter Object](parameter_object.md) \| [Reference Object](reference_object.md)] | A list of parameters that are applicable for all the operations described under this path. These parameters can be overridden at the operation level, but cannot be removed there. The list MUST NOT include duplicated parameters. A unique parameter is defined by a combination of a name and location. The list can use the [Reference Object](reference_object.md) to link to parameters that are defined at the [OpenAPI Object's components/parameters](components_object.md#parameters). |

This object MAY be extended with Specification Extensions.

### Path Item Object Example

```yaml
get:
  description: Returns pets based on ID
  summary: Find pets by ID
  operationId: getPetsById
  responses:
    "200":
      description: pet response
      content:
        "*/*":
          schema:
            type: array
            items:
              $ref: "#/components/schemas/Pet"
    default:
      description: error payload
      content:
        text/html:
          schema:
            $ref: "#/components/schemas/ErrorModel"
parameters:
  - name: id
    in: path
    description: ID of pet to use
    required: true
    schema:
      type: array
      items:
        type: string
    style: simple
```

## See Also

- [Path Item Object](path_item_object.md) - Detailed Path Item documentation
- [Operation Object](operation_object.md) - HTTP operation definitions
- [Parameter Object](parameter_object.md) - Parameter definitions
- [Server Object](server_object.md) - Server URL configuration
