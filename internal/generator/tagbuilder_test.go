package generator

import (
	"testing"

	"github.com/rowjak/godb-orm/internal/database"
)

func TestBuildGormTag_PrimaryKey(t *testing.T) {
	tb := NewTagBuilder()

	col := database.ColumnMetadata{
		Name:            "id",
		RawType:         "int unsigned",
		IsPrimaryKey:    true,
		IsAutoIncrement: true,
		IsNullable:      false,
	}

	tag := tb.BuildGormTag(col)
	expected := `gorm:"primaryKey;autoIncrement;column:id;type:int unsigned"`

	if tag != expected {
		t.Errorf("BuildGormTag() = %q; want %q", tag, expected)
	}
}

func TestBuildGormTag_WithDefault(t *testing.T) {
	tb := NewTagBuilder()

	defaultVal := "'active'"
	col := database.ColumnMetadata{
		Name:         "status",
		RawType:      "enum('active','inactive')",
		IsPrimaryKey: false,
		IsNullable:   false,
		DefaultValue: &defaultVal,
	}

	tag := tb.BuildGormTag(col)

	if !contains(tag, "default:'active'") {
		t.Errorf("BuildGormTag() = %q; should contain default value", tag)
	}
	if !contains(tag, "column:status") {
		t.Errorf("BuildGormTag() = %q; should contain column name", tag)
	}
}

func TestBuildGormTag_NotNull(t *testing.T) {
	tb := NewTagBuilder()

	col := database.ColumnMetadata{
		Name:       "email",
		RawType:    "varchar(255)",
		IsNullable: false,
	}

	tag := tb.BuildGormTag(col)

	if !contains(tag, "not null") {
		t.Errorf("BuildGormTag() = %q; should contain 'not null'", tag)
	}
}

func TestBuildGormTag_Nullable(t *testing.T) {
	tb := NewTagBuilder()

	col := database.ColumnMetadata{
		Name:       "description",
		RawType:    "text",
		IsNullable: true,
	}

	tag := tb.BuildGormTag(col)

	if contains(tag, "not null") {
		t.Errorf("BuildGormTag() = %q; should NOT contain 'not null'", tag)
	}
}

func TestBuildJSONTag(t *testing.T) {
	tb := NewTagBuilder()

	col := database.ColumnMetadata{
		Name: "created_at",
	}

	tag := tb.BuildJSONTag(col)
	expected := `json:"created_at"`

	if tag != expected {
		t.Errorf("BuildJSONTag() = %q; want %q", tag, expected)
	}
}

func TestBuildAllTags(t *testing.T) {
	tb := NewTagBuilder()

	col := database.ColumnMetadata{
		Name:            "id",
		RawType:         "int",
		IsPrimaryKey:    true,
		IsAutoIncrement: true,
	}

	tags := tb.BuildAllTags(col)

	if !contains(tags, "gorm:") {
		t.Errorf("BuildAllTags() = %q; should contain gorm tag", tags)
	}
	if !contains(tags, "json:") {
		t.Errorf("BuildAllTags() = %q; should contain json tag", tags)
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"id", "ID"},
		{"user_id", "UserID"},
		{"created_at", "CreatedAt"},
		{"first_name", "FirstName"},
		{"api_key", "APIKey"},
		{"json_data", "JSONData"},
		{"http_status", "HTTPStatus"},
		{"url", "URL"},
		{"some_column", "SomeColumn"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ToPascalCase(tt.input)
			if result != tt.expected {
				t.Errorf("ToPascalCase(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildStructField(t *testing.T) {
	tb := NewTagBuilder()
	tm := NewTypeMapper()

	col := database.ColumnMetadata{
		Name:            "id",
		RawType:         "int unsigned",
		IsPrimaryKey:    true,
		IsAutoIncrement: true,
		IsNullable:      false,
	}

	field := tb.BuildStructField(col, tm)

	if field.Name != "ID" {
		t.Errorf("StructField.Name = %q; want %q", field.Name, "ID")
	}
	if field.Type != "uint32" {
		t.Errorf("StructField.Type = %q; want %q", field.Type, "uint32")
	}
	if !contains(field.Tags, "primaryKey") {
		t.Errorf("StructField.Tags = %q; should contain primaryKey", field.Tags)
	}
}

func TestBuildStructField_WithEnum(t *testing.T) {
	tb := NewTagBuilder()
	tm := NewTypeMapper()

	col := database.ColumnMetadata{
		Name:       "status",
		DataType:   "enum",
		RawType:    "enum('active','inactive')",
		IsNullable: false,
		EnumValues: []string{"active", "inactive"},
	}

	field := tb.BuildStructField(col, tm)

	if field.Type != "string" {
		t.Errorf("StructField.Type = %q; want %q", field.Type, "string")
	}
	if field.Comment != "// enum('active','inactive')" {
		t.Errorf("StructField.Comment = %q; want enum comment", field.Comment)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
