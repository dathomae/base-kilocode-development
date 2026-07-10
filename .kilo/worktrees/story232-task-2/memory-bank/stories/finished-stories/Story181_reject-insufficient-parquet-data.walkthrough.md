# Story181 Walkthrough and Post-Mortem

## Overview

This story replaced the zero-default behavior in [`ReadParquetToTensor`](src/chronocone_planning_language/src/main/go/io/parquet_reader.go:90) with strict exact-match validation.

Previously, [`ReadParquetToTensor`](src/chronocone_planning_language/src/main/go/io/parquet_reader.go:90) allocated a dense `[]float64` slice initialized to `0.0` and wrote values only for rows present in the parquet file. Unspecified combinations silently defaulted to `0.0`, which violated the [No Manufactured Data](memory-bank/concepts.md) principle.

The new implementation enforces that the parquet file must contain **exactly** one row per valid dimension combination, no duplicates, no extraneous rows, and no unexpected columns. Any deviation produces a descriptive error, and the function never manufactures data.

---

## Design Decisions

### Deferred Duplicate Detection

Duplicates are detected during row processing by marking a `seen []bool` tracker at the computed flat index. Instead of erroring immediately on the first duplicate, the code records only the first duplicate index and continues processing all remaining rows. This deferred approach serves two purposes:

1. It allows the full row count to be accumulated, enabling accurate detection of **excess rows** (which is reported before duplicates).
2. It ensures that if the file is both over-size and contains duplicates, the most actionable error (excess rows) is surfaced first.

### Error Priority Ordering

Validation errors are reported in a strict priority order:

1. **Excess rows** — If `totalRowsRead > totalSize`, the file has more rows than the Cartesian product of dimension sizes. This is reported first because it indicates the most fundamental mismatch.
2. **Duplicates** — If any flat index was visited more than once, a duplicate error is reported with the dimension values of the first duplicate encountered.
3. **Missing combinations** — If `totalRowsRead == totalSize` but some indices were never seen, missing combinations are reported with up to 5 specific examples.

This ordering ensures that the error message points to the root cause most directly. For example, a file with 5 rows for a 2x2 tensor might be "excess rows" (if all 4 combos exist plus an extra duplicate) rather than "missing combos".

### `seen []bool` Tracker

The `seen` slice (line 177) is allocated alongside the `data` slice and tracks which flat indices have been visited. This provides O(1) combination existence checking per row and O(totalSize) final validation for missing combinations, with no additional memory overhead beyond one byte per expected element.

### `dimIndexToValues` Helper

The [`dimIndexToValues`](src/chronocone_planning_language/src/main/go/io/parquet_reader.go:313) helper reverse-computes dimension values from a flat index by walking backward through the shape dimensions, using modulo and division to extract each dimension's index. This is used to produce human-readable dimension values in error messages for both duplicate and missing combination errors, without requiring any additional state to be maintained during row processing.

---

## Validation Checks Implemented

### 1. Column-Count Exactness

Lines 124-142 verify that `schema.NumColumns() == len(dimensions) + 1`. If the parquet file has more columns than expected, the code iterates the column names, identifies which ones are not in the `dimensions` slice and are not `"Value"`, and returns an error listing the unexpected columns.

### 2. Duplicate Combination Detection

Lines 261-283 check the `seen` tracker during row processing. If a flat index has already been visited, the first duplicate index is recorded (line 263). After all rows are processed, if a duplicate was found, [`dimIndexToValues`](src/chronocone_planning_language/src/main/go/io/parquet_reader.go:313) is used to compute the dimension values and a descriptive error is returned.

### 3. Excess Row Count Detection

Lines 275-277 check whether `totalRowsRead > totalSize`. If so, an error is returned stating the expected and actual row counts.

### 4. Missing Combination Detection

Lines 286-303 scan the `seen` slice after processing. Any unvisited index indicates a missing combination. The code collects up to 5 specific missing examples, formatting each as `[Dimension1=value1, Dimension2=value2]`, and returns an error with the total missing count, expected rows, found rows, and the examples.

---

## Tests Added

Seven new test functions were added to [`parquet_reader_test.go`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go):

| Test Function                                                                                                                         | What It Verifies                                                                                                                                    |
| ------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| [`TestReadParquetToTensor_NonExistentFile`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:77)                | Returns an error when the specified parquet file does not exist.                                                                                    |
| [`TestReadParquetToTensor_ExactMatchSucceeds`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:93)             | A parquet file with exactly the right columns, row count, and all unique combinations loads without error and returns correct shape and data.       |
| [`TestReadParquetToTensor_MissingDataReturnsError`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:122)       | A parquet file with fewer rows than the Cartesian product returns an error containing expected count, found count, and missing combination details. |
| [`TestReadParquetToTensor_DuplicateRowsReturnsError`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:152)     | A parquet file containing a duplicate combination returns an error indicating a duplicate or missing problem.                                       |
| [`TestReadParquetToTensor_ExtraRowsReturnsError`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:183)         | A parquet file with all expected combinations plus an extra duplicate row returns an error indicating the row count exceeds the expected count.     |
| [`TestReadParquetToTensor_ExtraColumnsReturnsError`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:214)      | A parquet file with an extra unexpected column returns an error listing the extra column name.                                                      |
| [`TestReadParquetToTensor_ExplicitZeroValueSucceeds`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:244)     | A parquet file where one combination has value `0.0` loads successfully, proving explicit zeros are distinguishable from missing data.              |
| [`TestReadParquetToTensor_ErrorListsMissingCombinations`](src/chronocone_planning_language/src/main/go/io/parquet_reader_test.go:277) | The error message for missing data includes specific missing dimension values (e.g., `"East"`, `"2023"`).                                           |

---

## Changes from Original Behavior

| Aspect                 | Old Behavior                                     | New Behavior                                                                                                             |
| ---------------------- | ------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------ |
| Missing combinations   | Silently default to `0.0`                        | Descriptive error with count and up to 5 missing examples                                                                |
| Duplicate combinations | Last write wins (overwrites data)                | Descriptive error with dimension values of first duplicate                                                               |
| Extra rows             | Extra rows overwrite data or are ignored         | Descriptive error stating expected vs. found row count                                                                   |
| Extra columns          | Ignored                                          | Descriptive error listing unexpected column names                                                                        |
| Data initialization    | `make([]float64, totalSize)` (zero-filled)       | `make([]float64, totalSize)` plus `seen []bool` tracker                                                                  |
| Docstring              | Stated "Unspecified combinations default to 0.0" | States "All expected combinations must be present exactly once; missing, duplicate, or extraneous rows produce an error" |

---

## Post-Mortem

### What Went Well

- **Clear acceptance criteria** in the story file made it straightforward to enumerate the validation checks and test cases needed.
- **Deferred duplicate detection** proved to be a good design choice: by continuing processing after finding a duplicate, we can accurately detect whether the file is also over-size and report the more fundamental "excess rows" error first.
- **Helper function extraction**: [`dimIndexToValues`](src/chronocone_planning_language/src/main/go/io/parquet_reader.go:313) is simple, pure, and reusable for both duplicate and missing-combination error messages.
- **Test coverage**: The 7 new tests cover exact-match success, missing data, duplicates, extra rows, extra columns, explicit zeros, and detailed error messages.
- **Docstring alignment**: The function documentation was updated in the same commit as the implementation, ensuring the contract described in comments matches the actual behavior.

### What Could Be Improved

- **Error message formatting**: The missing-combination error message uses a flat list of examples. For tensors with many dimensions, the examples can become long. A future improvement might group missing combinations by dimension or provide a more structured error type.
- **Performance for very large tensors**: The `seen []bool` tracker requires `O(totalSize)` memory. For extremely high-cardinality dimensions, this could be large. A sparse alternative (e.g., map[int]bool) could be considered, though it would complicate duplicate detection.
- **First duplicate only**: Only the first duplicate index is recorded. If a file has multiple distinct duplicates, only one is reported. Recording all duplicates would require additional memory but could provide a more complete diagnostic.

### Lessons Learned

- **Zero is a valid value, not a sentinel**: The old code conflated "missing" with "zero." Making them distinct via a separate `seen` tracker is a classic pattern that avoids silent data corruption.
- **Error priority matters**: Users debugging a bad parquet file benefit from seeing the most fundamental problem first. Ordering validation checks so that structural problems (wrong column count, wrong row count) are caught before logical problems (duplicates, missing combos) reduces confusion.
- **Continue reading for better diagnostics**: Continuing to process rows after finding a duplicate (rather than erroring immediately) provides more accurate totals for the final error message. The small cost of extra processing pays off in diagnostic quality.

### Technical Debt and Follow-Up Items

- **Structured errors**: Currently all validation failures return `fmt.Errorf`. Consider defining typed errors (e.g., `ParquetSchemaError`, `ParquetDataError`) so callers can programmatically distinguish between schema and data problems.
- **Configurable missing-example limit**: The hardcoded limit of 5 missing examples in the error message (line 291) could be made configurable.
- **Integration test verification**: Task 2 of the story (running `connector`, `synthesis`, and compiler integration tests) is pending. These must pass before the story is fully complete.
- **Performance benchmarks**: No benchmarks exist for `ReadParquetToTensor`. Adding benchmarks would help evaluate the impact of the `seen` tracker and future optimizations.
