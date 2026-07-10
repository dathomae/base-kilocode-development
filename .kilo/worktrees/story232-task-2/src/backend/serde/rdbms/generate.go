package rdbms

// This file is used to generate Go-Jet models from the schema.sql file.
// It uses an embedded Postgres instance to host the schema during generation.

//go:generate go run github.com/go-jet/jet/v2/cmd/jet -source=Postgres -host=localhost -port=5438 -user=postgres -password=postgres -dbname=postgres -schema=public -path=./gen
