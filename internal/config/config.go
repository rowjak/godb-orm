package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// DBConfig holds the database connection configuration
type DBConfig struct {
	Host     string `yaml:"host" mapstructure:"host"`
	Port     int    `yaml:"port" mapstructure:"port"`
	User     string `yaml:"user" mapstructure:"user"`
	Password string `yaml:"password" mapstructure:"password"`
	DBName   string `yaml:"dbname" mapstructure:"dbname"`
	Driver   string `yaml:"driver" mapstructure:"driver"`
}

// GeneratorConfig holds generator-specific options
type GeneratorConfig struct {
	Tables    string `yaml:"tables" mapstructure:"tables"`
	OutputDir string `yaml:"output_dir" mapstructure:"output_dir"`
}

// Config holds the complete application configuration
type Config struct {
	Database  DBConfig        `yaml:"database" mapstructure:"database"`
	Generator GeneratorConfig `yaml:"generator" mapstructure:"generator"`
}

// configDir returns the configuration directory path
func configDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".godb-orm"), nil
}

// configFilePath returns the full path to the config file
func configFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// SaveConfig saves the configuration to ~/.godb-orm/config.yaml
func SaveConfig(cfg *Config) error {
	dir, err := configDir()
	if err != nil {
		return err
	}

	// Create config directory if not exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath, err := configFilePath()
	if err != nil {
		return err
	}

	v := viper.New()
	v.SetConfigType("yaml")

	// Set values
	v.Set("database.host", cfg.Database.Host)
	v.Set("database.port", cfg.Database.Port)
	v.Set("database.user", cfg.Database.User)
	v.Set("database.password", cfg.Database.Password)
	v.Set("database.dbname", cfg.Database.DBName)
	v.Set("database.driver", cfg.Database.Driver)
	v.Set("generator.tables", cfg.Generator.Tables)
	v.Set("generator.output_dir", cfg.Generator.OutputDir)

	// Write config file
	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig loads the configuration from ~/.godb-orm/config.yaml
func LoadConfig() (*Config, error) {
	configPath, err := configFilePath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Database: DBConfig{
			Host:   "localhost",
			Port:   3306,
			User:   "root",
			Driver: "mysql",
		},
		Generator: GeneratorConfig{
			Tables:    "*",
			OutputDir: "./output",
		},
	}
}
