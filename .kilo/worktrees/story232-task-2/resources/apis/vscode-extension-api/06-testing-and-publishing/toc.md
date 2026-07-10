# Testing and Publishing

This section covers testing extensions and publishing them to the VSCode Marketplace.

## Table of Contents

**[testing.md](testing.md)**
: Writing and running tests for VSCode extensions

**[publishing.md](publishing.md)**
: Packaging and publishing extensions to the marketplace

## Quick Reference

### Running Tests

```bash
# Run tests from command line
npm test

# Run tests in VSCode
# Press F5 with test file open
```

### Packaging Extension

```bash
# Install vsce
npm install -g @vscode/vsce

# Package as VSIX
vsce package

# Publish to marketplace
vsce publish
```

### Test Structure

```typescript
import * as vscode from "vscode";
import * as assert from "assert";

suite("Extension Test Suite", () => {
  test("Sample test", async () => {
    const extension = vscode.extensions.getExtension("publisher.extensionName");
    assert.ok(extension);
  });
});
```

### CI/CD Example

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm install
      - run: npm test
```
