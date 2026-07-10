# Parser Tests

Parser tests verify that your language parser correctly processes source code.

## Overview

### What Are Parser Tests?

Parser tests verify:

- Correct tokenization
- Proper AST structure
- Error recovery
- Edge case handling

### Test Approach

Parser tests typically use a "golden master" approach:

1. Parse input source code
2. Compare output to expected AST representation

## Basic Parser Test

### Test Class

```kotlin
class MyParserTest : ParsingTestCase(
    "parser",      // Data path
    "my",          // File extension
    MyParserDefinition() // Parser definition
) {
    override fun getTestDataPath(): String = "src/test/testData"

    fun testSimpleExpression() {
        doTest(true) // Compare with expected file
    }
}
```

### Test Data Files

```
testData/parser/
├── simpleExpression.my
└── simpleExpression.txt
```

**simpleExpression.my**:

```
1 + 2
```

**simpleExpression.txt**:

```
MyFile
  BinaryExpression
    NumberLiteral
      1
    Plus
    NumberLiteral
      2
```

## Running Parser Tests

### doTest Method

```kotlin
fun testBasic() {
    // Parse and compare with .txt file
    doTest(true)
}

fun testNoErrors() {
    // Parse and ensure no errors
    doTest(true, false) // compare, ensureNoErrors
}

fun testWithErrors() {
    // Parse with expected errors
    doTest(true, true)
}
```

### Inline Test

```kotlin
fun testInline() {
    val input = "1 + 2"
    val expected = """
        MyFile
          BinaryExpression
            NumberLiteral
              1
            Plus
            NumberLiteral
              2
    """.trimIndent()

    // Parse and check
    val file = createPsiFile("test", input)
    checkParsing(file, expected)
}
```

## AST Verification

### Checking Tree Structure

```kotlin
fun testTreeStructure() {
    val file = myFixture.configureByText("test.my", """
        function test() {
            return 42;
        }
    """.trimIndent())

    // Check root
    assertTrue(file is MyPsiFile)

    // Check children
    val children = file.children
    assertEquals(1, children.size)

    // Check function
    val function = children[0] as? MyFunction
    assertNotNull(function)
    assertEquals("test", function?.name)
}
```

### Element Type Checking

```kotlin
fun testElementTypes() {
    val file = parseFile("x = 1")

    // Check element types
    val assignment = file.firstChild
    assertEquals(MyElementTypes.ASSIGNMENT, assignment.node.elementType)

    val identifier = assignment.firstChild
    assertEquals(MyElementTypes.IDENTIFIER, identifier.node.elementType)
}
```

## Error Recovery Testing

### Testing Error Handling

```kotlin
fun testMissingSemicolon() {
    // Input with error
    doTest(true) // Will show error in output
}
```

**missingSemicolon.my**:

```
function test() {
    return 42
}
```

**missingSemicolon.txt**:

```
MyFile
  Function
    PsiElement(function)
    Identifier(test)
    PsiElement(()
    PsiElement())
    Block
      ReturnStatement
        PsiElement(return)
        NumberLiteral(42)
        <error>PsiElement(})</error>
```

### Checking Error Elements

```kotlin
fun testErrorElements() {
    val file = parseFile("function { }")

    // Find error elements
    val errors = PsiTreeUtil.findChildrenOfType(
        file,
        PsiErrorElement::class.java
    )

    assertTrue(errors.isNotEmpty())
}
```

## Token Testing

### Lexer Tests

```kotlin
class MyLexerTest : LexerTestCase() {
    override fun createLexer(): Lexer = MyLexer()

    override fun getDirPath(): String = "src/test/testData/lexer"

    fun testKeywords() {
        doTest("function return if else")
    }

    fun testNumbers() {
        doTest("123 45.67 0xFF")
    }

    fun testStrings() {
        doTest("\"hello\" 'world'")
    }
}
```

### Token Output Format

**input.my**:

```
function test()
```

**output.txt**:

```
FUNCTION 'function'
WHITESPACE ' '
IDENTIFIER 'test'
LPAREN '('
RPAREN ')'
```

## Expression Parsing

### Testing Expressions

```kotlin
fun testBinaryExpression() {
    doTest("1 + 2 * 3")
}

fun testPrecedence() {
    doTest("1 + 2 * 3 == 7")
}

fun testParentheses() {
    doTest("(1 + 2) * 3")
}
```

### Testing Operator Precedence

```kotlin
fun testOperatorPrecedence() {
    val file = parseFile("a + b * c")

    // Should parse as: a + (b * c)
    val addition = file.firstChild as MyBinaryExpression
    assertEquals("+", addition.operator)

    val left = addition.left as MyIdentifier
    assertEquals("a", left.name)

    val right = addition.right as MyBinaryExpression
    assertEquals("*", right.operator)
}
```

## Statement Parsing

### Testing Statements

```kotlin
fun testIfStatement() {
    doTest("""
        if (x > 0) {
            return x;
        }
    """.trimIndent())
}

fun testWhileStatement() {
    doTest("""
        while (x < 10) {
            x = x + 1;
        }
    """.trimIndent())
}
```

## Declaration Parsing

### Testing Declarations

```kotlin
fun testFunctionDeclaration() {
    doTest("""
        function add(a, b) {
            return a + b;
        }
    """.trimIndent())
}

fun testClassDeclaration() {
    doTest("""
        class MyClass {
            function method() {}
        }
    """.trimIndent())
}
```

## Comment Handling

### Testing Comments

```kotlin
fun testLineComment() {
    doTest("// This is a comment\nx = 1")
}

fun testBlockComment() {
    doTest("/* comment */ x = 1")
}

fun testDocumentationComment() {
    doTest("""
        /**
         * Documentation
         */
        function test() {}
    """.trimIndent())
}
```

## String Literal Parsing

### Testing Strings

```kotlin
fun testSimpleString() {
    doTest("\"hello world\"")
}

fun testEscapeSequences() {
    doTest("\"hello\\nworld\\t!\"")
}

fun testInterpolation() {
    doTest("\"Hello, ${name}!\"")
}
```

## Best Practices

### Parser Test Guidelines

1. **Test all constructs**: Cover all language features
2. **Test edge cases**: Empty inputs, malformed code
3. **Test error recovery**: Ensure parser handles errors gracefully
4. **Keep expected files updated**: Regenerate when grammar changes
5. **Use descriptive names**: Name tests after what they test

### Example: Complete Parser Test

```kotlin
class MyCompleteParserTest : ParsingTestCase(
    "parser",
    "my",
    MyParserDefinition()
) {
    override fun getTestDataPath(): String = "src/test/testData"

    // Expressions
    fun testLiteral() = doTest(true)
    fun testBinaryExpression() = doTest(true)
    fun testUnaryExpression() = doTest(true)
    fun testCallExpression() = doTest(true)

    // Statements
    fun testIfStatement() = doTest(true)
    fun testWhileStatement() = doTest(true)
    fun testReturnStatement() = doTest(true)

    // Declarations
    fun testFunctionDeclaration() = doTest(true)
    fun testClassDeclaration() = doTest(true)
    fun testVariableDeclaration() = doTest(true)

    // Error cases
    fun testMissingBrace() = doTest(true)
    fun testInvalidToken() = doTest(true)
    fun testIncompleteExpression() = doTest(true)

    // Edge cases
    fun testEmptyFile() = doTest(true)
    fun testOnlyComments() = doTest(true)
    fun testNestedStructures() = doTest(true)
}
```

## Related Topics

- [Lexer and Parser](../05-language-support/lexer-and-parser.md) - Parser implementation
- [Testing Overview](testing-overview.md) - Testing concepts
- [Highlighting Tests](highlighting-tests.md) - Syntax highlighting tests
