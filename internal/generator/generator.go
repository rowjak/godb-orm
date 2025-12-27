package generator

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"text/template"

	"github.com/rowjak/godb-orm/internal/database"
)

// Generator handles the generation of Go struct files from database tables
type Generator struct {
	introspector database.DBIntrospector
	typeMapper   *TypeMapper
	tagBuilder   *TagBuilder
	namingConv   *NamingConverter
	packageName  string
}

// GeneratorConfig holds configuration for the generator
type GeneratorConfig struct {
	PackageName string
}

// NewGenerator creates a new Generator instance
func NewGenerator(introspector database.DBIntrospector) *Generator {
	return &Generator{
		introspector: introspector,
		typeMapper:   NewTypeMapper(),
		tagBuilder:   NewTagBuilder(),
		namingConv:   NewNamingConverter(),
		packageName:  "models",
	}
}

// NewGeneratorWithConfig creates a new Generator with custom configuration
func NewGeneratorWithConfig(introspector database.DBIntrospector, cfg GeneratorConfig) *Generator {
	g := NewGenerator(introspector)
	if cfg.PackageName != "" {
		g.packageName = cfg.PackageName
	}
	return g
}

// GeneratedFile represents a generated Go file
type GeneratedFile struct {
	FileName    string
	PackageName string
	StructName  string
	TableName   string
	Imports     string
	Fields      []StructField
	Content     string
}

// Generate generates Go struct code for a table and returns formatted bytes
// This is the main entry point as specified in Tahap 3 Tugas 3
func (g *Generator) Generate(tableName string) ([]byte, error) {
	// Get table metadata
	meta, err := g.introspector.GetTableMetadata(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table metadata: %w", err)
	}

	// Build struct fields
	var fields []StructField
	for _, col := range meta.Columns {
		field := g.tagBuilder.BuildStructField(col, g.typeMapper)
		// Use strcase-based naming for field names
		field.Name = g.namingConv.ToGoFieldName(col.Name)
		fields = append(fields, field)
	}

	// Detect required imports using smart import detection
	importMgr := DetectRequiredImports(fields)

	// Build template data
	templateData := &TemplateData{
		PackageName: g.packageName,
		Imports:     importMgr.GenerateImportBlock(),
		StructName:  g.namingConv.ToGoStructName(tableName),
		TableName:   tableName,
		Fields:      fields,
		HasTime:     importMgr.Has(WellKnownImports.Time),
		HasJSON:     importMgr.Has(WellKnownImports.Datatypes),
		HasUUID:     importMgr.Has(WellKnownImports.UUID),
	}

	// Render template
	tmpl, err := template.New("struct").Parse(StructTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, templateData); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	// Format with go/format for proper indentation
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// If formatting fails, return unformatted with warning in content
		// This allows debugging of template issues
		return buf.Bytes(), fmt.Errorf("go/format failed (returning unformatted): %w", err)
	}

	return formatted, nil
}

// GenerateString generates Go struct code and returns as string
func (g *Generator) GenerateString(tableName string) (string, error) {
	bytes, err := g.Generate(tableName)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// GenerateToFile generates and writes the Go struct to a file
// File name uses snake_case as specified in Tahap 3 Tugas 4
func (g *Generator) GenerateToFile(tableName, outputDir string) (string, error) {
	// Generate formatted code
	content, err := g.Generate(tableName)
	if err != nil {
		return "", err
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate file name using snake_case
	fileName := g.namingConv.ToFileName(tableName)
	filePath := filepath.Join(outputDir, fileName)

	// Write file
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// GenerateAll generates Go structs for all tables
func (g *Generator) GenerateAll(outputDir string) ([]string, error) {
	tables, err := g.introspector.GetTables()
	if err != nil {
		return nil, fmt.Errorf("failed to get tables: %w", err)
	}

	var filePaths []string
	for _, table := range tables {
		filePath, err := g.GenerateToFile(table, outputDir)
		if err != nil {
			return filePaths, fmt.Errorf("failed to generate %s: %w", table, err)
		}
		filePaths = append(filePaths, filePath)
	}

	return filePaths, nil
}

// ToStructName converts a table name to a Go struct name (uses NamingConverter)
// Kept for backward compatibility
func ToStructName(tableName string) string {
	nc := NewNamingConverter()
	return nc.ToGoStructName(tableName)
}

// toSnakeCase converts a string to snake_case (uses NamingConverter)
// Kept for backward compatibility
// func toSnakeCase(s string) string {
// 	nc := NewNamingConverter()
// 	return nc.ToSnakeCaseStrcase(s)
// }
