package cmd

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Context returns the context text for the SQLite tools.
// The resulting string contains all of schemas in the database.
func Context(ctx context.Context, db *sql.DB) (string, error) {
	// Build the markdown output
	var out strings.Builder
	out.WriteString(`# START INSTRUCTIONS: Database Tools

You have access to tools for interacting with a SQLite database.
The Exec tool only accepts valid SQLite3 statements.
The Query tool only accepts valid SQLite3 queries.
Display all results from these tools and their schemas in markdown format.
If the user refers to creating or modifying tables assume they mean a SQLite3 table and not writing a table
in a markdown file.

# END INSTRUCTIONS: Database Tools
`)

	// Add the schemas section
	schemas, err := getSchemas(ctx, db)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve schemas: %w", err)
	}
	if schemas != "" {
		out.WriteString("# START CURRENT DATABASE SCHEMAS\n")
		out.WriteString(schemas)
		out.WriteString("\n# END CURRENT DATABASE SCHEMAS\n")
	} else {
		out.WriteString("# DATABASE HAS NO TABLES\n")
	}

	return out.String(), nil
}

// getSchemas returns an SQL string containing all schemas in the database.
func getSchemas(ctx context.Context, db *sql.DB) (string, error) {
	query := "SELECT sql FROM sqlite_master WHERE type IN ('table', 'index', 'view', 'trigger') AND name NOT LIKE 'sqlite_%' ORDER BY name"

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return "", fmt.Errorf("failed to query sqlite_master: %w", err)
	}
	defer rows.Close()

	var out strings.Builder
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return "", fmt.Errorf("failed to scan schema: %w", err)
		}
		if schema != "" {
			out.WriteString(fmt.Sprintf("\n%s\n", schema))
		}
	}

	if rows.Err() != nil {
		return "", fmt.Errorf("error iterating over schemas: %w", rows.Err())
	}

	return out.String(), nil
}
