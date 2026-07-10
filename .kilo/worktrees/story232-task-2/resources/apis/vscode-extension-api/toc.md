# VSCode Extension API Documentation

This directory contains comprehensive documentation for developing extensions for Visual Studio Code using the Extension API.

## Table of Contents

**[overview.md](overview.md)**
: Introduction to the VSCode Extension API and extension architecture

**[getting-started.md](getting-started.md)**
: Setting up your development environment and creating your first extension

### Core Topics

**[01-extension-anatomy/](01-extension-anatomy/)**
: Extension manifest, activation events, and lifecycle management

**[02-core-capabilities/](02-core-capabilities/)**
: Commands, menus, keybindings, and configuration

**[03-language-features/](03-language-features/)**
: Programmatic language features including completion, hover, diagnostics, and navigation

**[04-ui-components/](04-ui-components/)**
: User interface components including webviews, tree views, status bar, and notifications

### Advanced Topics

**[05-advanced-topics/](05-advanced-topics/)**
: Debug adapters, task providers, terminal integration, and file system providers

**[06-testing-and-publishing/](06-testing-and-publishing/)**
: Testing extensions and publishing to the marketplace

## Quick Navigation

### For Beginners

1. Start with [overview.md](overview.md) to understand the extension architecture
2. Follow [getting-started.md](getting-started.md) to create your first extension
3. Read [01-extension-anatomy/](01-extension-anatomy/) to understand extension structure

### For Specific Tasks

- **Adding commands**: See [02-core-capabilities/commands.md](02-core-capabilities/commands.md)
- **Creating menus**: See [02-core-capabilities/menus.md](02-core-capabilities/menus.md)
- **Adding keybindings**: See [02-core-capabilities/keybindings.md](02-core-capabilities/keybindings.md)
- **Code completion**: See [03-language-features/completion.md](03-language-features/completion.md)
- **Diagnostics**: See [03-language-features/diagnostics.md](03-language-features/diagnostics.md)
- **Custom views**: See [04-ui-components/tree-views.md](04-ui-components/tree-views.md)
- **Webviews**: See [04-ui-components/webviews.md](04-ui-components/webviews.md)
- **Debug support**: See [05-advanced-topics/debug-adapters.md](05-advanced-topics/debug-adapters.md)
- **Tasks**: See [05-advanced-topics/tasks.md](05-advanced-topics/tasks.md)
- **Testing**: See [06-testing-and-publishing/testing.md](06-testing-and-publishing/testing.md)
- **Publishing**: See [06-testing-and-publishing/publishing.md](06-testing-and-publishing/publishing.md)

## Extension Architecture Overview

VSCode extensions run in a separate Node.js process called the Extension Host. This isolation ensures that extensions cannot block the main editor process, maintaining VSCode's responsiveness.

### Key Concepts

**Extension Manifest**: The `package.json` file declares extension metadata, dependencies, and contributions.

**Activation Events**: Extensions activate lazily based on triggers like opening files or executing commands.

**Extension Context**: Provides access to storage, subscriptions, and extension resources during activation.

**Providers**: Many features use a provider pattern where extensions register objects that VSCode calls when needed.

### Common Extension Types

**Language Extensions**: Provide language support including completion, hover, diagnostics, and navigation.

**Theme Extensions**: Contribute color themes, file icon themes, and product icon themes.

**Debugger Extensions**: Implement debug adapters for specific runtimes or languages.

**Tool Extensions**: Integrate external tools, linters, formatters, or build systems.

## External Resources

The official VSCode documentation is available at:

- [VSCode Extension API](https://code.visualstudio.com/api)
- [Extension Guidelines](https://code.visualstudio.com/api/references/extension-guidelines)
- [Extension Samples](https://github.com/microsoft/vscode-extension-samples)
- [VSCode Marketplace](https://marketplace.visualstudio.com/vscode)
