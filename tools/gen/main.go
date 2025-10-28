// Package main contains a tool to generate GORM models from database schema.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/gen"

	"github.com/xelarion/go-layout/internal/infra/config"
	"github.com/xelarion/go-layout/internal/infra/database"
	"github.com/xelarion/go-layout/internal/infra/logger"
)

// TypeMapper handles the mapping of database types to Go types
type TypeMapper struct {
	// Maps base types (without array suffix) to their Go equivalents
	arrayTypeMap map[string]string
	// Maps regular types to their Go equivalents
	typeMap map[string]string
}

// NewTypeMapper creates a new type mapper with predefined mappings
func NewTypeMapper() *TypeMapper {
	return &TypeMapper{
		arrayTypeMap: map[string]string{
			"character varying": "pq.StringArray",
			"varchar":           "pq.StringArray",
			"text":              "pq.StringArray",
			"integer":           "pq.Int32Array",
			"int":               "pq.Int32Array",
			"int4":              "pq.Int32Array",
			"bigint":            "pq.Int64Array",
			"int8":              "pq.Int64Array",
			"boolean":           "pq.BoolArray",
			"bool":              "pq.BoolArray",
			"numeric":           "pq.Float64Array",
			"decimal":           "pq.Float64Array",
		},
		typeMap: map[string]string{
			"json":                   "datatypes.JSON",
			"jsonb":                  "datatypes.JSON",
			"uuid":                   "datatypes.UUID",
			"date":                   "datatypes.Date",
			"time":                   "datatypes.Time",
			"time without time zone": "datatypes.Time",
			"time with time zone":    "datatypes.Time",
		},
	}
}

// MapFieldType maps a database type to its corresponding Go type
func (tm *TypeMapper) MapFieldType(field gen.Field) string {
	types := field.GORMTag["type"]
	if len(types) == 0 {
		return field.Type
	}

	typeStr := strings.ToLower(types[0])

	// Handle array types
	if strings.HasSuffix(typeStr, "[]") {
		baseType := strings.TrimSuffix(typeStr, "[]")
		// Remove length information in parentheses
		if idx := strings.Index(baseType, "("); idx != -1 {
			baseType = baseType[:idx]
		}
		baseType = strings.TrimSpace(baseType)

		if goType, ok := tm.arrayTypeMap[baseType]; ok {
			return goType
		}
		return "pq.StringArray" // default for unknown array types
	}

	// Handle non-array types
	if goType, ok := tm.typeMap[typeStr]; ok {
		return goType
	}

	return field.Type
}

func main() {
	// Parse command line flags
	var tableName string
	flag.StringVar(&tableName, "table", "", "Specify a single table to generate model for")
	flag.Parse()

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

	// Create type mapper
	typeMapper := NewTypeMapper()

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

	// Get tables to generate models for
	var tables []string
	if tableName != "" {
		// Generate model for a specific table
		tables = []string{tableName}
		fmt.Printf("Generating model for table: %s\n", tableName)
	} else {
		// Find all tables (exclude system tables and partition tables)
		db.DB.Raw(`
			SELECT t.table_name
			FROM information_schema.tables t
			WHERE t.table_schema = 'public'
			AND t.table_name NOT IN ('goose_db_version', 'schema_migrations')
			AND NOT EXISTS (
				SELECT 1 FROM pg_inherits i
				JOIN pg_class c ON i.inhrelid = c.oid
				JOIN pg_namespace n ON c.relnamespace = n.oid
				WHERE n.nspname = 'public' AND c.relname = t.table_name
			)
		`).Find(&tables)
		fmt.Printf("Generating models for all tables (excluding partition child tables): %v\n", tables)
	}

	// Check if specified table exists
	if tableName != "" {
		var exists bool
		db.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = ?)", tableName).Scan(&exists)
		if !exists {
			fmt.Printf("Error: Table '%s' does not exist in the database\n", tableName)
			return
		}
	}

	for _, table := range tables {
		g.GenerateModel(table, gen.FieldModify(func(field gen.Field) gen.Field {
			// change id and _id fields to uint
			if field.Type == "int64" && (field.ColumnName == "id" || strings.HasSuffix(field.ColumnName, "_id")) {
				field.Type = "uint"
			}

			// Map database type to Go type
			mappedType := typeMapper.MapFieldType(field)
			if mappedType != field.Type {
				field.Type = mappedType
			}

			// Remove default tag fields to allow zero values
			field.GORMTag.Remove("default")

			return field
		}))
	}

	// Execute the generator
	g.Execute()

	if tableName != "" {
		fmt.Printf("Model generation completed for table '%s'. Generated model is in: %s\n", tableName, outputDir)
	} else {
		fmt.Println("Model generation completed for all tables. Generated models are in:", outputDir)
	}
	fmt.Println("You can now create extended models in internal/model/ directory.")
}
