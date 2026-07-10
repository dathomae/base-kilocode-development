# Inspection Profiles

Inspection profiles define which inspections are enabled and their severity levels.

## Overview

### What Are Inspection Profiles?

Inspection profiles provide:

- Configuration of enabled/disabled inspections
- Severity level settings per inspection
- Scope-based inspection filtering
- Profile import/export

### Profile Hierarchy

| Level   | Description               |
| ------- | ------------------------- |
| Project | Project-specific settings |
| IDE     | Global IDE settings       |
| Default | Built-in defaults         |

## Working with Profiles

### Accessing Profiles

```kotlin
// Get inspection profile manager
val profileManager = InspectionProfileManager.getInstance(project)

// Get current profile
val currentProfile = profileManager.currentProfile

// Get specific profile
val profile = profileManager.getProfile("MyProfile")
```

### Modifying Profiles

```kotlin
fun enableInspection(profile: InspectionProfile, shortName: String) {
    val tool = InspectionToolRegistrar.getInstance().getInspectionTool(shortName, project)
    tool?.let {
        profile.enableTool(it.shortName, project)
    }
}

fun disableInspection(profile: InspectionProfile, shortName: String) {
    profile.disableTool(shortName, project)
}

fun setSeverity(profile: InspectionProfile, shortName: String, severity: Severity) {
    profile.setErrorLevel(shortName, severity, project)
}
```

## Creating Custom Profiles

### Programmatic Profile Creation

```kotlin
fun createCustomProfile(): InspectionProfile {
    val profileManager = InspectionProfileManager.getInstance(project)

    // Create new profile
    val profile = InspectionProfileImpl("My Custom Profile")

    // Configure inspections
    val inspections = InspectionToolRegistrar.getInstance()

    for (toolWrapper in inspections) {
        val shortName = toolWrapper.shortName

        when {
            shouldEnable(shortName) -> profile.enableTool(shortName, project)
            shouldDisable(shortName) -> profile.disableTool(shortName, project)
        }

        // Set severity
        val severity = getDesiredSeverity(shortName)
        profile.setErrorLevel(shortName, severity, project)
    }

    // Add to manager
    profileManager.addProfile(profile)

    return profile
}
```

### Profile from Template

```kotlin
fun createProfileFromTemplate(templateName: String): InspectionProfile {
    val profileManager = InspectionProfileManager.getInstance(project)

    // Get template profile
    val template = profileManager.getProfile(templateName)
        ?: throw IllegalArgumentException("Template not found: $templateName")

    // Create copy
    val newProfile = template.clone()
    newProfile.name = "My Profile from $templateName"

    // Customize
    customizeProfile(newProfile)

    profileManager.addProfile(newProfile)
    return newProfile
}
```

## Severity Levels

### Built-in Severities

| Severity       | Description                       |
| -------------- | --------------------------------- |
| ERROR          | Red underline, blocks compilation |
| WARNING        | Yellow underline, informational   |
| WEAK WARNING   | Gray underline, minor issues      |
| SERVER PROBLEM | External server issues            |
| INFORMATION    | Information only, no highlight    |
| TYPO           | Spelling mistakes                 |

### Custom Severities

```kotlin
fun createCustomSeverity(): Severity {
    return Severity(
        "MY_SEVERITY",
        "My Custom Severity",
        TextAttributes()
            .apply { effectType = EffectType.WAVE_UNDERSCORE }
            .apply { effectColor = Color.MAGENTA }
    )
}

fun registerCustomSeverity(severity: Severity) {
    val registrar = SeverityRegistrar.getSeverityRegistrar(project)
    registrar.registerSeverity(severity)
}
```

## Scope-Based Configuration

### Configuring Per Scope

```kotlin
fun configureForScope(
    profile: InspectionProfile,
    shortName: String,
    scope: NamedScope,
    severity: Severity
) {
    // Get scope tool state
    val state = profile.getToolDefaultState(shortName, project)

    // Create scope override
    val override = ScopeToolState(state, scope, severity)

    // Add to profile
    profile.addScopeToolState(shortName, override)
}
```

### Example: Test Scope Configuration

```kotlin
fun configureTestScope(profile: InspectionProfile) {
    val testScope = NamedScope(
        "Tests",
        PackageSetFactory.getInstance().create("test:*..*")
    )

    // Disable some inspections in tests
    val disabledInTests = listOf(
        "UnusedDeclaration",
        "ConstantConditions"
    )

    for (inspection in disabledInTests) {
        configureForScope(
            profile,
            inspection,
            testScope,
                Severity.INFORMATION
        )
    }
}
```

## Profile Storage

### Exporting Profiles

```kotlin
fun exportProfile(profile: InspectionProfile, file: File) {
    val element = Element("profile")

    // Write profile to XML
    profile.writeExternal(element)

    // Save to file
    JDOMUtil.write(element, file)
}
```

### Importing Profiles

```kotlin
fun importProfile(file: File): InspectionProfile {
    val element = JDOMUtil.load(file)

    val profile = InspectionProfileImpl("Imported Profile")
    profile.readExternal(element)

    return profile
}
```

## Profile XML Format

### Example Profile XML

```xml
<profile version="1.0">
  <option name="myName" value="My Custom Profile" />

  <inspection_tool class="UnusedDeclaration" enabled="true" level="WARNING">
    <option name="MAIN_CLASS" value="true" />
    <scope name="Tests" level="INFORMATION" enabled="true" />
  </inspection_tool>

  <inspection_tool class="ConstantConditions" enabled="true" level="WARNING">
    <option name="SUGGEST_NULLABLE_ANNOTATIONS" value="true" />
  </inspection_tool>

  <inspection_tool class="MyCustomInspection" enabled="false" level="ERROR" />
</profile>
```

## Inspection Options

### Accessing Inspection Options

```kotlin
fun getInspectionOptions(profile: InspectionProfile, shortName: String): Map<String, Any> {
    val state = profile.getToolDefaultState(shortName, project)
    val options = mutableMapOf<String, Any>()

    // Get configurable fields
    val tool = state.tool

    for (field in tool.javaClass.fields) {
        if (field.isAnnotationPresent(JvmField::class.java)) {
            options[field.name] = field.get(tool)
        }
    }

    return options
}

fun setInspectionOption(
    profile: InspectionProfile,
    shortName: String,
    optionName: String,
    value: Any
) {
    val state = profile.getToolDefaultState(shortName, project)
    val tool = state.tool

    val field = tool.javaClass.getField(optionName)
    field.set(tool, value)
}
```

## Profile Management UI

### Opening Profile Settings

```kotlin
fun openInspectionSettings(project: Project) {
    ShowSettingsUtil.getInstance().showSettingsDialog(
        project,
        InspectionsConfigurable::class.java
    )
}

fun openInspectionSettingsForTool(project: Project, shortName: String) {
    val configurable = ShowSettingsUtil.getInstance()
        .findProjectConfigurable(project, InspectionsConfigurable::class.java)

    configurable?.let {
        ShowSettingsUtil.getInstance().showSettingsDialog(
            project,
            InspectionsConfigurable::class.java
        ) { configurable ->
            configurable.selectInspection(shortName)
        }
    }
}
```

## Best Practices

### Profile Organization

1. **Use descriptive names**: Name profiles after their purpose
2. **Document settings**: Add comments explaining custom configurations
3. **Version control**: Export profiles to share with team
4. **Scope appropriately**: Use scopes for different code areas

### Example: Team Profile Setup

```kotlin
fun setupTeamProfile(project: Project) {
    val profileManager = InspectionProfileManager.getInstance(project)

    // Check for existing team profile
    val existing = profileManager.getProfile("Team Standard")
    if (existing != null) return

    // Create team profile
    val profile = InspectionProfileImpl("Team Standard")

    // Enable critical inspections
    val criticalInspections = listOf(
        "ConstantConditions",
        "UnusedDeclaration",
        "SpellCheckingInspection"
    )

    for (shortName in criticalInspections) {
        profile.enableTool(shortName, project)
        profile.setErrorLevel(shortName, Severity.ERROR, project)
    }

    // Configure test scope
    configureTestScope(profile)

    // Add and set as current
    profileManager.addProfile(profile)
    profileManager.currentProfile = profile
}
```

## Related Topics

- [Inspections](inspections.md) - Local inspections
- [Global Inspections](global-inspections.md) - Project-wide analysis
- [Annotators](annotators.md) - Semantic highlighting
