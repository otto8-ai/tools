package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func ListTables(ctx context.Context, db *sql.DB) (string, error) {
	tables, err := listTables(ctx, db)
	if err != nil {
		return "", fmt.Errorf("failed to list tables: %w", err)
	}

	if len(tables) == 0 {
		return "No tables found in the database.", nil
	}

	// Build the markdown list
	var sb strings.Builder
	sb.WriteString("# Database Tables\n")
	for _, table := range tables {
		sb.WriteString(fmt.Sprintf("- **%s**\n", table))
	}

	return sb.String(), nil
}

func listTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("error iterating over table names: %w", rows.Err())
	}

	return tables, nil
}
