package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type Output struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
}

// Query executes a SQL query (e.g., SELECT) and returns the result formatted in JSON
func Query(ctx context.Context, db *sql.DB, query string) (string, error) {
	if query == "" {
		return "", fmt.Errorf("empty query")
	}

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("error executing query: %v", err)
	}
	defer rows.Close()

	// Retrieve column names
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("error retrieving columns: %v", err)
	}

	var output = Output{
		Columns: columns,
		Rows:    []map[string]any{},
	}

	// Prepare a slice of interface{} for each row's column values
	values := make([]interface{}, len(columns))
	valuePointers := make([]interface{}, len(columns))
	for i := range values {
		valuePointers[i] = &values[i]
	}

	// Fetch rows and write their contents
	for rows.Next() {
		err := rows.Scan(valuePointers...)
		if err != nil {
			return "", fmt.Errorf("error scanning row: %w", err)
		}

		// Convert values to strings
		rowData := map[string]any{}
		for i, val := range values {
			rowData[columns[i]] = val
		}
		output.Rows = append(output.Rows, rowData)
	}

	content, err := json.Marshal(output)
	return string(content), err
}
