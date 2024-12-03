package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Query executes a SQL query (e.g., SELECT) and returns the result formatted in Markdown.
func Query(ctx context.Context, db *sql.DB, query string) (string, error) {
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

	var output bytes.Buffer

	// Write the Markdown header row
	output.WriteString("| " + strings.Join(columns, " | ") + " |\n")
	output.WriteString("| " + strings.Repeat("--- | ", len(columns)) + "\n")

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
		rowData := make([]string, len(columns))
		for i, val := range values {
			if val == nil {
				rowData[i] = "NULL"
			} else {
				rowData[i] = fmt.Sprintf("%v", val)
			}
		}
		output.WriteString("| " + strings.Join(rowData, " | ") + " |\n")
	}

	return output.String(), nil
}
