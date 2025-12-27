package database

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq"
	"github.com/rowjak/godb-orm/internal/config"
)

// PostgresIntrospector implements database introspection for PostgreSQL
type PostgresIntrospector struct {
	BaseIntrospector
	currentSchema string
}

// NewPostgresIntrospector creates a new PostgreSQL introspector
func NewPostgresIntrospector(cfg *config.DBConfig) *PostgresIntrospector {
	return &PostgresIntrospector{
		BaseIntrospector: BaseIntrospector{cfg: cfg},
		currentSchema:    "public", // Default schema
	}
}

// GetSchemas returns a list of available schemas in the database
func (p *PostgresIntrospector) GetSchemas() ([]string, error) {
	query := `
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name NOT IN ('pg_catalog', 'information_schema', 'pg_toast')
		ORDER BY schema_name
	`

	rows, err := p.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}
		schemas = append(schemas, schemaName)
	}

	return schemas, nil
}

// SetSchema sets the current schema to use for table queries
func (p *PostgresIntrospector) SetSchema(schema string) {
	p.currentSchema = schema
}

// GetCurrentSchema returns the currently selected schema
func (p *PostgresIntrospector) GetCurrentSchema() string {
	return p.currentSchema
}

// Connect establishes a connection to the PostgreSQL database
func (p *PostgresIntrospector) Connect() error {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.cfg.Host,
		p.cfg.Port,
		p.cfg.User,
		p.cfg.Password,
		p.cfg.DBName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	p.db = db
	return nil
}

// GetTables returns a list of table names in the database
func (p *PostgresIntrospector) GetTables() ([]string, error) {
	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = $1 AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`

	rows, err := p.db.Query(query, p.currentSchema)
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}

	return tables, nil
}

// GetColumns returns column metadata for a specific table
func (p *PostgresIntrospector) GetColumns(tableName string) ([]ColumnMetadata, error) {
	// Main query for column information with udt_name for custom types
	query := `
		SELECT 
			c.column_name,
			c.data_type,
			c.udt_name,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			c.ordinal_position,
			COALESCE(pgd.description, '') as column_comment
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_statio_all_tables st 
			ON c.table_schema = st.schemaname AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pgd 
			ON pgd.objoid = st.relid AND pgd.objsubid = c.ordinal_position
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	rows, err := p.db.Query(query, p.currentSchema, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnMetadata
	for rows.Next() {
		var (
			columnName       string
			dataType         string
			udtName          string
			isNullable       string
			columnDefault    sql.NullString
			charMaxLength    sql.NullInt64
			numericPrecision sql.NullInt64
			numericScale     sql.NullInt64
			ordinalPosition  int
			columnComment    string
		)

		err := rows.Scan(
			&columnName,
			&dataType,
			&udtName,
			&isNullable,
			&columnDefault,
			&charMaxLength,
			&numericPrecision,
			&numericScale,
			&ordinalPosition,
			&columnComment,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		// Use udt_name for more specific type information
		// PostgreSQL udt_name gives us internal types like int4, int8, varchar, etc.
		rawType := p.buildRawType(dataType, udtName, charMaxLength, numericPrecision, numericScale)

		col := ColumnMetadata{
			Name:            columnName,
			DataType:        p.normalizeDataType(dataType, udtName),
			RawType:         rawType,
			IsNullable:      isNullable == "YES",
			OrdinalPosition: ordinalPosition,
			Comment:         columnComment,
		}

		// Handle default value
		if columnDefault.Valid {
			col.DefaultValue = &columnDefault.String
			// Detect auto-increment (serial/bigserial)
			if strings.Contains(columnDefault.String, "nextval") {
				col.IsAutoIncrement = true
			}
		}

		// Handle character max length
		if charMaxLength.Valid {
			length := int(charMaxLength.Int64)
			col.CharMaxLength = &length
		}

		// Handle numeric precision
		if numericPrecision.Valid {
			precision := int(numericPrecision.Int64)
			col.NumericPrecision = &precision
		}

		// Handle numeric scale
		if numericScale.Valid {
			scale := int(numericScale.Int64)
			col.NumericScale = &scale
		}

		columns = append(columns, col)
	}

	// Get primary key information
	pkColumns, err := p.getPrimaryKeyColumns(tableName)
	if err != nil {
		return nil, err
	}

	// Mark primary key columns
	for i := range columns {
		if _, ok := pkColumns[columns[i].Name]; ok {
			columns[i].IsPrimaryKey = true
		}
	}

	return columns, nil
}

// getPrimaryKeyColumns returns a set of column names that are primary keys
func (p *PostgresIntrospector) getPrimaryKeyColumns(tableName string) (map[string]bool, error) {
	// Use schema-qualified name for regclass
	qualifiedName := fmt.Sprintf("%s.%s", p.currentSchema, tableName)
	query := `
		SELECT a.attname
		FROM pg_index i
		JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
		WHERE i.indrelid = $1::regclass AND i.indisprimary
	`

	rows, err := p.db.Query(query, qualifiedName)
	if err != nil {
		return nil, fmt.Errorf("failed to query primary keys: %w", err)
	}
	defer rows.Close()

	pkColumns := make(map[string]bool)
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, fmt.Errorf("failed to scan primary key column: %w", err)
		}
		pkColumns[columnName] = true
	}

	return pkColumns, nil
}

// normalizeDataType normalizes PostgreSQL data types to common names
func (p *PostgresIntrospector) normalizeDataType(dataType, udtName string) string {
	// Map udt_name to standard types
	switch udtName {
	case "int2":
		return "smallint"
	case "int4":
		return "integer"
	case "int8":
		return "bigint"
	case "float4":
		return "real"
	case "float8":
		return "double precision"
	case "bool":
		return "boolean"
	case "varchar", "bpchar":
		return "varchar"
	case "timestamptz":
		return "timestamptz"
	case "timestamp":
		return "timestamp"
	case "jsonb":
		return "jsonb"
	case "json":
		return "json"
	case "uuid":
		return "uuid"
	case "bytea":
		return "bytea"
	default:
		// For ARRAY types, dataType is 'ARRAY' and udt_name starts with '_'
		if dataType == "ARRAY" && strings.HasPrefix(udtName, "_") {
			return "[]" + udtName[1:] // e.g., "_int4" -> "[]int4"
		}
		return dataType
	}
}

// buildRawType constructs the full type string with size information
func (p *PostgresIntrospector) buildRawType(dataType, udtName string, charMaxLength, numericPrecision, numericScale sql.NullInt64) string {
	normalizedType := p.normalizeDataType(dataType, udtName)

	// Add size information for varchar/char
	if (normalizedType == "varchar" || normalizedType == "character varying") && charMaxLength.Valid {
		return fmt.Sprintf("varchar(%d)", charMaxLength.Int64)
	}

	// Add precision/scale for numeric types
	if (normalizedType == "numeric" || normalizedType == "decimal") && numericPrecision.Valid {
		if numericScale.Valid && numericScale.Int64 > 0 {
			return fmt.Sprintf("numeric(%d,%d)", numericPrecision.Int64, numericScale.Int64)
		}
		return fmt.Sprintf("numeric(%d)", numericPrecision.Int64)
	}

	return normalizedType
}

// GetTableMetadata returns full metadata for a specific table
func (p *PostgresIntrospector) GetTableMetadata(tableName string) (*TableMetadata, error) {
	columns, err := p.GetColumns(tableName)
	if err != nil {
		return nil, err
	}

	// Get table comment using schema-qualified name
	qualifiedName := fmt.Sprintf("%s.%s", p.currentSchema, tableName)
	var tableComment sql.NullString
	query := `
		SELECT obj_description($1::regclass, 'pg_class')
	`
	err = p.db.QueryRow(query, qualifiedName).Scan(&tableComment)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get table comment: %w", err)
	}

	meta := &TableMetadata{
		Schema:  p.currentSchema,
		Name:    tableName,
		Columns: columns,
	}

	if tableComment.Valid {
		meta.Comment = tableComment.String
	}

	return meta, nil
}
