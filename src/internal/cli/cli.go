package cli

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/tomventa/palude/internal/database"
)

// RunCLI starts the interactive CLI loop for the Database
func RunCLI(db *database.Database) {
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "💬 Enter your query: ",
		HistoryFile:     "/tmp/palude_history.tmp",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		fmt.Printf("Failed to initialize readline: %v\n", err)
		return
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		query := strings.TrimSpace(line)
		if query == "" {
			continue
		}

		if strings.ToLower(query) == "exit" || strings.ToLower(query) == "quit" {
			fmt.Println("👋 Goodbye!")
			break
		}

		if err := db.ProcessQuery(query); err != nil {
			fmt.Printf("❌ Error: %v\n\n", err)
		}
	}
}

// PrintDatabaseInfo prints parsed database connection info
func PrintDatabaseInfo(dsn string) {
	var user, host, port, dbname string
	user = "?"
	host = "?"
	port = "?"
	dbname = "?"

	at := strings.Index(dsn, "@tcp(")
	if at > 0 {
		user = dsn[:at]
		remain := dsn[at+5:]
		close := strings.Index(remain, ")")
		if close > 0 {
			hostport := remain[:close]
			colon := strings.Index(hostport, ":")
			if colon > 0 {
				host = hostport[:colon]
				port = hostport[colon+1:]
			} else {
				host = hostport
			}
			remain = remain[close+1:]
			if strings.HasPrefix(remain, "/") {
				remain = remain[1:]
				end := strings.IndexAny(remain, "?&")
				if end > 0 {
					dbname = remain[:end]
				} else {
					dbname = remain
				}
			}
		}
	}
	fmt.Printf("\n📦 Database connection info:\n")
	fmt.Printf("  User:     %s\n", user)
	fmt.Printf("  Host:     %s\n", host)
	fmt.Printf("  Port:     %s\n", port)
	fmt.Printf("  Database: %s\n\n", dbname)
}

// CheckOllamaStatus checks if Ollama is running and prints status
func CheckOllamaStatus(url string) {
	client := http.Client{Timeout: 1200 * time.Millisecond}
	resp, err := client.Get(strings.TrimRight(url, "/") + "/api/tags")
	if err != nil {
		fmt.Printf("⚠️  Ollama status: not reachable at %s\n", url)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Printf("🤖 Ollama status: running at %s\n\n", url)
	} else {
		fmt.Printf("⚠️  Ollama status: HTTP %d at %s\n", resp.StatusCode, url)
	}
}
