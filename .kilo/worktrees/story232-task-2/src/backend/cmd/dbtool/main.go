package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jackc/pgx/v5"
)

// applySchemaInternal applies the embedded schema.sql to the database connection.
// The schema is embedded at compile time via schema_embed.go to ensure dbtool
// can run before go-jet model generation without importing the serde/rdbms package.
func applySchemaInternal(ctx context.Context, conn *pgx.Conn) error {
	if _, err := conn.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("schema execution failed: %w", err)
	}
	return nil
}

func main() {
	port := flag.Uint("port", 5438, "Port to run Postgres on")
	applySchema := flag.Bool("apply-schema", false, "Apply schema.sql after starting")
	keepAlive := flag.Bool("keep-alive", false, "Keep the database running until interrupted")
	flag.Parse()

	config := embeddedpostgres.DefaultConfig().
		Port(uint32(*port)).
		Logger(os.Stdout).
		Username("postgres").
		Password("postgres").
		Database("postgres")

	database := embeddedpostgres.NewDatabase(config)

	fmt.Printf("Starting embedded Postgres on port %d...\n", *port)
	if err := database.Start(); err != nil {
		log.Fatalf("Failed to start database: %v", err)
	}
	defer database.Stop()

	if *applySchema {
		fmt.Println("Applying schema...")
		ctx := context.Background()
		conn, err := pgx.Connect(ctx, fmt.Sprintf("postgres://postgres:postgres@localhost:%d/postgres", *port))
		if err != nil {
			log.Fatalf("Failed to connect: %v", err)
		}
		defer conn.Close(ctx)

		if err := applySchemaInternal(ctx, conn); err != nil {
			log.Fatalf("Failed to apply schema: %v", err)
		}
		fmt.Println("Schema applied successfully.")
	}

	if *keepAlive {
		fmt.Println("Database is running. Press Ctrl+C to stop.")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		fmt.Println("Shutting down...")
	}
}
