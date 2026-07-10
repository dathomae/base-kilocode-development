# IntelliJ Platform Overview

The IntelliJ Platform is an open-source framework for building IDEs and developer tools. JetBrains uses this platform as the foundation for their IDEs including IntelliJ IDEA, PyCharm, WebStorm, PhpStorm, and others. You can use the same platform to create plugins that extend these IDEs or build your own standalone development tools.

## What Is the IntelliJ Platform?

The IntelliJ Platform provides a comprehensive set of building blocks for creating developer tools. At its core, it offers:

- **Project Management**: A unified model for representing projects, modules, and their dependencies
- **Virtual File System**: An abstraction layer for working with files regardless of their physical location
- **Program Structure Interface (PSI)**: A parsed representation of source code that enables code analysis and manipulation
- **Editor Components**: Rich text editing with syntax highlighting, code completion, and navigation
- **UI Framework**: A Swing-based UI toolkit with platform-specific theming and accessibility support

The platform is written primarily in Kotlin and Java. Plugins can be written in either language, though Kotlin is now the preferred choice for new development.

## Supported IDEs

Plugins built on the IntelliJ Platform can target any JetBrains IDE. The available target products include:

| Product        | Focus                     |
| -------------- | ------------------------- |
| IntelliJ IDEA  | Java and JVM languages    |
| PyCharm        | Python                    |
| WebStorm       | JavaScript and TypeScript |
| PhpStorm       | PHP                       |
| RubyMine       | Ruby                      |
| GoLand         | Go                        |
| CLion          | C and C++                 |
| Rider          | .NET languages            |
| DataGrip       | Databases and SQL         |
| Android Studio | Android development       |

A single plugin can target multiple IDEs by specifying compatibility in its configuration. Some plugins are designed to work across all JetBrains IDEs, while others target specific products.

## Plugin Architecture

Plugins extend the platform through well-defined extension points. The architecture follows these principles:

### Declarative Configuration

Plugins declare their contributions in a `plugin.xml` file. This declarative approach allows the platform to load plugins efficiently and manage dependencies between them. The configuration specifies:

- Extension points the plugin implements
- Actions the plugin provides
- Services the plugin registers
- Dependencies on other plugins

### Service-Oriented Design

The platform uses a service-oriented architecture. Services are singleton or per-project components that provide specific functionality. Services can be:

- **Application-level**: One instance shared across the entire IDE
- **Project-level**: One instance per open project
- **Module-level**: One instance per module within a project

### Event-Driven Communication

Plugins respond to platform events through listeners and message buses. This decoupled design allows plugins to react to user actions, file changes, and other events without tight coupling to platform internals.

## Key Concepts

### Actions

Actions represent user-triggered operations such as menu items, toolbar buttons, and keyboard shortcuts. The actions system provides a unified way to define, register, and handle user commands.

### Services

Services encapsulate business logic and state. They provide a clean separation between UI code and functionality, and they support dependency injection through the platform's service registry.

### Extension Points

Extension points define places where plugins can contribute functionality. The platform provides hundreds of built-in extension points, and plugins can define their own for other plugins to use.

### PSI (Program Structure Interface)

PSI is the abstract syntax tree representation of source code. It provides a language-agnostic API for working with code structure, enabling features like navigation, refactoring, and analysis.

## Development Workflow

The typical plugin development workflow involves:

1. **Setup**: Create a project using the IntelliJ Platform Plugin template
2. **Configure**: Define plugin metadata and dependencies in `plugin.xml`
3. **Implement**: Write code to implement the desired functionality
4. **Test**: Use the built-in testing framework to verify behavior
5. **Debug**: Run a sandboxed IDE instance with the plugin loaded
6. **Build**: Package the plugin as a JAR or ZIP file
7. **Publish**: Distribute through the JetBrains Marketplace or other channels

## Platform Versions

The IntelliJ Platform evolves with each IDE release. Key considerations for version compatibility include:

- **API Changes**: The platform API changes between major versions. Some changes are backward compatible, while others require plugin updates.
- **Deprecation Policy**: Deprecated APIs are typically removed after two major releases.
- **Version Numbering**: Platform versions correspond to IDE releases (e.g., 2023.1, 2023.2).

When developing plugins, you specify a target platform version. The plugin will be compatible with that version and typically with later versions, though testing across versions is recommended.

## Next Steps

- [Getting Started](getting_started.md) - Set up your development environment
- [Core Concepts](01-core-concepts/) - Learn about the fundamental building blocks
- [Plugin Configuration](09-deployment/plugin-configuration.md) - Understand plugin metadata and dependencies
