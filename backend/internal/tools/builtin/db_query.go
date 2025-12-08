package builtin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jacklau/prism/internal/llm"
)

// DatabaseQueryTool executes read-only SQL queries against a user's sandbox database
type DatabaseQueryTool struct {
	db *sql.DB
}

// NewDatabaseQueryTool creates a new database query tool
func NewDatabaseQueryTool(db *sql.DB) *DatabaseQueryTool {
	return &DatabaseQueryTool{db: db}
}

func (t *DatabaseQueryTool) Name() string {
	return "query_database"
}

func (t *DatabaseQueryTool) Description() string {
	return "Execute a read-only SQL query against the database. Only SELECT queries are allowed. Returns the query results as a list of rows."
}

func (t *DatabaseQueryTool) Parameters() llm.JSONSchema {
	return llm.JSONSchema{
		Type: "object",
		Properties: map[string]llm.JSONProperty{
			"query": {
				Type:        "string",
				Description: "The SQL SELECT query to execute",
			},
			"limit": {
				Type:        "integer",
				Description: "Maximum number of rows to return (default: 100, max: 1000)",
				Default:     100,
			},
		},
		Required: []string{"query"},
	}
}

func (t *DatabaseQueryTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	query, ok := params["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query parameter is required")
	}

	// Validate that the query is a SELECT statement (read-only)
	trimmedQuery := strings.TrimSpace(strings.ToUpper(query))
	if !strings.HasPrefix(trimmedQuery, "SELECT") {
		return nil, fmt.Errorf("only SELECT queries are allowed")
	}

	// Check for dangerous keywords
	dangerousKeywords := []string{"INSERT", "UPDATE", "DELETE", "DROP", "CREATE", "ALTER", "TRUNCATE", "EXEC", "EXECUTE"}
	for _, keyword := range dangerousKeywords {
		if strings.Contains(trimmedQuery, keyword) {
			return nil, fmt.Errorf("query contains disallowed keyword: %s", keyword)
		}
	}

	limit := 100
	if l, ok := params["limit"].(float64); ok {
		limit = int(l)
		if limit < 1 {
			limit = 1
		}
		if limit > 1000 {
			limit = 1000
		}
	}

	// Add LIMIT if not present
	if !strings.Contains(strings.ToUpper(query), "LIMIT") {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	}

	rows, err := t.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Collect results
	results := make([]map[string]interface{}, 0)
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))

	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			// Convert []byte to string for readability
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return map[string]interface{}{
		"columns":   columns,
		"rows":      results,
		"row_count": len(results),
	}, nil
}

func (t *DatabaseQueryTool) RequiresConfirmation() bool {
	return false // Read-only queries don't need confirmation
}
