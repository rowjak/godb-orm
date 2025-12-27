package generator

import (
	"regexp"
	"strings"
)

// TypeMapping represents a type with its import requirement
type TypeMapping struct {
	GoType     string
	ImportPath string // empty if no import needed
	IsSlice    bool   // true for types like []byte that shouldn't get pointer prefix
}

// TypeMapper handles database type to Go type conversion
type TypeMapper struct {
	// typeMap contains known type mappings
	typeMap map[string]TypeMapping
}

// NewTypeMapper creates a new TypeMapper instance
func NewTypeMapper() *TypeMapper {
	tm := &TypeMapper{
		typeMap: make(map[string]TypeMapping),
	}
	tm.initTypeMappings()
	return tm
}

// initTypeMappings initializes all known type mappings
func (tm *TypeMapper) initTypeMappings() {
	// Integer types
	tm.typeMap["int"] = TypeMapping{GoType: "int32"}
	tm.typeMap["integer"] = TypeMapping{GoType: "int32"}
	tm.typeMap["smallint"] = TypeMapping{GoType: "int16"}
	tm.typeMap["mediumint"] = TypeMapping{GoType: "int32"}
	tm.typeMap["bigint"] = TypeMapping{GoType: "int64"}
	tm.typeMap["tinyint"] = TypeMapping{GoType: "int8"}
	tm.typeMap["serial"] = TypeMapping{GoType: "int32"}
	tm.typeMap["bigserial"] = TypeMapping{GoType: "int64"}
	tm.typeMap["smallserial"] = TypeMapping{GoType: "int16"}

	// Unsigned integer types (MySQL)
	tm.typeMap["int unsigned"] = TypeMapping{GoType: "uint32"}
	tm.typeMap["integer unsigned"] = TypeMapping{GoType: "uint32"}
	tm.typeMap["smallint unsigned"] = TypeMapping{GoType: "uint16"}
	tm.typeMap["mediumint unsigned"] = TypeMapping{GoType: "uint32"}
	tm.typeMap["bigint unsigned"] = TypeMapping{GoType: "uint64"}
	tm.typeMap["tinyint unsigned"] = TypeMapping{GoType: "uint8"}

	// Float/Decimal types
	tm.typeMap["decimal"] = TypeMapping{GoType: "float64"}
	tm.typeMap["numeric"] = TypeMapping{GoType: "float64"}
	tm.typeMap["float"] = TypeMapping{GoType: "float32"}
	tm.typeMap["double"] = TypeMapping{GoType: "float64"}
	tm.typeMap["double precision"] = TypeMapping{GoType: "float64"}
	tm.typeMap["real"] = TypeMapping{GoType: "float32"}
	tm.typeMap["money"] = TypeMapping{GoType: "float64"}

	// String types
	tm.typeMap["varchar"] = TypeMapping{GoType: "string"}
	tm.typeMap["char"] = TypeMapping{GoType: "string"}
	tm.typeMap["character"] = TypeMapping{GoType: "string"}
	tm.typeMap["character varying"] = TypeMapping{GoType: "string"}
	tm.typeMap["text"] = TypeMapping{GoType: "string"}
	tm.typeMap["longtext"] = TypeMapping{GoType: "string"}
	tm.typeMap["mediumtext"] = TypeMapping{GoType: "string"}
	tm.typeMap["tinytext"] = TypeMapping{GoType: "string"}
	tm.typeMap["citext"] = TypeMapping{GoType: "string"}

	// Date/Time types
	tm.typeMap["timestamp"] = TypeMapping{GoType: "time.Time", ImportPath: "time"}
	tm.typeMap["timestamptz"] = TypeMapping{GoType: "time.Time", ImportPath: "time"}
	tm.typeMap["timestamp with time zone"] = TypeMapping{GoType: "time.Time", ImportPath: "time"}
	tm.typeMap["timestamp without time zone"] = TypeMapping{GoType: "time.Time", ImportPath: "time"}
	tm.typeMap["datetime"] = TypeMapping{GoType: "time.Time", ImportPath: "time"}
	tm.typeMap["date"] = TypeMapping{GoType: "time.Time", ImportPath: "time"}
	tm.typeMap["time"] = TypeMapping{GoType: "string"} // time without date is better as string
	tm.typeMap["time with time zone"] = TypeMapping{GoType: "string"}
	tm.typeMap["time without time zone"] = TypeMapping{GoType: "string"}
	tm.typeMap["year"] = TypeMapping{GoType: "int16"}
	tm.typeMap["interval"] = TypeMapping{GoType: "string"}

	// Boolean types
	tm.typeMap["bool"] = TypeMapping{GoType: "bool"}
	tm.typeMap["boolean"] = TypeMapping{GoType: "bool"}
	tm.typeMap["tinyint(1)"] = TypeMapping{GoType: "bool"} // MySQL boolean

	// JSON types
	tm.typeMap["json"] = TypeMapping{GoType: "datatypes.JSON", ImportPath: "gorm.io/datatypes"}
	tm.typeMap["jsonb"] = TypeMapping{GoType: "datatypes.JSON", ImportPath: "gorm.io/datatypes"}

	// UUID type
	tm.typeMap["uuid"] = TypeMapping{GoType: "uuid.UUID", ImportPath: "github.com/google/uuid"}

	// Binary types
	tm.typeMap["bytea"] = TypeMapping{GoType: "[]byte", IsSlice: true}
	tm.typeMap["blob"] = TypeMapping{GoType: "[]byte", IsSlice: true}
	tm.typeMap["tinyblob"] = TypeMapping{GoType: "[]byte", IsSlice: true}
	tm.typeMap["mediumblob"] = TypeMapping{GoType: "[]byte", IsSlice: true}
	tm.typeMap["longblob"] = TypeMapping{GoType: "[]byte", IsSlice: true}
	tm.typeMap["binary"] = TypeMapping{GoType: "[]byte", IsSlice: true}
	tm.typeMap["varbinary"] = TypeMapping{GoType: "[]byte", IsSlice: true}
	tm.typeMap["bit"] = TypeMapping{GoType: "[]byte", IsSlice: true}

	// ENUM type (handled specially, but default to string)
	tm.typeMap["enum"] = TypeMapping{GoType: "string"}

	// SET type (MySQL)
	tm.typeMap["set"] = TypeMapping{GoType: "string"}

	// PostgreSQL specific types
	tm.typeMap["inet"] = TypeMapping{GoType: "string"}
	tm.typeMap["cidr"] = TypeMapping{GoType: "string"}
	tm.typeMap["macaddr"] = TypeMapping{GoType: "string"}
	tm.typeMap["macaddr8"] = TypeMapping{GoType: "string"}
	tm.typeMap["xml"] = TypeMapping{GoType: "string"}
	tm.typeMap["point"] = TypeMapping{GoType: "string"}
	tm.typeMap["line"] = TypeMapping{GoType: "string"}
	tm.typeMap["lseg"] = TypeMapping{GoType: "string"}
	tm.typeMap["box"] = TypeMapping{GoType: "string"}
	tm.typeMap["path"] = TypeMapping{GoType: "string"}
	tm.typeMap["polygon"] = TypeMapping{GoType: "string"}
	tm.typeMap["circle"] = TypeMapping{GoType: "string"}
}

// GetGoType converts a database type to a Go type
// dbType is the database column type (e.g., "varchar(255)", "int unsigned")
// isNullable indicates if the column allows NULL values
// Returns the Go type and any required import path
func (tm *TypeMapper) GetGoType(dbType string, isNullable bool) (string, string, string) {
	// Normalize the type: lowercase and trim
	normalizedType := strings.ToLower(strings.TrimSpace(dbType))

	// Extract base type without size specification
	baseType := tm.extractBaseType(normalizedType)

	// Check for unsigned integers (MySQL specific)
	if strings.Contains(normalizedType, "unsigned") {
		unsignedKey := baseType + " unsigned"
		if mapping, ok := tm.typeMap[unsignedKey]; ok {
			goType := tm.applyNullable(mapping.GoType, isNullable, mapping.IsSlice)
			return goType, mapping.ImportPath, ""
		}
	}

	// Check if it's a tinyint(1) which is boolean in MySQL
	if strings.HasPrefix(normalizedType, "tinyint(1)") && !strings.Contains(normalizedType, "unsigned") {
		mapping := tm.typeMap["tinyint(1)"]
		goType := tm.applyNullable(mapping.GoType, isNullable, mapping.IsSlice)
		return goType, mapping.ImportPath, ""
	}

	// Check exact match first
	if mapping, ok := tm.typeMap[normalizedType]; ok {
		goType := tm.applyNullable(mapping.GoType, isNullable, mapping.IsSlice)
		return goType, mapping.ImportPath, ""
	}

	// Check base type match
	if mapping, ok := tm.typeMap[baseType]; ok {
		goType := tm.applyNullable(mapping.GoType, isNullable, mapping.IsSlice)
		return goType, mapping.ImportPath, ""
	}

	// Fallback: return interface{} with comment
	comment := "// unknown type: " + dbType
	goType := tm.applyNullable("interface{}", isNullable, false)
	return goType, "", comment
}

// GetGoTypeSimple is a simpler version that returns just the Go type
func (tm *TypeMapper) GetGoTypeSimple(dbType string, isNullable bool) string {
	goType, _, _ := tm.GetGoType(dbType, isNullable)
	return goType
}

// extractBaseType extracts the base type from a type with size specification
// e.g., "varchar(255)" -> "varchar", "decimal(10,2)" -> "decimal"
func (tm *TypeMapper) extractBaseType(dbType string) string {
	// Remove anything after opening parenthesis
	if idx := strings.Index(dbType, "("); idx != -1 {
		return strings.TrimSpace(dbType[:idx])
	}
	return dbType
}

// applyNullable returns the Go type (GORM handles nullable with zero values)
func (tm *TypeMapper) applyNullable(goType string, _ bool, _ bool) string {
	// GORM automatically handles NULL values with Go zero values
	// No pointer prefix needed
	return goType
}

// ParseEnumValues extracts enum values from a MySQL enum definition
// e.g., "enum('active','inactive','pending')" -> ["active", "inactive", "pending"]
func ParseEnumValues(columnType string) []string {
	// Match enum('value1','value2',...)
	re := regexp.MustCompile(`enum\s*\(\s*(.+)\s*\)`)
	matches := re.FindStringSubmatch(strings.ToLower(columnType))
	if len(matches) < 2 {
		return nil
	}

	// Extract values
	valuesPart := matches[1]
	var values []string

	// Parse each quoted value
	valueRe := regexp.MustCompile(`'([^']*)'`)
	valueMatches := valueRe.FindAllStringSubmatch(valuesPart, -1)
	for _, m := range valueMatches {
		if len(m) >= 2 {
			values = append(values, m[1])
		}
	}

	return values
}

// FormatEnumComment creates a comment string for enum values
func FormatEnumComment(values []string) string {
	if len(values) == 0 {
		return ""
	}
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = "'" + v + "'"
	}
	return "// enum(" + strings.Join(quoted, ",") + ")"
}
