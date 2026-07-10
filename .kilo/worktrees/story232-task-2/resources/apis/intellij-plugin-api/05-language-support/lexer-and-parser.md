# Lexer and Parser

Lexers and parsers convert source code text into a PSI tree. This is a fundamental component of language support.

## Overview

### Lexing vs. Parsing

| Stage  | Input      | Output   | Purpose                          |
| ------ | ---------- | -------- | -------------------------------- |
| Lexer  | Characters | Tokens   | Break text into meaningful units |
| Parser | Tokens     | PSI Tree | Build hierarchical structure     |

### Token Types

Tokens represent the smallest meaningful units:

- Keywords (`if`, `else`, `for`)
- Identifiers (`myVariable`, `MyClass`)
- Literals (`"string"`, `42`)
- Operators (`+`, `-`, `==`)
- Brackets (`(`, `)`, `{`, `}`)

## Implementing a Lexer

### Using JFlex

JFlex is the recommended tool for generating lexers.

#### JFlex File Structure

```
// mylexer.flex
package com.example;

import com.intellij.lexer.FlexLexer;
import com.intellij.psi.tree.IElementType;

%%

%class MyLexer
%implements FlexLexer
%unicode
%function advance
%type IElementType
%eof{ return null;
%eof}

// Character classes
WHITE_SPACE=[ \t\n\r]+
LETTER=[a-zA-Z]
DIGIT=[0-9]
IDENTIFIER={LETTER}({LETTER}|{DIGIT})*
NUMBER={DIGIT}+
STRING=\"([^\"\n]|\\\")*\"

%%

<YYINITIAL> {
  {WHITE_SPACE}        { return MyTokenTypes.WHITE_SPACE; }
  "if"                 { return MyTokenTypes.IF; }
  "else"               { return MyTokenTypes.ELSE; }
  "for"                { return MyTokenTypes.FOR; }
  "while"              { return MyTokenTypes.WHILE; }
  "return"             { return MyTokenTypes.RETURN; }
  "class"              { return MyTokenTypes.CLASS; }
  "function"           { return MyTokenTypes.FUNCTION; }
  {IDENTIFIER}         { return MyTokenTypes.IDENTIFIER; }
  {NUMBER}             { return MyTokenTypes.NUMBER; }
  {STRING}             { return MyTokenTypes.STRING; }
  "+"                  { return MyTokenTypes.PLUS; }
  "-"                  { return MyTokenTypes.MINUS; }
  "=="                 { return MyTokenTypes.EQ; }
  "="                  { return MyTokenTypes.ASSIGN; }
  "("                  { return MyTokenTypes.LPAREN; }
  ")"                  { return MyTokenTypes.RPAREN; }
  "{"                  { return MyTokenTypes.LBRACE; }
  "}"                  { return MyTokenTypes.RBRACE; }
  "//"[^\n]*           { return MyTokenTypes.LINE_COMMENT; }
  "/*"([^*]|\*+[^*/])*\*+\/  { return MyTokenTypes.BLOCK_COMMENT; }
  .                    { return MyTokenTypes.BAD_CHARACTER; }
}
```

#### Generating the Lexer

Add to `build.gradle.kts`:

```kotlin
tasks {
    generateLexer {
        sourceFile.set(file("src/main/grammar/mylexer.flex"))
        targetDir.set("src/main/gen/com/example"))
        targetClass.set("MyLexer")
    }
}
```

### Manual Lexer Implementation

```kotlin
class MyManualLexer : LexerBase() {
    private var buffer: CharSequence? = null
    private var startOffset = 0
    private var endOffset = 0
    private var bufferEnd = 0
    private var state = 0

    override fun start(buffer: CharSequence, startOffset: Int, endOffset: Int, initialState: Int) {
        this.buffer = buffer
        this.startOffset = startOffset
        this.endOffset = startOffset
        this.bufferEnd = endOffset
        this.state = initialState
        advance()
    }

    override fun advance(): IElementType? {
        if (endOffset >= bufferEnd) return null

        startOffset = endOffset
        val char = buffer!![endOffset]

        return when {
            char.isWhitespace() -> {
                while (endOffset < bufferEnd && buffer!![endOffset].isWhitespace()) {
                    endOffset++
                }
                MyTokenTypes.WHITE_SPACE
            }
            char.isLetter() -> {
                while (endOffset < bufferEnd && buffer!![endOffset].isLetterOrDigit()) {
                    endOffset++
                }
                val text = buffer!!.subSequence(startOffset, endOffset).toString()
                when (text) {
                    "if" -> MyTokenTypes.IF
                    "else" -> MyTokenTypes.ELSE
                    "for" -> MyTokenTypes.FOR
                    "while" -> MyTokenTypes.WHILE
                    "return" -> MyTokenTypes.RETURN
                    else -> MyTokenTypes.IDENTIFIER
                }
            }
            char.isDigit() -> {
                while (endOffset < bufferEnd && buffer!![endOffset].isDigit()) {
                    endOffset++
                }
                MyTokenTypes.NUMBER
            }
            else -> {
                endOffset++
                when (char) {
                    '+' -> MyTokenTypes.PLUS
                    '-' -> MyTokenTypes.MINUS
                    '=' -> if (endOffset < bufferEnd && buffer!![endOffset] == '=') {
                        endOffset++
                        MyTokenTypes.EQ
                    } else {
                        MyTokenTypes.ASSIGN
                    }
                    '(' -> MyTokenTypes.LPAREN
                    ')' -> MyTokenTypes.RPAREN
                    '{' -> MyTokenTypes.LBRACE
                    '}' -> MyTokenTypes.RBRACE
                    else -> MyTokenTypes.BAD_CHARACTER
                }
            }
        }
    }

    override fun getTokenType(): IElementType? {
        return if (startOffset >= bufferEnd) null else tokenType
    }

    override fun getTokenStart(): Int = startOffset
    override fun getTokenEnd(): Int = endOffset
    override fun getBufferSequence(): CharSequence = buffer!!
    override fun getBufferEnd(): Int = bufferEnd
    override fun getState(): Int = state
}
```

## Implementing a Parser

### Using Grammar Kit

Grammar Kit is the recommended tool for generating parsers.

#### BNF Grammar File

```
// mygrammar.bnf
{
  parserClass="com.example.MyParser"
  extends="com.intellij.lang.PsiParser"
  psiClassPrefix="My"
  psiImplClassSuffix="Impl"
  psiPackage="com.example.psi"
  psiImplPackage="com.example.psi.impl"
  elementTypeHolderClass="com.example.MyElementTypes"
  elementTypeClass="com.example.MyElementType"
  tokenTypeClass="com.example.MyTokenType"
  tokenTypeFactory="com.example.MyTokenTypeFactory"
}

file ::=
  declaration*

declaration ::=
    classDeclaration
  | functionDeclaration
  | variableDeclaration

classDeclaration ::=
  'class' IDENTIFIER '{' classBody '}'

classBody ::=
  declaration*

functionDeclaration ::=
  'function' IDENTIFIER '(' parameterList? ')' block

parameterList ::=
  parameter (',' parameter)*

parameter ::=
  IDENTIFIER

variableDeclaration ::=
  IDENTIFIER '=' expression

block ::=
  '{' statement* '}'

statement ::=
    ifStatement
  | forStatement
  | whileStatement
  | returnStatement
  | expressionStatement

ifStatement ::=
  'if' '(' expression ')' block ('else' block)?

forStatement ::=
  'for' '(' expression? ';' expression? ';' expression? ')' block

whileStatement ::=
  'while' '(' expression ')' block

returnStatement ::=
  'return' expression?

expressionStatement ::=
  expression

expression ::=
  additiveExpression

additiveExpression ::=
  multiplicativeExpression (('+' | '-') multiplicativeExpression)*

multiplicativeExpression ::=
  primaryExpression (('*' | '/') primaryExpression)*

primaryExpression ::=
    NUMBER
  | STRING
  | IDENTIFIER
  | '(' expression ')'
  | callExpression

callExpression ::=
  IDENTIFIER '(' argumentList? ')'

argumentList ::=
  expression (',' expression)*
```

#### Generating the Parser

Add to `build.gradle.kts`:

```kotlin
tasks {
    generateParser {
        sourceFile.set(file("src/main/grammar/mygrammar.bnf"))
        targetRoot.set("src/main/gen")
        pathToParser.set("com/example/MyParser.java")
        pathToPsiRoot.set("com/example/psi")
    }
}
```

### Manual Parser Implementation

```kotlin
class MyParser : PsiParser {
    override fun parse(root: IElementType, builder: PsiBuilder): ASTNode {
        val rootMarker = builder.mark()

        while (!builder.eof()) {
            parseDeclaration(builder)
        }

        rootMarker.done(root)
        return builder.treeBuilt
    }

    private fun parseDeclaration(builder: PsiBuilder) {
        when (builder.tokenType) {
            MyTokenTypes.CLASS -> parseClassDeclaration(builder)
            MyTokenTypes.FUNCTION -> parseFunctionDeclaration(builder)
            MyTokenTypes.IDENTIFIER -> parseVariableDeclaration(builder)
            else -> builder.error("Expected declaration")
        }
    }

    private fun parseClassDeclaration(builder: PsiBuilder) {
        val marker = builder.mark()

        builder.advanceLexer() // 'class'

        if (builder.tokenType != MyTokenTypes.IDENTIFIER) {
            builder.error("Expected identifier")
        } else {
            builder.advanceLexer()
        }

        if (builder.tokenType != MyTokenTypes.LBRACE) {
            builder.error("Expected '{'")
        } else {
            builder.advanceLexer()
            parseClassBody(builder)
            if (builder.tokenType != MyTokenTypes.RBRACE) {
                builder.error("Expected '}'")
            } else {
                builder.advanceLexer()
            }
        }

        marker.done(MyElementTypes.CLASS)
    }

    private fun parseClassBody(builder: PsiBuilder) {
        while (builder.tokenType != MyTokenTypes.RBRACE && !builder.eof()) {
            parseDeclaration(builder)
        }
    }

    private fun parseFunctionDeclaration(builder: PsiBuilder) {
        val marker = builder.mark()

        builder.advanceLexer() // 'function'

        if (builder.tokenType != MyTokenTypes.IDENTIFIER) {
            builder.error("Expected identifier")
        } else {
            builder.advanceLexer()
        }

        if (builder.tokenType != MyTokenTypes.LPAREN) {
            builder.error("Expected '('")
        } else {
            builder.advanceLexer()
            if (builder.tokenType != MyTokenTypes.RPAREN) {
                parseParameterList(builder)
            }
            if (builder.tokenType != MyTokenTypes.RPAREN) {
                builder.error("Expected ')'")
            } else {
                builder.advanceLexer()
            }
        }

        parseBlock(builder)

        marker.done(MyElementTypes.FUNCTION)
    }

    private fun parseBlock(builder: PsiBuilder) {
        if (builder.tokenType != MyTokenTypes.LBRACE) {
            builder.error("Expected '{'")
            return
        }

        val marker = builder.mark()
        builder.advanceLexer()

        while (builder.tokenType != MyTokenTypes.RBRACE && !builder.eof()) {
            parseStatement(builder)
        }

        if (builder.tokenType != MyTokenTypes.RBRACE) {
            builder.error("Expected '}'")
        } else {
            builder.advanceLexer()
        }

        marker.done(MyElementTypes.BLOCK)
    }
}
```

## Error Handling

### Error Recovery

```kotlin
private fun parseStatement(builder: PsiBuilder) {
    when (builder.tokenType) {
        MyTokenTypes.IF -> parseIfStatement(builder)
        MyTokenTypes.FOR -> parseForStatement(builder)
        MyTokenTypes.WHILE -> parseWhileStatement(builder)
        MyTokenTypes.RETURN -> parseReturnStatement(builder)
        else -> {
            // Try to parse expression, recover on error
            val marker = builder.mark()
            parseExpression(builder)
            marker.done(MyElementTypes.EXPRESSION_STATEMENT)
        }
    }
}
```

### Expected Token Helper

```kotlin
private fun expect(builder: PsiBuilder, type: IElementType, message: String): Boolean {
    if (builder.tokenType == type) {
        builder.advanceLexer()
        return true
    }
    builder.error(message)
    return false
}

// Usage
expect(builder, MyTokenTypes.IDENTIFIER, "Expected identifier")
```

## Related Topics

- [Language Registration](language-registration.md) - Registering languages
- [Syntax Highlighting](syntax-highlighting.md) - Visual styling
- [PSI Elements](../03-psi/psi-elements.md) - PSI architecture
