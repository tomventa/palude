package tableprint

import (
	"fmt"
	"strings"
)

// PrintTable prints a table with headers and rows in a formatted way
func PrintTable(headers []string, rows [][]string) {
	// Calculate max width for each column
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	fmt.Print("  |")
	for i, h := range headers {
		fmt.Printf(" %-*s |", widths[i], h)
	}
	fmt.Println()
	fmt.Print("  +")
	for _, w := range widths {
		fmt.Print(strings.Repeat("-", w+2) + "+")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Print("  |")
		for i, cell := range row {
			fmt.Printf(" %-*s |", widths[i], cell)
		}
		fmt.Println()
	}
}
