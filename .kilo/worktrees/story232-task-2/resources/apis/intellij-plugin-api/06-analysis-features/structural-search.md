# Structural Search

Structural Search allows finding code patterns based on structural templates rather than text matching.

## Overview

### What Is Structural Search?

Structural search provides:

- Pattern-based code searching
- Template matching with variables
- Replace functionality
- Integration with inspections

### Use Cases

| Use Case           | Description                      |
| ------------------ | -------------------------------- |
| Code review        | Find problematic patterns        |
| Refactoring        | Locate code to transform         |
| Migration          | Update deprecated patterns       |
| Custom inspections | Create pattern-based inspections |

## Basic Concepts

### Search Templates

Templates define the structure to match:

```kotlin
// Simple variable pattern
val $x$ = $expr$

// Method call pattern
$obj$.$method$($args$)

// Class pattern
class $Class$ extends $Base$ {
    $members$
}
```

### Template Variables

Variables capture matched elements:

| Variable Type | Syntax   | Description         |
| ------------- | -------- | ------------------- |
| Any           | `$x$`    | Matches any element |
| Identifier    | `$name$` | Matches identifiers |
| Expression    | `$expr$` | Matches expressions |
| Statement     | `$stmt$` | Matches statements  |

## Using Structural Search

### Programmatic Search

```kotlin
fun performStructuralSearch(
    project: Project,
    template: String,
    scope: GlobalSearchScope
): List<MatchResult> {
    // Parse template
    val searchTemplate = StructuralSearchTemplate(template)

    // Create search options
    val options = SearchOptions(
        searchTemplate,
        scope,
        caseSensitive = true
    )

    // Execute search
    val results = StructuralSearch.search(project, options)

    return results.toList()
}
```

### Search with Constraints

```kotlin
fun searchWithConstraints(project: Project): List<MatchResult> {
    val template = StructuralSearchTemplate("val $x$ = $expr$")

    // Add constraints to variable
    val xVar = template.getVariable("x")
    xVar.addConstraint(
        VariableConstraint.text("x", Regex("[a-z].*"))
    )

    val exprVar = template.getVariable("expr")
    exprVar.addConstraint(
        VariableConstraint.type("expr", "java.lang.String")
    )

    return StructuralSearch.search(
        project,
        SearchOptions(template, GlobalSearchScope.projectScope(project))
    ).toList()
}
```

## Structural Replace

### Basic Replace

```kotlin
fun performStructuralReplace(
    project: Project,
    searchTemplate: String,
    replaceTemplate: String,
    scope: GlobalSearchScope
) {
    val search = StructuralSearchTemplate(searchTemplate)
    val replace = StructuralReplaceTemplate(replaceTemplate)

    val options = ReplaceOptions(search, replace, scope)

    // Execute replace
    StructuralSearch.replace(project, options)
}
```

### Example: Convert to Property

```kotlin
fun convertGettersToProperties(project: Project) {
    // Search pattern: obj.getFoo()
    val search = StructuralSearchTemplate("$obj$.get$Prop$()")

    // Replace pattern: obj.foo
    val replace = StructuralReplaceTemplate("$obj$.$prop$")

    // Configure variable mapping
    replace.mapVariable("Prop", "prop") { propValue ->
        // Convert "Foo" to "foo"
        propValue.replaceFirstChar { it.lowercase() }
    }

    val options = ReplaceOptions(
        search,
        replace,
        GlobalSearchScope.projectScope(project)
    )

    StructuralSearch.replace(project, options)
}
```

## Creating Search Templates

### Template Syntax

```
// Variable: $name$
$variable$

// Typed variable: $name$:type
$variable$:MyType

// Optional element: $name$?
$variable$?

// Quantifiers
$name$*      // Zero or more
$name$+      // One or more
$name${0,5}  // Between 0 and 5
```

### Common Patterns

```kotlin
// Find all if statements
if ($cond$) $then$

// Find method calls with specific name
$obj$.$method$($args$)
// Constraint: method == "toString"

// Find class fields
class $Class$ {
    $FieldType$ $FieldName$;
}

// Find try-catch blocks
try {
    $body$
} catch ($Exception$ $e$) {
    $handler$
}
```

## Variable Constraints

### Text Constraints

```kotlin
// Exact match
variable.addConstraint(VariableConstraint.textEquals("expectedValue"))

// Regex match
variable.addConstraint(VariableConstraint.textMatches(Regex("[a-z]+")))

// Within set
variable.addConstraint(VariableConstraint.textIn(setOf("val1", "val2")))
```

### Type Constraints

```kotlin
// Exact type
variable.addConstraint(VariableConstraint.typeEquals("java.lang.String"))

// Subtype
variable.addConstraint(VariableConstraint.typeSubtype("java.lang.Exception"))

// Supertype
variable.addConstraint(VariableConstraint.typeSupertype("java.lang.Object"))
```

### Reference Constraints

```kotlin
// Reference to specific element
val targetClass = JavaPsiFacade.getInstance(project)
    .findClass("com.example.TargetClass", GlobalSearchScope.allScope(project))

variable.addConstraint(VariableConstraint.referencesTo(targetClass))

// Reference from specific element
variable.addConstraint(VariableConstraint.referencedFrom(targetClass))
```

## Structural Search Inspections

### Creating SSInspection

```kotlin
class MyStructuralSearchInspection : StructuralSearchInspectionTool() {
    override fun getDisplayName() = "My structural search inspection"

    override fun getShortName() = "MySSInspection"

    override fun getStaticDescription() = """
        Detects problematic code patterns using structural search.
    """.trimIndent()

    override fun getSearchTemplate(): String {
        return "if ($cond$) { return $value$; }"
    }

    override fun getReplaceTemplate(): String? {
        return "return $cond$ ? $value$ : null;"
    }

    override fun getVariableConstraints(): List<VariableConstraintInfo> {
        return listOf(
            VariableConstraintInfo(
                "cond",
                VariableConstraint.typeSubtype("java.lang.Boolean")
            )
        )
    }
}
```

### Registering SSInspection

```xml
<extensions defaultExtensionNs="com.intellij">
    <localInspection
        shortName="MySSInspection"
        displayName="My structural search inspection"
        groupName="My plugin"
        enabledByDefault="true"
        implementationClass="com.example.MyStructuralSearchInspection"/>
</extensions>
```

## Advanced Patterns

### Nested Patterns

```kotlin
// Find if inside for loop
for ($var$ in $collection$) {
    if ($cond$) {
        $body$
    }
}

// Find method with specific annotation
@$Annotation$
fun $method$($params$): $returnType$
```

### Context Constraints

```kotlin
// Must be inside specific class
variable.addConstraint(
    VariableConstraint.inClass("com.example.MyClass")
)

// Must be in method
variable.addConstraint(
    VariableConstraint.inMethod("myMethod")
)

// Must be at top level
variable.addConstraint(
    VariableConstraint.atTopLevel()
)
```

## Performance Considerations

### Efficient Searches

```kotlin
fun efficientSearch(project: Project): List<MatchResult> {
    // Use narrow scope
    val scope = GlobalSearchScope.fileScope(project, "src/main/kotlin")

    // Use specific constraints
    val template = StructuralSearchTemplate("$obj$.get$Prop$()")
    val propVar = template.getVariable("Prop")
    propVar.addConstraint(VariableConstraint.textMatches(Regex("[A-Z][a-z]+")))

    // Limit results
    val options = SearchOptions(template, scope, maxResults = 100)

    return StructuralSearch.search(project, options).toList()
}
```

### Caching Results

```kotlin
class CachedStructuralSearch(private val project: Project) {
    private val cache = ConcurrentHashMap<String, List<MatchResult>>()

    fun search(template: String): List<MatchResult> {
        return cache.getOrPut(template) {
            performSearch(template)
        }
    }

    private fun performSearch(template: String): List<MatchResult> {
        val searchTemplate = StructuralSearchTemplate(template)
        val options = SearchOptions(
            searchTemplate,
            GlobalSearchScope.projectScope(project)
        )
        return StructuralSearch.search(project, options).toList()
    }

    fun clearCache() {
        cache.clear()
    }
}
```

## Related Topics

- [Inspections](inspections.md) - Local inspections
- [Find Usages](../05-language-support/find-usages.md) - Usage search
- [PSI Navigation](../03-psi/navigating-psi.md) - Navigating PSI
