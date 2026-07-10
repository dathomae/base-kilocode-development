package main

import (
	"flag"
	"log"
	"strings"

	"github.com/go-jet/jet/v2/generator/metadata"
	postgres_gen "github.com/go-jet/jet/v2/generator/postgres"
	"github.com/go-jet/jet/v2/generator/template"
	"github.com/go-jet/jet/v2/postgres"
	"unicode"
)

// toGoIdentifier is a simplified version of Go-Jet's internal identifier conversion.
// It converts snake_case to PascalCase.
func toGoIdentifier(name string) string {
	runes := []rune(name)
	result := make([]rune, 0, len(runes))
	upperNext := true
	for _, r := range runes {
		if r == '_' {
			upperNext = true
			continue
		}
		if upperNext {
			result = append(result, unicode.ToUpper(r))
			upperNext = false
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func main() {
	destPath := flag.String("path", "./serde/rdbms/gen", "Destination path for generated files")
	host := flag.String("host", "127.0.0.1", "Database host")
	port := flag.Int("port", 5438, "Database port")
	user := flag.String("user", "postgres", "Database user")
	password := flag.String("password", "postgres", "Database password")
	dbname := flag.String("dbname", "postgres", "Database name")
	schema := flag.String("schema", "public", "Database schema")
	flag.Parse()

	log.Printf("Starting programmatic Go-Jet generation for schema %s...", *schema)

	// Create a custom template that handles 'citext' silently
	customTemplate := template.Default(postgres.Dialect).
		UseSchema(func(schemaMetaData metadata.Schema) template.Schema {
			return template.DefaultSchema(schemaMetaData).
				UseModel(template.DefaultModel().
					UseTable(func(table metadata.Table) template.TableModel {
						return template.DefaultTableModel(table).
							UseField(func(column metadata.Column) template.TableModelField {
								// Silence 'citext' warnings in Models
								if strings.ToLower(column.DataType.Name) == "citext" {
									var tags []string
									if column.IsPrimaryKey {
										tags = append(tags, `sql:"primary_key"`)
									}
									typeName := "string"
									if column.IsNullable {
										typeName = "*string"
									}
									return template.TableModelField{
										Name: toGoIdentifier(column.Name),
										Type: template.Type{Name: typeName},
										Tags: tags,
									}
								}
								return template.DefaultTableModelField(column)
							})
					}),
				).
				UseSQLBuilder(template.DefaultSQLBuilder().
					UseTable(func(table metadata.Table) template.TableSQLBuilder {
						return template.DefaultTableSQLBuilder(table).
							UseColumn(func(column metadata.Column) template.TableSQLBuilderColumn {
								// Silence 'citext' warnings in SQL Builder
								if strings.ToLower(column.DataType.Name) == "citext" {
									return template.TableSQLBuilderColumn{
										Name: toGoIdentifier(column.Name),
										Type: "String",
									}
								}
								return template.DefaultTableSQLBuilderColumn(column)
							})
					}),
				)
		})

	err := postgres_gen.Generate(*destPath, postgres_gen.DBConnection{
		Host:       *host,
		Port:       *port,
		User:       *user,
		Password:   *password,
		DBName:     *dbname,
		SchemaName: *schema,
		SslMode:    "disable",
	}, customTemplate)

	if err != nil {
		log.Fatalf("Generation failed: %v", err)
	}

	log.Println("Generation complete.")
}
