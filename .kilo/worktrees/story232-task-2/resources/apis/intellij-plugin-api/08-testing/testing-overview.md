# Testing Overview

Testing is essential for ensuring plugin reliability and correctness.

## Overview

### Why Test Plugins?

Testing provides:

- Confidence in code changes
- Early bug detection
- Documentation of expected behavior
- Regression prevention

### Test Types

| Type               | Description                      |
| ------------------ | -------------------------------- |
| Light Tests        | Fast, isolated unit tests        |
| Heavy Tests        | Full IDE integration tests       |
| Parser Tests       | Language parsing verification    |
| Completion Tests   | Code completion testing          |
| Highlighting Tests | Syntax highlighting verification |

## Test Setup

### Dependencies

Add test dependencies to `build.gradle.kts`:

```kotlin
dependencies {
    testImplementation("com.intellij:platform-test-framework")
    testImplementation("junit:junit:4.13.2")
    testImplementation("org.mockito:mockito-core:4.0.0")
}
```

### Test Directory Structure

```
src/
├── main/
│   └── kotlin/
│       └── com/example/
└── test/
    ├── kotlin/
    │   └── com/example/
    │       ├── MyLightTest.kt
    │       └── MyHeavyTest.kt
    └── testData/
        └── parsing/
            └── test1.my
```

## Test Base Classes

### LightPlatformCodeInsightFixtureTestCase

For most plugin tests:

```kotlin
class MyLightTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    fun testSimple() {
        // Test code
    }
}
```

### BasePlatformTestCase

For project-level tests:

```kotlin
class MyProjectTest : BasePlatformTestCase() {
    override fun setUp() {
        super.setUp()
        // Setup code
    }

    fun testProject() {
        // Test code
    }
}
```

## Test Fixtures

### What Are Fixtures?

Fixtures provide:

- Test project setup
- File creation utilities
- Editor manipulation
- PSI access

### Common Fixtures

| Fixture             | Purpose               |
| ------------------- | --------------------- |
| `myFixture`         | Main fixture accessor |
| `myFixture.file`    | Current PSI file      |
| `myFixture.editor`  | Current editor        |
| `myFixture.project` | Current project       |

### Using Fixtures

```kotlin
class MyFixtureTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testFileCreation() {
        // Create file
        val file = myFixture.addFileToProject(
            "test.my",
            "content"
        )

        // Access file
        assertNotNull(file)
        assertEquals("test.my", file.name)
    }

    fun testEditor() {
        // Configure editor
        myFixture.configureByText(
            "test.my",
            "content<caret>"
        )

        // Check caret position
        val caret = myFixture.editor.caretModel
        assertEquals(7, caret.offset)
    }
}
```

## Test Configuration

### Plugin Configuration

```kotlin
class MyTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun setUp() {
        super.setUp()

        // Register extension
        myFixture.registerExtension(
            MyExtensionPoint::class.java,
            MyExtension()
        )
    }
}
```

### Test Service Registration

```kotlin
class MyServiceTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun setUp() {
        super.setUp()

        // Register service
        registerApplicationService(
            MyService::class.java,
            MyServiceImpl()
        )
    }

    protected fun <T : Any> registerApplicationService(
        serviceClass: Class<T>,
        implementation: T
    ) {
        ApplicationManager.getApplication()
            .replaceService(serviceClass, implementation, testRootDisposable)
    }
}
```

## Assertions

### Standard Assertions

```kotlin
import org.junit.Assert.*

fun testAssertions() {
    // Equality
    assertEquals(expected, actual)

    // Boolean
    assertTrue(condition)
    assertFalse(condition)

    // Null checks
    assertNotNull(value)
    assertNull(value)

    // Same instance
    assertSame(expected, actual)
}
```

### IntelliJ Assertions

```kotlin
import com.intellij.testFramework.assertions.Assertions.assertThat

fun testIntelliJAssertions() {
    // PSI assertions
    assertThat(psiFile).hasText("expected content")

    // Collection assertions
    assertThat(list).hasSize(3)
    assertThat(list).contains("item")

    // String assertions
    assertThat(text).contains("substring")
    assertThat(text).startsWith("prefix")
}
```

## Test Data

### Test Data Files

Store test data in `src/test/testData`:

```
testData/
├── parsing/
│   ├── simple.my
│   └── complex.my
├── completion/
│   └── basic.my
└── highlighting/
    └── errors.my
```

### Loading Test Data

```kotlin
class MyTestDataTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData"

    fun testLoadFile() {
        // Load file content
        val content = loadFile("parsing/simple.my")

        // Use in test
        myFixture.configureByText("test.my", content)
    }

    protected fun loadFile(path: String): String {
        return File(testDataPath, path).readText()
    }
}
```

## Running Tests

### Running from IDE

1. Right-click test class or method
2. Select "Run" or "Debug"
3. View results in test runner

### Running from Command Line

```bash
# Run all tests
./gradlew test

# Run specific test
./gradlew test --tests "com.example.MyTest"

# Run with verbose output
./gradlew test --info
```

### Test Configuration

```kotlin
// build.gradle.kts
tasks.test {
    useJUnit()

    testLogging {
        events("passed", "skipped", "failed")
    }

    jvmArgs = listOf(
        "-Xmx2g",
        "-XX:+HeapDumpOnOutOfMemoryError"
    )
}
```

## Best Practices

### Test Guidelines

1. **Test one thing**: Each test should verify one behavior
2. **Use descriptive names**: Test names should describe what is tested
3. **Keep tests independent**: Tests should not depend on each other
4. **Clean up resources**: Dispose resources in tearDown
5. **Use test data files**: Keep test content in separate files

### Example: Complete Test

```kotlin
class MyCompleteTest : LightPlatformCodeInsightFixtureTestCase() {
    private lateinit var myService: MyService

    override fun getTestDataPath(): String = "src/test/testData"

    override fun setUp() {
        super.setUp()

        // Setup service
        myService = MyServiceImpl()
        registerApplicationService(MyService::class.java, myService)
    }

    override fun tearDown() {
        // Cleanup
        myService = TODO()
        super.tearDown()
    }

    fun testSimpleOperation() {
        // Arrange
        val input = "test input"

        // Act
        val result = myService.process(input)

        // Assert
        assertEquals("expected output", result)
    }

    fun testWithFile() {
        // Load test data
        val content = loadFile("testInput.my")

        // Configure file
        myFixture.configureByText("test.my", content)

        // Perform test
        val psiFile = myFixture.file
        assertNotNull(psiFile)
    }

    fun testWithCaret() {
        // Configure with caret marker
        myFixture.configureByText("test.my", "content<caret>")

        // Check caret position
        val offset = myFixture.editor.caretModel.offset
        assertEquals(7, offset)
    }
}
```

## Related Topics

- [Light Tests](light-tests.md) - Light test framework
- [Heavy Tests](heavy-tests.md) - Heavy test framework
- [Test Data and Fixtures](test-data-and-fixtures.md) - Test data management
