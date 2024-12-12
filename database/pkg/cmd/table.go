package cmd

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

func ListTables(ctx context.Context, db *sql.DB) (string, error) {
	tables, err := listTables(ctx, db)
	if err != nil {
		return "", fmt.Errorf("failed to list tables: %w", err)
	}

	content, err := json.Marshal(tables)
	return string(content), err
}

type tables struct {
	Tables []Table `json:"tables"`
}

type Table struct {
	Name string `json:"name,omitempty"`
}

func listTables(ctx context.Context, db *sql.DB) (tables, error) {
	rows, err := db.QueryContext(ctx, "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%';")
	if err != nil {
		return tables{}, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	var tables tables
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return tables, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables.Tables = append(tables.Tables, Table{
			Name: tableName,
		})
	}
	if rows.Err() != nil {
		return tables, fmt.Errorf("error iterating over table names: %w", rows.Err())
	}

	return tables, nil
}
