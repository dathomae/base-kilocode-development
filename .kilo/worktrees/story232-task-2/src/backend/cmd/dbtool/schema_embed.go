package main

import _ "embed"

// schemaSQL is embedded at compile time. dbtool must NOT import the
// serde/rdbms package because dbtool runs *before* go-jet has
// generated the model files in serde/rdbms/gen/. Embedding the
// schema file directly (rather than walking up the filesystem)
// makes the dependency explicit to the Go toolchain and survives
// packaging into a single binary.
//
// schema.sql is copied into this directory to work around Go's
// embed directive restriction: Go forbids relative paths with ".." in
// embed patterns and does not support symlinks. The canonical schema.sql
// resides in serde/rdbms/; this copy is generated as part of the build
// process to keep both in sync.
//
//go:embed schema.sql
var schemaSQL string
