package main

import (
	"fmt"
	"log"

	"github.com/joho/godotenv"

	"github.com/tomventa/palude/internal/cli"
	"github.com/tomventa/palude/internal/config"
	"github.com/tomventa/palude/internal/database"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading: %v", err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := database.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("🗄️  Palude - Natural Language Database Query Tool")

	// (moved to cli package)
	// Print database info
	cli.PrintDatabaseInfo(cfg.DatabaseURL)

	// (moved to cli package)
	// Check Ollama status
	cli.CheckOllamaStatus(cfg.OllamaURL)

	// Start the interactive CLI
	cli.RunCLI(db)
}
