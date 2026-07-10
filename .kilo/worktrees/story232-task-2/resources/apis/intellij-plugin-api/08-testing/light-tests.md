# Light Tests

Light tests are fast, isolated unit tests that don't require a full IDE environment.

## Overview

### What Are Light Tests?

Light tests provide:

- Fast execution
- Minimal setup
- Isolated test environment
- No project required

### When to Use Light Tests

| Use Case               | Suitable |
| ---------------------- | -------- |
| PSI operations         | Yes      |
| Simple inspections     | Yes      |
| Parser tests           | Yes      |
| Multi-file analysis    | No       |
| Project-level features | No       |

## Basic Light Test

### Test Class

```kotlin
class MyLightTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    fun testSimple() {
        // Configure file
        myFixture.configureByText("test.my", "content")

        // Test logic
        val file = myFixture.file
        assertNotNull(file)
    }
}
```

### Test with Caret

```kotlin
fun testWithCaret() {
    myFixture.configureByText("test.my", """
        function test() {
            <caret>
        }
    """.trimIndent())

    // Check caret position
    val offset = myFixture.editor.caretModel.offset
    assertTrue(offset > 0)
}
```

### Test with Selection

```kotlin
fun testWithSelection() {
    myFixture.configureByText("test.my", """
        function <selection>test</selection>() {
        }
    """.trimIndent())

    // Check selection
    val selection = myFixture.editor.selectionModel
    assertEquals("test", selection.selectedText)
}
```

## File Operations

### Creating Files

```kotlin
fun testCreateFile() {
    // Create single file
    val file = myFixture.addFileToProject(
        "src/main.my",
        "content"
    )

    assertNotNull(file)
}

fun testCreateMultipleFiles() {
    // Create multiple files
    myFixture.addFileToProject("src/a.my", "content a")
    myFixture.addFileToProject("src/b.my", "content b")

    // Find files
    val aFile = myFixture.findFileInTempDir("src/a.my")
    assertNotNull(aFile)
}
```

### Creating Directories

```kotlin
fun testCreateDirectory() {
    // Create directory
    val dir = myFixture.tempDirFixture.findOrCreateDir("src/main")

    assertNotNull(dir)
    assertTrue(dir.isDirectory)
}
```

## PSI Testing

### PSI Element Access

```kotlin
fun testPsiElement() {
    myFixture.configureByText("test.my", """
        function myFunction() {
            return 42;
        }
    """.trimIndent())

    // Get PSI file
    val file = myFixture.file as MyPsiFile

    // Find element at offset
    val element = file.findElementAt(10)
    assertNotNull(element)

    // Navigate PSI tree
    val parent = element?.parent
    assertNotNull(parent)
}
```

### PSI Tree Navigation

```kotlin
fun testPsiNavigation() {
    myFixture.configureByText("test.my", """
        class MyClass {
            function method() {}
        }
    """.trimIndent())

    val file = myFixture.file as MyPsiFile

    // Find all functions
    val functions = file.children.filterIsInstance<MyFunction>()
    assertEquals(1, functions.size)

    // Check function name
    assertEquals("method", functions[0].name)
}
```

## Inspection Testing

### Basic Inspection Test

```kotlin
class MyInspectionTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData/inspection"

    fun testWarning() {
        // Configure file
        myFixture.configureByText("test.my", """
            function unusedFunction() {}
        """.trimIndent())

        // Enable inspection
        myFixture.enableInspections(MyInspection::class.java)

        // Check highlights
        myFixture.checkHighlighting()
    }
}
```

### Inspection with Quick Fix

```kotlin
fun testQuickFix() {
    myFixture.configureByText("test.my", """
        function <warning>test</warning>() {}
    """.trimIndent())

    myFixture.enableInspections(MyInspection::class.java)

    // Apply quick fix
    val intention = myFixture.findSingleIntention("Fix description")
    myFixture.launchAction(intention)

    // Check result
    myFixture.checkResult("""
        // Fixed content
    """.trimIndent())
}
```

## Completion Testing

### Basic Completion Test

```kotlin
class MyCompletionTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData/completion"

    fun testBasicCompletion() {
        myFixture.configureByText("test.my", """
            function test() {
                <caret>
            }
        """.trimIndent())

        // Invoke completion
        val completions = myFixture.completeBasic()

        // Check results
        assertNotNull(completions)
        assertTrue(completions.any { it.lookupString == "test" })
    }
}
```

### Completion with Lookup

```kotlin
fun testCompletionSelection() {
    myFixture.configureByText("test.my", """
        function test() {
            te<caret>
        }
    """.trimIndent())

    // Invoke completion
    myFixture.completeBasic()

    // Select item
    myFixture.selectItem(myFixture.lookupElements?.indexOfFirst {
        it.lookupString == "test"
    } ?: 0)

    // Check result
    myFixture.checkResult("""
        function test() {
            test<caret>
        }
    """.trimIndent())
}
```

## Find Usages Testing

```kotlin
class MyFindUsagesTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testFindUsages() {
        // Create files
        myFixture.addFileToProject("a.my", """
            function myFunction() {}
        """.trimIndent())

        myFixture.configureByText("b.my", """
            myFunction();
            myFunction();
        """.trimIndent())

        // Find usages
        val usages = myFixture.findUsages(myFixture.elementAtCaret)

        assertEquals(2, usages.size)
    }
}
```

## Rename Testing

```kotlin
class MyRenameTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testRename() {
        myFixture.configureByText("test.my", """
            function myFunction() {
                myFunction();
            }
        """.trimIndent())

        // Rename element
        myFixture.renameElementAtCaret("newFunction")

        // Check result
        myFixture.checkResult("""
            function newFunction() {
                newFunction();
            }
        """.trimIndent())
    }
}
```

## Testing Actions

```kotlin
class MyActionTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testAction() {
        myFixture.configureByText("test.my", "content")

        // Create and run action
        val action = MyAction()
        myFixture.testAction(action)

        // Check result
        myFixture.checkResult("new content")
    }

    fun testActionAvailability() {
        myFixture.configureByText("test.my", "content")

        val action = MyAction()
        val presentation = myFixture.testAction(action)

        // Check if action is enabled
        assertTrue(presentation.isEnabled)
    }
}
```

## Performance Testing

```kotlin
class MyPerformanceTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testPerformance() {
        // Generate large file
        val content = buildString {
            repeat(10000) {
                appendLine("function func$it() {}")
            }
        }

        myFixture.configureByText("test.my", content)

        // Measure time
        val start = System.currentTimeMillis()

        // Perform operation
        myFixture.enableInspections(MyInspection::class.java)

        val duration = System.currentTimeMillis() - start

        // Assert performance
        assertTrue("Operation took too long: ${duration}ms", duration < 5000)
    }
}
```

## Best Practices

### Light Test Guidelines

1. **Use for unit tests**: Light tests are ideal for isolated tests
2. **Keep tests fast**: Avoid heavy operations
3. **Use test data files**: Store complex test data externally
4. **Clean up in tearDown**: Dispose resources properly
5. **Test edge cases**: Include boundary conditions

### Example: Complete Light Test

```kotlin
class MyCompleteLightTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    private lateinit var myService: MyService

    override fun setUp() {
        super.setUp()
        myService = MyServiceImpl()
    }

    fun testSimpleOperation() {
        // Arrange
        myFixture.configureByText("test.my", "input")

        // Act
        val result = myService.process(myFixture.file)

        // Assert
        assertEquals("expected", result)
    }

    fun testWithCaretPosition() {
        myFixture.configureByText("test.my", """
            function test() {
                <caret>
            }
        """.trimIndent())

        val offset = myFixture.editor.caretModel.offset
        assertTrue(offset > 0)
    }

    fun testInspection() {
        myFixture.configureByText("test.my", """
            <warning>deprecatedFunction</warning>()
        """.trimIndent())

        myFixture.enableInspections(MyInspection::class.java)
        myFixture.checkHighlighting()
    }
}
```

## Related Topics

- [Testing Overview](testing-overview.md) - Testing concepts
- [Heavy Tests](heavy-tests.md) - Full IDE tests
- [Test Data and Fixtures](test-data-and-fixtures.md) - Test data management
