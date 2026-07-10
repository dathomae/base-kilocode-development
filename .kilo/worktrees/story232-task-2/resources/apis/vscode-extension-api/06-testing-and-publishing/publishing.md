# Publishing Extensions

After developing and testing your extension, you can package it as a VSIX file and publish it to the VSCode Marketplace for others to install.

## Prerequisites

Before publishing, ensure your extension is ready:

**Complete manifest**: All required fields in `package.json` are filled

**README**: A comprehensive README.md explaining the extension

**License**: A LICENSE file with your chosen license

**Icon**: An icon for the marketplace (128x128 pixels minimum)

**Changelog**: A CHANGELOG.md documenting version history

## VSCE Tool

The `vsce` (Visual Studio Code Extension Manager) tool packages and publishes extensions:

```bash
# Install vsce globally
npm install -g @vscode/vsce

# Or use npx without installing
npx @vscode/vsce --help
```

## Packaging

Create a VSIX package file:

```bash
# Create a .vsix file
vsce package

# This creates: extension-name-1.0.0.vsix
```

### Installing from VSIX

Users can install the VSIX file:

```bash
code --install-extension extension-name-1.0.0.vsix
```

Or through VSCode:

1. Open Command Palette
2. Run "Extensions: Install from VSIX..."
3. Select the VSIX file

## Publishing to Marketplace

### Create a Publisher

1. Go to [Visual Studio Marketplace](https://marketplace.visualstudio.com/)
2. Sign in with your Microsoft or GitHub account
3. Create a publisher ID

### Login with VSCE

```bash
# Login with your publisher
vsce login <publisher-id>

# You'll be prompted for a Personal Access Token
```

### Create a Personal Access Token

1. Go to [Azure DevOps](https://dev.azure.com/)
2. Click User Settings > Personal Access Tokens
3. Create a new token with:
   - Name: "VSCode Extension Publishing"
   - Organization: All accessible organizations
   - Scopes: Marketplace > Manage

### Publish

```bash
# Publish the extension
vsce publish

# Publish a specific version
vsce publish 1.0.1

# Publish with major/minor/patch increment
vsce publish major  # 1.0.0 -> 2.0.0
vsce publish minor  # 1.0.0 -> 1.1.0
vsce publish patch  # 1.0.0 -> 1.0.1
```

## Manifest Requirements

Ensure your `package.json` has all required fields:

```json
{
  "name": "my-extension",
  "displayName": "My Extension",
  "description": "A brief description of what the extension does",
  "version": "1.0.0",
  "publisher": "my-publisher-id",
  "license": "MIT",
  "icon": "images/icon.png",
  "engines": {
    "vscode": "^1.74.0"
  },
  "categories": ["Programming Languages", "Linters"],
  "keywords": ["javascript", "linting"],
  "repository": {
    "type": "git",
    "url": "https://github.com/username/my-extension.git"
  },
  "bugs": {
    "url": "https://github.com/username/my-extension/issues"
  },
  "homepage": "https://github.com/username/my-extension#readme"
}
```

### Required Fields

| Field            | Description                                 |
| ---------------- | ------------------------------------------- |
| `name`           | Extension identifier (lowercase, no spaces) |
| `displayName`    | Human-readable name                         |
| `description`    | Brief description                           |
| `version`        | Semantic version                            |
| `publisher`      | Your publisher ID                           |
| `engines.vscode` | Minimum VSCode version                      |

### Recommended Fields

| Field        | Description             |
| ------------ | ----------------------- |
| `license`    | SPDX license identifier |
| `icon`       | Path to icon file       |
| `categories` | Marketplace categories  |
| `keywords`   | Search keywords         |
| `repository` | Source code repository  |
| `bugs`       | Issue tracker URL       |

## README Guidelines

A good README includes:

```markdown
# My Extension

Brief description of what the extension does.

## Features

- Feature 1: Description
- Feature 2: Description

## Requirements

List any prerequisites or dependencies.

## Installation

1. Open VSCode
2. Go to Extensions
3. Search for "My Extension"
4. Click Install

## Usage

How to use the extension after installation.

## Configuration

List available settings with descriptions.

## Known Issues

Any known limitations or issues.

## Release Notes

### 1.0.0

Initial release.

## License

MIT
```

## Versioning

Follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

```bash
# Increment version
vsce publish major  # Breaking changes
vsce publish minor  # New features
vsce publish patch  # Bug fixes
```

## Pre-release Versions

Publish pre-release versions for testing:

```json
{
  "version": "1.1.0-beta.1"
}
```

```bash
# Publish as pre-release
vsce publish --pre-release
```

## Unpublishing

Remove an extension from the marketplace:

```bash
# Unpublish completely
vsce unpublish <publisher>.<extension>

# Unpublish a specific version
vsce unpublish <publisher>.<extension>@1.0.0
```

## CI/CD Publishing

Automate publishing with GitHub Actions:

```yaml
# .github/workflows/publish.yml
name: Publish Extension

on:
  release:
    types: [created]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: "18"

      - name: Install dependencies
        run: npm ci

      - name: Compile
        run: npm run compile

      - name: Publish to Marketplace
        run: vsce publish -p ${{ secrets.VSCE_PAT }}
```

Store your Personal Access Token as a repository secret named `VSCE_PAT`.

## Alternative Marketplaces

### Open VSX Registry

An open-source alternative to the VSCode Marketplace:

```bash
# Install the Open VSX CLI
npm install -g ovsx

# Publish to Open VSX
ovsx publish -p <token>
```

Get a token from [Open VSX](https://open-vsx.org/).

## Best Practices

**Test before publishing**: Run all tests and manually verify functionality.

**Write good documentation**: A clear README helps users understand and use your extension.

**Use semantic versioning**: Users should understand what changes between versions.

**Include screenshots**: Visual elements help users understand what the extension does.

**Keep it small**: Minimize extension size by excluding unnecessary files.

**Update regularly**: Keep your extension compatible with new VSCode versions.

**Respond to issues**: Address user feedback and bug reports promptly.

## Common Issues

### Missing Publisher

```
Error: Missing publisher name
```

Add the `publisher` field to `package.json`.

### Invalid Version

```
Error: Invalid version
```

Ensure version follows semantic versioning (e.g., `1.0.0`).

### Missing README

```
Error: README.md not found
```

Create a README.md file in the extension root.

### Large Extension Size

```
Warning: Extension size exceeds 5MB
```

Exclude unnecessary files using `.vscodeignore`:

```
.vscodeignore
.vscode/**
.vscode-test/**
src/**
.gitignore
.eslintrc.json
**/*.ts
**/*.map
node_modules/**
```

## Next Steps

- [Testing](testing.md): Write tests for your extension
- [VSCode Publishing Documentation](https://code.visualstudio.com/api/working-with-extensions/publishing-extension)
