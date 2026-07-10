# Story 016: IntelliJ Plugin Workspace Setup

## Overview

Set up the IntelliJ IDEA plugin workspace under the chronochrome directory. This includes creating the basic project structure for a Kotlin-based IntelliJ Platform plugin, configuring the build system with Gradle, and establishing the development environment for the ChronoCone IntelliJ plugin.

## References

- [Architecture](../specifications/architecture/)
- [IntelliJ Platform SDK Documentation](https://plugins.jetbrains.com/docs/intellij/)
- [Story001_workspace-setup.md](Story001_workspace-setup.md) - Original workspace setup for reference

## Dependencies

### Prerequisites

- [Story001_workspace-setup.md](Story001_workspace-setup.md) - Done (base workspace structure)

### Dependents

- [Story017_timeseries-parser-kotlin.md](Story017_timeseries-parser-kotlin.md) - Not Started (requires IntelliJ plugin project structure)

## Tasks

1. Create IntelliJ plugin project directory structure - Not Started
   a. Create `chronochrome/intellij/` directory under project root
   b. Create `src/main/kotlin/com/chronocone/plugin/` source directory
   c. Create `src/main/resources/META-INF/` for plugin configuration
   d. Create `src/test/kotlin/com/chronocone/plugin/` test directory
   e. Create `src/test/resources/` for test resources

2. Configure Gradle build for IntelliJ plugin - Not Started
   a. Create `build.gradle.kts` with IntelliJ Platform Gradle Plugin
   b. Configure Kotlin JVM plugin and toolchain
   c. Set up IntelliJ Platform extension with target IDE version
   d. Configure dependencies (IntelliJ Platform SDK, test libraries)
   e. Configure test framework (JUnit 5)
   f. Create `settings.gradle.kts` for project settings

3. Create plugin configuration files - Not Started
   a. Create `src/main/resources/META-INF/plugin.xml` with plugin descriptor
   b. Define plugin ID, name, version, and vendor
   c. Define plugin description and change notes
   d. Declare extension points and dependencies
   e. Configure idea-version compatibility

4. Set up basic plugin service structure - Not Started
   a. Create `ChronoConePlugin.kt` main plugin class
   b. Create `plugin.xml` service registration
   c. Create basic plugin lifecycle handlers (init, dispose)
   d. Write unit tests for plugin initialization
   e. Verify plugin loads in IntelliJ IDEA

5. Configure development environment - Not Started
   a. Configure `runIde` Gradle task for testing
   b. Set up IntelliJ IDEA run configuration
   c. Configure code style and inspection settings
   d. Add `.gitignore` entries for IntelliJ-specific files
   e. Document development setup in README.md

6. Add project to multi-root workspace - Not Started
   a. Update root workspace configuration to include intellij folder
   b. Configure workspace-specific settings for Kotlin
   c. Verify workspace opens correctly in IntelliJ IDEA
   d. Document workspace setup in project README

7. Verify build and test infrastructure - Not Started
   a. Run `./gradlew build` to verify project compiles
   b. Run `./gradlew test` to verify test framework works
   c. Run `./gradlew verifyPlugin` to validate plugin configuration
   d. Run `./gradlew runIde` to verify plugin loads in sandbox IDE
   e. Document build commands in README.md

## Decisions

1. **Gradle with Kotlin DSL**: Use Gradle with Kotlin DSL (`build.gradle.kts`) for build configuration. This provides type-safe configuration and better IDE support than Groovy DSL.

2. **IntelliJ Platform Gradle Plugin**: Use the official IntelliJ Platform Gradle Plugin (version 2.x) for plugin development. This handles dependency management, sandbox IDE configuration, and plugin packaging.

3. **Target IDE version**: Target IntelliJ IDEA Community Edition as the primary platform. This ensures the plugin works in the free version of IntelliJ IDEA.

4. **Kotlin version**: Use the latest stable Kotlin version compatible with the IntelliJ Platform version. This ensures access to modern Kotlin features.

5. **Package structure**: Use `com.chronocone.plugin` as the base package. This follows Java/Kotlin package naming conventions and provides a clear namespace.

## Notes

- The IntelliJ Platform SDK documentation should be consulted for best practices on plugin structure and configuration.
- Consider using the IntelliJ Platform Plugin Template as a starting point for project structure.
- The plugin should be designed to work with both IntelliJ IDEA Community and Ultimate editions.
- Future stories will add time series parsing functionality and editor support.
- Consider setting up continuous integration for the plugin build.
