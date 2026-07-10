# Tree Components

Tree components display hierarchical data structures in the IDE.

## Overview

### What Are Tree Components?

Tree components provide:

- Hierarchical data visualization
- Expandable/collapsible nodes
- Selection and navigation
- Drag and drop support
- Context menus

### Common Uses

| Use Case       | Description           |
| -------------- | --------------------- |
| Project View   | File system tree      |
| Structure View | Code structure tree   |
| Find Results   | Search result tree    |
| Custom Views   | Plugin-specific trees |

## Basic Tree

### Creating a Simple Tree

```kotlin
class MyTreePanel : JPanel(BorderLayout()) {
    private val tree: JTree

    init {
        // Create tree model
        val root = DefaultMutableTreeNode("Root")
        root.add(DefaultMutableTreeNode("Child 1"))
        root.add(DefaultMutableTreeNode("Child 2"))

        val model = DefaultTreeModel(root)
        tree = JTree(model)

        // Configure tree
        tree.isRootVisible = true
        tree.showsRootHandles = true

        // Add to panel
        add(JBScrollPane(tree), BorderLayout.CENTER)
    }
}
```

### Tree with Custom Nodes

```kotlin
class MyTreeNode(val data: MyData) : DefaultMutableTreeNode(data) {
    override fun toString(): String = data.displayName

    fun updateChildren() {
        removeAllChildren()
        for (child in data.children) {
            add(MyTreeNode(child))
        }
    }
}
```

## IntelliJ Tree Components

### Using TreeBuilder

```kotlin
class MyTreeStructure : AbstractTreeStructure() {
    private val root: MyRoot = MyRoot()

    override fun getRootElement(): Any = root

    override fun getChildElements(element: Any): Array<Any> {
        return when (element) {
            is MyRoot -> element.children.toTypedArray()
            is MyNode -> element.children.toTypedArray()
            else -> emptyArray()
        }
    }

    override fun getParentElement(element: Any): Any? {
        return when (element) {
            is MyNode -> element.parent
            else -> null
        }
    }

    override fun commit() {}

    override fun hasSomethingToCommit(): Boolean = false

    override fun createDescriptor(element: Any, parent: ItemDescriptor?): ItemDescriptor {
        return MyNodeDescriptor(element, parent)
    }
}
```

## Tree Selection

### Handling Selection

```kotlin
class MyTreePanel : JPanel(BorderLayout()) {
    private val tree: JTree

    init {
        tree.addTreeSelectionListener { event ->
            val node = event.path.lastPathComponent
            handleSelection(node)
        }
    }

    private fun handleSelection(node: Any) {
        when (node) {
            is MyTreeNode -> {
                val data = node.data
                // Handle selection
            }
        }
    }
}
```

### Getting Selected Elements

```kotlin
fun getSelectedNodes(tree: JTree): List<MyData> {
    val nodes = mutableListOf<MyData>()

    for (path in tree.selectionPaths ?: emptyArray()) {
        val node = path.lastPathComponent
        if (node is MyTreeNode) {
            nodes.add(node.data)
        }
    }

    return nodes
}
```

## Tree Actions

### Double-Click Action

```kotlin
tree.addMouseListener(object : MouseAdapter() {
    override fun mouseClicked(e: MouseEvent) {
        if (e.clickCount == 2) {
            val path = tree.getPathForLocation(e.x, e.y)
            path?.let { handleDoubleClick(it) }
        }
    }
})
```

### Context Menu

```kotlin
tree.componentPopupMenu = JPopupMenu().apply {
    add(JMenuItem("Open").apply {
        addActionListener { openSelected() }
    })
    add(JMenuItem("Delete").apply {
        addActionListener { deleteSelected() }
    })
}
```

## Tree Expansion

### Expanding Nodes

```kotlin
fun expandAll(tree: JTree) {
    var row = 0
    while (row < tree.rowCount) {
        tree.expandRow(row)
        row++
    }
}

fun collapseAll(tree: JTree) {
    for (row in tree.rowCount - 1 downTo 0) {
        tree.collapseRow(row)
    }
}
```

## Tree Rendering

### Custom Cell Renderer

```kotlin
class MyTreeCellRenderer : DefaultTreeCellRenderer() {
    override fun getTreeCellRendererComponent(
        tree: JTree,
        value: Any,
        sel: Boolean,
        expanded: Boolean,
        leaf: Boolean,
        row: Int,
        hasFocus: Boolean
    ): Component {
        super.getTreeCellRendererComponent(tree, value, sel, expanded, leaf, row, hasFocus)

        if (value is MyTreeNode) {
            val data = value.data
            icon = data.icon
            text = data.displayName
            toolTipText = data.tooltip
        }

        return this
    }
}
```

## Best Practices

1. **Lazy loading**: Load children on demand
2. **Background updates**: Update tree off EDT
3. **Preserve state**: Remember expansion and selection
4. **Provide actions**: Support context menu and keyboard
5. **Use appropriate icons**: Match IDE conventions

## Related Topics

- [Tables and Lists](tables-and-lists.md) - Data display
- [Tool Windows](tool-windows.md) - Tool window content
- [Actions System](../01-core-concepts/actions-system.md) - Action registration
