# Plugin Configuration

Plugin configuration defines metadata, dependencies, and extensions in the plugin manifest.

## Overview

### What Is Plugin Configuration?

Plugin configuration includes:

- Plugin metadata (name, version, vendor)
- Dependency declarations
- Extension registrations
- Resource declarations

### Configuration Files

| File                | Purpose              |
| ------------------- | -------------------- |
| `plugin.xml`        | Main plugin manifest |
| `gradle.properties` | Build configuration  |
| `build.gradle.kts`  | Build script         |

## plugin.xml Structure

### Basic Structure

```xml
<idea-plugin>
    <!-- Plugin metadata -->
    <id>com.example.myplugin</id>
    <name>My Plugin</name>
    <version>1.0.0</version>
    <vendor email="support@example.com" url="https://example.com">My Company</vendor>

    <!-- Description -->
    <description><![CDATA[
        Plugin description with HTML support.
    ]]></description>

    <!-- Change notes -->
    <change-notes><![CDATA[
        <ul>
            <li>Added feature X</li>
            <li>Fixed bug Y</li>
        </ul>
    ]]></change-notes>

    <!-- Compatibility -->
    <idea-version since-build="223.0" until-build="231.*"/>

    <!-- Dependencies -->
    <depends>com.intellij.modules.platform</depends>

    <!-- Extensions -->
    <extensions defaultExtensionNs="com.intellij">
        <!-- Extension declarations -->
    </extensions>

    <!-- Actions -->
    <actions>
        <!-- Action declarations -->
    </actions>
</idea-plugin>
```

## Plugin Metadata

### Required Elements

```xml
<idea-plugin>
    <!-- Unique identifier -->
    <id>com.example.myplugin</id>

    <!-- Display name -->
    <name>My Plugin</name>

    <!-- Version -->
    <version>1.0.0</version>

    <!-- Vendor information -->
    <vendor email="support@example.com"
            url="https://example.com">My Company</vendor>
</idea-plugin>
```

### Optional Elements

```xml
<idea-plugin>
    <!-- Category for marketplace -->
    <category>Code Tools</category>

    <!-- Resource bundle -->
    <resource-bundle>messages.MyBundle</resource-bundle>

    <!-- Description -->
    <description><![CDATA[
        <p>Plugin description.</p>
        <p>Supports HTML formatting.</p>
    ]]></description>

    <!-- Change notes -->
    <change-notes><![CDATA[
        <h2>Version 1.0.0</h2>
        <ul>
            <li>Initial release</li>
        </ul>
    ]]></change-notes>
</idea-plugin>
```

## Version Compatibility

### Idea-Version Element

```xml
<idea-plugin>
    <!-- Specify compatible versions -->
    <idea-version
        since-build="223.0"    <!-- Minimum version -->
        until-build="231.*"/>  <!-- Maximum version -->
</idea-plugin>
```

### Version Numbers

| IDE Version | Build Number |
| ----------- | ------------ |
| 2022.3      | 223.\*       |
| 2023.1      | 231.\*       |
| 2023.2      | 232.\*       |
| 2023.3      | 233.\*       |

### Compatibility Guidelines

```xml
<!-- Support single major version -->
<idea-version since-build="231.0" until-build="231.*"/>

<!-- Support multiple major versions -->
<idea-version since-build="223.0" until-build="233.*"/>

<!-- Open-ended support -->
<idea-version since-build="231.0"/>
```

## Dependencies

### Module Dependencies

```xml
<idea-plugin>
    <!-- Required modules -->
    <depends>com.intellij.modules.platform</depends>
    <depends>com.intellij.modules.java</depends>

    <!-- Optional dependency -->
    <depends optional="true"
             config-file="java-features.xml">
        com.intellij.modules.java
    </depends>

    <!-- Plugin dependency -->
    <depends>com.example.otherplugin</depends>
</idea-plugin>
```

### Available Modules

| Module                               | Description    |
| ------------------------------------ | -------------- |
| `com.intellij.modules.platform`      | Core platform  |
| `com.intellij.modules.java`          | Java support   |
| `com.intellij.modules.python`        | Python support |
| `com.intellij.modules.kotlin`        | Kotlin support |
| `com.intellij.modules.androidstudio` | Android Studio |

### Conditional Configuration

```xml
<!-- Main plugin.xml -->
<idea-plugin>
    <depends optional="true"
             config-file="java-support.xml">
        com.intellij.modules.java
    </depends>
</idea-plugin>

<!-- java-support.xml -->
<idea-plugin>
    <extensions defaultExtensionNs="com.intellij">
        <!-- Java-specific extensions -->
    </extensions>
</idea-plugin>
```

## Extensions

### Registering Extensions

```xml
<idea-plugin>
    <extensions defaultExtensionNs="com.intellij">
        <!-- Service -->
        <applicationService
            serviceImplementation="com.example.MyService"/>

        <!-- Action -->
        <action
            id="MyAction"
            class="com.example.MyAction"/>

        <!-- Tool window -->
        <toolWindow
            id="MyToolWindow"
            factoryClass="com.example.MyToolWindowFactory"/>
    </extensions>
</idea-plugin>
```

### Custom Extension Points

```xml
<idea-plugin>
    <!-- Define extension point -->
    <extensionPoints>
        <extensionPoint
            name="myExtension"
            interface="com.example.MyExtension"
            dynamic="true"/>
    </extensionPoints>

    <!-- Use extension point -->
    <extensions defaultExtensionNs="com.example.myplugin">
        <myExtension implementation="com.example.MyExtensionImpl"/>
    </extensions>
</idea-plugin>
```

## Actions

### Registering Actions

```xml
<idea-plugin>
    <actions>
        <!-- Simple action -->
        <action id="MyAction"
                class="com.example.MyAction"
                text="My Action"
                description="Action description">
            <keyboard-shortcut keymap="$default" first-keystroke="ctrl alt M"/>
        </action>

        <!-- Action group -->
        <group id="MyGroup" text="My Actions" popup="true">
            <action id="Action1" class="com.example.Action1"/>
            <action id="Action2" class="com.example.Action2"/>
            <separator/>
            <action id="Action3" class="com.example.Action3"/>
        </group>

        <!-- Add to existing menu -->
        <action id="MyAction" class="com.example.MyAction">
            <add-to-group group-id="ToolsMenu" anchor="last"/>
        </action>
    </actions>
</idea-plugin>
```

## Application Components

### Services

```xml
<idea-plugin>
    <extensions defaultExtensionNs="com.intellij">
        <!-- Application-level service -->
        <applicationService
            serviceInterface="com.example.MyService"
            serviceImplementation="com.example.MyServiceImpl"/>

        <!-- Project-level service -->
        <projectService
            serviceInterface="com.example.MyProjectService"
            serviceImplementation="com.example.MyProjectServiceImpl"/>
    </extensions>
</idea-plugin>
```

### Listeners

```xml
<idea-plugin>
    <extensions defaultExtensionNs="com.intellij">
        <!-- Application listener -->
        <applicationListener
            class="com.example.MyListener"
            topic="com.intellij.openapi.project.ProjectManagerListener"/>
    </extensions>
</idea-plugin>
```

## Resource Bundle

### Defining Bundle

```xml
<idea-plugin>
    <resource-bundle>messages.MyBundle</resource-bundle>
</idea-plugin>
```

### Using Bundle

```properties
# messages/MyBundle.properties
action.my.text=My Action
action.my.description=Action description
```

```kotlin
class MyAction : AnAction() {
    override fun update(e: AnActionEvent) {
        e.presentation.text = MyBundle.message("action.my.text")
    }
}
```

## Build Configuration

### gradle.properties

```properties
pluginGroup=com.example
pluginName=My Plugin
pluginVersion=1.0.0
pluginSinceBuild=231.0
pluginUntilBuild=233.*

platformType=IC
platformVersion=2023.1
```

### build.gradle.kts

```kotlin
plugins {
    id("java")
    id("org.jetbrains.intellij") version "1.15.0"
}

group = properties("pluginGroup")
version = properties("pluginVersion")

intellij {
    type.set(properties("platformType"))
    version.set(properties("platformVersion"))

    plugins.set(listOf(
        "java",
        "Kotlin"
    ))
}

tasks {
    patchPluginXml {
        sinceBuild.set(properties("pluginSinceBuild"))
        untilBuild.set(properties("pluginUntilBuild"))
    }

    buildPlugin {
        archiveBaseName.set(properties("pluginName"))
    }
}
```

## Best Practices

### Configuration Guidelines

1. **Use unique ID**: Ensure plugin ID is unique
2. **Version correctly**: Follow semantic versioning
3. **Declare dependencies**: List all required modules
4. **Test compatibility**: Verify with target versions
5. **Document changes**: Keep change notes updated

### Example: Complete plugin.xml

```xml
<idea-plugin>
    <id>com.example.myplugin</id>
    <name>My Plugin</name>
    <version>1.0.0</version>
    <vendor email="support@example.com" url="https://example.com">
        My Company
    </vendor>

    <category>Code Tools</category>
    <resource-bundle>messages.MyBundle</resource-bundle>

    <description><![CDATA[
        <p>My Plugin provides useful features.</p>
        <ul>
            <li>Feature 1</li>
            <li>Feature 2</li>
        </ul>
    ]]></description>

    <change-notes><![CDATA[
        <h2>Version 1.0.0</h2>
        <ul>
            <li>Initial release</li>
        </ul>
    ]]></change-notes>

    <idea-version since-build="231.0" until-build="233.*"/>

    <depends>com.intellij.modules.platform</depends>
    <depends optional="true" config-file="java-support.xml">
        com.intellij.modules.java
    </depends>

    <extensionPoints>
        <extensionPoint name="myExtension"
                        interface="com.example.MyExtension"/>
    </extensionPoints>

    <extensions defaultExtensionNs="com.intellij">
        <applicationService
            serviceImplementation="com.example.MyService"/>
        <projectService
            serviceImplementation="com.example.MyProjectService"/>
        <toolWindow id="MyToolWindow"
                    anchor="right"
                    factoryClass="com.example.MyToolWindowFactory"/>
    </extensions>

    <actions>
        <action id="MyAction"
                class="com.example.MyAction"
                text="My Action">
            <add-to-group group-id="ToolsMenu" anchor="last"/>
        </action>
    </actions>
</idea-plugin>
```

## Related Topics

- [Plugin Compatibility](plugin-compatibility.md) - Version compatibility
- [Publishing](publishing.md) - Publishing to Marketplace
- [Plugin Verifier](plugin-verifier.md) - Verification tool
