// Package main contains a tool to generate GORM models from database schema.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gen"

	"github.com/xelarion/go-layout/pkg/config"
	"github.com/xelarion/go-layout/pkg/database"
	"github.com/xelarion/go-layout/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	zapLogger, err := logger.New(&cfg.Log)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		return
	}
	defer zapLogger.Sync()

	log := zapLogger.Logger

	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get working directory: %v\n", err)
		return
	}

	// Create output directory
	outputDir := filepath.Join(workDir, "internal", "model", "gen")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Failed to create output directory: %v\n", err)
		return
	}

	fmt.Printf("Using output directory: %s\n", outputDir)

	// Connect to database
	db, err := database.NewPostgres(&cfg.PG, log)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}

	// Create gen configuration
	g := gen.NewGenerator(gen.Config{
		OutPath:           outputDir,
		ModelPkgPath:      "gen",
		FieldNullable:     true,
		FieldCoverable:    false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
	})

	// Use database connection
	g.UseDB(db.DB)

	// Find all tables (exclude system tables)
	var tables []string
	db.DB.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' AND table_name NOT IN ('goose_db_version', 'schema_migrations')").Find(&tables)

	for _, table := range tables {
		g.GenerateModel(table, gen.FieldModify(func(field gen.Field) gen.Field {
			// change id and _id fields to uint
			if field.Type == "int64" && (field.ColumnName == "id" || strings.HasSuffix(field.ColumnName, "_id")) {
				field.Type = "uint"
			}

			// change array fields to pq.StringArray / pg.XXXArray
			types := field.GORMTag["type"]
			/**
			character varying(n)[] -> pq.StringArray
			varchar(n)[] -> pq.StringArray
			text[] -> pq.StringArray
			integer[] -> pq.Int32Array
			bigint[] -> pq.Int64Array
			boolean[] -> pq.BoolArray
			numeric[] -> pq.Float64Array
			*/
			if len(types) > 0 && strings.HasSuffix(types[0], "[]") {
				// Remove length information and array suffix
				baseType := strings.TrimSuffix(types[0], "[]")
				// Remove length information in parentheses
				if idx := strings.Index(baseType, "("); idx != -1 {
					baseType = baseType[:idx]
				}
				// Trim spaces
				baseType = strings.TrimSpace(baseType)

				switch baseType {
				case "character varying", "varchar", "text":
					field.Type = "pq.StringArray"
				case "integer", "int", "int4":
					field.Type = "pq.Int32Array"
				case "bigint", "int8":
					field.Type = "pq.Int64Array"
				case "boolean", "bool":
					field.Type = "pq.BoolArray"
				case "numeric", "decimal":
					field.Type = "pq.Float64Array"
				default:
					field.Type = "pq.StringArray" // default to string array for unknown types
				}
			}

			return field
		}))
	}

	// Execute the generator
	g.Execute()

	fmt.Println("Model generation completed. Generated models are in:", outputDir)
	fmt.Println("You can now create extended models in internal/model/ directory.")
}
