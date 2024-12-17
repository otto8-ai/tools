package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"obot-platform/database/pkg/cmd"

	"github.com/gptscript-ai/go-gptscript"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

var workspaceID = os.Getenv("DATABASE_WORKSPACE_ID")

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}
	command := os.Args[1]

	g, err := gptscript.NewGPTScript()
	if err != nil {
		fmt.Printf("Error creating GPTScript: %v\n", err)
		os.Exit(1)
	}
	defer g.Close()

	var (
		ctx             = context.Background()
		dbFileName      = "acorn.db"
		dbWorkspacePath = "/databases/" + dbFileName
	)

	// Read the database file from the workspace
	initialDBData, err := g.ReadFileInWorkspace(ctx, dbWorkspacePath, gptscript.ReadFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	})
	var notFoundErr *gptscript.NotFoundInWorkspaceError
	if err != nil && !errors.As(err, &notFoundErr) {
		fmt.Printf("Error reading DB file: %v\n", err)
		os.Exit(1)
	}

	// Create a temporary file for the SQLite database
	dbFile, err := os.CreateTemp("", dbFileName)
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer dbFile.Close()
	defer os.Remove(dbFile.Name())

	// Write the data to the temporary file
	if initialDBData != nil {
		if err := os.WriteFile(dbFile.Name(), initialDBData, 0644); err != nil {
			fmt.Printf("Error writing to temp file: %v\n", err)
			os.Exit(1)
		}
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", dbFile.Name())
	if err != nil {
		fmt.Printf("Error opening DB: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Run the requested command
	var result string
	switch command {
	case "listTables":
		result, err = cmd.ListTables(ctx, db)
	case "exec":
		result, err = cmd.Exec(ctx, db, os.Getenv("STATEMENT"))
		if err == nil {
			err = saveWorkspaceDB(ctx, g, dbWorkspacePath, dbFile, initialDBData)
		}
	case "query":
		result, err = cmd.Query(ctx, db, os.Getenv("QUERY"))
	case "context":
		result, err = cmd.Context(ctx, db)
	default:
		err = fmt.Errorf("unknown command: %s", command)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Print(result)
}

// saveWorkspaceDB saves the updated database file to the workspace if the content of the database has changed.
func saveWorkspaceDB(
	ctx context.Context,
	g *gptscript.GPTScript,
	dbWorkspacePath string,
	dbFile *os.File,
	initialDBData []byte,
) error {
	updatedDBData, err := os.ReadFile(dbFile.Name())
	if err != nil {
		return fmt.Errorf("Error reading updated DB file: %v", err)
	}

	if hash(initialDBData) == hash(updatedDBData) {
		return nil
	}

	if err := g.WriteFileInWorkspace(ctx, dbWorkspacePath, updatedDBData, gptscript.WriteFileInWorkspaceOptions{
		WorkspaceID: workspaceID,
	}); err != nil {
		return fmt.Errorf("Error writing updated DB file to workspace: %v", err)
	}

	return nil
}

// hash computes the SHA-256 hash of the given data and returns it as a hexadecimal string
func hash(data []byte) string {
	if data == nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
