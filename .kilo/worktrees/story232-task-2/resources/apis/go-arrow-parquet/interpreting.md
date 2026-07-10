# Interpreting Parquet Files with Go Arrow

The `pqarrow` package provides tools for interpreting Parquet files and mapping their schemas to Arrow schemas. This is crucial when you need to understand the structure of a Parquet file before reading its data, or when you need to translate between Parquet and Arrow types.

## Schema Translation

You can translate a Parquet schema to an Arrow schema using `FromParquet`, and vice versa using `ToParquet`.

```go
package main

import (
	"fmt"
	"os"

	"github.com/apache/arrow/go/v16/arrow"
	"github.com/apache/arrow/go/v16/parquet"
	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleSchemaTranslation() {
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

	// Get the Parquet schema
	parquetSchema := rdr.MetaData().Schema

	// Translate the Parquet schema to an Arrow schema
	arrProps := pqarrow.ArrowReadProperties{}
	kvMeta := rdr.MetaData().KeyValueMetadata()

	arrowSchema, err := pqarrow.FromParquet(parquetSchema, &arrProps, kvMeta)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Arrow Schema: %v\n", arrowSchema)

	// Translate the Arrow schema back to a Parquet schema
	writerProps := parquet.NewWriterProperties()
	writerArrProps := pqarrow.NewArrowWriterProperties()

	newParquetSchema, err := pqarrow.ToParquet(arrowSchema, writerProps, writerArrProps)
	if err != nil {
		panic(err)
	}

	fmt.Printf("New Parquet Schema: %v\n", newParquetSchema)
}
```

## Schema Manifest

For more complex scenarios, such as dealing with nested types, you can use a `SchemaManifest`. A `SchemaManifest` represents a full manifest for mapping a Parquet schema to an Arrow Schema.

```go
package main

import (
	"fmt"
	"os"

	"github.com/apache/arrow/go/v16/parquet/file"
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleSchemaManifest() {
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

	// Get the Parquet schema
	parquetSchema := rdr.MetaData().Schema
	kvMeta := rdr.MetaData().KeyValueMetadata()

	// Create a SchemaManifest
	arrProps := pqarrow.ArrowReadProperties{}
	manifest, err := pqarrow.NewSchemaManifest(parquetSchema, kvMeta, &arrProps)
	if err != nil {
		panic(err)
	}

	// Inspect the manifest
	fmt.Printf("Origin Arrow Schema: %v\n", manifest.OriginSchema)

	// Get the corresponding Field for a given column index
	colIndex := 0
	field, err := manifest.GetColumnField(colIndex)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Field for column %d: %v\n", colIndex, field.Field.Name)

	// Check if the field is a leaf column
	if field.IsLeaf() {
		fmt.Printf("Field %s is a leaf column\n", field.Field.Name)
	}

	// Get the parent field if it is a nested column
	parent := manifest.GetParent(field)
	if parent != nil {
		fmt.Printf("Parent field: %v\n", parent.Field.Name)
	}
}
```

## Storing Arrow Schema in Parquet Metadata

When writing Parquet files, you can store the Arrow schema in the file's metadata. This allows `pqarrow` to automatically reconstruct the exact Arrow schema when reading the file back, preserving types that might otherwise be ambiguous in Parquet (e.g., distinguishing between different timestamp units or dictionary-encoded types).

```go
package main

import (
	"github.com/apache/arrow/go/v16/parquet/pqarrow"
)

func ExampleStoreSchema() {
	// Set up Arrow writer properties to store the schema
	arrProps := pqarrow.NewArrowWriterProperties(
		pqarrow.WithStoreSchema(),
	)

	// Use arrProps when creating a FileWriter or calling WriteTable
	// ...
}
```
