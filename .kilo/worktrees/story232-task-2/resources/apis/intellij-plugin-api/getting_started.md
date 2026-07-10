# Getting Started with IntelliJ Plugin Development

This guide walks you through setting up a development environment and creating your first IntelliJ Platform plugin.

## Prerequisites

Before you begin, ensure you have the following:

- **JDK 17 or later**: The IntelliJ Platform requires Java 17 as of the 2023.2 release. Earlier versions may support Java 11.
- **IntelliJ IDEA**: The Community Edition is sufficient for plugin development. Ultimate Edition provides additional features.
- **Gradle**: The build system for plugin projects. Gradle 8.x is recommended.

## Creating a Plugin Project

The recommended way to create a new plugin project is using the IntelliJ Platform Plugin template.

### Using the New Project Wizard

1. Open IntelliJ IDEA
2. Select **File > New > Project**
3. In the New Project dialog:
   - Select **IntelliJ Platform Plugin** from the generators list
   - Enter a project name and location
   - Choose the build system (Gradle with Kotlin DSL is recommended)
   - Select the target IDE (IntelliJ IDEA Community is a good default)
4. Click **Create**

### Project Structure

A newly created plugin project has this structure:

```
my-plugin/
├── src/
│   └── main/
│       ├── kotlin/           # Source code
│       └── resources/
│           └── META-INF/
│               └── plugin.xml # Plugin configuration
├── build.gradle.kts          # Build configuration
├── settings.gradle.kts       # Gradle settings
└── gradle.properties         # Gradle properties
```

### Build Configuration

The `build.gradle.kts` file configures the plugin build. Key sections include:

```kotlin
plugins {
    id("java")
    id("org.jetbrains.kotlin.jvm") version "1.9.0"
    id("org.jetbrains.intellij") version "1.16.0"
}

group = "com.example"
version = "1.0.0"

repositories {
    mavenCentral()
}

intellij {
    version.set("2023.2")
    type.set("IC") // IntelliJ IDEA Community
    plugins.set(listOf("java"))
}

tasks {
    patchPluginXml {
        sinceBuild.set("232")
        untilBuild.set("241.*")
    }
}
```

The `intellij` block specifies:

- **version**: The target platform version
- **type**: The IDE type (IC for Community, IU for Ultimate)
- **plugins**: Required bundled plugins

## Plugin Configuration

The `plugin.xml` file in `src/main/resources/META-INF/` defines the plugin's metadata and extensions.

### Basic Structure

```xml
<idea-plugin>
    <id>com.example.my-plugin</id>
    <name>My Plugin</name>
    <vendor>My Company</vendor>
    <version>1.0.0</version>

    <description><![CDATA[
        A brief description of what the plugin does.
    ]]></description>

    <depends>com.intellij.modules.platform</depends>

    <extensions defaultExtensionNs="com.intellij">
        <!-- Extension declarations go here -->
    </extensions>

    <actions>
        <!-- Action declarations go here -->
    </actions>
</idea-plugin>
```

### Required Elements

- **id**: A unique identifier for the plugin, typically a reverse domain name
- **name**: The display name shown in the plugin manager
- **version**: The plugin version using semantic versioning
- **description**: A detailed description of the plugin's functionality

### Dependencies

The `<depends>` element declares dependencies on other plugins or modules:

```xml
<!-- Required for all plugins -->
<depends>com.intellij.modules.platform</depends>

<!-- Required for Java support -->
<depends>com.intellij.modules.java</depends>

<!-- Optional dependency -->
<depends optional="true" config-file="java-features.xml">com.intellij.modules.java</depends>
```

## Creating Your First Action

Actions are the primary way plugins add functionality. An action represents a user-triggered operation.

### Define the Action Class

Create a Kotlin class that extends `AnAction`:

```kotlin
package com.example.myplugin

import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.ui.Messages

class HelloAction : AnAction() {
    override fun actionPerformed(e: AnActionEvent) {
        val project = e.project ?: return
        Messages.showMessageDialog(
            project,
            "Hello from my plugin!",
            "Greeting",
            Messages.getInformationIcon()
        )
    }

    override fun update(e: AnActionEvent) {
        // Enable only when a project is open
        e.presentation.isEnabled = e.project != null
    }
}
```

### Register the Action

Add the action to `plugin.xml`:

```xml
<actions>
    <action id="com.example.myplugin.HelloAction"
            class="com.example.myplugin.HelloAction"
            text="Say Hello"
            description="Displays a greeting message">
        <add-to-group group-id="ToolsMenu" anchor="first"/>
        <keyboard-shortcut keymap="$default" first-keystroke="ctrl alt H"/>
    </action>
</actions>
```

This registers the action in the Tools menu with a keyboard shortcut.

## Running and Debugging

### Running the Plugin

Use the Gradle task to run a sandboxed IDE with your plugin:

```bash
./gradlew runIde
```

This launches a new instance of IntelliJ IDEA with your plugin installed. The sandbox instance uses a separate configuration directory, so it does not affect your development environment.

### Debugging

To debug your plugin:

1. Add breakpoints in your code
2. Run the `runIde` task in debug mode from your IDE
3. The sandbox instance will start with the debugger attached

### Hot Reload

For faster development iterations, use the `runIde` task with continuous mode:

```bash
./gradlew runIde --continuous
```

Note that changes to `plugin.xml` require restarting the sandbox IDE.

## Testing

The IntelliJ Platform provides a testing framework for plugins. Tests run in a sandboxed environment with access to platform components.

### Test Dependencies

Add test dependencies to `build.gradle.kts`:

```kotlin
dependencies {
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.0")
    testImplementation("org.jetbrains.kotlin:kotlin-test")
}

tasks.test {
    useJUnitPlatform()
}
```

### Writing a Simple Test

```kotlin
package com.example.myplugin

import com.intellij.testFramework.fixtures.LightJavaCodeInsightFixtureTestCase

class HelloActionTest : LightJavaCodeInsightFixtureTestCase() {
    fun testActionIsEnabled() {
        val action = HelloAction()
        val event = myFixture.testAction(action)
        assertTrue(event.presentation.isEnabled)
    }
}
```

Run tests with:

```bash
./gradlew test
```

## Building the Plugin

### Building a Distribution

To create a distributable plugin archive:

```bash
./gradlew buildPlugin
```

The output is located in `build/distributions/` as a ZIP file.

### Verifying Compatibility

Use the Plugin Verifier to check compatibility across IDE versions:

```bash
./gradlew verifyPlugin
```

This task checks your plugin against the target platform version and reports any API compatibility issues.

## Publishing

### JetBrains Marketplace

The primary distribution channel is the JetBrains Marketplace:

1. Create an account at [plugins.jetbrains.com](https://plugins.jetbrains.com)
2. Add your plugin repository or upload manually
3. Configure plugin details and screenshots
4. Submit for review

### Manual Distribution

You can also distribute the plugin ZIP file directly. Users can install it via **Settings > Plugins > Gear Icon > Install Plugin from Disk**.

## Common Issues

### Class Not Found Errors

Ensure your classes are in the correct package and the package structure matches the directory structure. Check that the class name in `plugin.xml` matches the fully qualified class name.

### Plugin Not Loading

Check the IDE log file for errors. The log is located at:

- **Linux**: `~/.cache/JetBrains/<IDE>/log/idea.log`
- **macOS**: `~/Library/Logs/JetBrains/<IDE>/idea.log`
- **Windows**: `%LOCALAPPDATA%\JetBrains\<IDE>\log\idea.log`

### API Incompatibility

If the Plugin Verifier reports incompatibility, check the API documentation for the target version. Use the `@ApiStatus.ScheduledForRemoval` annotations to identify deprecated APIs.

## Next Steps

- [Core Concepts](01-core-concepts/) - Learn about the platform architecture
- [Actions System](01-core-concepts/actions-system.md) - Deep dive into actions
- [Services](01-core-concepts/services.md) - Create reusable service components
- [Testing](08-testing/) - Comprehensive testing strategies
