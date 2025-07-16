package utils

import "strings"

// IsReadOnlyQuery checks if a SQL query is read-only (safe to execute)
func IsReadOnlyQuery(query string) bool {
	// Trim whitespace and convert to lowercase
	cleanQuery := strings.TrimSpace(strings.ToLower(query))

	// Remove common SQL comments
	cleanQuery = strings.ReplaceAll(cleanQuery, "--", "")
	cleanQuery = strings.ReplaceAll(cleanQuery, "/*", "")
	cleanQuery = strings.ReplaceAll(cleanQuery, "*/", "")
	cleanQuery = strings.TrimSpace(cleanQuery)

	// Check if it starts with SELECT (allowing for WITH clauses)
	return strings.HasPrefix(cleanQuery, "select") ||
		strings.HasPrefix(cleanQuery, "with") ||
		strings.HasPrefix(cleanQuery, "show") ||
		strings.HasPrefix(cleanQuery, "describe") ||
		strings.HasPrefix(cleanQuery, "desc") ||
		strings.HasPrefix(cleanQuery, "explain")
}

// CleanSQLResponse removes markdown formatting from generated SQL
func CleanSQLResponse(sqlQuery string) string {
	// Clean up the response
	cleaned := strings.TrimSpace(sqlQuery)
	cleaned = strings.TrimPrefix(cleaned, "```sql")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}
