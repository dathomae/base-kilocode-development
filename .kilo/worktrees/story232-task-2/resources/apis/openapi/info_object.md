# Info Object

The object provides metadata about the API. The metadata MAY be used by the clients if needed, and MAY be presented in editing or documentation generation tools for convenience.

## Fixed Fields

| Field Name     |               Type                | Description                                                                                                                                                                             |
| -------------- | :-------------------------------: | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| title          |             `string`              | **REQUIRED**. The title of the API.                                                                                                                                                     |
| summary        |             `string`              | A short summary of the API.                                                                                                                                                             |
| description    |             `string`              | A description of the API. [CommonMark syntax](https://spec.commonmark.org/) MAY be used for rich text representation.                                                                   |
| termsOfService |             `string`              | A URL to the Terms of Service for the API. MUST be in the format of a URL.                                                                                                              |
| contact        | [Contact Object](#contact-object) | The contact information for the exposed API.                                                                                                                                            |
| license        | [License Object](#license-object) | The license information for the exposed API.                                                                                                                                            |
| version        |             `string`              | **REQUIRED**. The version of the OpenAPI document (which is distinct from the [OpenAPI Specification version](overview.md#versions-and-deprecation) or the API implementation version). |

This object MAY be extended with Specification Extensions.

## Contact Object

Contact information for the exposed API.

### Fixed Fields

| Field Name |   Type   | Description                                                                                      |
| ---------- | :------: | ------------------------------------------------------------------------------------------------ |
| name       | `string` | The identifying name of the contact person/organization.                                         |
| url        | `string` | The URL pointing to the contact information. MUST be in the format of a URL.                     |
| email      | `string` | The email address of the contact person/organization. MUST be in the format of an email address. |

This object MAY be extended with Specification Extensions.

### Contact Object Example

```yaml
name: API Support
url: https://www.example.com/support
email: support@example.com
```

## License Object

License information for the exposed API.

### Fixed Fields

| Field Name |   Type   | Description                                                                                                                             |
| ---------- | :------: | --------------------------------------------------------------------------------------------------------------------------------------- |
| name       | `string` | **REQUIRED**. The license name used for the API.                                                                                        |
| identifier | `string` | An [SPDX](https://spdx.org/licenses/) license expression for the API. The `identifier` field is mutually exclusive of the `url` field.  |
| url        | `string` | A URL to the license used for the API. MUST be in the format of a URL. The `url` field is mutually exclusive of the `identifier` field. |

This object MAY be extended with Specification Extensions.

### License Object Examples

Using SPDX identifier:

```yaml
name: MIT
identifier: MIT
```

Using URL:

```yaml
name: Apache 2.0
url: https://www.apache.org/licenses/LICENSE-2.0.html
```

## Info Object Example

```yaml
title: Sample Pet Store App
summary: A pet store manager.
description: This is a sample server Petstore server. You can find out more about Swagger at [http://swagger.io](http://swagger.io) or on [irc.freenode.net, #swagger](http://swagger.io/irc/).
termsOfService: https://example.com/terms/
contact:
  name: API Support
  url: https://www.example.com/support
  email: support@example.com
license:
  name: Apache 2.0
  url: https://www.apache.org/licenses/LICENSE-2.0.html
version: 1.0.1
```

## See Also

- [OpenAPI Object](openapi_object.md) - The root object containing this Info Object
- [External Documentation Object](external_documentation_object.md) - Additional external documentation
