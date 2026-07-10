package rdbms

import (
	"context"
	"path/filepath"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"
)

// TestSchemaExecution proves that the consolidated schema.sql can be executed
// against a fresh PostgreSQL instance without errors.
func TestSchemaExecution(t *testing.T) {
	// Configure embedded postgres on a specific port to avoid conflicts
	tempDir := t.TempDir()
	database := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(5436). // Using a unique port for this proof-of-concept test
		RuntimePath(filepath.Join(tempDir, "runtime")).
		DataPath(filepath.Join(tempDir, "data")).
		Logger(nil).
		Username("postgres").
		Password("postgres").
		Database("postgres"))

	if err := database.Start(); err != nil {
		t.Fatalf("Failed to start embedded postgres: %v", err)
	}
	defer database.Stop()

	// Connect to the embedded database
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5436/postgres")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Execute the schema using the shared helper
	// This proves the helper and the schema.sql are both valid.
	if err := ApplySchema(ctx, conn); err != nil {
		t.Fatalf("Schema execution failed: %v", err)
	}

	t.Log("Consolidated schema.sql executed successfully using shared ApplySchema helper")
}

// TestBitemporalDefaultsPopulation verifies that the bitemporal columns are 
// correctly populated with defaults.
func TestBitemporalDefaultsPopulation(t *testing.T) {
	tempDir := t.TempDir()
	database := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(5437).
		RuntimePath(filepath.Join(tempDir, "runtime")).
		DataPath(filepath.Join(tempDir, "data")).
		Logger(nil).
		Username("postgres").
		Password("postgres").
		Database("postgres"))

	if err := database.Start(); err != nil {
		t.Fatalf("Failed to start embedded postgres: %v", err)
	}
	defer database.Stop()

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5437/postgres")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	if err := ApplySchema(ctx, conn); err != nil {
		t.Fatalf("Schema execution failed: %v", err)
	}

	// Insert a row and check bitemporal defaults
	// Using the language table as a representative bitemporal table
	var txStart, txEnd, mode string
	err = conn.QueryRow(ctx, `
		INSERT INTO language (language, country, display_name, insertion_metadata)
		VALUES ('fr', '', 'French', '{"reason": "test"}')
		RETURNING transaction_start::text, transaction_end::text, mode::text
	`).Scan(&txStart, &txEnd, &mode)

	if err != nil {
		t.Fatalf("Failed to insert and retrieve bitemporal defaults: %v", err)
	}

	if txStart == "" {
		t.Error("transaction_start was not populated")
	}
	if txEnd != "3000-01-01 00:00:00+00" {
		t.Errorf("transaction_end: expected 3000-01-01 00:00:00+00, got %s", txEnd)
	}
	if mode != "active" {
		t.Errorf("expected mode 'active', got %q", mode)
	}
}
