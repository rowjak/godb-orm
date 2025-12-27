package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rowjak/godb-orm/internal/config"
)

// MySQLIntrospector implements database introspection for MySQL
type MySQLIntrospector struct {
	BaseIntrospector
}

// NewMySQLIntrospector creates a new MySQL introspector
func NewMySQLIntrospector(cfg *config.DBConfig) *MySQLIntrospector {
	return &MySQLIntrospector{
		BaseIntrospector: BaseIntrospector{cfg: cfg},
	}
}

// Connect establishes a connection to the MySQL database
func (m *MySQLIntrospector) Connect() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true",
		m.cfg.User,
		m.cfg.Password,
		m.cfg.Host,
		m.cfg.Port,
		m.cfg.DBName,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	m.db = db
	return nil
}

// GetTables returns a list of table names in the database
func (m *MySQLIntrospector) GetTables() ([]string, error) {
	query := `
		SELECT TABLE_NAME 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_TYPE = 'BASE TABLE'
		ORDER BY TABLE_NAME
	`

	rows, err := m.db.Query(query, m.cfg.DBName)
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
func (m *MySQLIntrospector) GetColumns(tableName string) ([]ColumnMetadata, error) {
	query := `
		SELECT 
			COLUMN_NAME,
			DATA_TYPE,
			COLUMN_TYPE,
			IS_NULLABLE,
			COLUMN_KEY,
			EXTRA,
			COLUMN_DEFAULT,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			COLUMN_COMMENT,
			ORDINAL_POSITION
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := m.db.Query(query, m.cfg.DBName, tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to query columns: %w", err)
	}
	defer rows.Close()

	var columns []ColumnMetadata
	for rows.Next() {
		var (
			columnName       string
			dataType         string
			columnType       string
			isNullable       string
			columnKey        sql.NullString
			extra            sql.NullString
			columnDefault    sql.NullString
			charMaxLength    sql.NullInt64
			numericPrecision sql.NullInt64
			numericScale     sql.NullInt64
			columnComment    sql.NullString
			ordinalPosition  int
		)

		err := rows.Scan(
			&columnName,
			&dataType,
			&columnType,
			&isNullable,
			&columnKey,
			&extra,
			&columnDefault,
			&charMaxLength,
			&numericPrecision,
			&numericScale,
			&columnComment,
			&ordinalPosition,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan column: %w", err)
		}

		col := ColumnMetadata{
			Name:            columnName,
			DataType:        dataType,
			RawType:         columnType,
			IsNullable:      isNullable == "YES",
			IsPrimaryKey:    columnKey.Valid && columnKey.String == "PRI",
			IsAutoIncrement: extra.Valid && strings.Contains(extra.String, "auto_increment"),
			OrdinalPosition: ordinalPosition,
		}

		// Handle default value
		if columnDefault.Valid {
			col.DefaultValue = &columnDefault.String
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

		// Handle column comment
		if columnComment.Valid {
			col.Comment = columnComment.String
		}

		// Detect unsigned integers
		col.IsUnsigned = strings.Contains(strings.ToLower(columnType), "unsigned")

		// Parse ENUM values if it's an enum type
		if strings.ToLower(dataType) == "enum" {
			col.EnumValues = parseEnumValues(columnType)
		}

		columns = append(columns, col)
	}

	return columns, nil
}

// GetTableMetadata returns full metadata for a specific table
func (m *MySQLIntrospector) GetTableMetadata(tableName string) (*TableMetadata, error) {
	columns, err := m.GetColumns(tableName)
	if err != nil {
		return nil, err
	}

	// Get table comment
	var tableComment sql.NullString
	query := `
		SELECT TABLE_COMMENT 
		FROM information_schema.TABLES 
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`
	err = m.db.QueryRow(query, m.cfg.DBName, tableName).Scan(&tableComment)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get table comment: %w", err)
	}

	meta := &TableMetadata{
		Schema:  m.cfg.DBName,
		Name:    tableName,
		Columns: columns,
	}

	if tableComment.Valid {
		meta.Comment = tableComment.String
	}

	return meta, nil
}

// parseEnumValues extracts enum values from a MySQL COLUMN_TYPE
// e.g., "enum('active','inactive','pending')" -> ["active", "inactive", "pending"]
func parseEnumValues(columnType string) []string {
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
