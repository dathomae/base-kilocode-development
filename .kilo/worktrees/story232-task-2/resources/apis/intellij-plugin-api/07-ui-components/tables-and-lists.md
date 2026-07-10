# Tables and Lists

Tables and lists display structured data in rows and columns or single-column format.

## Overview

### What Are Tables and Lists?

These components provide:

- Structured data display
- Sorting and filtering
- Selection and editing
- Custom rendering

### Component Types

| Type      | Description               |
| --------- | ------------------------- |
| JTable    | Multi-column table        |
| JBList    | Single-column list        |
| JBTable   | IntelliJ-styled table     |
| TableView | Advanced table with model |

## Lists

### Basic List

```kotlin
class MyListPanel : JPanel(BorderLayout()) {
    private val list: JBList<String>

    init {
        // Create list model
        val model = DefaultListModel<String>()
        model.addElement("Item 1")
        model.addElement("Item 2")
        model.addElement("Item 3")

        // Create list
        list = JBList(model)
        list.selectionMode = ListSelectionModel.SINGLE_SELECTION

        // Add selection listener
        list.addListSelectionListener { event ->
            if (!event.valueIsAdjusting) {
                val selected = list.selectedValue
                handleSelection(selected)
            }
        }

        add(JBScrollPane(list), BorderLayout.CENTER)
    }

    private fun handleSelection(item: String?) {
        // Handle selection
    }
}
```

### List with Custom Renderer

```kotlin
class MyListCellRenderer : JPanel(), ListCellRenderer<MyItem> {
    private val nameLabel = JLabel()
    private val descriptionLabel = JLabel()

    init {
        layout = BorderLayout()
        add(nameLabel, BorderLayout.NORTH)
        add(descriptionLabel, BorderLayout.CENTER)
    }

    override fun getListCellRendererComponent(
        list: JList<out MyItem>,
        value: MyItem,
        index: Int,
        isSelected: Boolean,
        cellHasFocus: Boolean
    ): Component {
        nameLabel.text = value.name
        descriptionLabel.text = value.description

        background = if (isSelected) {
            list.selectionBackground
        } else {
            list.background
        }

        foreground = if (isSelected) {
            list.selectionForeground
        } else {
            list.foreground
        }

        return this
    }
}

// Usage
list.cellRenderer = MyListCellRenderer()
```

### List Actions

```kotlin
class MyListPanel : JPanel(BorderLayout()) {
    private val list: JBList<MyItem>

    init {
        // ... list setup

        // Double-click
        list.addMouseListener(object : MouseAdapter() {
            override fun mouseClicked(e: MouseEvent) {
                if (e.clickCount == 2) {
                    val index = list.locationToIndex(e.point)
                    if (index >= 0) {
                        openItem(list.model.getElementAt(index))
                    }
                }
            }
        })

        // Keyboard
        list.inputMap.put(KeyStroke.getKeyStroke(KeyEvent.VK_ENTER, 0), "open")
        list.actionMap.put("open", object : AbstractAction() {
            override fun actionPerformed(e: ActionEvent) {
                list.selectedValue?.let { openItem(it) }
            }
        })
    }
}
```

## Tables

### Basic Table

```kotlin
class MyTablePanel : JPanel(BorderLayout()) {
    private val table: JBTable

    init {
        // Create table model
        val model = object : AbstractTableModel() {
            private val data = listOf(
                arrayOf("Row 1, Col 1", "Row 1, Col 2"),
                arrayOf("Row 2, Col 1", "Row 2, Col 2")
            )

            override fun getRowCount() = data.size
            override fun getColumnCount() = 2
            override fun getValueAt(row: Int, col: Int) = data[row][col]

            override fun getColumnName(col: Int) = when (col) {
                0 -> "Column 1"
                1 -> "Column 2"
                else -> ""
            }
        }

        table = JBTable(model)
        table.rowSelectionAllowed = true

        add(JBScrollPane(table), BorderLayout.CENTER)
    }
}
```

### Table with Custom Model

```kotlin
class MyTableModel : AbstractTableModel() {
    private val items = mutableListOf<MyItem>()

    fun setItems(newItems: List<MyItem>) {
        items.clear()
        items.addAll(newItems)
        fireTableDataChanged()
    }

    fun getItem(row: Int): MyItem = items[row]

    override fun getRowCount() = items.size

    override fun getColumnCount() = 3

    override fun getValueAt(row: Int, col: Int): Any {
        val item = items[row]
        return when (col) {
            0 -> item.name
            1 -> item.value
            2 -> item.status
            else -> ""
        }
    }

    override fun getColumnName(col: Int): String {
        return when (col) {
            0 -> "Name"
            1 -> "Value"
            2 -> "Status"
            else -> ""
        }
    }

    override fun getColumnClass(col: Int): Class<*> {
        return when (col) {
            1 -> Integer::class.java
            else -> String::class.java
        }
    }
}
```

### Table Cell Renderer

```kotlin
class MyTableCellRenderer : DefaultTableCellRenderer() {
    override fun getTableCellRendererComponent(
        table: JTable,
        value: Any,
        isSelected: Boolean,
        hasFocus: Boolean,
        row: Int,
        col: Int
    ): Component {
        super.getTableCellRendererComponent(table, value, isSelected, hasFocus, row, col)

        // Custom rendering based on value
        when (col) {
            2 -> { // Status column
                when (value) {
                    "OK" -> foreground = JBColor.GREEN
                    "Warning" -> foreground = JBColor.YELLOW
                    "Error" -> foreground = JBColor.RED
                }
            }
        }

        return this
    }
}

// Usage
table.setDefaultRenderer(String::class.java, MyTableCellRenderer())
```

### Table Selection

```kotlin
// Single selection
table.selectionModel.selectionMode = ListSelectionModel.SINGLE_SELECTION

// Multiple selection
table.selectionModel.selectionMode = ListSelectionModel.MULTIPLE_INTERVAL_SELECTION

// Selection listener
table.selectionModel.addListSelectionListener { event ->
    if (!event.valueIsAdjusting) {
        val selectedRows = table.selectedRows
        for (row in selectedRows) {
            val item = (table.model as MyTableModel).getItem(row)
            // Process selected item
        }
    }
}
```

## TableView

### Using TableView

```kotlin
class MyTableViewPanel : JPanel(BorderLayout()) {
    private val tableView: TableView<MyItem>

    init {
        // Create column model
        val columns = arrayOf(
            ColumnInfo<MyItem, String>("Name") { it.name },
            ColumnInfo<MyItem, Int>("Value") { it.value },
            ColumnInfo<MyItem, String>("Status") { it.status }
        )

        // Create table view
        val model = ListTableModel<MyItem>(columns)
        tableView = TableView(model)

        // Set items
        model.items = getItems()

        add(JBScrollPane(tableView), BorderLayout.CENTER)
    }
}
```

### Custom Column Info

```kotlin
class MyColumnInfo : ColumnInfo<MyItem, String>("Status") {
    override fun valueOf(item: MyItem): String = item.status

    override fun getRenderer(item: MyItem): TableCellRenderer? {
        return object : DefaultTableCellRenderer() {
            override fun getTableCellRendererComponent(
                table: JTable,
                value: Any,
                isSelected: Boolean,
                hasFocus: Boolean,
                row: Int,
                col: Int
            ): Component {
                super.getTableCellRendererComponent(table, value, isSelected, hasFocus, row, col)

                when (value) {
                    "OK" -> foreground = JBColor.GREEN
                    "Error" -> foreground = JBColor.RED
                }

                return this
            }
        }
    }

    override fun isCellEditable(item: MyItem): Boolean = true

    override fun setValue(item: MyItem, value: String) {
        item.status = value
    }
}
```

## Sorting and Filtering

### Row Sorter

```kotlin
// Enable sorting
table.autoCreateRowSorter = true

// Custom sorter
val sorter = TableRowSorter<MyTableModel>(model)
table.rowSorter = sorter

// Sort by column
sorter.toggleSortOrder(0)

// Set filter
sorter.rowFilter = RowFilter<MyTableModel, Int> { entry ->
    entry.getStringValue(0).contains("filter")
}
```

### Filtering with Search

```kotlin
class FilterableTablePanel : JPanel(BorderLayout()) {
    private val table: JBTable
    private val model: MyTableModel
    private val searchField: JTextField

    init {
        model = MyTableModel()
        table = JBTable(model)
        table.autoCreateRowSorter = true

        searchField = JTextField()
        searchField.document.addDocumentListener(object : DocumentAdapter() {
            override fun textChanged(e: DocumentEvent) {
                updateFilter()
            }
        })

        add(searchField, BorderLayout.NORTH)
        add(JBScrollPane(table), BorderLayout.CENTER)
    }

    private fun updateFilter() {
        val text = searchField.text
        val sorter = table.rowSorter as TableRowSorter<*>

        if (text.isNullOrBlank()) {
            sorter.rowFilter = null
        } else {
            sorter.rowFilter = RowFilter.regexFilter("(?i)$text")
        }
    }
}
```

## Editing

### Cell Editing

```kotlin
class MyTableCellEditor : DefaultCellEditor(JTextField()) {
    override fun getCellEditorValue(): Any {
        return (component as JTextField).text
    }

    override fun stopCellEditing(): Boolean {
        val value = cellEditorValue as String
        if (value.isBlank()) {
            return false // Reject empty values
        }
        return super.stopCellEditing()
    }
}

// Usage
table.setDefaultEditor(String::class.java, MyTableCellEditor())
```

### ComboBox Editor

```kotlin
val comboBox = JComboBox(arrayOf("OK", "Warning", "Error"))
table.getColumnModel().getColumn(2).cellEditor = DefaultCellEditor(comboBox)
```

## Best Practices

### Table Guidelines

1. **Use appropriate model**: Choose model based on data complexity
2. **Implement proper rendering**: Match IDE styling
3. **Support sorting/filtering**: Enable common operations
4. **Handle large datasets**: Use lazy loading for big data
5. **Preserve selection**: Maintain selection across updates

### Example: Complete Table Panel

```kotlin
class MyCompleteTablePanel : JPanel(BorderLayout()) {
    private val table: JBTable
    private val model: MyTableModel

    init {
        model = MyTableModel()
        table = JBTable(model).apply {
            autoCreateRowSorter = true
            selectionMode = ListSelectionModel.SINGLE_SELECTION

            // Set column widths
            columnModel.getColumn(0).preferredWidth = 200
            columnModel.getColumn(1).preferredWidth = 100
            columnModel.getColumn(2).preferredWidth = 80

            // Selection listener
            selectionModel.addListSelectionListener { event ->
                if (!event.valueIsAdjusting) {
                    onSelectionChanged()
                }
            }

            // Double-click
            addMouseListener(object : MouseAdapter() {
                override fun mouseClicked(e: MouseEvent) {
                    if (e.clickCount == 2) {
                        openSelected()
                    }
                }
            })
        }

        // Toolbar
        val toolbar = createToolbar()
        add(toolbar, BorderLayout.NORTH)
        add(JBScrollPane(table), BorderLayout.CENTER)

        // Load data
        loadData()
    }

    private fun createToolbar(): JComponent {
        val group = DefaultActionGroup().apply {
            add(RefreshAction())
            addSeparator()
            add(AddAction())
            add(RemoveAction())
        }

        return ActionManager.getInstance()
            .createActionToolbar("MyTableToolbar", group, true)
            .component
    }

    fun loadData() {
        model.setItems(fetchItems())
    }

    private fun onSelectionChanged() {
        // Update UI based on selection
    }

    private fun openSelected() {
        val row = table.selectedRow
        if (row >= 0) {
            val item = model.getItem(table.convertRowIndexToModel(row))
            openItem(item)
        }
    }
}
```

## Related Topics

- [Tree Components](tree-components.md) - Hierarchical data
- [Tool Windows](tool-windows.md) - Tool window content
- [UI Guidelines](ui-guidelines.md) - Design best practices
