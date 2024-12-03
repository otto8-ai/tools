package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// GetSchema returns the schema of a table as a markdown string.
func GetSchema(ctx context.Context, db *sql.DB, tableName string) (string, error) {
	// Query to get the schema of the specified table
	query := fmt.Sprintf("PRAGMA table_info(%s);", tableName)
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to query schema for table %s: %w", tableName, err)
	}
	defer rows.Close()

	// Process schema details
	var schemaDetails []string
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dfltValue sql.NullString

		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err != nil {
			return "", fmt.Errorf("failed to scan schema for table %s: %w", tableName, err)
		}

		// Format column details
		column := fmt.Sprintf("- `%s` (%s)", name, ctype)
		if notnull == 1 {
			column += " NOT NULL"
		}
		if pk == 1 {
			column += " PRIMARY KEY"
		}
		if dfltValue.Valid {
			column += fmt.Sprintf(" DEFAULT %s", dfltValue.String)
		}
		schemaDetails = append(schemaDetails, column)
	}
	if rows.Err() != nil {
		return "", fmt.Errorf("error iterating over schema for table %s: %w", tableName, rows.Err())
	}

	// If no schema details are found
	if len(schemaDetails) == 0 {
		return fmt.Sprintf("No schema found for table `%s`.", tableName), nil
	}

	// Build the markdown output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Schema for Table: `%s`\n", tableName))
	sb.WriteString(strings.Join(schemaDetails, "\n"))
	return sb.String(), nil
}
