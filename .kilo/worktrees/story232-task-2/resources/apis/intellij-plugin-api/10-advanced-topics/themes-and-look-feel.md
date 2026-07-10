# Themes and Look and Feel

Custom themes allow plugins to modify the IDE's visual appearance.

## Overview

### What Are Themes?

Themes provide:

- Color schemes
- Editor colors
- UI styling
- Icon customization

### Theme Types

| Type         | Description                  |
| ------------ | ---------------------------- |
| Editor Theme | Syntax highlighting colors   |
| UI Theme     | Interface colors and styling |
| Icon Theme   | Custom icon sets             |

## Creating a Theme

### Theme Definition

```xml
<!-- META-INF/plugin.xml -->
<extensions defaultExtensionNs="com.intellij">
    <themeProvider id="my-theme" path="/themes/my-theme.theme.json"/>
</extensions>
```

### Theme JSON

```json
{
  "name": "My Theme",
  "dark": false,
  "author": "My Company",
  "editorScheme": "/themes/my-theme.xml",
  "colors": {
    "primary": "#4A90D9",
    "secondary": "#6C7A89",
    "accent": "#E74C3C"
  },
  "ui": {
    "*": {
      "background": "#FFFFFF",
      "foreground": "#2C3E50",
      "selectionBackground": "#4A90D9",
      "selectionForeground": "#FFFFFF"
    },
    "Button": {
      "background": "#4A90D9",
      "foreground": "#FFFFFF",
      "borderColor": "#3A80C9"
    },
    "Panel": {
      "background": "#F5F6F7"
    },
    "Tree": {
      "background": "#FFFFFF",
      "selectionBackground": "#E8F4FD"
    }
  }
}
```

## Editor Color Scheme

### Scheme Definition

```xml
<!-- themes/my-theme.xml -->
<scheme name="My Theme" version="142">
    <colors>
        <option name="LINE_NUMBERS_COLOR" value="9CA3AF"/>
        <option name="CARET_COLOR" value="2C3E50"/>
        <option name="SELECTION_BACKGROUND" value="4A90D9"/>
        <option name="SELECTION_FOREGROUND" value="FFFFFF"/>
    </colors>

    <attributes>
        <option name="TEXT">
            <value>
                <option name="FOREGROUND" value="2C3E50"/>
                <option name="BACKGROUND" value="FFFFFF"/>
            </value>
        </option>

        <option name="KEYWORD">
            <value>
                <option name="FOREGROUND" value="E74C3C"/>
                <option name="FONT_TYPE" value="1"/>
            </value>
        </option>

        <option name="STRING">
            <value>
                <option name="FOREGROUND" value="27AE60"/>
            </value>
        </option>

        <option name="COMMENT">
            <value>
                <option name="FOREGROUND" value="9CA3AF"/>
                <option name="FONT_TYPE" value="2"/>
            </value>
        </option>

        <option name="CLASS_NAME">
            <value>
                <option name="FOREGROUND" value="2980B9"/>
            </value>
        </option>

        <option name="FUNCTION_CALL">
            <value>
                <option name="FOREGROUND" value="8E44AD"/>
            </value>
        </option>
    </attributes>
</scheme>
```

## UI Customization

### UI Colors

```json
{
  "ui": {
    "Panel": {
      "background": "panelBackground"
    },
    "Button": {
      "background": "buttonBackground",
      "foreground": "buttonForeground",
      "borderColor": "buttonBorder",
      "focusedBorderColor": "accentColor"
    },
    "Label": {
      "foreground": "labelForeground",
      "disabledForeground": "disabledForeground"
    },
    "TextField": {
      "background": "fieldBackground",
      "foreground": "fieldForeground",
      "selectionBackground": "selectionBackground"
    },
    "Tree": {
      "background": "treeBackground",
      "foreground": "treeForeground",
      "selectionBackground": "treeSelectionBackground"
    },
    "List": {
      "background": "listBackground",
      "selectionBackground": "listSelectionBackground"
    },
    "Table": {
      "background": "tableBackground",
      "gridColor": "gridColor"
    }
  }
}
```

### Custom UI Components

```json
{
  "ui": {
    "MyCustomComponent": {
      "background": "#F0F0F0",
      "foreground": "#333333",
      "border": "1,1,1,1,#CCCCCC"
    }
  }
}
```

## Dark Theme Support

### Dark Theme Definition

```json
{
  "name": "My Dark Theme",
  "dark": true,
  "colors": {
    "primary": "#5C9CE6",
    "secondary": "#4A5568",
    "accent": "#F56565"
  },
  "ui": {
    "*": {
      "background": "#1A1A2E",
      "foreground": "#E2E8F0",
      "selectionBackground": "#5C9CE6",
      "selectionForeground": "#FFFFFF"
    }
  }
}
```

### Theme Metadata

```kotlin
class MyThemeProvider : ThemeMetadataProvider {
    override fun getMetadata(themeId: String): ThemeMetadata? {
        return when (themeId) {
            "my-light-theme" -> ThemeMetadata(
                id = "my-light-theme",
                name = "My Light Theme",
                isDark = false
            )
            "my-dark-theme" -> ThemeMetadata(
                id = "my-dark-theme",
                name = "My Dark Theme",
                isDark = true
            )
            else -> null
        }
    }
}
```

## Icon Customization

### Icon Theme

```xml
<extensions defaultExtensionNs="com.intellij">
    <iconProvider implementation="com.example.MyIconProvider"/>
</extensions>
```

### Icon Provider

```kotlin
class MyIconProvider : IconProvider() {
    override fun getIcon(
        element: PsiElement,
        flags: Int
    ): Icon? {
        return when (element) {
            is MyCustomElement -> MyIcons.CUSTOM_ICON
            else -> null
        }
    }
}
```

### Custom Icons

```kotlin
object MyIcons {
    val CUSTOM_ICON: Icon = IconLoader.getIcon(
        "/icons/custom.svg",
        MyIcons::class.java
    )

    val FILE_ICON: Icon = IconLoader.getIcon(
        "/icons/file.svg",
        MyIcons::class.java
    )
}
```

## Programmatic Theme Access

### Getting Current Theme

```kotlin
fun getCurrentTheme(): Theme? {
    return LafManager.getInstance().currentLookAndFeel as? Theme
}

fun isDarkTheme(): Boolean {
    return getCurrentTheme()?.isDark ?: false
}
```

### Theme Colors

```kotlin
fun getThemeColor(key: String): Color? {
    return UIManager.getColor(key)
}

fun getPrimaryColor(): Color {
    return getThemeColor("primary") ?: Color.BLUE
}
```

### Listening for Theme Changes

```kotlin
class MyThemeListener : LafManagerListener {
    override fun lookAndFeelChanged(source: LafManager) {
        val newTheme = source.currentLookAndFeel

        // Update UI for new theme
        updateColors(newTheme)
    }
}

// Register listener
LafManager.getInstance().addLafManagerListener(MyThemeListener())
```

## Theme-Specific Resources

### Conditional Resources

```kotlin
fun getThemedIcon(name: String): Icon {
    val isDark = isDarkTheme()
    val path = if (isDark) {
        "/icons/dark/$name.svg"
    } else {
        "/icons/light/$name.svg"
    }

    return IconLoader.getIcon(path, javaClass)
}
```

### Theme-Aware Colors

```kotlin
class ThemedColor(
    private val lightColor: Color,
    private val darkColor: Color
) : JBColor(lightColor, darkColor) {
    companion object {
        val PRIMARY = ThemedColor(
            Color(0x4A90D9),
            Color(0x5C9CE6)
        )

        val SECONDARY = ThemedColor(
            Color(0x6C7A89),
            Color(0x4A5568)
        )
    }
}
```

## Best Practices

### Theme Guidelines

1. **Support both modes**: Provide light and dark variants
2. **Use semantic colors**: Name colors by purpose
3. **Test contrast**: Ensure readability
4. **Follow conventions**: Match IDE patterns
5. **Document colors**: Explain color usage

### Example: Complete Theme

```json
{
  "name": "My Complete Theme",
  "dark": false,
  "author": "My Company",
  "editorScheme": "/themes/my-scheme.xml",
  "colors": {
    "primary": "#4A90D9",
    "secondary": "#6C7A89",
    "accent": "#E74C3C",
    "success": "#27AE60",
    "warning": "#F39C12",
    "error": "#E74C3C",
    "info": "#3498DB"
  },
  "ui": {
    "*": {
      "background": "#FFFFFF",
      "foreground": "#2C3E50",
      "selectionBackground": "#4A90D9",
      "selectionForeground": "#FFFFFF",
      "disabledForeground": "#9CA3AF"
    },
    "Button": {
      "background": "#4A90D9",
      "foreground": "#FFFFFF",
      "borderColor": "#3A80C9",
      "focusedBorderColor": "#2C70B9"
    },
    "Panel": {
      "background": "#F5F6F7"
    },
    "Editor": {
      "background": "#FFFFFF",
      "foreground": "#2C3E50",
      "lineNumberForeground": "#9CA3AF",
      "selectionBackground": "#E8F4FD"
    }
  }
}
```

## Related Topics

- [UI Guidelines](../07-ui-components/ui-guidelines.md) - UI design guidelines
- [Syntax Highlighting](../05-language-support/syntax-highlighting.md) - Editor highlighting
- [Tool Windows](../07-ui-components/tool-windows.md) - Tool window creation
