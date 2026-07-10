# Testing Extensions

Testing ensures your extension works correctly across different scenarios. VSCode provides a testing framework that runs tests in a special Extension Development Host environment.

## Test Project Structure

The extension generator creates a test structure:

```
my-extension/
├── src/
│   └── extension.ts
├── src/test/
│   ├── runTest.ts        # Test runner entry point
│   └── suite/
│       ├── index.ts      # Test suite setup
│       └── extension.test.ts  # Test files
├── .vscode/
│   └── launch.json       # Debug configuration for tests
└── package.json
```

## Test Runner

The test runner (`src/test/runTest.ts`) sets up the test environment:

```typescript
import * as path from "path";
import { runTests } from "@vscode/test-electron";

async function main() {
  try {
    const extensionDevelopmentPath = path.resolve(__dirname, "../../");
    const extensionTestsPath = path.resolve(__dirname, "./suite/index");

    await runTests({
      extensionDevelopmentPath,
      extensionTestsPath,
    });
  } catch (err) {
    console.error("Failed to run tests");
    process.exit(1);
  }
}

main();
```

## Test Suite Setup

The test suite index (`src/test/suite/index.ts`) configures Mocha:

```typescript
import * as path from "path";
import Mocha from "mocha";
import { glob } from "glob";

export async function run(): Promise<void> {
  const mocha = new Mocha({
    ui: "tdd",
    color: true,
  });

  const testsRoot = path.resolve(__dirname, ".");
  const files = await glob("**/**.test.js", { cwd: testsRoot });

  files.forEach((f) => mocha.addFile(path.resolve(testsRoot, f)));

  return new Promise((resolve, reject) => {
    try {
      mocha.run((failures) => {
        if (failures > 0) {
          reject(new Error(`${failures} tests failed.`));
        } else {
          resolve();
        }
      });
    } catch (err) {
      reject(err);
    }
  });
}
```

## Writing Tests

### Basic Test Structure

```typescript
import * as vscode from "vscode";
import * as assert from "assert";

suite("Extension Test Suite", () => {
  suiteSetup(async () => {
    // Runs once before all tests
  });

  setup(async () => {
    // Runs before each test
  });

  teardown(() => {
    // Runs after each test
  });

  suiteTeardown(() => {
    // Runs once after all tests
  });

  test("Sample test", () => {
    assert.strictEqual(1 + 1, 2);
  });
});
```

### Testing Extension Activation

```typescript
suite("Extension Activation", () => {
  test("Extension should be present", () => {
    const extension = vscode.extensions.getExtension("publisher.extensionName");
    assert.ok(extension);
  });

  test("Extension should activate", async () => {
    const extension = vscode.extensions.getExtension("publisher.extensionName");
    await extension?.activate();
    assert.ok(extension?.isActive);
  });

  test("Extension should register commands", async () => {
    const commands = await vscode.commands.getCommands(true);
    assert.ok(commands.includes("myExtension.hello"));
  });
});
```

### Testing Commands

```typescript
suite("Command Tests", () => {
  test("Hello command should show message", async () => {
    // Set up a listener for the message
    let messageShown = false;
    const disposable = vscode.window.onDidChangeActiveTextEditor(() => {
      // Check for message
    });

    // Execute the command
    await vscode.commands.executeCommand("myExtension.hello");

    // Wait for the message
    await new Promise((resolve) => setTimeout(resolve, 100));

    disposable.dispose();
  });
});
```

### Testing with Documents

```typescript
suite("Document Tests", () => {
  let document: vscode.TextDocument;

  setup(async () => {
    // Create a new document
    document = await vscode.workspace.openTextDocument({
      language: "javascript",
      content: "const x = 1;",
    });
    await vscode.window.showTextDocument(document);
  });

  teardown(async () => {
    // Close the document
    await vscode.commands.executeCommand("workbench.action.closeActiveEditor");
  });

  test("Document should be open", () => {
    assert.ok(vscode.window.activeTextEditor);
    assert.strictEqual(
      vscode.window.activeTextEditor?.document.languageId,
      "javascript",
    );
  });

  test("Should edit document", async () => {
    const editor = vscode.window.activeTextEditor!;

    await editor.edit((editBuilder) => {
      editBuilder.insert(new vscode.Position(0, 0), "// Comment\n");
    });

    assert.strictEqual(document.getText(), "// Comment\nconst x = 1;");
  });
});
```

### Testing Completion Provider

```typescript
suite("Completion Tests", () => {
  test("Should provide completions", async () => {
    const document = await vscode.workspace.openTextDocument({
      language: "javascript",
      content: "cons",
    });

    const position = new vscode.Position(0, 4);
    const completions =
      await vscode.commands.executeCommand<vscode.CompletionList>(
        "vscode.executeCompletionItemProvider",
        document.uri,
        position,
      );

    assert.ok(completions);
    assert.ok(completions.items.some((item) => item.label === "console"));
  });
});
```

### Testing Diagnostics

```typescript
suite("Diagnostic Tests", () => {
  test("Should report diagnostics", async () => {
    const document = await vscode.workspace.openTextDocument({
      language: "javascript",
      content: "const x = ;", // Syntax error
    });

    // Wait for diagnostics to be computed
    await new Promise((resolve) => setTimeout(resolve, 500));

    const diagnostics = vscode.languages.getDiagnostics(document.uri);
    assert.ok(diagnostics.length > 0);
  });
});
```

### Testing Configuration

```typescript
suite("Configuration Tests", () => {
  test("Should read configuration", async () => {
    const config = vscode.workspace.getConfiguration("myExtension");
    const value = config.get<boolean>("enableFeature");

    assert.strictEqual(value, true);
  });

  test("Should update configuration", async () => {
    const config = vscode.workspace.getConfiguration("myExtension");
    await config.update(
      "enableFeature",
      false,
      vscode.ConfigurationTarget.Global,
    );

    const newValue = config.get<boolean>("enableFeature");
    assert.strictEqual(newValue, false);

    // Reset
    await config.update(
      "enableFeature",
      true,
      vscode.ConfigurationTarget.Global,
    );
  });
});
```

## Running Tests

### From VSCode

1. Open a test file
2. Press F5 or use "Run > Start Debugging"
3. Tests run in a new Extension Development Host window

### From Command Line

```bash
npm test
```

This runs the test script defined in `package.json`:

```json
{
  "scripts": {
    "test": "node ./out/test/runTest.js"
  }
}
```

### With Specific Options

```bash
# Run with specific reporter
npm test -- --reporter spec

# Run specific test file
npm test -- --grep "Command Tests"
```

## Test Utilities

### Waiting for Conditions

```typescript
async function waitFor(
  condition: () => boolean,
  timeout: number = 5000,
): Promise<void> {
  const start = Date.now();
  while (!condition()) {
    if (Date.now() - start > timeout) {
      throw new Error("Timeout waiting for condition");
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
}

// Usage
await waitFor(() => vscode.window.activeTextEditor !== undefined);
```

### Creating Test Files

```typescript
async function createTestFile(
  content: string,
  language: string = "javascript",
): Promise<vscode.TextDocument> {
  const document = await vscode.workspace.openTextDocument({
    language,
    content,
  });
  await vscode.window.showTextDocument(document);
  return document;
}
```

### Cleaning Up

```typescript
async function closeAllEditors(): Promise<void> {
  await vscode.commands.executeCommand("workbench.action.closeAllEditors");
}
```

## Mocking

### Mocking VSCode API

For unit tests that don't need the full VSCode environment:

```typescript
// test/mocks/vscode.ts
export const mockVscode = {
  window: {
    showInformationMessage: sinon.stub(),
    showErrorMessage: sinon.stub(),
    activeTextEditor: undefined,
  },
  workspace: {
    getConfiguration: sinon.stub().returns({
      get: sinon.stub(),
    }),
  },
  commands: {
    registerCommand: sinon.stub(),
    executeCommand: sinon.stub(),
  },
};
```

### Using Sinon for Stubs

```typescript
import * as sinon from "sinon";

suite("With Stubs", () => {
  let sandbox: sinon.SinonSandbox;

  setup(() => {
    sandbox = sinon.createSandbox();
  });

  teardown(() => {
    sandbox.restore();
  });

  test("Should call showInformationMessage", async () => {
    const stub = sandbox.stub(vscode.window, "showInformationMessage");

    await vscode.commands.executeCommand("myExtension.hello");

    assert.ok(stub.calledOnce);
    assert.ok(stub.calledWith("Hello World!"));
  });
});
```

## CI/CD Integration

### GitHub Actions

```yaml
# .github/workflows/test.yml
name: Tests

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: "18"
          cache: "npm"

      - name: Install dependencies
        run: npm ci

      - name: Compile
        run: npm run compile

      - name: Run tests
        uses: coactions/setup-xvfb@v1
        with:
          run: npm test
```

### Azure Pipelines

```yaml
# azure-pipelines.yml
trigger:
  - main

pool:
  vmImage: "ubuntu-latest"

steps:
  - task: NodeTool@0
    inputs:
      versionSpec: "18.x"

  - script: npm ci
    displayName: "Install dependencies"

  - script: npm run compile
    displayName: "Compile"

  - script: npm test
    displayName: "Run tests"
    env:
      DISPLAY: ":99.0"
```

## Best Practices

**Test in isolation**: Each test should be independent and not rely on other tests.

**Clean up resources**: Close editors, delete files, and restore settings after tests.

**Use meaningful names**: Test names should describe what is being tested.

**Test edge cases**: Don't just test the happy path; test error conditions too.

**Keep tests fast**: Slow tests discourage running them frequently.

**Use CI**: Run tests automatically on every push and pull request.

## Next Steps

- [Publishing](publishing.md): Package and publish your extension
- [VSCode Testing Documentation](https://code.visualstudio.com/api/working-with-extensions/testing-extension)
