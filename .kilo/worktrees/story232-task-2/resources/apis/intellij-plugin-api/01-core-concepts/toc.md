# Core Concepts

This section covers the fundamental concepts of the IntelliJ Platform that every plugin developer should understand.

## Topics

**[intellij-platform.md](intellij-platform.md)**
: Overview of the IntelliJ Platform architecture and components

**[plugin-lifecycle.md](plugin-lifecycle.md)**
: How plugins are loaded, initialized, and unloaded

**[actions-system.md](actions-system.md)**
: The actions framework for handling user commands

**[services.md](services.md)**
: Creating and using services for business logic

**[extension-points.md](extension-points.md)**
: Extending the platform through extension points

## Recommended Reading Order

For developers new to the IntelliJ Platform:

1. Start with [intellij-platform.md](intellij-platform.md) to understand the overall architecture
2. Read [plugin-lifecycle.md](plugin-lifecycle.md) to understand how plugins integrate with the platform
3. Study [actions-system.md](actions-system.md) as actions are the most common way to add functionality
4. Learn about [services.md](services.md) for organizing business logic
5. Explore [extension-points.md](extension-points.md) to understand how to extend platform features

## Key Concepts

### Component Model

The IntelliJ Platform uses a component-based architecture. Components are managed by the platform and have well-defined lifecycles. Understanding this model is essential for proper resource management.

### Threading Model

The platform has specific threading rules:

- UI operations must run on the Event Dispatch Thread (EDT)
- Long-running operations should use background threads
- Read and write actions have specific threading requirements

### Service Registry

Services are registered in `plugin.xml` and retrieved through the service registry. This enables dependency injection and loose coupling between components.
