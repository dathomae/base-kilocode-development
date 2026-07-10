# excelize v2 API Reference

> **Module:** `github.com/xuri/excelize/v2`  
> **Latest stable:** v2.10.1 (Feb 2026)  
> **License:** BSD 3-Clause  
> **Go version:** 1.24+ (v2.10.x); earlier v2.x versions support older Go  
> **Stars/Importers:** 20K+ stars, 2,400+ importers  
> **Docs:** https://pkg.go.dev/github.com/xuri/excelize/v2 | https://xuri.me/excelize/

## Overview

Excelize provides a **high-level, cell-reference-based API** for creating, reading, and modifying Excel-compatible spreadsheet files. Unlike go-ods (which exposes raw XML structs), excelize abstracts away the file format internals behind a `File` struct with methods like `SetCellValue("Sheet1", "A1", 42)`.

The library supports:
- **Reading and writing** XLSX, XLSM, XLTM, XLTX, XLAM files
- **Formulas** with standard Excel A1 notation and a built-in formula calculation engine (`CalcCellValue`)
- **Streaming** write/read API for very large datasets
- **Charts, pictures, pivot tables, data validation, conditional formatting, sparklines**
- **Cell styling** (fonts, borders, fills, alignment, number formats)
- **In-memory generation** via `NewFile()` + `WriteToBuffer()`

---

## Core Types

### `type File struct`

The central type. Created via `NewFile()` or `OpenFile()`. All spreadsheet operations are methods on `*File`.

```go
f := excelize.NewFile()
defer f.Close()
```

**File lifecycle:**
- `NewFile()` — create a fresh workbook (always has one sheet "Sheet1")
- `OpenFile(filename)` — open an existing file from disk  
- `OpenReader(r)` — open from `io.Reader`
- `SaveAs(filename)` — save to a new path
- `Save()` — save in-place (original path from `OpenFile`)
- `Write(w)` — write to `io.Writer`
- `WriteTo(w)` — synonym for `Write`
- `WriteToBuffer()` — write to `*bytes.Buffer` (in-memory)
- `Close()` — release resources (call `defer f.Close()`)

---

## Sheet Management

### `f.NewSheet(name string) (int, error)`

Creates a new sheet. Returns the sheet index.

```go
idx, err := f.NewSheet("data_a")
```

### `f.DeleteSheet(name string) error`

Deletes a sheet by name. At least one visible sheet must remain.

```go
f.DeleteSheet("Sheet1")
```

### `f.GetSheetList() []string`

Returns the names of all sheets in the workbook.

```go
sheets := f.GetSheetList()
// ["Sheet1", "data_a", "expected_output"]
```

### `f.GetSheetIndex(name string) (int, error)`

Returns the 0-based index of a sheet by name.

```go
idx, _ := f.GetSheetIndex("data_a")
```

### `f.GetSheetName(index int) string`

Returns the sheet name at the given index.

### `f.SetActiveSheet(index int)`

Sets the active (visible on open) sheet.

### `f.SetSheetName(source, target string) error`

Renames a sheet.

### `f.GetSheetMap() map[int]string`

Returns a map of sheet index → sheet name.

### `f.CopySheet(from, to int) error`

Copies a sheet to a new position.

---

## Cell Value I/O

### `f.SetCellValue(sheet, cell string, value interface{}) error`

The primary write method. Accepts any Go type and serializes it appropriately:
- `string` → string cell
- `int`, `int64`, `uint` → integer cell
- `float64` → float cell  
- `bool` → boolean cell
- `time.Time` → date/time cell
- `nil` → empty cell

```go
f.SetCellValue("Sheet1", "A1", "Hello")
f.SetCellValue("Sheet1", "B1", 42)
f.SetCellValue("Sheet1", "C1", 3.14)
f.SetCellValue("Sheet1", "D1", true)
```

### `f.SetCellFloat(sheet, cell string, value float64, precision, bitSize int) error`

Writes a float value with explicit formatting control. Use `-1` for precision to use default formatting.

```go
f.SetCellFloat("Sheet1", "B1", 42.5, -1, 64)
```

### `f.SetCellStr(sheet, cell, value string) error`

Writes a string cell with explicit string type (prevents auto-interpretation as number). Use when you want literal text that looks numeric.

```go
f.SetCellStr("Sheet1", "A1", "00123")  // stored as text, not as 123
```

### `f.SetCellInt(sheet, cell string, value int) error`

Writes an integer value.

```go
f.SetCellInt("Sheet1", "A1", 100)
```

### `f.SetCellBool(sheet, cell string, value bool) error`

Writes a boolean value.

### `f.SetCellDefault(sheet, cell string, value string) error`

Writes a cell with default formatting (no type coercion).

### `f.GetCellValue(sheet, cell string, opts ...Options) (string, error)`

Reads a cell's value as a string. For numeric cells, returns the formatted number as a string. For formula cells, returns the cached/final computed result.

```go
val, err := f.GetCellValue("Sheet1", "A1")
// "Hello" for string, "42" for int, "42.5" for float
num, _ := strconv.ParseFloat(val, 64)
```

### `f.GetCellType(sheet, cell string) (CellType, error)`

Returns the cell's data type. `CellType` constants:
- `CellTypeUnset` — empty/unset
- `CellTypeBool` — boolean
- `CellTypeNumber` — numeric (int or float)
- `CellTypeString` — string
- `CellTypeSharedString` — shared string (internal)
- `CellTypeInlineStr` — inline string
- `CellTypeFormula` — formula
- `CellTypeDate` — date/time

---

## Batch Row/Column Operations

### `f.SetSheetRow(sheet, cell string, slice interface{}) error`

Writes an entire row starting at `cell`, expanding horizontally. The `slice` parameter should be a pointer to a `[]interface{}` or `[]string`.

```go
headers := []interface{}{"Region", "Date", "Value"}
f.SetSheetRow("Sheet1", "A1", &headers)

data := []interface{}{"North", "2024-01-01", 42.5}
f.SetSheetRow("Sheet1", "A2", &data)
```

### `f.SetSheetCol(sheet, cell string, slice interface{}) error`

Writes an entire column starting at `cell`, expanding vertically.

```go
values := []interface{}{"North", "South", "East", "West"}
f.SetSheetCol("Sheet1", "A1", &values)
```

### `f.GetRows(sheet string, opts ...Options) ([][]string, error)`

Returns all rows in a sheet as a `[][]string`. Each inner slice is a row, each element is a cell value (always string).

```go
rows, err := f.GetRows("Sheet1")
for r, row := range rows {
    for c, cell := range row {
        fmt.Printf("Row %d Col %d: %s\n", r, c, cell)
    }
}
```

### `f.GetCols(sheet string, opts ...Options) ([][]string, error)`

Returns all columns in a sheet.

### `f.Rows(sheet string) (*Rows, error)`

Returns a streaming row iterator for memory-efficient processing of large sheets.

```go
rows, _ := f.Rows("Sheet1")
for rows.Next() {
    row, _ := rows.Columns()
    for _, colCell := range row {
        fmt.Print(colCell, "\t")
    }
}
rows.Close()
```

---

## Formula Support

### `f.SetCellFormula(sheet, cell, formula string, opts ...FormulaOpts) error`

Sets a cell to contain a formula using standard Excel A1 notation.

```go
// Simple formula
f.SetCellFormula("Sheet1", "C1", "A1+B1")

// Cross-sheet reference with Excel notation
f.SetCellFormula("expected_output", "D2", "data_a!B2*2")

// IF formula
f.SetCellFormula("expected_output", "E2", `IF(data_a!B2>10, data_a!B2*2, "")`)
```

### `f.GetCellFormula(sheet, cell string) (string, error)`

Returns the formula string stored in a cell (empty string if not a formula).

### `f.CalcCellValue(sheet, cell string, opts ...Options) (string, error)`

Evaluates a formula cell and returns the computed result. This is excelize's built-in formula calculation engine — it can compute values **without opening the file in Excel**.

```go
result, err := f.CalcCellValue("expected_output", "E2")
num, _ := strconv.ParseFloat(result, 64)
```

### `FormulaOpts` struct

```go
type FormulaOpts struct {
    Type   *string  // "array" | "dataTable" | "shared"
    Ref    *string  // shared formula reference range
}
```

---

## Column/Row Reference Utilities

### `excelize.CoordinatesToCellName(col, row int, abs ...bool) (string, error)`

Converts 0-based column and 1-based row to an Excel cell reference.

```go
cell, _ := excelize.CoordinatesToCellName(0, 1)     // "A1"
cell, _ := excelize.CoordinatesToCellName(3, 5)     // "D5"
cell, _ := excelize.CoordinatesToCellName(26, 10)   // "AA10"
```

### `excelize.ColumnNumberToName(num int) (string, error)`

Converts a 1-based column number to a column letter.

```go
name, _ := excelize.ColumnNumberToName(1)   // "A"
name, _ := excelize.ColumnNumberToName(27)  // "AA"
```

### `excelize.ColumnNameToNumber(name string) (int, error)`

Converts a column letter to a 1-based column number.

```go
num, _ := excelize.ColumnNameToNumber("A")   // 1
num, _ := excelize.ColumnNameToNumber("AA")  // 27
```

### `excelize.CellNameToCoordinates(cell string) (int, int, error)`

Parses a cell reference into 0-based column and 1-based row.

```go
col, row, _ := excelize.CellNameToCoordinates("A1")  // col=0, row=1
col, row, _ := excelize.CellNameToCoordinates("D5")  // col=3, row=5
```

### `excelize.JoinCellName(col string, row int) (string, error)`

Joins a column letter and row number into a cell reference.

```go
cell, _ := excelize.JoinCellName("A", 1)  // "A1"
```

---

## File Creation and Persistence

### Creating from scratch

```go
f := excelize.NewFile()
// Default sheet "Sheet1" exists
f.SetCellValue("Sheet1", "A1", "Hello")
f.SaveAs("output.xlsx")
```

### Opening existing

```go
f, err := excelize.OpenFile("existing.xlsx")
f.GetCellValue("Sheet1", "A1")
f.Save()  // overwrites original
```

### In-memory generation (no disk write)

```go
f := excelize.NewFile()
f.SetCellValue("Sheet1", "A1", "Data")
buf, err := f.WriteToBuffer()
// buf.Bytes() is the XLSX file content, can be embedded or sent over network
```

### Opening from a byte slice / reader

```go
// From bytes (e.g., embedded template)
f, err := excelize.OpenReader(bytes.NewReader(rawBytes))
```

---

## Streaming API (for large datasets)

For writing huge worksheets with very low memory overhead:

```go
f := excelize.NewFile()
sw, err := f.NewStreamWriter("Sheet1")

// Write rows
sw.SetRow("A1", []interface{}{"Header1", "Header2", "Header3"})
sw.SetRow("A2", []interface{}{1, "data", 3.14})

// Flush writes the stream data to the sheet
sw.Flush()
f.SaveAs("large.xlsx")
```

---

## Complete Examples

### Example 1: Create a Spreadsheet with Headers and Data

```go
package main

import (
    "fmt"
    "github.com/xuri/excelize/v2"
)

func main() {
    f := excelize.NewFile()
    defer f.Close()

    // Create input data sheet
    idx, _ := f.NewSheet("data_a")
    f.SetActiveSheet(idx)

    // Write headers
    f.SetSheetRow("data_a", "A1", &[]interface{}{"Region", "Value"})
    // Write data rows
    f.SetSheetRow("data_a", "A2", &[]interface{}{"North", 10.0})
    f.SetSheetRow("data_a", "A3", &[]interface{}{"South", 15.0})

    // Delete default sheet
    f.DeleteSheet("Sheet1")

    if err := f.SaveAs("example.xlsx"); err != nil {
        fmt.Println(err)
    }
}
```

### Example 2: Read and Verify a Spreadsheet

```go
package main

import (
    "fmt"
    "strconv"
    "github.com/xuri/excelize/v2"
)

func main() {
    f, err := excelize.OpenFile("example.xlsx")
    if err != nil {
        fmt.Println(err)
        return
    }
    defer f.Close()

    // Get all rows from the expected_output sheet
    rows, _ := f.GetRows("expected_output")

    // rows[0] is header, skip it
    for r := 1; r < len(rows); r++ {
        row := rows[r]
        if len(row) < 4 {
            continue
        }

        // Read the flag column (B = index 1)
        flag := row[1]
        if flag == "final" {
            region := row[0]
            valueStr := row[2]
            value, _ := strconv.ParseFloat(valueStr, 64)
            fmt.Printf("%s: %v\n", region, value)
        }
    }
}
```

### Example 3: Cross-Sheet Formulas with CalcCellValue

```go
package main

import (
    "fmt"
    "strconv"
    "github.com/xuri/excelize/v2"
)

func main() {
    f := excelize.NewFile()
    defer f.Close()

    // Set up input sheet
    f.NewSheet("data_a")
    f.SetCellValue("data_a", "A1", "Region")
    f.SetCellValue("data_a", "B1", "Value")
    f.SetCellValue("data_a", "A2", "North")
    f.SetCellFloat("data_a", "B2", 10, -1, 64)
    f.DeleteSheet("Sheet1")

    // Set up output sheet with formula
    f.NewSheet("expected_output")
    f.SetCellValue("expected_output", "A1", "Region")
    f.SetCellValue("expected_output", "B1", "Value")
    f.SetCellValue("expected_output", "A2", "North")

    // Formula referencing input sheet
    f.SetCellFormula("expected_output", "B2", "data_a!B2*2")

    // Compute the formula result programmatically
    result, _ := f.CalcCellValue("expected_output", "B2")
    computed, _ := strconv.ParseFloat(result, 64)
    fmt.Printf("Computed: %v (expected: 20)\n", computed)

    f.SaveAs("computed.xlsx")
}
```

---

## Important Behaviors and Gotchas

### 1. Row Numbers Are 1-Based, Column Letters Start at A=0

- `SetCellValue("Sheet1", "A1", val)` — row 1, column A
- `CoordinatesToCellName(0, 1)` returns `"A1"` — col 0 → A, row 1 → 1
- `CellNameToCoordinates("A1")` returns `col=0, row=1`

### 2. `GetCellValue` Always Returns a String

For numeric cells, parse with `strconv.ParseFloat`. For formula cells, this returns the cached result (or `"0"` for un-cached formulas).

### 3. `SetCellValue` Auto-Detects Type

Passing `10` (int) writes an integer cell. Passing `10.0` (float64) writes a float cell. Passing `"10"` writes a string cell. Use `SetCellFloat`, `SetCellInt`, `SetCellStr` for explicit type control.

### 4. Formula Cells: Separate Static Value + Formula Pattern

Unlike go-ods, excelize **allows** a formula cell and a separate static value cell to coexist in the same row:
- Set a cell as a formula with `SetCellFormula`
- In a different column, place the statically computed value with `SetCellFloat`
- You can verify by reading the static value cell (no CalcCellValue needed for verification)

### 5. `CalcCellValue` — Built-in Formula Engine

Excelize v2 includes a formula calculation engine. You can call `CalcCellValue("sheet", "C2")` and get the computed result as a string, even for cross-sheet references. This eliminates the need for the "branch/final row" pattern used in go-ods — you can compute and verify formula results entirely in Go.

### 6. `NewFile()` Always Creates "Sheet1"

The default "Sheet1" can be renamed or deleted. Call `f.DeleteSheet("Sheet1")` if you only want custom sheets.

### 7. Error Handling

All methods return `error`. Always check errors, especially for sheet name mismatches and cell reference parsing.

### 8. Go Version Compatibility

v2.10.x requires Go 1.24+. For Go 1.22 compatibility, use an earlier v2 release (v2.7.x through v2.8.x, though exact cutoff varies — check specific release notes). Alternatively, bump your Go toolchain.

---

## Key API Surface Summary

| Operation | Method |
|-----------|--------|
| Create workbook | `excelize.NewFile()` |
| Open file | `excelize.OpenFile(path)` |
| Open from reader | `excelize.OpenReader(r)` |
| Save to disk | `f.SaveAs(path)` |
| Save in-place | `f.Save()` |
| Write to buffer | `f.WriteToBuffer()` |
| Close workbook | `f.Close()` |
| Create sheet | `f.NewSheet(name)` |
| Delete sheet | `f.DeleteSheet(name)` |
| List sheets | `f.GetSheetList()` |
| Write cell | `f.SetCellValue(sheet, cell, val)` |
| Write float | `f.SetCellFloat(sheet, cell, val, prec, bits)` |
| Write string | `f.SetCellStr(sheet, cell, val)` |
| Write row | `f.SetSheetRow(sheet, cell, &slice)` |
| Read cell | `f.GetCellValue(sheet, cell)` |
| Read rows | `f.GetRows(sheet)` |
| Set formula | `f.SetCellFormula(sheet, cell, formula)` |
| Get formula | `f.GetCellFormula(sheet, cell)` |
| Compute formula | `f.CalcCellValue(sheet, cell)` |
| Col→Name | `excelize.ColumnNumberToName(num)` |
| Name→Col | `excelize.ColumnNameToNumber(name)` |
| Coords→Cell | `excelize.CoordinatesToCellName(col, row)` |
| Cell→Coords | `excelize.CellNameToCoordinates(cell)` |
