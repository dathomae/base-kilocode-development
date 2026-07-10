package rdbms

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fergusstrange/embedded-postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "embedded-postgres-locks-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	database := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(5440).
		RuntimePath(filepath.Join(tempDir, "runtime")).
		DataPath(filepath.Join(tempDir, "data")).
		Logger(nil).
		Username("postgres").
		Password("postgres").
		Database("postgres"))

	if err := database.Start(); err != nil {
		fmt.Printf("Failed to start embedded postgres: %v\n", err)
		os.Exit(1)
	}

	testDB, err = sql.Open("pgx", "postgres://postgres:postgres@localhost:5440/postgres")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		database.Stop()
		os.Exit(1)
	}

	// Load schema
	schemaPath := filepath.Join("..", "..", "serde", "rdbms", "schema.sql")
	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		fmt.Printf("Failed to read schema.sql: %v\n", err)
		database.Stop()
		os.Exit(1)
	}

	if _, err := testDB.Exec(string(schema)); err != nil {
		fmt.Printf("Failed to execute schema.sql: %v\n", err)
		database.Stop()
		os.Exit(1)
	}

	code := m.Run()

	testDB.Close()
	database.Stop()
	os.Exit(code)
}
