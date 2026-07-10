# Editor Features

This section covers how to extend and interact with the IntelliJ Platform editor, including documents, code completion, intentions, and formatting.

## Topics

**[editor-basics.md](editor-basics.md)**
: Editor fundamentals and core concepts

**[document-model.md](document-model.md)**
: Working with documents and text content

**[code-completion.md](code-completion.md)**
: Implementing custom code completion

**[intentions.md](intentions.md)**
: Creating intention actions (quick fixes)

**[code-formatting.md](code-formatting.md)**
: Custom code formatting rules

**[code-folding.md](code-folding.md)**
: Implementing code folding regions

## Key Concepts

### Editor

The editor is the central component for displaying and editing text. It provides:

- Text display and editing
- Caret and selection management
- Scrolling and viewport control
- Inline elements (inlays, gutter markers)

### Document

Documents represent the text content being edited. They are separate from the editor and can be shared between multiple editors.

### Code Completion

Code completion provides suggestions as users type. Plugins can contribute custom completion providers for specific languages or contexts.

### Intentions

Intentions are quick fixes or enhancements suggested by the IDE. They appear as lightbulb icons and can modify code automatically.

## Recommended Reading Order

1. Start with [editor-basics.md](editor-basics.md) to understand the editor architecture
2. Read [document-model.md](document-model.md) for text manipulation
3. Study [code-completion.md](code-completion.md) for completion features
4. Explore [intentions.md](intentions.md) for quick fixes
5. Learn [code-formatting.md](code-formatting.md) for formatting rules
6. Reference [code-folding.md](code-folding.md) for folding support
