package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/rowjak/godb-orm/internal/config"
	"github.com/rowjak/godb-orm/internal/database"
	"github.com/rowjak/godb-orm/internal/generator"
	"github.com/spf13/cobra"
)

var (
	// Database connection flags
	host     string
	port     int
	user     string
	password string
	dbName   string
	driver   string

	// Generator flags
	table     string
	outputDir string

	// Configuration
	cfg *config.Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "godb-orm",
	Short: "Generate Go ORM structs from database tables",
	Long: `godb-orm is a CLI/GUI tool that generates Go ORM structs 
from your database tables. It supports MySQL and PostgreSQL databases.

Example usage:
  godb-orm --host localhost --port 3306 --user root --db mydb --driver mysql
  godb-orm -H localhost -P 3306 -u root -d mydb --driver mysql --table users`,
	Run: func(cmd *cobra.Command, args []string) {
		// Build configuration from flags
		cfg = &config.Config{
			Database: config.DBConfig{
				Host:     host,
				Port:     port,
				User:     user,
				Password: password,
				DBName:   dbName,
				Driver:   driver,
			},
			Generator: config.GeneratorConfig{
				Tables:    table,
				OutputDir: outputDir,
			},
		}

		// Display current configuration
		fmt.Println("🚀 GoDB-Orm - Database Model Generator")
		fmt.Println("======================================")
		fmt.Printf("Host:     %s\n", cfg.Database.Host)
		fmt.Printf("Port:     %d\n", cfg.Database.Port)
		fmt.Printf("User:     %s\n", cfg.Database.User)
		fmt.Printf("Database: %s\n", cfg.Database.DBName)
		fmt.Printf("Driver:   %s\n", cfg.Database.Driver)
		fmt.Printf("Tables:   %s\n", cfg.Generator.Tables)
		fmt.Printf("Output:   %s\n", cfg.Generator.OutputDir)
		fmt.Println("======================================")

		// Validate required fields
		if cfg.Database.DBName == "" {
			fmt.Println("❌ Error: Database name is required (--db or -d)")
			os.Exit(1)
		}

		// Save configuration for future use
		if err := config.SaveConfig(cfg); err != nil {
			fmt.Printf("⚠️  Warning: Could not save config: %v\n", err)
		} else {
			fmt.Println("✅ Configuration saved to ~/.godb-orm/config.yaml")
		}

		// TODO: Implement actual database connection and model generation
		fmt.Println("\n📋 CLI mode is ready. Model generation will be implemented in Stage 2.")

		// Generate models if all required parameters are present
		if cfg.Database.DBName != "" && cfg.Database.Driver != "" {
			fmt.Println("\n🔄 Connecting to database...")

			introspector, err := database.NewIntrospector(&cfg.Database)
			if err != nil {
				fmt.Printf("❌ Error: %v\n", err)
				os.Exit(1)
			}

			if err := introspector.Connect(); err != nil {
				fmt.Printf("❌ Error connecting to database: %v\n", err)
				os.Exit(1)
			}
			defer introspector.Close()

			fmt.Println("✅ Connected to database successfully!")

			gen := generator.NewGenerator(introspector)

			// Get tables to generate
			var tablesToGenerate []string
			if cfg.Generator.Tables == "*" || cfg.Generator.Tables == "" {
				tables, err := introspector.GetTables()
				if err != nil {
					fmt.Printf("❌ Error getting tables: %v\n", err)
					os.Exit(1)
				}
				tablesToGenerate = tables
				fmt.Printf("📋 Found %d tables\n", len(tables))
			} else {
				tablesToGenerate = splitTables(cfg.Generator.Tables)
			}

			// Generate models
			fmt.Printf("\n🛠️  Generating models to %s...\n", cfg.Generator.OutputDir)
			for _, tableName := range tablesToGenerate {
				filePath, err := gen.GenerateToFile(tableName, cfg.Generator.OutputDir)
				if err != nil {
					fmt.Printf("  ❌ %s: %v\n", tableName, err)
					continue
				}
				fmt.Printf("  ✅ %s -> %s\n", tableName, filePath)
			}

			fmt.Println("\n🎉 Model generation complete!")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Load existing config as defaults
	existingCfg, _ := config.LoadConfig()

	// Database connection flags
	rootCmd.Flags().StringVarP(&host, "host", "H", existingCfg.Database.Host, "Database host")
	rootCmd.Flags().IntVarP(&port, "port", "P", existingCfg.Database.Port, "Database port")
	rootCmd.Flags().StringVarP(&user, "user", "u", existingCfg.Database.User, "Database user")
	rootCmd.Flags().StringVarP(&password, "pass", "p", existingCfg.Database.Password, "Database password")
	rootCmd.Flags().StringVarP(&dbName, "db", "d", existingCfg.Database.DBName, "Database name")
	rootCmd.Flags().StringVar(&driver, "driver", existingCfg.Database.Driver, "Database driver (mysql/postgres)")

	// Generator flags
	rootCmd.Flags().StringVarP(&table, "table", "t", existingCfg.Generator.Tables, "Table name(s) to generate (* for all)")
	rootCmd.Flags().StringVarP(&outputDir, "out", "o", existingCfg.Generator.OutputDir, "Output directory for generated files")
}

// splitTables splits a comma-separated list of table names
func splitTables(tables string) []string {
	var result []string
	for _, t := range strings.Split(tables, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}
