package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"

	"sqlite/pkg/cmd"

	"github.com/gptscript-ai/go-gptscript"
	_ "github.com/ncruces/go-sqlite3/driver"
	_ "github.com/ncruces/go-sqlite3/embed"
)

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
		ctx               = context.Background()
		fileName          = "otto8.db"
		workspaceFileName = "files/" + fileName
	)

	// Read the database file from the workspace
	data, err := g.ReadFileInWorkspace(ctx, workspaceFileName)
	var notFoundErr *gptscript.NotFoundInWorkspaceError
	if err != nil && !errors.As(err, &notFoundErr) {
		fmt.Printf("Error reading DB file: %v\n", err)
		os.Exit(1)
	}

	// Calculate hash of the initial data
	initialHash := hash(data)

	// Create a temporary file for the SQLite database
	tmpFile, err := os.CreateTemp("", fileName)
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpFile.Name())

	// Write the data to the temporary file
	if data != nil {
		if err := os.WriteFile(tmpFile.Name(), data, 0644); err != nil {
			fmt.Printf("Error writing to temp file: %v\n", err)
			os.Exit(1)
		}
	}

	// Open the SQLite database
	db, err := sql.Open("sqlite3", tmpFile.Name())
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
	case "getSchema":
		result, err = cmd.GetSchema(ctx, db, os.Getenv("TABLE_NAME"))
	case "exec":
		result, err = cmd.Exec(ctx, db, os.Getenv("STATEMENT"))
	case "query":
		result, err = cmd.Query(ctx, db, os.Getenv("QUERY"))
	case "context":
		result, err = cmd.Context(ctx, db)
	default:
		err = fmt.Errorf("Unknown command: %s", command)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Close the database before writing back the updated file
	if err := db.Close(); err != nil {
		fmt.Printf("Error closing DB: %v\n", err)
		os.Exit(1)
	}

	// Read the updated file
	updatedData, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		fmt.Printf("Error reading updated DB file: %v\n", err)
		os.Exit(1)
	}

	// Calculate hash of the updated data
	updatedHash := hash(updatedData)

	// Write the updated file back to the workspace only if the hash has changed
	if initialHash != updatedHash {
		if err := g.WriteFileInWorkspace(ctx, workspaceFileName, updatedData); err != nil {
			fmt.Printf("Error writing updated DB file to workspace: %v\n", err)
			os.Exit(1)
		}
	}

	fmt.Print(result)
}

// hash computes the SHA-256 hash of the given data and returns it as a hexadecimal string
func hash(data []byte) string {
	if data == nil {
		return ""
	}
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
