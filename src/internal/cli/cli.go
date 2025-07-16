package cli

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/tomventa/palude/internal/database"
)

// RunCLI starts the interactive CLI loop for the Database
func RunCLI(db *database.Database) {
	fmt.Println("🗄️  Palude - Natural Language Database Query Tool")
	fmt.Printf("Type 'exit' to quit\n\n")

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
