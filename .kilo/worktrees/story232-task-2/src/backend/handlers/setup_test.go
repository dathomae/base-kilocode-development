package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	tempDir, err := os.MkdirTemp("", "embedded-postgres-handlers-*")
	if err != nil {
		fmt.Printf("Failed to create temp dir: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	database := embeddedpostgres.NewDatabase(embeddedpostgres.DefaultConfig().
		Port(5441).
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

	testDB, err = sql.Open("pgx", "postgres://postgres:postgres@localhost:5441/postgres")
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		database.Stop()
		os.Exit(1)
	}

	// Load schema
	schemaPath := filepath.Join("..", "serde", "rdbms", "schema.sql")
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

func createTestAPI(t *testing.T, db *sql.DB) (huma.API, http.Handler) {
	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("Test API", "0.1.0"))

	RegisterFamilyRoutes(api, db)
	RegisterNamespaceRoutes(api, db)
	RegisterCategoryRoutes(api, db)
	RegisterElementRoutes(api, db)

	return api, router
}

func doRequest(t *testing.T, router http.Handler, method, path string, body interface{}) (*http.Response, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewReader(jsonBody)
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	resp := w.Result()
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	resp.Body.Close()

	return resp, respBody
}
