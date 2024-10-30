package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/gptscript-ai/go-gptscript"
)

type output struct {
	Results []subqueryResults `json:"subqueryResults"`
}

type subqueryResults struct {
	Subquery        string     `json:"subquery"`
	ResultDocuments []document `json:"resultDocuments"`
}

type document struct {
	ID       string            `json:"id"`
	Content  string            `json:"content,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type hit struct {
	URL     string `json:"url,omitempty"`
	Content string `json:"content,omitempty"`
}

func main() {
	var (
		output            output
		out               = gptscript.GetEnv("OUTPUT", "")
		client, clientErr = gptscript.NewGPTScript()
		ctx               = context.Background()
	)

	// This is ugly code, I know. Beauty comes later.

	if clientErr != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to create gptscript client: %v\n", clientErr)
	}

	if err := json.Unmarshal([]byte(out), &output); err != nil {
		fmt.Print(out)
		return
	}

	var (
		outDocs []hit
		wg      sync.WaitGroup
	)

	for _, result := range output.Results {
		if len(outDocs) >= 10 {
			break
		}
		for _, doc := range result.ResultDocuments {
			outDocs = append(outDocs, hit{
				URL:     doc.Metadata["url"],
				Content: doc.Content,
			})
			index := len(outDocs) - 1
			if index < 3 && clientErr == nil {
				size, _ := strconv.Atoi(doc.Metadata["fileSize"])
				workspaceID := doc.Metadata["workspaceID"]
				filename := doc.Metadata["workspaceFileName"]
				if size > 5_000 && size < 100_000 {
					_, _ = fmt.Fprintf(os.Stderr, "reading file in workspace: %s\n", filename)
					wg.Add(1)
					go func() {
						defer wg.Done()

						content, err := client.ReadFileInWorkspace(ctx, filename, gptscript.ReadFileInWorkspaceOptions{
							WorkspaceID: workspaceID,
						})
						if err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "failed to read file in workspace: %v\n", err)
						}

						outDocs[index].Content = string(content)
					}()
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "file size is not within the range: %s %s %d\n", workspaceID, filename, size)
				}
			}
		}
	}
	wg.Wait()
	if len(outDocs) == 0 {
		_, _ = fmt.Println("no relevant documents found")
		return
	}
	_ = json.NewEncoder(os.Stdout).Encode(outDocs)
}
