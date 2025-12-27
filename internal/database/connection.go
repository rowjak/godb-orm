package database

import (
	"database/sql"
	"fmt"

	"github.com/rowjak/godb-orm/internal/config"
)

// NewIntrospector creates a new database introspector based on the driver
func NewIntrospector(cfg *config.DBConfig) (DBIntrospector, error) {
	switch cfg.Driver {
	case "mysql":
		return NewMySQLIntrospector(cfg), nil
	case "postgres", "postgresql":
		return NewPostgresIntrospector(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Driver)
	}
}

// BaseIntrospector provides common functionality for database introspection
type BaseIntrospector struct {
	cfg *config.DBConfig
	db  *sql.DB
}

// Close closes the database connection
func (b *BaseIntrospector) Close() error {
	if b.db != nil {
		return b.db.Close()
	}
	return nil
}

// DB returns the underlying database connection
func (b *BaseIntrospector) DB() *sql.DB {
	return b.db
}
