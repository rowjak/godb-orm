package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/rowjak/godb-orm/internal/config"
	"github.com/rowjak/godb-orm/internal/database"
	"github.com/rowjak/godb-orm/internal/generator"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

// Common errors for bridge methods
var (
	ErrNotConnected = errors.New("database not connected")
	ErrNoInspector  = errors.New("database inspector not initialized")
)

// ColumnInfo represents column information for the frontend
type ColumnInfo struct {
	Name            string   `json:"name"`
	DataType        string   `json:"dataType"`
	RawType         string   `json:"rawType"`
	GoType          string   `json:"goType"`
	IsNullable      bool     `json:"isNullable"`
	IsPrimaryKey    bool     `json:"isPrimaryKey"`
	IsAutoIncrement bool     `json:"isAutoIncrement"`
	DefaultValue    *string  `json:"defaultValue"`
	EnumValues      []string `json:"enumValues,omitempty"`
	Comment         string   `json:"comment,omitempty"`
}

// ConnectionStatus represents the current connection status
type ConnectionStatus struct {
	Connected    bool   `json:"connected"`
	Driver       string `json:"driver"`
	Host         string `json:"host"`
	DatabaseName string `json:"databaseName"`
	Error        string `json:"error,omitempty"`
}

// App struct holds the application state
type App struct {
	ctx          context.Context
	mu           sync.RWMutex
	introspector database.DBIntrospector
	dbConfig     *config.DBConfig
	generator    *generator.Generator
	connected    bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// Startup is called when the app starts
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx

	// Try to load saved configuration
	cfg, err := config.LoadConfig()
	if err == nil && cfg.Database.DBName != "" {
		a.dbConfig = &cfg.Database
	}
}

// Greet returns a greeting for the given name (kept for testing)
func (a *App) Greet(name string) string {
	return "Hello " + name + ", welcome to godb-orm!"
}

// GetSavedConfig returns the saved database configuration
func (a *App) GetSavedConfig() *config.DBConfig {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.dbConfig
}

// GetConnectionStatus returns the current connection status
func (a *App) GetConnectionStatus() ConnectionStatus {
	a.mu.RLock()
	defer a.mu.RUnlock()

	status := ConnectionStatus{
		Connected: a.connected,
	}

	if a.dbConfig != nil {
		status.Driver = a.dbConfig.Driver
		status.Host = a.dbConfig.Host
		status.DatabaseName = a.dbConfig.DBName
	}

	return status
}

// ConnectDB attempts to connect to the database with the given configuration
// This is the main method called from frontend to establish a connection
func (a *App) ConnectDB(cfg config.DBConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Close existing connection if any
	if a.introspector != nil {
		a.introspector.Close()
		a.introspector = nil
		a.generator = nil
		a.connected = false
	}

	// Create new introspector based on driver
	introspector, err := database.NewIntrospector(&cfg)
	if err != nil {
		return fmt.Errorf("failed to create introspector: %w", err)
	}

	// Attempt connection
	if err := introspector.Connect(); err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Store state
	a.introspector = introspector
	a.dbConfig = &cfg
	a.generator = generator.NewGenerator(introspector)
	a.connected = true

	// Save configuration for future use
	fullCfg := &config.Config{
		Database: cfg,
		Generator: config.GeneratorConfig{
			Tables:    "*",
			OutputDir: "./models",
		},
	}
	if err := config.SaveConfig(fullCfg); err != nil {
		// Log warning but don't fail the connection
		log.Printf("Warning: Could not save config: %v", err)
	}

	return nil
}

// DisconnectDB closes the database connection
func (a *App) DisconnectDB() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.introspector != nil {
		if err := a.introspector.Close(); err != nil {
			return fmt.Errorf("failed to close connection: %w", err)
		}
		a.introspector = nil
		a.generator = nil
		a.connected = false
	}

	return nil
}

// IsPostgres returns true if the connected database is PostgreSQL
func (a *App) IsPostgres() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.dbConfig != nil && a.dbConfig.Driver == "postgres"
}

// FetchSchemas returns a list of schemas for PostgreSQL databases
func (a *App) FetchSchemas() ([]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.introspector == nil {
		return nil, ErrNotConnected
	}

	// Check if it's a PostgreSQL connection
	if pgIntrospector, ok := a.introspector.(*database.PostgresIntrospector); ok {
		return pgIntrospector.GetSchemas()
	}

	// For MySQL/other databases, return empty (no schema concept)
	return []string{}, nil
}

// SetSchema sets the current schema for PostgreSQL databases
func (a *App) SetSchema(schema string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.connected || a.introspector == nil {
		return ErrNotConnected
	}

	// Check if it's a PostgreSQL connection
	if pgIntrospector, ok := a.introspector.(*database.PostgresIntrospector); ok {
		pgIntrospector.SetSchema(schema)
		return nil
	}

	// For MySQL/other databases, ignore (no schema concept)
	return nil
}

// GetCurrentSchema returns the current schema for PostgreSQL databases
func (a *App) GetCurrentSchema() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.introspector == nil {
		return ""
	}

	// Check if it's a PostgreSQL connection
	if pgIntrospector, ok := a.introspector.(*database.PostgresIntrospector); ok {
		return pgIntrospector.GetCurrentSchema()
	}

	// For MySQL, return database name
	if a.dbConfig != nil {
		return a.dbConfig.DBName
	}
	return ""
}

// FetchTables returns a list of table names from the connected database
func (a *App) FetchTables() ([]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.introspector == nil {
		return nil, ErrNotConnected
	}

	tables, err := a.introspector.GetTables()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tables: %w", err)
	}

	return tables, nil
}

// FetchTableSchema returns detailed column information for a specific table
func (a *App) FetchTableSchema(tableName string) ([]ColumnInfo, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.introspector == nil {
		return nil, ErrNotConnected
	}

	columns, err := a.introspector.GetColumns(tableName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch schema for table %s: %w", tableName, err)
	}

	// Create type mapper for Go type conversion
	typeMapper := generator.NewTypeMapper()

	// Convert to ColumnInfo for frontend
	var columnInfos []ColumnInfo
	for _, col := range columns {
		goType, _, _ := typeMapper.GetGoType(col.RawType, col.IsNullable)

		info := ColumnInfo{
			Name:            col.Name,
			DataType:        col.DataType,
			RawType:         col.RawType,
			GoType:          goType,
			IsNullable:      col.IsNullable,
			IsPrimaryKey:    col.IsPrimaryKey,
			IsAutoIncrement: col.IsAutoIncrement,
			DefaultValue:    col.DefaultValue,
			EnumValues:      col.EnumValues,
			Comment:         col.Comment,
		}
		columnInfos = append(columnInfos, info)
	}

	return columnInfos, nil
}

// GetCodePreview generates and returns the Go struct code for a table
func (a *App) GetCodePreview(tableName string) (string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.generator == nil {
		return "", ErrNotConnected
	}

	code, err := a.generator.GenerateString(tableName)
	if err != nil {
		return "", fmt.Errorf("failed to generate code for table %s: %w", tableName, err)
	}

	return code, nil
}

// GetCodePreviewMultiple generates code preview for multiple tables
func (a *App) GetCodePreviewMultiple(tableNames []string) (map[string]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.generator == nil {
		return nil, ErrNotConnected
	}

	results := make(map[string]string)
	for _, tableName := range tableNames {
		code, err := a.generator.GenerateString(tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to generate code for table %s: %w", tableName, err)
		}
		results[tableName] = code
	}

	return results, nil
}

// SaveCodeToFile saves the generated code for a table to a file
func (a *App) SaveCodeToFile(tableName string, filePath string) error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.generator == nil {
		return ErrNotConnected
	}

	// Generate the code
	code, err := a.generator.Generate(tableName)
	if err != nil {
		return fmt.Errorf("failed to generate code for table %s: %w", tableName, err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write to file
	if err := os.WriteFile(filePath, code, 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	return nil
}

// SaveAllToDirectory saves all tables to a directory
func (a *App) SaveAllToDirectory(outputDir string) ([]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.generator == nil {
		return nil, ErrNotConnected
	}

	filePaths, err := a.generator.GenerateAll(outputDir)
	if err != nil {
		return nil, fmt.Errorf("failed to generate all tables: %w", err)
	}

	return filePaths, nil
}

// SaveSelectedToDirectory saves selected tables to a directory
func (a *App) SaveSelectedToDirectory(tableNames []string, outputDir string) ([]string, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if !a.connected || a.generator == nil {
		return nil, ErrNotConnected
	}

	var filePaths []string
	for _, tableName := range tableNames {
		filePath, err := a.generator.GenerateToFile(tableName, outputDir)
		if err != nil {
			return filePaths, fmt.Errorf("failed to generate %s: %w", tableName, err)
		}
		filePaths = append(filePaths, filePath)
	}

	return filePaths, nil
}

// StartGUI launches the Wails GUI application
func StartGUI() {
	app := NewApp()

	err := wails.Run(&options.App{
		Title:  "GoDB-Orm - Database Model Generator",
		Width:  1280,
		Height: 800,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.Startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatal("Error starting GUI:", err)
	}
}
