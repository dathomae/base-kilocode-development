# Reading Parquet Files with Go Arrow

The `pqarrow` package provides several ways to read Parquet files into Arrow data structures. You can read an entire file into an Arrow Table at once, or you can read it progressively using a `FileReader` or `RecordReader`.

## Reading an Entire File into a Table

The simplest way to read a Parquet file is to use the `ReadTable` convenience function. This function reads the entire file into an `arrow.Table`.

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/arrow/go/v16/arrow/memory"
	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleReadTable() {
	// Open the Parquet file
	f, err := os.Open("example.parquet")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Create a Parquet reader
	rdr, err := file.NewParquetReader(f)
	if err != nil {
		panic(err)
	}
	defer rdr.Close()

	// Set up Arrow read properties
	arrProps := pqarrow.ArrowReadProperties{
		Parallel:  true,
		BatchSize: 1024,
	}

	// Read the entire file into an Arrow Table
	ctx := context.Background()
	mem := memory.NewGoAllocator()

	// Note: ReadTable takes a parquet.ReaderAtSeeker, which os.File implements
	tbl, err := pqarrow.ReadTable(ctx, f, rdr.MetaData().GetReaderProperties(), arrProps, mem)
	if err != nil {
		panic(err)
	}
	defer tbl.Release()

	fmt.Printf("Read table with %d rows and %d columns\n", tbl.NumRows(), tbl.NumCols())
}
```

## Progressive Reading with FileReader

If you need more control over the reading process, such as reading specific columns or row groups, you can use `FileReader`.

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/arrow/go/v16/arrow/memory"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleFileReader() {
	// Open the Parquet file
	f, err := os.Open("example.parquet")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Create a Parquet reader
	rdr, err := file.NewParquetReader(f)
	if err != nil {
		panic(err)
	}
	defer rdr.Close()

	// Set up Arrow read properties
	arrProps := pqarrow.ArrowReadProperties{
		Parallel:  false,
		BatchSize: 1024,
	}

	// Create a pqarrow FileReader
	mem := memory.NewGoAllocator()
	fileReader, err := pqarrow.NewFileReader(rdr, arrProps, mem)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Read only specific columns (e.g., column 0 and 2) and specific row groups (e.g., row group 0)
	colIndices := []int{0, 2}
	rowGroups := []int{0}

	tbl, err := fileReader.ReadRowGroups(ctx, colIndices, rowGroups)
	if err != nil {
		panic(err)
	}
	defer tbl.Release()

	fmt.Printf("Read subset table with %d rows and %d columns\n", tbl.NumRows(), tbl.NumCols())
}
```

## Streaming with RecordReader

For streaming large files without loading everything into memory, you can use a `RecordReader`.

```go
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/apache/arrow/go/v16/arrow/memory"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleRecordReader() {
	// Open the Parquet file
	f, err := os.Open("example.parquet")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Create a Parquet reader
	rdr, err := file.NewParquetReader(f)
	if err != nil {
		panic(err)
	}
	defer rdr.Close()

	// Create a pqarrow FileReader
	mem := memory.NewGoAllocator()
	fileReader, err := pqarrow.NewFileReader(rdr, pqarrow.ArrowReadProperties{BatchSize: 1024}, mem)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	// Get a RecordReader for all columns and all row groups
	recordReader, err := fileReader.GetRecordReader(ctx, nil, nil)
	if err != nil {
		panic(err)
	}
	defer recordReader.Release()

	// Iterate over the record batches
	for recordReader.Next() {
		rec := recordReader.Record()
		fmt.Printf("Read batch with %d rows\n", rec.NumRows())
		// Process the record batch...
	}

	if err := recordReader.Err(); err != nil {
		panic(err)
	}
}
```
