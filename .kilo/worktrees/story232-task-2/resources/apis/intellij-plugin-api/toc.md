# IntelliJ Platform Plugin SDK Documentation

This directory contains comprehensive documentation for developing plugins for JetBrains IDEs using the IntelliJ Platform SDK.

## Table of Contents

**[overview.md](overview.md)**
: Introduction to the IntelliJ Platform and plugin development

**[getting_started.md](getting_started.md)**
: Setting up your development environment and creating your first plugin

### Core Topics

**[01-core-concepts/](01-core-concepts/)**
: Fundamental concepts including the platform architecture, plugin lifecycle, actions, services, and extension points

**[02-project-structure/](02-project-structure/)**
: Project model, virtual file system, indexing, and workspace model

**[03-psi/](03-psi/)**
: Program Structure Interface (PSI) for parsing and manipulating source code

**[04-editor-features/](04-editor-features/)**
: Editor integration including documents, completion, intentions, and formatting

**[05-language-support/](05-language-support/)**
: Creating custom language support including lexers, parsers, and refactoring

### Advanced Topics

**[06-analysis-features/](06-analysis-features/)**
: Code analysis features including annotators, inspections, and structural search

**[07-ui-components/](07-ui-components/)**
: User interface components including tool windows, dialogs, and tree views

**[08-testing/](08-testing/)**
: Testing strategies for plugins including light and heavy tests

**[09-deployment/](09-deployment/)**
: Plugin configuration, verification, publishing, and compatibility

**[10-advanced-topics/](10-advanced-topics/)**
: Advanced features including run configurations, debuggers, and themes

**[11-reference/](11-reference/)**
: Reference materials including API changes, common patterns, and troubleshooting

## Quick Navigation

### For Beginners

1. Start with [overview.md](overview.md) to understand the platform
2. Follow [getting_started.md](getting_started.md) to set up your environment
3. Read [01-core-concepts/](01-core-concepts/) to learn fundamental concepts

### For Specific Tasks

- **Adding menu items or keyboard shortcuts**: See [01-core-concepts/actions-system.md](01-core-concepts/actions-system.md)
- **Working with files**: See [02-project-structure/virtual-file-system.md](02-project-structure/virtual-file-system.md)
- **Parsing or modifying code**: See [03-psi/](03-psi/)
- **Adding code completion**: See [04-editor-features/code-completion.md](04-editor-features/code-completion.md)
- **Creating a custom language**: See [05-language-support/](05-language-support/)
- **Adding code inspections**: See [06-analysis-features/inspections.md](06-analysis-features/inspections.md)
- **Creating tool windows**: See [07-ui-components/tool-windows.md](07-ui-components/tool-windows.md)
- **Testing your plugin**: See [08-testing/](08-testing/)
- **Publishing your plugin**: See [09-deployment/publishing.md](09-deployment/publishing.md)

## External Resources

The official JetBrains documentation is available at:

- [IntelliJ Platform SDK DevGuide](https://plugins.jetbrains.com/docs/intellij/welcome.html)
- [IntelliJ Platform Explorer](https://plugins.jetbrains.com/intellij-platform-explorer/)
- [JetBrains Marketplace](https://plugins.jetbrains.com/)
