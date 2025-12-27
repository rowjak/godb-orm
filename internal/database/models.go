package database

// ColumnMetadata represents metadata for a database column
type ColumnMetadata struct {
	Name             string   // Column name
	DataType         string   // Normalized data type (e.g., varchar, int)
	RawType          string   // Original DB type with size (e.g., varchar(255), int unsigned)
	IsNullable       bool     // Whether the column allows NULL values
	IsPrimaryKey     bool     // Whether the column is a primary key
	IsAutoIncrement  bool     // Whether the column auto-increments
	DefaultValue     *string  // Default value if any (nil if no default)
	EnumValues       []string // Enum values for ENUM types
	IsUnsigned       bool     // For MySQL unsigned integers
	CharMaxLength    *int     // Maximum character length for string types
	NumericPrecision *int     // Precision for numeric types
	NumericScale     *int     // Scale for numeric types
	Comment          string   // Column comment if any
	OrdinalPosition  int      // Position of the column in the table
}

// TableMetadata represents metadata for a database table
type TableMetadata struct {
	Schema  string           // Schema/Database name
	Name    string           // Table name
	Columns []ColumnMetadata // List of columns
	Comment string           // Table comment if any
}

// DBIntrospector defines the interface for database introspection
type DBIntrospector interface {
	// Connect establishes a connection to the database
	Connect() error

	// Close closes the database connection
	Close() error

	// GetTables returns a list of table names in the database
	GetTables() ([]string, error)

	// GetColumns returns column metadata for a specific table
	GetColumns(tableName string) ([]ColumnMetadata, error)

	// GetTableMetadata returns full metadata for a specific table
	GetTableMetadata(tableName string) (*TableMetadata, error)
}
