package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/gptscript-ai/go-gptscript"
)

const FilesDir = "files"

var (
	FileEnv     = os.Getenv("FILENAME")
	MaxFileSize = 1 << 16
)

func main() {
	if len(os.Args) == 1 {
		fmt.Printf(`
Subcommands: read, write
env: FILENAME, CONTENT, GPTSCRIPT_WORKSPACE_DIR
Usage: go run main.go <path>\n`)
		return
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cmd := os.Args[1]
	if cmd == "read" && (FileEnv == "" || strings.HasSuffix(FileEnv, "/")) {
		cmd = "list"
	}

	switch cmd {
	case "input":
		input(ctx)
		return
	case "list":
		if err := list(ctx, FileEnv); err != nil {
			fmt.Printf("Failed to list %s: %v\n", FileEnv, err)
			return
		}
	case "read":
		if err := read(ctx, FileEnv); err != nil {
			fmt.Printf("Failed to read %s: %v\n", FileEnv, err)
			return
		}
	case "write":
		content := gptscript.GetEnv("CONTENT", "")
		if err := write(ctx, FileEnv, content); err != nil {
			fmt.Printf("Failed to write %s: %v\n", FileEnv, err)
			return
		}
		fmt.Printf("Wrote %d bytes\n", len(content))
	}
}

type data struct {
	Prompt       string            `json:"prompt,omitempty"`
	Explain      *explain          `json:"explain,omitempty"`
	Improve      *explain          `json:"improve,omitempty"`
	ChangedFiles map[string]string `json:"changedFiles,omitempty"`
}

type explain struct {
	Filename  string `json:"filename,omitempty"`
	Selection string `json:"selection,omitempty"`
}

func inBackTicks(s string) string {
	return "\n```\n" + s + "\n```\n"
}

func input(ctx context.Context) {
	var (
		input = gptscript.GetEnv("INPUT", "")
		data  data
	)

	if err := json.Unmarshal([]byte(input), &data); err != nil {
		fmt.Print(input)
		return
	}

	if data.Explain != nil {
		fmt.Printf(`Explain the following selection from the "%s" workspace file: %s`,
			data.Explain.Filename, inBackTicks(data.Explain.Selection))
	}

	if data.Improve != nil {
		if data.Improve.Selection == "" {
			fmt.Printf(`Refering to the workspace file "%s", %s
Write any suggested changes back to the file`, data.Improve.Filename, data.Prompt)
		} else {
			fmt.Printf(`Refering to the below selection from the workspace file "%s", %s: %s
Write any suggested changes back to the file.`,
				data.Improve.Filename, data.Prompt, inBackTicks(data.Improve.Selection))
		}
	}

	if len(data.ChangedFiles) > 0 {
		var printed bool
		c, err := gptscript.NewGPTScript()
		for filename, content := range data.ChangedFiles {
			if err == nil {
				if err := c.WriteFileInWorkspace(ctx, path.Join(FilesDir, filename), []byte(content)); err == nil {
					if !printed {
						printed = true
						fmt.Println("The following files have been changed in the workspace:")
					}
					fmt.Printf("File: %s\n%s\n", filename, inBackTicks(content))
				}
			}
		}
		fmt.Println("")
	}

	if data.Prompt != "" {
		fmt.Print(data.Prompt)
	}
}

func list(ctx context.Context, filename string) error {
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	files, err := client.ListFilesInWorkspace(ctx, gptscript.ListFilesInWorkspaceOptions{
		Prefix: path.Join(FilesDir, filename),
	})
	if err != nil {
		return err
	}

	for _, file := range files {
		p := strings.TrimPrefix(file, FilesDir+"/")
		if p != "" {
			fmt.Println(p)
		}
	}

	return nil
}

func read(ctx context.Context, filename string) error {
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	data, err := client.ReadFileInWorkspace(ctx, path.Join(FilesDir, filename))
	if err != nil {
		return err
	}

	if len(data) > MaxFileSize {
		return fmt.Errorf("file size exceeds %d bytes", MaxFileSize)
	}

	if utf8.Valid(data) {
		fmt.Println(string(data))
		return nil
	}

	return fmt.Errorf("file is not valid UTF-8")
}

func write(ctx context.Context, filename, content string) error {
	client, err := gptscript.NewGPTScript()
	if err != nil {
		return err
	}

	return client.WriteFileInWorkspace(ctx, path.Join(FilesDir, filename), []byte(content))
}
