# Highlighting Tests

Highlighting tests verify syntax highlighting and semantic annotation.

## Overview

### What Are Highlighting Tests?

Highlighting tests verify:

- Syntax highlighting colors
- Error highlighting
- Warning annotations
- Semantic highlighting

### Test Approach

1. Configure file with markers
2. Enable inspections/annotators
3. Check highlighting matches markers

## Basic Highlighting Test

### Test Class

```kotlin
class MyHighlightingTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData/highlighting"

    fun testSimpleHighlighting() {
        myFixture.configureByText("test.my", """
            function test() {
                return 42;
            }
        """.trimIndent())

        myFixture.checkHighlighting()
    }
}
```

### Highlighting Markers

Use markers to indicate expected highlights:

| Marker       | Description       |
| ------------ | ----------------- |
| `<warning>`  | Warning highlight |
| `<error>`    | Error highlight   |
| `<info>`     | Info highlight    |
| `</warning>` | End marker        |

```kotlin
fun testWarningHighlighting() {
    myFixture.configureByText("test.my", """
        function test() {
            val unused = <warning>1</warning>;
        }
    """.trimIndent())

    myFixture.enableInspections(UnusedVariableInspection::class.java)
    myFixture.checkHighlighting()
}
```

## Error Highlighting

### Syntax Errors

```kotlin
fun testSyntaxError() {
    myFixture.configureByText("test.my", """
        function test( {
            <error>}</error>
    """.trimIndent())

    myFixture.checkHighlighting()
}
```

### Semantic Errors

```kotlin
fun testUndefinedVariable() {
    myFixture.configureByText("test.my", """
        function test() {
            return <error>undefinedVar</error>;
        }
    """.trimIndent())

    myFixture.enableInspections(UndeclaredVariableInspection::class.java)
    myFixture.checkHighlighting()
}
```

### Type Errors

```kotlin
fun testTypeError() {
    myFixture.configureByText("test.my", """
        function test(): number {
            return <error>"string"</error>;
        }
    """.trimIndent())

    myFixture.enableInspections(TypeMismatchInspection::class.java)
    myFixture.checkHighlighting()
}
```

## Warning Highlighting

### Unused Code

```kotlin
fun testUnusedVariable() {
    myFixture.configureByText("test.my", """
        function test() {
            val <warning>x</warning> = 1;
        }
    """.trimIndent())

    myFixture.enableInspections(UnusedVariableInspection::class.java)
    myFixture.checkHighlighting()
}
```

### Deprecated Code

```kotlin
fun testDeprecatedUsage() {
    myFixture.configureByText("test.my", """
        @deprecated
        function oldFunction() {}

        function test() {
            <warning>oldFunction</warning>();
        }
    """.trimIndent())

    myFixture.enableInspections(DeprecationInspection::class.java)
    myFixture.checkHighlighting()
}
```

## Info Highlighting

### Information Annotations

```kotlin
fun testInfoHighlighting() {
    myFixture.configureByText("test.my", """
        function test() {
            // <info>TODO: implement</info>
        }
    """.trimIndent())

    myFixture.enableInspections(TodoCommentInspection::class.java)
    myFixture.checkHighlighting()
}
```

## Strict Highlighting

### Strict Mode

```kotlin
fun testStrictHighlighting() {
    myFixture.configureByText("test.my", """
        function test() {
            val <warning descr="Unused variable">x</warning> = 1;
        }
    """.trimIndent())

    myFixture.enableInspections(UnusedVariableInspection::class.java)

    // Strict checking - must match exactly
    myFixture.checkHighlighting(true, false, true)
}
```

### Checking Descriptions

```kotlin
fun testHighlightingDescription() {
    myFixture.configureByText("test.my", """
        function test() {
            val <warning descr="Variable 'x' is never used">x</warning> = 1;
        }
    """.trimIndent())

    myFixture.enableInspections(UnusedVariableInspection::class.java)
    myFixture.checkHighlighting(true, false, true)
}
```

## Syntax Highlighting

### Testing Text Attributes

```kotlin
fun testSyntaxHighlighting() {
    myFixture.configureByText("test.my", """
        function test() {
            return 42;
        }
    """.trimIndent())

    // Get highlighter
    val highlighter = myFixture.editor.highlighter

    // Check keyword highlighting
    val keywordOffset = myFixture.file.text.indexOf("function")
    val attributes = highlighter.getTokenType(keywordOffset)

    assertEquals(MyTokenTypes.FUNCTION_KEYWORD, attributes)
}
```

### Testing Colors

```kotlin
fun testHighlightingColors() {
    myFixture.configureByText("test.my", """
        val x = "string";
    """.trimIndent())

    // Get text attributes for string
    val stringStart = myFixture.file.text.indexOf("\"")
    val highlighter = myFixture.editor.highlighter

    // Check string is highlighted correctly
    val tokenType = highlighter.getTokenType(stringStart)
    assertEquals(MyTokenTypes.STRING_LITERAL, tokenType)
}
```

## Annotator Testing

### Testing Annotators

```kotlin
class MyAnnotatorTest : LightPlatformCodeInsightFixtureTestCase() {
    fun testAnnotator() {
        myFixture.configureByText("test.my", """
            function test() {
                // TODO: implement
            }
        """.trimIndent())

        // Annotator runs automatically
        val annotations = myFixture.doHighlighting()

        // Check for TODO annotation
        assertTrue(annotations.any {
            it.description?.contains("TODO") == true
        })
    }
}
```

### Custom Annotator Test

```kotlin
fun testCustomAnnotator() {
    myFixture.configureByText("test.my", """
        @MyAnnotation
        function test() {}
    """.trimIndent())

    val annotations = myFixture.doHighlighting()

    // Check for custom annotation
    val myAnnotation = annotations.find {
        it.severity == HighlightSeverity.INFORMATION &&
        it.text == "MyAnnotation"
    }

    assertNotNull(myAnnotation)
}
```

## Multiple Highlighting Types

### Mixed Highlights

```kotlin
fun testMixedHighlighting() {
    myFixture.configureByText("test.my", """
        function test() {
            val <warning>x</warning> = <error>undefinedVar</error>;
            // <info>TODO: fix</info>
        }
    """.trimIndent())

    myFixture.enableInspections(
        UnusedVariableInspection::class.java,
        UndeclaredVariableInspection::class.java,
        TodoCommentInspection::class.java
    )

    myFixture.checkHighlighting()
}
```

## Highlighting Pass Testing

### Testing Highlighting Passes

```kotlin
fun testHighlightingPass() {
    myFixture.configureByText("test.my", """
        function test() {
            return 42;
        }
    """.trimIndent())

    // Run highlighting passes
    myFixture.doHighlighting()

    // Check results
    val highlights = myFixture.editor.markupModel.allHighlighters
    assertTrue(highlights.isNotEmpty())
}
```

## Quick Fix Testing

### Testing Quick Fixes

```kotlin
fun testQuickFix() {
    myFixture.configureByText("test.my", """
        function test() {
            val <warning>x</warning> = 1;
        }
    """.trimIndent())

    myFixture.enableInspections(UnusedVariableInspection::class.java)

    // Find quick fix
    val intention = myFixture.findSingleIntention("Remove unused variable")
    assertNotNull(intention)

    // Apply quick fix
    myFixture.launchAction(intention)

    // Check result
    myFixture.checkResult("""
        function test() {
        }
    """.trimIndent())
}
```

## Best Practices

### Highlighting Test Guidelines

1. **Use markers**: Clearly indicate expected highlights
2. **Test all severities**: Errors, warnings, info
3. **Test descriptions**: Verify highlight messages
4. **Test quick fixes**: Ensure fixes work correctly
5. **Test edge cases**: Empty files, malformed code

### Example: Complete Highlighting Test

```kotlin
class MyCompleteHighlightingTest : LightPlatformCodeInsightFixtureTestCase() {
    override fun getTestDataPath(): String = "src/test/testData/highlighting"

    // Syntax errors
    fun testMissingBrace() {
        myFixture.configureByText("test.my", """
            function test() {
                <error>}</error>
        """.trimIndent())
        myFixture.checkHighlighting()
    }

    // Semantic errors
    fun testUndefinedVariable() {
        myFixture.configureByText("test.my", """
            function test() {
                return <error>undefinedVar</error>;
            }
        """.trimIndent())
        myFixture.enableInspections(UndeclaredVariableInspection::class.java)
        myFixture.checkHighlighting()
    }

    // Warnings
    fun testUnusedVariable() {
        myFixture.configureByText("test.my", """
            function test() {
                val <warning descr="Variable 'x' is never used">x</warning> = 1;
            }
        """.trimIndent())
        myFixture.enableInspections(UnusedVariableInspection::class.java)
        myFixture.checkHighlighting(true, false, true)
    }

    // Info
    fun testTodoComment() {
        myFixture.configureByText("test.my", """
            function test() {
                // <info>TODO: implement</info>
            }
        """.trimIndent())
        myFixture.enableInspections(TodoCommentInspection::class.java)
        myFixture.checkHighlighting()
    }

    // Quick fix
    fun testQuickFixRemovesUnusedVariable() {
        myFixture.configureByText("test.my", """
            function test() {
                val <warning>x</warning> = 1;
            }
        """.trimIndent())

        myFixture.enableInspections(UnusedVariableInspection::class.java)

        val intention = myFixture.findSingleIntention("Remove variable")
        myFixture.launchAction(intention)

        myFixture.checkResult("""
            function test() {
            }
        """.trimIndent())
    }
}
```

## Related Topics

- [Annotators](../06-analysis-features/annotators.md) - Annotator implementation
- [Inspections](../06-analysis-features/inspections.md) - Inspection implementation
- [Testing Overview](testing-overview.md) - Testing concepts
