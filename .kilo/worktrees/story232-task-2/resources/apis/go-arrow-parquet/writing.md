# Writing Parquet Files with Go Arrow

The `pqarrow` package provides several ways to write Arrow data structures to Parquet files. You can write an entire Arrow Table at once, or you can write it progressively using a `FileWriter`.

## Writing an Entire Table

The simplest way to write an Arrow Table to a Parquet file is to use the `WriteTable` convenience function.

```go
package main

import (
	"os"

	"github.com/apache/arrow/go/v16/arrow"
	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleWriteTable(tbl arrow.Table) {
	// Create a new Parquet file
	f, err := os.Create("example.parquet")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Set up Parquet writer properties
	props := parquet.NewWriterProperties(
		parquet.WithCompression(parquet.CompressionSnappy),
		parquet.WithVersion(parquet.V2_LATEST),
	)

	// Set up Arrow writer properties
	arrProps := pqarrow.NewArrowWriterProperties(
		pqarrow.WithStoreSchema(),
	)

	// Write the entire table to the file
	chunkSize := int64(1024) // Number of rows per row group
	err = pqarrow.WriteTable(tbl, f, chunkSize, props, arrProps)
	if err != nil {
		panic(err)
	}
}
```

## Progressive Writing with FileWriter

If you need more control over the writing process, such as writing specific columns or records, you can use `FileWriter`.

```go
package main

import (
	"os"

	"github.com/apache/arrow/go/v16/arrow"
	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleFileWriter(schema *arrow.Schema, records []arrow.Record) {
	// Create a new Parquet file
	f, err := os.Create("example.parquet")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Set up Parquet writer properties
	props := parquet.NewWriterProperties(
		parquet.WithCompression(parquet.CompressionSnappy),
		parquet.WithVersion(parquet.V2_LATEST),
	)

	// Set up Arrow writer properties
	arrProps := pqarrow.NewArrowWriterProperties(
		pqarrow.WithStoreSchema(),
	)

	// Create a pqarrow FileWriter
	fileWriter, err := pqarrow.NewFileWriter(schema, f, props, arrProps)
	if err != nil {
		panic(err)
	}
	defer fileWriter.Close()

	// Write records to the file
	for _, rec := range records {
		// WriteBuffered will either append to an existing row group or create a new
		// one based on the record length and max row group length.
		err = fileWriter.WriteBuffered(rec)
		if err != nil {
			panic(err)
		}
	}
}
```

## Writing Columns Directly

You can also write columns directly to a Parquet file using `FileWriter.WriteColumnData` or `FileWriter.WriteColumnChunked`.

```go
package main

import (
	"os"

	"github.com/apache/arrow/go/v16/arrow"
	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleWriteColumnData(schema *arrow.Schema, colData []arrow.Array) {
	// Create a new Parquet file
	f, err := os.Create("example.parquet")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Set up Parquet writer properties
	props := parquet.NewWriterProperties(
		parquet.WithCompression(parquet.CompressionSnappy),
		parquet.WithVersion(parquet.V2_LATEST),
	)

	// Set up Arrow writer properties
	arrProps := pqarrow.NewArrowWriterProperties(
		pqarrow.WithStoreSchema(),
	)

	// Create a pqarrow FileWriter
	fileWriter, err := pqarrow.NewFileWriter(schema, f, props, arrProps)
	if err != nil {
		panic(err)
	}
	defer fileWriter.Close()

	// Create a new row group
	fileWriter.NewRowGroup()

	// Write columns to the file
	for _, col := range colData {
		err = fileWriter.WriteColumnData(col)
		if err != nil {
			panic(err)
		}
	}
}
```
