package generator

import (
	"fmt"
	"strings"

	"github.com/rowjak/godb-orm/internal/database"
)

// TagBuilder handles GORM tag generation
type TagBuilder struct{}

// NewTagBuilder creates a new TagBuilder instance
func NewTagBuilder() *TagBuilder {
	return &TagBuilder{}
}

// BuildGormTag generates a GORM struct tag for a column
func (tb *TagBuilder) BuildGormTag(col database.ColumnMetadata) string {
	var parts []string

	// Primary key
	if col.IsPrimaryKey {
		parts = append(parts, "primaryKey")
	}

	// Auto increment
	if col.IsAutoIncrement {
		parts = append(parts, "autoIncrement")
	}

	// Column name
	parts = append(parts, fmt.Sprintf("column:%s", col.Name))

	// Type (always include for schema sync)
	parts = append(parts, fmt.Sprintf("type:%s", col.RawType))

	// Default value
	if col.DefaultValue != nil {
		defaultVal := *col.DefaultValue
		// Clean up default values
		defaultVal = tb.cleanDefaultValue(defaultVal)
		if defaultVal != "" {
			parts = append(parts, fmt.Sprintf("default:%s", defaultVal))
		}
	}

	// Not null constraint (only if not nullable and not primary key)
	if !col.IsNullable && !col.IsPrimaryKey {
		parts = append(parts, "not null")
	}

	return fmt.Sprintf(`gorm:"%s"`, strings.Join(parts, ";"))
}

// cleanDefaultValue cleans up default values for GORM tag
func (tb *TagBuilder) cleanDefaultValue(defaultVal string) string {
	// Remove function-like defaults for auto-increment (handled separately)
	if strings.Contains(defaultVal, "nextval") {
		return ""
	}

	// Remove CURRENT_TIMESTAMP type defaults (GORM handles these automatically)
	lower := strings.ToLower(defaultVal)
	if strings.Contains(lower, "current_timestamp") ||
		strings.Contains(lower, "now()") ||
		strings.Contains(lower, "current_date") {
		return ""
	}

	// Remove parentheses wrapping from PostgreSQL defaults
	if strings.HasPrefix(defaultVal, "(") && strings.HasSuffix(defaultVal, ")") {
		defaultVal = defaultVal[1 : len(defaultVal)-1]
	}

	// Handle NULL default
	if strings.ToUpper(defaultVal) == "NULL" {
		return ""
	}

	return defaultVal
}

// BuildJSONTag generates a JSON struct tag for a column
func (tb *TagBuilder) BuildJSONTag(col database.ColumnMetadata) string {
	// Use snake_case column name for JSON
	return fmt.Sprintf(`json:"%s"`, col.Name)
}

// BuildAllTags generates all struct tags for a column
func (tb *TagBuilder) BuildAllTags(col database.ColumnMetadata) string {
	tags := []string{
		tb.BuildGormTag(col),
		tb.BuildJSONTag(col),
	}
	return strings.Join(tags, " ")
}

// StructField represents a Go struct field with its metadata
type StructField struct {
	Name       string // Go field name (PascalCase)
	Type       string // Go type
	Tags       string // Struct tags
	Comment    string // Field comment (for enums, unknown types, etc.)
	ImportPath string // Required import path if any
}

// BuildStructField creates a complete struct field from column metadata
func (tb *TagBuilder) BuildStructField(col database.ColumnMetadata, typeMapper *TypeMapper) StructField {
	// Get Go type
	goType, importPath, typeComment := typeMapper.GetGoType(col.RawType, col.IsNullable)

	// Build field
	field := StructField{
		Name:       ToPascalCase(col.Name),
		Type:       goType,
		Tags:       tb.BuildAllTags(col),
		ImportPath: importPath,
	}

	// Add enum comment if this is an enum type
	if len(col.EnumValues) > 0 {
		field.Comment = FormatEnumComment(col.EnumValues)
	} else if typeComment != "" {
		field.Comment = typeComment
	} else if col.Comment != "" {
		field.Comment = "// " + col.Comment
	}

	return field
}

// ToPascalCase converts snake_case or other formats to PascalCase
func ToPascalCase(s string) string {
	// Handle common acronyms
	acronyms := map[string]string{
		"id":   "ID",
		"url":  "URL",
		"api":  "API",
		"http": "HTTP",
		"json": "JSON",
		"xml":  "XML",
		"sql":  "SQL",
		"uuid": "UUID",
		"ip":   "IP",
		"html": "HTML",
		"css":  "CSS",
	}

	// Split by underscore or existing caps
	var result strings.Builder
	words := strings.Split(s, "_")

	for _, word := range words {
		if word == "" {
			continue
		}

		// Check if it's a known acronym
		lower := strings.ToLower(word)
		if acronym, ok := acronyms[lower]; ok {
			result.WriteString(acronym)
		} else {
			// Capitalize first letter
			result.WriteString(strings.ToUpper(word[:1]))
			if len(word) > 1 {
				result.WriteString(strings.ToLower(word[1:]))
			}
		}
	}

	return result.String()
}
