# Completion Tests

Completion tests verify that code completion works correctly for your language.

## Overview

### What Are Completion Tests?

Completion tests verify:

- Completion item appearance
- Completion item sorting
- Insertion behavior
- Context sensitivity

### Test Approach

1. Position caret at completion point
2. Invoke completion
3. Verify completion items
4. Optionally test insertion

## Basic Completion Test

### Test Class

```kotlin
class MyCompletionTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData/completion"

    fun testBasicCompletion() {
        myFixture.configureByText("test.my", """
            function test() {
                <caret>
            }
        """.trimIndent())

        val completions = myFixture.completeBasic()

        assertNotNull(completions)
        assertTrue(completions.any { it.lookupString == "test" })
    }
}
```

### Testing Completion Items

```kotlin
fun testCompletionItems() {
    myFixture.configureByText("test.my", "str<caret>")

    val completions = myFixture.completeBasic()

    // Check specific items exist
    assertNotNull(completions)

    val strings = completions.map { it.lookupString }
    assertTrue(strings.contains("string"))
    assertTrue(strings.contains("strlen"))
}
```

## Completion Types

### Basic Completion

```kotlin
fun testBasicCompletion() {
    myFixture.configureByText("test.my", "x<caret>")
    val completions = myFixture.completeBasic()
    // Test basic completions
}
```

### Smart Completion

```kotlin
fun testSmartCompletion() {
    myFixture.configureByText("test.my", """
        function test(): string {
            return <caret>
        }
    """.trimIndent())

    val completions = myFixture.completeBasic()

    // Smart completion should filter by type
    assertTrue(completions.all {
        it.lookupString.startsWith("\"") ||
        it.lookupString == "test"
    })
}
```

### Statement Completion

```kotlin
fun testStatementCompletion() {
    myFixture.configureByText("test.my", """
        if (true) <caret>
    """.trimIndent())

    myFixture.type('\n') // Trigger statement completion

    myFixture.checkResult("""
        if (true) {
            <caret>
        }
    """.trimIndent())
}
```

## Testing Lookup Elements

### Checking Lookup Properties

```kotlin
fun testLookupProperties() {
    myFixture.configureByText("test.my", "my<caret>")

    val completions = myFixture.completeBasic()
    val item = completions.find { it.lookupString == "myFunction" }

    assertNotNull(item)

    // Check icon
    assertNotNull(item?.icon)

    // Check type text
    assertEquals("void", item?.typeText)

    // Check tail text
    assertEquals("()", item?.tailText)
}
```

### Checking Priorities

```kotlin
fun testCompletionPriority() {
    myFixture.configureByText("test.my", "te<caret>")

    val completions = myFixture.completeBasic()

    // Check that local variables come before functions
    val testVar = completions.indexOfFirst { it.lookupString == "testVar" }
    val testFunc = completions.indexOfFirst { it.lookupString == "testFunc" }

    assertTrue(testVar < testFunc)
}
```

## Testing Insertion

### Basic Insertion

```kotlin
fun testInsertion() {
    myFixture.configureByText("test.my", "my<caret>")

    myFixture.completeBasic()
    myFixture.selectItem(0) // Select first item

    myFixture.checkResult("myFunction<caret>")
}
```

### Insertion with Parameters

```kotlin
fun testParameterInsertion() {
    myFixture.configureByText("test.my", "myFunc<caret>")

    myFixture.completeBasic()
    myFixture.selectItem(0)

    // Should insert with parameter placeholders
    myFixture.checkResult("myFunction(<caret>)")
}
```

### Insertion with Imports

```kotlin
fun testImportInsertion() {
    myFixture.addFileToProject("lib.my", """
        export function importedFunction() {}
    """.trimIndent())

    myFixture.configureByText("test.my", "imported<caret>")

    myFixture.completeBasic()
    myFixture.selectItem(0)

    // Should add import statement
    myFixture.checkResult("""
        import { importedFunction } from './lib';
        importedFunction<caret>
    """.trimIndent())
}
```

## Context-Sensitive Completion

### Member Completion

```kotlin
fun testMemberCompletion() {
    myFixture.configureByText("test.my", """
        class MyClass {
            function method1() {}
            function method2() {}
        }

        val obj = MyClass();
        obj.<caret>
    """.trimIndent())

    val completions = myFixture.completeBasic()

    // Should show class members
    assertTrue(completions.any { it.lookupString == "method1" })
    assertTrue(completions.any { it.lookupString == "method2" })
}
```

### Scope-Based Completion

```kotlin
fun testLocalVariables() {
    myFixture.configureByText("test.my", """
        function test() {
            val localVar = 1;
            val otherVar = 2;
            <caret>
        }
    """.trimIndent())

    val completions = myFixture.completeBasic()

    // Should include local variables
    assertTrue(completions.any { it.lookupString == "localVar" })
    assertTrue(completions.any { it.lookupString == "otherVar" })
}
```

### Keyword Completion

```kotlin
fun testKeywordCompletion() {
    myFixture.configureByText("test.my", "fun<caret>")

    val completions = myFixture.completeBasic()

    assertTrue(completions.any { it.lookupString == "function" })
}
```

## Testing Completion Filtering

### Type-Based Filtering

```kotlin
fun testTypeFiltering() {
    myFixture.configureByText("test.my", """
        function getString(): string { return ""; }
        function getNumber(): number { return 0; }

        val s: string = <caret>
    """.trimIndent())

    val completions = myFixture.completeBasic()

    // Should prefer string-returning functions
    val stringFunc = completions.indexOfFirst { it.lookupString == "getString" }
    val numberFunc = completions.indexOfFirst { it.lookupString == "getNumber" }

    assertTrue(stringFunc < numberFunc)
}
```

### Name Filtering

```kotlin
fun testNameFiltering() {
    myFixture.configureByText("test.my", "getStr<caret>")

    val completions = myFixture.completeBasic()

    // Should only show items matching "getStr"
    assertTrue(completions.all {
        it.lookupString.lowercase().contains("getstr")
    })
}
```

## Testing Completion Contributors

### Custom Contributor Test

```kotlin
class MyCompletionContributorTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testContributor() {
        myFixture.configureByText("test.my", "custom<caret>")

        // Register contributor
        myFixture.registerCompletionContributor(
            MyCompletionContributor()
        )

        val completions = myFixture.completeBasic()

        // Check custom completions
        assertTrue(completions.any { it.lookupString == "customItem" })
    }
}
```

## Completion Test Patterns

### Test Data Pattern

```kotlin
fun testFromFile() {
    // Load test data
    myFixture.configureByFile("completion/basic.my")

    // Invoke completion
    val completions = myFixture.completeBasic()

    // Check results
    checkCompletion(completions, listOf("item1", "item2", "item3"))
}

private fun checkCompletion(
    completions: Array<LookupElement>?,
    expected: List<String>
) {
    assertNotNull(completions)
    val actual = completions!!.map { it.lookupString }
    assertEquals(expected.sorted(), actual.sorted())
}
```

## Best Practices

### Completion Test Guidelines

1. **Test all contexts**: Different completion locations
2. **Test filtering**: Verify type and name filtering
3. **Test insertion**: Check inserted text
4. **Test imports**: Verify automatic imports
5. **Test priorities**: Check item ordering

### Example: Complete Completion Test

```kotlin
class MyCompleteCompletionTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData/completion"

    // Basic completions
    fun testVariableCompletion() {
        myFixture.configureByText("test.my", """
            val myVariable = 1;
            my<caret>
        """.trimIndent())

        val completions = myFixture.completeBasic()
        assertTrue(completions.any { it.lookupString == "myVariable" })
    }

    fun testFunctionCompletion() {
        myFixture.configureByText("test.my", """
            function myFunction() {}
            my<caret>
        """.trimIndent())

        val completions = myFixture.completeBasic()
        val func = completions.find { it.lookupString == "myFunction" }

        assertNotNull(func)
        assertEquals("()", func?.tailText)
    }

    fun testMemberCompletion() {
        myFixture.configureByText("test.my", """
            class MyClass {
                val property = 1;
                function method() {}
            }

            val obj = MyClass();
            obj.<caret>
        """.trimIndent())

        val completions = myFixture.completeBasic()

        assertTrue(completions.any { it.lookupString == "property" })
        assertTrue(completions.any { it.lookupString == "method" })
    }

    fun testCompletionInsertion() {
        myFixture.configureByText("test.my", """
            function test() {}
            te<caret>
        """.trimIndent())

        myFixture.completeBasic()
        myFixture.selectItem(0)

        myFixture.checkResult("""
            function test() {}
            test<caret>
        """.trimIndent())
    }
}
```

## Related Topics

- [Code Completion](../04-editor-features/code-completion.md) - Completion implementation
- [Testing Overview](testing-overview.md) - Testing concepts
- [Light Tests](light-tests.md) - Light test framework
