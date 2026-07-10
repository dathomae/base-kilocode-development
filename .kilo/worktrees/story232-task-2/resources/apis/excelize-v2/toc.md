# excelize v2 API Documentation - Table of Contents

API documentation for `github.com/xuri/excelize/v2` — a mature, high-level Go library for reading, writing, and manipulating Excel (XLSX/XLSM/XLTM/XLTX) files.

**[api-reference.md](api-reference.md)**
: Comprehensive API reference covering the `File` type, cell I/O (`SetCellValue`, `GetCellValue`, `SetCellFloat`, `SetCellStr`), sheet management (`NewSheet`, `DeleteSheet`, `GetSheetList`), row/column operations, formula support (`SetCellFormula`, `CalcCellValue`), file lifecycle (`NewFile`, `OpenFile`, `SaveAs`, `WriteToBuffer`), utility functions (`CoordinatesToCellName`, `ColumnNumberToName`), and complete code examples for generator and verifier patterns.

**[migration-guide.md](migration-guide.md)**
: Side-by-side migration guide from go-ods v1.0.7 to excelize v2. Covers conceptual mapping (Table→Sheet, TableRow/TableCell→cell references), API translation for every go-ods operation used in the ChronoCone codebase, formula syntax differences (ODF bracket notation vs Excel A1 notation), and a rewritten version of the core generator/verifier pattern.
