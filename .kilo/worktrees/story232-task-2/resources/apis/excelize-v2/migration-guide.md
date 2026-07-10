# Migration Guide: go-ods v1.0.7 → excelize v2

> **Target module:** `github.com/xuri/excelize/v2`  
> **Source module:** `github.com/AlexJarrah/go-ods` (v1.0.7)

This guide maps every go-ods v1.0.7 pattern used in the ChronoCone `expected_ods_generator` package to its excelize v2 equivalent.

---

## Conceptual Mapping

| go-ods Concept | excelize v2 Equivalent |
|----------------|------------------------|
| `ods.ODS` struct (holds all content) | `*excelize.File` (central handle) |
| `Content.Body.Spreadsheet.Table[i]` | Sheet accessed by name: `f.SetSheetRow("sheet", ...)` |
| `Table.Name` | Sheet name (string parameter on every method) |
| `Table.TableRow[]` | Rows accessed via `f.GetRows("sheet")` or written via `f.SetSheetRow` |
| `TableRow.TableCell[]` | Cell reference strings: `"A1"`, `"B2"`, etc. |
| `TableCell.P` (display text) | Value returned by `f.GetCellValue("sheet", "A1")` |
| `TableCell.Value` (raw value attr) | Parsed from `f.GetCellValue` via `strconv.ParseFloat` |
| `TableCell.ValueType` ("float", "string") | Auto-detected; explicit via `SetCellFloat`/`SetCellStr` |
| `TableCell.Formula` with `of:=` prefix | `f.SetCellFormula("sheet", "A1", formula)` — standard Excel syntax |
| `ods.Uncompress(content, 20)` | Not needed — excelize handles cell expansion internally |
| Template file requirement | `NewFile()` creates from scratch; no template needed |
| Cross-sheet: `[$SheetName.ColRow]` | Cross-sheet: `SheetName!ColRow` (Excel A1 notation) |

---

## Operation-by-Operation Migration

### 1. Creating a New Spreadsheet

**go-ods (template-based):**
```go
import ods "github.com/AlexJarrah/go-ods"

data, files, err := ods.Read("template.ods")
data.Content = ods.Uncompress(data.Content, 20)

// Rename default sheet
data.Content.Body.Spreadsheet.Table[0].Name = "data_a"
data.Content.Body.Spreadsheet.Table[0].TableRow = nil

// Add a new sheet
data.Content.Body.Spreadsheet.Table = append(
    data.Content.Body.Spreadsheet.Table,
    ods.Table{Name: "expected_output"},
)

// Save
err = ods.Write("output.ods", data, files)
```

**excelize (from scratch):**
```go
import "github.com/xuri/excelize/v2"

f := excelize.NewFile()
defer f.Close()

// Create sheets (default "Sheet1" exists)
f.NewSheet("data_a")
f.NewSheet("expected_output")
f.DeleteSheet("Sheet1")  // remove default

// Save
err := f.SaveAs("output.xlsx")
```

### 2. Writing a Header Row

**go-ods:**
```go
sheet := &data.Content.Body.Spreadsheet.Table[sheetIdx]
headerRow := ods.TableRow{
    TableCell: []ods.TableCell{
        {P: "Region", ValueType: "string"},
        {P: "Value", ValueType: "string"},
    },
}
sheet.TableRow = append(sheet.TableRow, headerRow)
```

**excelize:**
```go
headers := []interface{}{"Region", "Value"}
f.SetSheetRow("data_a", "A1", &headers)
```

### 3. Writing a Data Row with Mixed Types

**go-ods:**
```go
dataRow := ods.TableRow{
    TableCell: []ods.TableCell{
        {P: "North", ValueType: "string"},
        {ValueType: "float", Value: "10", P: "10"},
    },
}
sheet.TableRow = append(sheet.TableRow, dataRow)
```

**excelize:**
```go
// Option A: SetSheetRow with mixed interface{} types
row := []interface{}{"North", 10.0}
f.SetSheetRow("data_a", "A2", &row)

// Option B: Set individual cells (more explicit)
f.SetCellValue("data_a", "A2", "North")
f.SetCellFloat("data_a", "B2", 10.0, -1, 64)
```

### 4. Writing a Formula Cell

**go-ods (ODF `of:=` prefix + bracket notation):**
```go
ods.TableCell{
    ValueType: "float",
    Formula:   "of:=[$data_a.B2]*2",
    Value:     "0",
    P:         "",
}
```

**excelize (standard Excel notation):**
```go
f.SetCellFormula("expected_output", "D2", "data_a!B2*2")
```

### 5. Writing an IF Formula

**go-ods:**
```go
ods.TableCell{
    ValueType: "string",
    Formula:   `of:=IF([$data_a.B2]=10,"OK","Not OK")`,
    Value:     "",
    P:         "",
}
```

**excelize:**
```go
f.SetCellFormula("expected_output", "E2", `IF(data_a!B2=10,"OK","Not OK")`)
```

### 6. Finding a Sheet

**go-ods:**
```go
var outSheet *ods.Table
for i := range data.Content.Body.Spreadsheet.Table {
    if data.Content.Body.Spreadsheet.Table[i].Name == "expected_output" {
        outSheet = &data.Content.Body.Spreadsheet.Table[i]
        break
    }
}
```

**excelize:**
```go
// Sheets are accessed by name directly — no lookup needed
// Just use the sheet name string:
f.SetCellValue("expected_output", "A1", "value")

// If you need to check existence:
idx, err := f.GetSheetIndex("expected_output")
if err != nil {
    // sheet doesn't exist
}
```

### 7. Reading Cell Values (Verification)

**go-ods:**
```go
// Iterate rows manually, filter by flag column
for r := 1; r < len(outSheet.TableRow); r++ {
    row := outSheet.TableRow[r]
    flag := row.TableCell[1].P
    if flag == "final" {
        region := row.TableCell[0].P
        valueStr := row.TableCell[2].Value
        value, _ := strconv.ParseFloat(valueStr, 64)
        fmt.Printf("%s: %v\n", region, value)
    }
}
```

**excelize:**
```go
// GetRows returns the entire sheet as [][]string
rows, _ := f.GetRows("expected_output")
for r := 1; r < len(rows); r++ {
    row := rows[r]
    if len(row) < 3 {
        continue
    }
    flag := row[1]  // flag column
    if flag == "final" {
        region := row[0]
        value, _ := strconv.ParseFloat(row[2], 64)
        fmt.Printf("%s: %v\n", region, value)
    }
}
```

### 8. Formula Evaluation (Computed Value)

**go-ods (not possible in code — requires Calc):**
```go
// Cannot evaluate formulas programmatically
// Must use "branch/final" row pattern:
//   branch row: formula (human verifiable in Calc)
//   final row: static value (machine verifiable)
```

**excelize (`CalcCellValue` — evaluate formulas in Go):**
```go
// Compute any formula cell programmatically
result, err := f.CalcCellValue("expected_output", "E2")
if err != nil {
    // handle error
}
computed, _ := strconv.ParseFloat(result, 64)

// The branch/final pattern is optional with excelize
// You can compute and verify formulas entirely in Go
```

### 9. Column Letter Conversion

**go-ods (custom helper):**
```go
func colLetter(col int) string {
    result := ""
    col++
    for col > 0 {
        col--
        result = string(rune('A'+col%26)) + result
        col /= 26
    }
    return result
}
```

**excelize (built-in):**
```go
// 0-based column to letter
cell, _ := excelize.CoordinatesToCellName(col, 1)  // gives "A21"
// cell is like "A1", col part is "A"

// Or directly:
name, _ := excelize.ColumnNumberToName(col + 1)  // 1-based!
// col=0 → "A", col=25 → "Z", col=26 → "AA"
```

**Important:** `ColumnNumberToName` takes a **1-based** column number. `CoordinatesToCellName` takes a **0-based** column number.

### 10. In-Memory Generation (No Disk File)

**go-ods (requires zip reader context):**
```go
buf := new(bytes.Buffer)
err := ods.WriteTo(buf, data, files)
```

**excelize (straightforward buffer):**
```go
buf, err := f.WriteToBuffer()
// buf.Bytes() contains the complete XLSX file
```

### 11. Opening from Bytes (Template Embedding)

**go-ods:**
```go
raw := embeddedTemplate
data, files, err := ods.ReadFrom(bytes.NewReader(raw), int64(len(raw)))
data.Content = ods.Uncompress(data.Content, 20)
```

**excelize:**
```go
f, err := excelize.OpenReader(bytes.NewReader(rawBytes))
```

---

## Formula Syntax Translation

| go-ods (ODF Bracket Notation) | excelize (Excel A1 Notation) |
|-------------------------------|------------------------------|
| `of:=[$data_a.A2]` | `data_a!A2` |
| `of:=[$data_a.A2]+[$data_b.A2]` | `data_a!A2+data_b!A2` |
| `of:=IF([$data_a.A2]=10,[$data_a.A2]*2,\"\")` | `IF(data_a!A2=10,data_a!A2*2,"")` |
| `of:=AND([$data_a.A2]>=5,[$data_b.B2]<10)` | `AND(data_a!A2>=5,data_b!B2<10)` |
| `of:=IF(AND([$data_a.A2]>=5,[$data_b.B2]<10),100,0)` | `IF(AND(data_a!A2>=5,data_b!B2<10),100,0)` |

**Key differences:**
- No `of:=` prefix in excelize
- Sheet references use `!` instead of `$.` prefix and `.` separator
- Excel uses standard `SheetName!CellRef` notation
- Column letters and row numbers are identical between ODF and Excel

### Column Letter Mapping (same in both)

| Column Index | ODF/Excel Letter |
|-------------|-----------------|
| 0 | A |
| 1 | B |
| 2 | C |
| ... | ... |
| 25 | Z |
| 26 | AA |

### Row Number Mapping

| go-ods struct index | ODF formula row | Excel formula row |
|---------------------|----------------|-------------------|
| 0 (header row) | 1 | 1 |
| 1 (first data row) | 2 | 2 |
| n | n+1 | n+1 |

Both use the same 1-based row numbering in formulas. The header row at struct index 0 is row 1 in spreadsheet formulas.

---

## Complete Generator Rewrite (Side-by-Side)

### go-ods Generator Pattern

```go
func GenerateODS(..., data *ods.ODS) error {
    // 1. Set up input sheets
    sheet := &data.Content.Body.Spreadsheet.Table[0]
    sheet.Name = "data_a"
    sheet.TableRow = append(sheet.TableRow, ods.TableRow{
        TableCell: []ods.TableCell{
            {P: "Region", ValueType: "string"},
            {P: "Value", ValueType: "string"},
        },
    })
    sheet.TableRow = append(sheet.TableRow, ods.TableRow{
        TableCell: []ods.TableCell{
            {P: "North", ValueType: "string"},
            {ValueType: "float", Value: "10", P: "10"},
        },
    })

    // 2. Set up output sheet
    outSheet := ods.Table{Name: "expected_output"}
    data.Content.Body.Spreadsheet.Table = append(
        data.Content.Body.Spreadsheet.Table, outSheet)
    outPtr := &data.Content.Body.Spreadsheet.Table[len(...)-1]

    // 3. Write headers
    outPtr.TableRow = append(outPtr.TableRow, ods.TableRow{
        TableCell: []ods.TableCell{
            {P: "Region", ValueType: "string"},
            {P: "Flag", ValueType: "string"},
            {P: "Value", ValueType: "string"},
        },
    })

    // 4. Write formula row
    outPtr.TableRow = append(outPtr.TableRow, ods.TableRow{
        TableCell: []ods.TableCell{
            {P: "North", ValueType: "string"},
            {P: "branch", ValueType: "string"},
            {
                ValueType: "float",
                Formula:   "of:=[$data_a.B2]*2",
                Value:     "0",
                P:         "",
            },
        },
    })

    // 5. Write final (static) verification row
    outPtr.TableRow = append(outPtr.TableRow, ods.TableRow{
        TableCell: []ods.TableCell{
            {P: "North", ValueType: "string"},
            {P: "final", ValueType: "string"},
            {ValueType: "float", Value: "20", P: "20"},
        },
    })

    return nil
}
```

### excelize Generator Pattern

```go
func GenerateXLSX(...) (*excelize.File, error) {
    f := excelize.NewFile()

    // 1. Set up input sheet
    f.SetSheetName("Sheet1", "data_a")  // rename default

    // Write header row (row 1)
    f.SetSheetRow("data_a", "A1", &[]interface{}{"Region", "Value"})

    // Write data row (row 2)
    f.SetSheetRow("data_a", "A2", &[]interface{}{"North", 10.0})

    // 2. Create output sheet
    f.NewSheet("expected_output")

    // 3. Write headers (row 1)
    f.SetSheetRow("expected_output", "A1",
        &[]interface{}{"Region", "Flag", "Value"})

    // 4. Write formula row (row 2)
    f.SetCellValue("expected_output", "A2", "North")
    f.SetCellValue("expected_output", "B2", "branch")

    // Cross-sheet formula in Excel A1 notation
    f.SetCellFormula("expected_output", "C2", "data_a!B2*2")

    // 5. Write static verification value (row 3)
    // Note: with CalcCellValue, this row may not be needed at all
    f.SetCellValue("expected_output", "A3", "North")
    f.SetCellValue("expected_output", "B3", "final")
    f.SetCellFloat("expected_output", "C3", 20.0, -1, 64)

    return f, nil
}
```

### excelize Generator Pattern (Without Branch/Final — Using CalcCellValue)

```go
func GenerateXLSX(...) (*excelize.File, error) {
    f := excelize.NewFile()

    // (Set up input sheet as above...)

    // Output sheet — single row with formula
    f.NewSheet("expected_output")
    f.SetSheetRow("expected_output", "A1",
        &[]interface{}{"Region", "Metric", "Computed Value", "Expected Value"})

    f.SetCellValue("expected_output", "A2", "North")
    f.SetCellValue("expected_output", "B2", "Double")
    f.SetCellFormula("expected_output", "C2", "data_a!B2*2")
    f.SetCellFloat("expected_output", "D2", 20.0, -1, 64)

    // Verify programmatically (no Calc needed):
    computed, _ := f.CalcCellValue("expected_output", "C2")
    val, _ := strconv.ParseFloat(computed, 64)
    if val == 20.0 {
        fmt.Println("PASS")
    }

    return f, nil
}
```

---

## Summary of Benefits

| Aspect | go-ods v1.0.7 | excelize v2 |
|--------|--------------|-------------|
| Template file required | Yes | No (`NewFile()`) |
| Formula evaluation in Go | No | Yes (`CalcCellValue`) |
| Column letter conversion | Custom helper | Built-in (`CoordinatesToCellName`) |
| Branch/final row pattern | Required | Optional (use `CalcCellValue` instead) |
| Row iteration | Manual index + bounds check | `GetRows()` returns `[][]string` |
| Cell write API | `TableCell{P: ..., Value: ..., ValueType: ...}` | `SetCellValue("sheet", "A1", val)` |
| ZIP/namespace management | Must track `*zip.ReadCloser` | Fully internal |
| In-memory generation | `WriteTo(w, data, files)` | `WriteToBuffer()` |
| Formula syntax | ODF `of:=[$Sheet.ColRow]` | Excel `Sheet!ColRow` |
