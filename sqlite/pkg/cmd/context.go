package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Context returns the context text for the SQLite tools.
// This contains all of the schemas of the tables in the database in markdown format.
func Context(ctx context.Context, db *sql.DB) (string, error) {
	tables, err := listTables(ctx, db)
	if err != nil {
		return "", fmt.Errorf("failed to list tables: %w", err)
	}

	// Build the markdown output
	var sb strings.Builder
	sb.WriteString(`# START INSTRUCTIONS: SQLite Tools
You have access to tools for interacting with a SQLite database.
The Exec tool only accepts valid SQLite3 statements.
The Query tool only accepts valid SQLite3 queries.
The Query tool returns query results in markdown format.
Display all results from these tools in markdown format.

`)
	for i, table := range tables {
		if i == 0 {
			sb.WriteString("\n# START CURRENT TABLE SCHEMAS\n")
		}
		schema, err := GetSchema(ctx, db, table)
		if err != nil {
			return "", fmt.Errorf("failed to get schema for table %s: %w", table, err)
		}
		sb.WriteString(schema + "\n")
		if i == len(tables)-1 {
			sb.WriteString("# END CURRENT TABLE SCHEMAS\n")
		}
	}

	sb.WriteString("# END INSTRUCTIONS: SQLite Tools\n")
	return sb.String(), nil
}
