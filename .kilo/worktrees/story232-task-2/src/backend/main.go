package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"chronocone/backend/handlers"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	router := chi.NewMux()
	api := humachi.New(router, huma.DefaultConfig("ChronoCone Backend API", "0.1.0"))

	handlers.RegisterFamilyRoutes(api, db)
	handlers.RegisterNamespaceRoutes(api, db)
	handlers.RegisterCategoryRoutes(api, db)
	handlers.RegisterElementRoutes(api, db)

	fmt.Println("ChronoCone backend listening on :8080")
	fmt.Println("API docs: http://localhost:8080/docs")
	http.ListenAndServe(":8080", router)
}
