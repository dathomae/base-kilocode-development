# Publishing

Publishing plugins to the JetBrains Marketplace makes them available to all users.

## Overview

### What Is JetBrains Marketplace?

JetBrains Marketplace provides:

- Plugin distribution
- Version management
- User reviews and ratings
- Download statistics

### Publishing Options

| Method | Description          |
| ------ | -------------------- |
| Gradle | Automated publishing |
| Web UI | Manual upload        |
| API    | Programmatic upload  |

## Preparing for Publishing

### Plugin Requirements

1. **Unique plugin ID**: Must be unique across marketplace
2. **Complete metadata**: Name, description, vendor
3. **Version number**: Semantic versioning
4. **Change notes**: Description of changes
5. **Screenshots**: Visual documentation

### plugin.xml Preparation

```xml
<idea-plugin>
    <id>com.example.myplugin</id>
    <name>My Plugin</name>
    <version>1.0.0</version>

    <vendor email="support@example.com"
            url="https://example.com">My Company</vendor>

    <description><![CDATA[
        <p>Plugin description with HTML support.</p>
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
</idea-plugin>
```

## Gradle Publishing

### Configuration

```kotlin
// build.gradle.kts
intellij {
    type.set("IC")
    version.set("2023.1")
}

tasks {
    publishPlugin {
        token.set(System.getenv("PUBLISH_TOKEN"))

        // Optional: specify channel
        channels.set(listOf("stable"))
    }
}
```

### Publishing

```bash
# Set token
export PUBLISH_TOKEN=your_token_here

# Publish
./gradlew publishPlugin
```

### Publishing Channels

```kotlin
tasks {
    publishPlugin {
        // Stable channel (default)
        channels.set(listOf("stable"))

        // Beta channel
        channels.set(listOf("beta"))

        // EAP channel
        channels.set(listOf("eap"))
    }
}
```

## Web UI Publishing

### Steps

1. Log in to [JetBrains Marketplace](https://plugins.jetbrains.com/)
2. Click "Add Plugin"
3. Upload plugin ZIP file
4. Fill in plugin details
5. Submit for review

### Required Information

| Field        | Description          |
| ------------ | -------------------- |
| Plugin ID    | Unique identifier    |
| Name         | Display name         |
| Category     | Plugin category      |
| Description  | Detailed description |
| Change notes | Version changes      |
| License      | License type         |
| Source code  | Repository URL       |

## Plugin Signing

### Why Sign Plugins?

Plugin signing:

- Verifies plugin authenticity
- Protects users from tampering
- Required for marketplace

### Setting Up Signing

```kotlin
// build.gradle.kts
tasks {
    signPlugin {
        certificateChain.set(System.getenv("CERTIFICATE_CHAIN"))
        privateKey.set(System.getenv("PRIVATE_KEY"))
        password.set(System.getenv("PRIVATE_KEY_PASSWORD"))
    }

    buildPlugin {
        dependsOn(signPlugin)
    }
}
```

### Generating Keys

```bash
# Generate key pair
./gradlew generateKey

# Keys stored in:
# - certificateChain.crt
# - privateKey.pem
```

## Version Management

### Semantic Versioning

```
MAJOR.MINOR.PATCH

Examples:
1.0.0 - Initial release
1.1.0 - New features
1.1.1 - Bug fixes
2.0.0 - Breaking changes
```

### Updating Versions

```kotlin
// gradle.properties
pluginVersion=1.0.0

// For updates
pluginVersion=1.1.0
```

### Change Notes

```xml
<change-notes><![CDATA[
    <h2>Version 1.1.0</h2>
    <ul>
        <li>Added feature X</li>
        <li>Fixed bug Y</li>
        <li>Improved performance</li>
    </ul>

    <h2>Version 1.0.0</h2>
    <ul>
        <li>Initial release</li>
    </ul>
]]></change-notes>
```

## Plugin Approval

### Review Process

1. Submit plugin
2. JetBrains reviews plugin
3. Plugin approved or rejected
4. Feedback provided if rejected

### Common Rejection Reasons

| Reason               | Solution                 |
| -------------------- | ------------------------ |
| Missing description  | Add detailed description |
| No change notes      | Document version changes |
| Compatibility issues | Fix verifier errors      |
| Security concerns    | Address security issues  |
| Poor quality         | Improve code quality     |

## Plugin Statistics

### Viewing Statistics

1. Log in to Marketplace
2. Navigate to your plugin
3. View "Statistics" tab

### Available Metrics

| Metric       | Description     |
| ------------ | --------------- |
| Downloads    | Total downloads |
| Active users | Current users   |
| Ratings      | User ratings    |
| Reviews      | User reviews    |

## Marketing

### Improving Visibility

1. **Clear description**: Explain plugin benefits
2. **Screenshots**: Show plugin in action
3. **Video demos**: Demonstrate features
4. **Documentation**: Provide user guides
5. **Support**: Respond to issues

### Categories

Choose appropriate categories:

- Code Tools
- Languages
- Frameworks
- Build Tools
- VCS
- UI
- Other

## Best Practices

### Publishing Guidelines

1. **Test thoroughly**: Verify before publishing
2. **Document changes**: Clear change notes
3. **Support users**: Respond to feedback
4. **Update regularly**: Fix bugs, add features
5. **Maintain compatibility**: Support current IDE versions

### Example: Complete Publishing Setup

```kotlin
// build.gradle.kts
plugins {
    id("org.jetbrains.intellij") version "1.15.0"
}

group = "com.example"
version = "1.0.0"

intellij {
    type.set("IC")
    version.set("2023.1")
    plugins.set(listOf("java"))
}

tasks {
    buildPlugin {
        archiveBaseName.set("my-plugin")
    }

    verifyPlugin {
        // Verify plugin structure
    }

    signPlugin {
        certificateChain.set(System.getenv("CERTIFICATE_CHAIN"))
        privateKey.set(System.getenv("PRIVATE_KEY"))
        password.set(System.getenv("PRIVATE_KEY_PASSWORD"))
    }

    publishPlugin {
        token.set(System.getenv("PUBLISH_TOKEN"))
        channels.set(listOf("stable"))
    }

    // Run verification before publishing
    publishPlugin {
        dependsOn(verifyPlugin)
    }
}
```

## Related Topics

- [Plugin Configuration](plugin-configuration.md) - Plugin manifest
- [Plugin Verifier](plugin-verifier.md) - Verification tool
- [Plugin Compatibility](plugin-compatibility.md) - Version compatibility
