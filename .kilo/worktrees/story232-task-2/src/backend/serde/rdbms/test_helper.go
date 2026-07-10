package rdbms

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgx/v5"
)

// ApplySchema reads the consolidated schema.sql and executes it against the provided connection.
// It uses a robust path discovery mechanism to find the schema.sql file regardless of where
// the test is being run from.
func ApplySchema(ctx context.Context, conn *pgx.Conn) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get caller information")
	}
	
	schemaPath := filepath.Join(filepath.Dir(filename), "schema.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file at %s: %w", schemaPath, err)
	}

	if _, err := conn.Exec(ctx, string(schemaBytes)); err != nil {
		return fmt.Errorf("schema execution failed: %w", err)
	}

	return nil
}

// ApplySchemaTx is like ApplySchema but uses a transaction.
func ApplySchemaTx(ctx context.Context, tx pgx.Tx) error {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("failed to get caller information")
	}
	
	schemaPath := filepath.Join(filepath.Dir(filename), "schema.sql")
	schemaBytes, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file at %s: %w", schemaPath, err)
	}

	if _, err := tx.Exec(ctx, string(schemaBytes)); err != nil {
		return fmt.Errorf("schema execution failed: %w", err)
	}

	return nil
}
