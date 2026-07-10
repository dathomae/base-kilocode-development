package domain

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"

	"chronocone/backend/serde/rdbms"
)

var suiteConn pgx.Conn

type txKey struct{}

// TestMain acts as the package-scoped fixture. It spins up embedded Postgres,
// applies the schema once, runs all tests, and tears everything down.
func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "embedded-postgres-domain-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	database := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(5433).
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

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5433/postgres")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		database.Stop()
		os.Exit(1)
	}
	suiteConn = *conn

	if err := rdbms.ApplySchema(ctx, conn); err != nil {
		fmt.Printf("Failed to apply schema: %v\n", err)
		conn.Close(ctx)
		database.Stop()
		os.Exit(1)
	}

	code := m.Run()

	conn.Close(ctx)
	if err := database.Stop(); err != nil {
		fmt.Printf("Failed to stop embedded postgres: %v\n", err)
	}

	os.Exit(code)
}

// RunTx opens a fresh transaction for each test, ensuring per-test isolation.
// Each test gets its own transaction_timestamp() value, enabling bitemporal
// staleness tests.
func RunTx(t *testing.T, testLogic func(t *testing.T, ctx context.Context)) {
	t.Helper()

	ctx := context.Background()
	testTx, err := suiteConn.Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	defer testTx.Rollback(ctx)

	testCtx := context.WithValue(ctx, txKey{}, testTx)
	testLogic(t, testCtx)
}

// TestSetupTx_TransactionStartIsFresh verifies each RunTx invocation gets a
// distinct transaction_timestamp() value.
func TestSetupTx_TransactionStartIsFresh(t *testing.T) {
	var ts1, ts2 string

	RunTx(t, func(t *testing.T, ctx context.Context) {
		testTx := ctx.Value(txKey{}).(pgx.Tx)
		if err := testTx.QueryRow(ctx, "SELECT transaction_timestamp()::text").Scan(&ts1); err != nil {
			t.Fatalf("Failed to query transaction_timestamp: %v", err)
		}
	})

	RunTx(t, func(t *testing.T, ctx context.Context) {
		testTx := ctx.Value(txKey{}).(pgx.Tx)
		if err := testTx.QueryRow(ctx, "SELECT transaction_timestamp()::text").Scan(&ts2); err != nil {
			t.Fatalf("Failed to query transaction_timestamp: %v", err)
		}
	})

	if ts1 == ts2 {
		t.Errorf("transaction_timestamp values must differ across RunTx invocations: got %s == %s", ts1, ts2)
	}
	if ts1 >= ts2 {
		t.Errorf("second transaction_timestamp must be after the first: got %s >= %s", ts1, ts2)
	}
}
