package generator

import (
	"testing"
)

func TestNewTypeMapper(t *testing.T) {
	tm := NewTypeMapper()
	if tm == nil {
		t.Fatal("NewTypeMapper returned nil")
	}
	if tm.typeMap == nil {
		t.Fatal("typeMap is nil")
	}
}

func TestGetGoType_Integers(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		dbType     string
		isNullable bool
		expected   string
	}{
		// Non-nullable integers
		{"int", false, "int32"},
		{"integer", false, "int32"},
		{"smallint", false, "int16"},
		{"mediumint", false, "int32"},
		{"bigint", false, "int64"},
		{"tinyint", false, "int8"},

		// Nullable integers
		{"int", true, "*int32"},
		{"bigint", true, "*int64"},
		{"smallint", true, "*int16"},

		// Unsigned integers
		{"int unsigned", false, "uint32"},
		{"bigint unsigned", false, "uint64"},
		{"smallint unsigned", false, "uint16"},
		{"tinyint unsigned", false, "uint8"},

		// Nullable unsigned
		{"int unsigned", true, "*uint32"},
		{"bigint unsigned", true, "*uint64"},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := tm.GetGoTypeSimple(tt.dbType, tt.isNullable)
			if result != tt.expected {
				t.Errorf("GetGoType(%q, %v) = %q; want %q", tt.dbType, tt.isNullable, result, tt.expected)
			}
		})
	}
}

func TestGetGoType_Floats(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		dbType     string
		isNullable bool
		expected   string
	}{
		{"decimal", false, "float64"},
		{"decimal(10,2)", false, "float64"},
		{"numeric", false, "float64"},
		{"float", false, "float32"},
		{"double", false, "float64"},
		{"double precision", false, "float64"},
		{"real", false, "float32"},

		// Nullable
		{"decimal", true, "*float64"},
		{"float", true, "*float32"},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := tm.GetGoTypeSimple(tt.dbType, tt.isNullable)
			if result != tt.expected {
				t.Errorf("GetGoType(%q, %v) = %q; want %q", tt.dbType, tt.isNullable, result, tt.expected)
			}
		})
	}
}

func TestGetGoType_Strings(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		dbType     string
		isNullable bool
		expected   string
	}{
		{"varchar", false, "string"},
		{"varchar(255)", false, "string"},
		{"char", false, "string"},
		{"char(10)", false, "string"},
		{"text", false, "string"},
		{"longtext", false, "string"},
		{"mediumtext", false, "string"},
		{"tinytext", false, "string"},

		// Nullable
		{"varchar(255)", true, "*string"},
		{"text", true, "*string"},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := tm.GetGoTypeSimple(tt.dbType, tt.isNullable)
			if result != tt.expected {
				t.Errorf("GetGoType(%q, %v) = %q; want %q", tt.dbType, tt.isNullable, result, tt.expected)
			}
		})
	}
}

func TestGetGoType_DateTime(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		dbType       string
		isNullable   bool
		expectedType string
	}{
		{"timestamp", false, "time.Time"},
		{"timestamptz", false, "time.Time"},
		{"timestamp with time zone", false, "time.Time"},
		{"datetime", false, "time.Time"},
		{"date", false, "time.Time"},
		{"time", false, "string"}, // time without date is string

		// Nullable
		{"timestamp", true, "*time.Time"},
		{"datetime", true, "*time.Time"},
		{"date", true, "*time.Time"},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := tm.GetGoTypeSimple(tt.dbType, tt.isNullable)
			if result != tt.expectedType {
				t.Errorf("GetGoType(%q, %v) = %q; want %q", tt.dbType, tt.isNullable, result, tt.expectedType)
			}
		})
	}
}

func TestGetGoType_Boolean(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		dbType     string
		isNullable bool
		expected   string
	}{
		{"bool", false, "bool"},
		{"boolean", false, "bool"},
		{"tinyint(1)", false, "bool"}, // MySQL boolean

		// Nullable
		{"bool", true, "*bool"},
		{"boolean", true, "*bool"},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			result := tm.GetGoTypeSimple(tt.dbType, tt.isNullable)
			if result != tt.expected {
				t.Errorf("GetGoType(%q, %v) = %q; want %q", tt.dbType, tt.isNullable, result, tt.expected)
			}
		})
	}
}

func TestGetGoType_Special(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		dbType         string
		isNullable     bool
		expectedType   string
		expectedImport string
	}{
		// JSON types
		{"json", false, "datatypes.JSON", "gorm.io/datatypes"},
		{"jsonb", false, "datatypes.JSON", "gorm.io/datatypes"},
		{"json", true, "*datatypes.JSON", "gorm.io/datatypes"},

		// UUID
		{"uuid", false, "uuid.UUID", "github.com/google/uuid"},
		{"uuid", true, "*uuid.UUID", "github.com/google/uuid"},

		// Binary types (no pointer for slices)
		{"bytea", false, "[]byte", ""},
		{"bytea", true, "[]byte", ""}, // slices don't get pointer
		{"blob", false, "[]byte", ""},
		{"blob", true, "[]byte", ""},
		{"binary", false, "[]byte", ""},
		{"varbinary", false, "[]byte", ""},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			goType, importPath, _ := tm.GetGoType(tt.dbType, tt.isNullable)
			if goType != tt.expectedType {
				t.Errorf("GetGoType(%q, %v) type = %q; want %q", tt.dbType, tt.isNullable, goType, tt.expectedType)
			}
			if importPath != tt.expectedImport {
				t.Errorf("GetGoType(%q, %v) import = %q; want %q", tt.dbType, tt.isNullable, importPath, tt.expectedImport)
			}
		})
	}
}

func TestGetGoType_Unknown(t *testing.T) {
	tm := NewTypeMapper()

	tests := []struct {
		dbType          string
		isNullable      bool
		expectedType    string
		expectedComment string
	}{
		{"custom_type", false, "interface{}", "// unknown type: custom_type"},
		{"unknown_type", true, "*interface{}", "// unknown type: unknown_type"},
		{"my_special_type", false, "interface{}", "// unknown type: my_special_type"},
	}

	for _, tt := range tests {
		t.Run(tt.dbType, func(t *testing.T) {
			goType, _, comment := tm.GetGoType(tt.dbType, tt.isNullable)
			if goType != tt.expectedType {
				t.Errorf("GetGoType(%q, %v) type = %q; want %q", tt.dbType, tt.isNullable, goType, tt.expectedType)
			}
			if comment != tt.expectedComment {
				t.Errorf("GetGoType(%q, %v) comment = %q; want %q", tt.dbType, tt.isNullable, comment, tt.expectedComment)
			}
		})
	}
}

func TestParseEnumValues(t *testing.T) {
	tests := []struct {
		columnType string
		expected   []string
	}{
		{"enum('active','inactive')", []string{"active", "inactive"}},
		{"ENUM('pending','approved','rejected')", []string{"pending", "approved", "rejected"}},
		{"enum('yes','no')", []string{"yes", "no"}},
		{"enum('a','b','c')", []string{"a", "b", "c"}},
		{"varchar(255)", nil}, // not an enum
		{"int", nil},          // not an enum
	}

	for _, tt := range tests {
		t.Run(tt.columnType, func(t *testing.T) {
			result := ParseEnumValues(tt.columnType)
			if len(result) != len(tt.expected) {
				t.Errorf("ParseEnumValues(%q) = %v; want %v", tt.columnType, result, tt.expected)
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("ParseEnumValues(%q)[%d] = %q; want %q", tt.columnType, i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestFormatEnumComment(t *testing.T) {
	tests := []struct {
		values   []string
		expected string
	}{
		{[]string{"active", "inactive"}, "// enum('active','inactive')"},
		{[]string{"yes", "no"}, "// enum('yes','no')"},
		{[]string{}, ""},
		{nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := FormatEnumComment(tt.values)
			if result != tt.expected {
				t.Errorf("FormatEnumComment(%v) = %q; want %q", tt.values, result, tt.expected)
			}
		})
	}
}
