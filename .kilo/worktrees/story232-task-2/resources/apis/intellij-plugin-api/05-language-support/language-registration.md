# Language Registration

Registering a custom language is the first step in creating language support. This involves defining the language, file type, and basic infrastructure.

## Defining a Language

### Language Class

Create a class extending `Language`:

```kotlin
class MyLanguage private constructor() : Language("MyLanguage") {
    companion object {
        @JvmStatic
        val INSTANCE = MyLanguage()
    }
}
```

### Language Properties

```kotlin
class MyLanguage private constructor() : Language("MyLanguage") {
    override fun isCaseSensitive(): Boolean = true

    override fun getDisplayName(): String = "My Language"

    // Associated file types
    override fun getAssociatedFileType(): FileType? = MyFileType.INSTANCE
}
```

## File Type Definition

### FileType Implementation

```kotlin
class MyFileType private constructor() : FileType {
    override fun getName(): String = "My File"

    override fun getDescription(): String = "My language file"

    override fun getDefaultExtension(): String = "myf"

    override fun getIcon(): Icon = AllIcons.FileTypes.Text

    override fun isBinary(): Boolean = false

    companion object {
        @JvmStatic
        val INSTANCE = MyFileType()
    }
}
```

### FileTypeFactory

```kotlin
class MyFileTypeFactory : FileTypeFactory() {
    override fun createFileTypes(fileTypes: FileTypeConsumer) {
        fileTypes.consume(MyFileType.INSTANCE, "myf;mylang")
    }
}
```

## Registering Extensions

### plugin.xml Configuration

```xml
<extensions defaultExtensionNs="com.intellij">
    <!-- Language definition -->
    <lang.parserDefinition
        language="MyLanguage"
        implementationClass="com.example.MyParserDefinition"/>

    <!-- File type -->
    <fileType
        name="My File"
        implementationClass="com.example.MyFileType"
        fieldName="INSTANCE"
        language="MyLanguage"
        extensions="myf;mylang"/>

    <!-- Syntax highlighter -->
    <lang.syntaxHighlighterFactory
        language="MyLanguage"
        implementationClass="com.example.MySyntaxHighlighterFactory"/>
</extensions>
```

## Parser Definition

### Implementing ParserDefinition

```kotlin
class MyParserDefinition : ParserDefinition {
    override fun getLanguage(): Language = MyLanguage.INSTANCE

    override fun getFileNodeType(): IFileElementType {
        return MyElementTypes.FILE
    }

    override fun getCommentTokens(): TokenSet {
        return MyTokenSets.COMMENTS
    }

    override fun getStringLiteralElements(): TokenSet {
        return MyTokenSets.STRINGS
    }

    override fun createLexer(project: Project?): Lexer {
        return MyLexer()
    }

    override fun createParser(project: Project?): PsiParser {
        return MyParser()
    }

    override fun createFile(viewProvider: FileViewProvider): PsiFile {
        return MyFileImpl(viewProvider)
    }

    override fun spaceExistenceTypeBetweenTokens(left: ASTNode, right: ASTNode): ParserDefinition.SpaceRequirements {
        return ParserDefinition.SpaceRequirements.MAY
    }
}
```

## Element Types

### Token Types

```kotlin
object MyTokenTypes {
    // Keywords
    val IF = MyTokenType("IF", "if")
    val ELSE = MyTokenType("ELSE", "else")
    val FOR = MyTokenType("FOR", "for")
    val WHILE = MyTokenType("WHILE", "while")
    val RETURN = MyTokenType("RETURN", "return")

    // Literals
    val STRING = MyTokenType("STRING", "string literal")
    val NUMBER = MyTokenType("NUMBER", "number literal")
    val IDENTIFIER = MyTokenType("IDENTIFIER", "identifier")

    // Operators
    val PLUS = MyTokenType("PLUS", "+")
    val MINUS = MyTokenType("MINUS", "-")
    val EQ = MyTokenType("EQ", "==")
    val ASSIGN = MyTokenType("ASSIGN", "=")

    // Brackets
    val LPAREN = MyTokenType("LPAREN", "(")
    val RPAREN = MyTokenType("RPAREN", ")")
    val LBRACE = MyTokenType("LBRACE", "{")
    val RBRACE = MyTokenType("RBRACE", "}")
}
```

### Element Types

```kotlin
object MyElementTypes {
    val FILE = IFileElementType(MyLanguage.INSTANCE)

    // Declarations
    val CLASS = MyElementType("CLASS")
    val FUNCTION = MyElementType("FUNCTION")
    val VARIABLE = MyElementType("VARIABLE")

    // Statements
    val IF_STATEMENT = MyElementType("IF_STATEMENT")
    val FOR_STATEMENT = MyElementType("FOR_STATEMENT")
    val RETURN_STATEMENT = MyElementType("RETURN_STATEMENT")

    // Expressions
    val BINARY_EXPRESSION = MyElementType("BINARY_EXPRESSION")
    val CALL_EXPRESSION = MyElementType("CALL_EXPRESSION")
    val REFERENCE_EXPRESSION = MyElementType("REFERENCE_EXPRESSION")
}
```

### Token Sets

```kotlin
object MyTokenSets {
    val KEYWORDS = TokenSet.create(
        MyTokenTypes.IF,
        MyTokenTypes.ELSE,
        MyTokenTypes.FOR,
        MyTokenTypes.WHILE,
        MyTokenTypes.RETURN
    )

    val COMMENTS = TokenSet.create(
        MyTokenTypes.LINE_COMMENT,
        MyTokenTypes.BLOCK_COMMENT
    )

    val STRINGS = TokenSet.create(
        MyTokenTypes.STRING
    )

    val OPERATORS = TokenSet.create(
        MyTokenTypes.PLUS,
        MyTokenTypes.MINUS,
        MyTokenTypes.EQ,
        MyTokenTypes.ASSIGN
    )
}
```

## PSI Elements

### Base Element Types

```kotlin
class MyTokenType(debugName: String, val text: String) : IElementType(debugName, MyLanguage.INSTANCE)

class MyElementType(debugName: String) : IElementType(debugName, MyLanguage.INSTANCE)
```

### PSI Element Implementation

```kotlin
interface MyNamedElement : PsiNameIdentifierOwner {
    val nameIdentifier: PsiIdentifier?
}

class MyClassImpl(node: ASTNode) : MyNamedElementImpl(node), MyClass {
    override val nameIdentifier: PsiIdentifier?
        get() = findChildByType(MyTokenTypes.IDENTIFIER)

    override fun getName(): String? = nameIdentifier?.text

    override fun setName(name: String): PsiElement {
        nameIdentifier?.replace(MyElementFactory.createIdentifier(name))
        return this
    }
}
```

### PSI File Implementation

```kotlin
class MyFileImpl(viewProvider: FileViewProvider) : PsiFileBase(viewProvider, MyLanguage.INSTANCE) {
    override fun getFileType(): FileType = MyFileType.INSTANCE

    override fun toString(): String = "My File"

    val declarations: List<MyDeclaration>
        get() = findChildrenByClass(MyDeclaration::class.java).toList()
}
```

## Element Factory

### Creating Elements Programmatically

```kotlin
object MyElementFactory {
    fun createFile(project: Project, text: String): MyFile {
        return PsiFileFactory.getInstance(project)
            .createFileFromText("dummy.myf", MyLanguage.INSTANCE, text) as MyFile
    }

    fun createIdentifier(project: Project, name: String): PsiIdentifier {
        val file = createFile(project, "class $name {}")
        return file.findChildByClass(MyClass::class.java)?.nameIdentifier!!
    }

    fun createClass(project: Project, name: String): MyClass {
        val file = createFile(project, "class $name {}")
        return file.findChildByClass(MyClass::class.java)!!
    }
}
```

## Related Topics

- [Lexer and Parser](lexer-and-parser.md) - Implementing parsing
- [Syntax Highlighting](syntax-highlighting.md) - Visual styling
- [PSI Elements](../03-psi/psi-elements.md) - PSI architecture
