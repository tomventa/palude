package database

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"github.com/tomventa/palude/internal/config"
	"github.com/tomventa/palude/internal/ollama"
	"github.com/tomventa/palude/internal/tableprint"
	"github.com/tomventa/palude/internal/utils"
)

// Database represents the database connection and operations
type Database struct {
	db           *sql.DB
	config       *config.Config
	ollamaClient *ollama.Client
}

// New creates a new Database instance
func New(cfg *config.Config) (*Database, error) {
	db, err := sql.Open("mysql", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{
		db:           db,
		config:       cfg,
		ollamaClient: ollama.New(cfg),
	}, nil
}

// Close closes the database connection
func (d *Database) Close() error {
	return d.db.Close()
}

// getSchema retrieves and formats the database schema
func (d *Database) getSchema() (string, error) {
	rows, err := d.db.Query("SHOW TABLES")
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var schema strings.Builder
	schema.WriteString("Database Schema:\n\n")

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			continue
		}
		tables = append(tables, tableName)
	}

	for _, tableName := range tables {
		// Use SHOW CREATE TABLE to get the complete table definition
		var tableCreate string
		var dummy string // MySQL returns table name as first column
		err := d.db.QueryRow("SHOW CREATE TABLE "+tableName).Scan(&dummy, &tableCreate)
		if err != nil {
			continue
		}

		schema.WriteString(fmt.Sprintf("Table: %s\n", tableName))
		schema.WriteString(tableCreate)
		schema.WriteString("\n\n")
	}

	return schema.String(), nil
}

// ProcessQuery handles the natural language to SQL conversion and execution
func (d *Database) ProcessQuery(userQuery string) error {
	schema, err := d.getSchema()
	// fmt.Printf("📜 Current Schema:\n%s\n", schema)
	if err != nil {
		return fmt.Errorf("failed to get schema: %w", err)
	}

	var sqlQuery string
	var lastError string

	for attempt := 1; attempt <= d.config.MaxAttempts; attempt++ {
		if attempt == 1 {
			fmt.Printf("🤖 Generating SQL query...\n")
		} else {
			fmt.Printf("🤖 Generating SQL query (attempt %d/%d)...\n", attempt, d.config.MaxAttempts)
		}

		var prompt string
		if attempt == 1 {
			// First attempt - basic prompt
			prompt = fmt.Sprintf(`
			Given this MySQL database schema:

%s

Convert this natural language query to a SQL SELECT statement:
"%s"

Requirements:
- Only return the SQL statement, nothing else
- Use proper MySQL syntax
- Only generate READ queries (SELECT statements)
- Do not include any explanations or markdown

SQL:
		`, schema, userQuery)
		} else {
			// Retry attempt - include the previous error
			prompt = fmt.Sprintf(`Given this MySQL database schema:

%s

I tried to convert this natural language query to SQL:
"%s"

The previous SQL query was:
%s

But it failed with this MySQL error:
%s

Please fix the SQL query to resolve this error.

Requirements:
- Only return the corrected SQL statement, nothing else
- Use proper MySQL syntax
- Only generate READ queries (SELECT statements)
- Do not include any explanations or markdown

Corrected SQL:`, schema, userQuery, sqlQuery, lastError)
		}

		generatedSQL, err := d.ollamaClient.Query(prompt)
		if err != nil {
			return fmt.Errorf("failed to query Ollama on attempt %d: %w", attempt, err)
		}

		// Clean up the response
		sqlQuery = utils.CleanSQLResponse(generatedSQL)

		fmt.Printf("📝 Generated SQL: %s\n\n", sqlQuery)

		// Ask for confirmation only on first attempt or if query is not a simple SELECT
		shouldAskConfirmation := attempt == 1 || !utils.IsReadOnlyQuery(sqlQuery)

		if shouldAskConfirmation {
			fmt.Print("❓ Execute this query? (y/N): ")
			scanner := bufio.NewScanner(os.Stdin)
			if !scanner.Scan() {
				return fmt.Errorf("failed to read confirmation")
			}

			confirmation := strings.ToLower(strings.TrimSpace(scanner.Text()))
			if confirmation != "y" && confirmation != "yes" {
				fmt.Printf("❌ Query execution cancelled.\n\n")
				return nil
			}
		} else {
			fmt.Println("🔄 Auto-executing read-only retry query...")
		}

		// Try to execute the query
		err = d.executeQuery(sqlQuery)
		if err == nil {
			// Success!
			return nil
		}

		// Store the error for the next attempt
		lastError = err.Error()
		fmt.Printf("❌ Query failed: %v\n", err)

		if attempt < d.config.MaxAttempts {
			fmt.Printf("🔄 Trying to auto-fix the issue...\n\n")
		}
	}

	return fmt.Errorf("failed to generate working SQL after %d attempts. Last error: %s", d.config.MaxAttempts, lastError)
}

// executeQuery executes a SQL query and displays the results
func (d *Database) executeQuery(sqlQuery string) error {
	// Execute the query
	rows, err := d.db.Query(sqlQuery)
	if err != nil {
		return err
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	fmt.Println("📊 Results:")
	fmt.Println(strings.Repeat("-", 50))

	// Print header
	for i, col := range columns {
		if i > 0 {
			fmt.Print(" | ")
		}
		fmt.Printf("%-15s", col)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 50))

	// Print rows
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	rowCount := 0
	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		for i, val := range values {
			if i > 0 {
				fmt.Print(" | ")
			}
			if val != nil {
				switch v := val.(type) {
				case []byte:
					fmt.Printf("%-15s", string(v))
				default:
					fmt.Printf("%-15v", v)
				}
			} else {
				fmt.Printf("%-15s", "NULL")
			}
		}
		fmt.Println()
		rowCount++
	}

	fmt.Printf("\n✅ Query executed successfully. %d rows returned.\n\n", rowCount)
	return nil
}

// ShowTablesInfo prints all tables, their columns, and row counts
func (d *Database) ShowTablesInfo() {
	rows, err := d.db.Query("SHOW TABLES")
	if err != nil {
		fmt.Printf("Error fetching tables: %v\n", err)
		return
	}
	defer rows.Close()

	tables := []string{}
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err == nil {
			tables = append(tables, table)
		}
	}

	if len(tables) == 0 {
		fmt.Println("No tables found.")
		return
	}

	for _, table := range tables {
		fmt.Printf("\n📋 Table: %s\n", table)

		// Get columns
		colRows, err := d.db.Query("SHOW COLUMNS FROM `" + table + "`")
		if err != nil {
			fmt.Printf("  Error fetching columns: %v\n", err)
			continue
		}

		// Prepare headers and rows for tableprint.PrintTable
		headers := []string{"Field", "Type", "Null", "Key", "Default", "Extra"}
		var rowsData [][]string
		for colRows.Next() {
			var field, colType, null, key string
			var def sql.NullString
			var extra string
			if err := colRows.Scan(&field, &colType, &null, &key, &def, &extra); err == nil {
				defVal := "NULL"
				if def.Valid {
					defVal = def.String
				}
				rowsData = append(rowsData, []string{field, colType, null, key, defVal, extra})
			}
		}
		colRows.Close()

		tableprint.PrintTable(headers, rowsData)

		// Get row count
		var count int
		cntRow := d.db.QueryRow("SELECT COUNT(*) FROM `" + table + "`")
		if err := cntRow.Scan(&count); err == nil {
			fmt.Printf("  Rows: %d\n", count)
		} else {
			fmt.Printf("  Rows: ? (error: %v)\n", err)
		}
	}
	fmt.Println()
}
