# Project Structure

This section covers how the IntelliJ Platform represents projects, files, and their organization.

## Topics

**[project-model.md](project-model.md)**
: Projects, modules, and their relationships

**[virtual-file-system.md](virtual-file-system.md)**
: The Virtual File System abstraction

**[indexing.md](indexing.md)**
: File indexing for fast searching

**[workspace-model.md](workspace-model.md)**
: The modern workspace model for project structure

## Key Concepts

### Project Hierarchy

The IntelliJ Platform organizes code in a hierarchy:

- **Project**: The top-level container for all files and settings
- **Module**: A logical grouping of source files and dependencies
- **Content Root**: A directory containing source files
- **Source Root**: A directory marked as containing source code

### Virtual File System

The Virtual File System (VFS) provides a uniform API for accessing files regardless of their location. Files can be:

- On the local filesystem
- In JAR files
- In version control
- Generated dynamically

### Indexing

The platform maintains indexes for fast searching across files. Understanding indexing is essential for plugins that need to search or analyze code.

## Recommended Reading Order

1. Start with [project-model.md](project-model.md) to understand the project hierarchy
2. Read [virtual-file-system.md](virtual-file-system.md) for file access patterns
3. Study [indexing.md](indexing.md) for search functionality
4. Explore [workspace-model.md](workspace-model.md) for modern project structure APIs
