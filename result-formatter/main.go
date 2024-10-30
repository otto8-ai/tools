package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
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
	ID       string         `json:"id"`
	Content  string         `json:"content,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type hit struct {
	URL     string `json:"url,omitempty"`
	Content string `json:"content,omitempty"`
}

type inputContent struct {
	Documents []document `json:"documents"`
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
		_, _ = fmt.Fprintf(os.Stderr, "failed to unmarshal output: %v\n", err)
		fmt.Print(out)
		return
	}

	var (
		outDocs      []hit
		wg           sync.WaitGroup
		fullyFetched = map[string]struct{}{}
		budget       = 120_000
	)

	for _, result := range output.Results {
		if len(outDocs) >= 10 {
			break
		}
		for _, doc := range result.ResultDocuments {
			filename, _ := doc.Metadata["workspaceFileName"].(string)
			if _, ok := fullyFetched[filename]; ok {
				continue
			}

			url, _ := doc.Metadata["url"].(string)
			outDocs = append(outDocs, hit{
				URL:     url,
				Content: doc.Content,
			})

			index := len(outDocs) - 1

			if index < 3 && clientErr == nil {
				fileSize, _ := doc.Metadata["fileSize"].(string)
				size, _ := strconv.Atoi(fileSize)
				workspaceID, _ := doc.Metadata["workspaceID"].(string)
				if size > 5_000 && size < budget && workspaceID != "" {
					_, _ = fmt.Fprintf(os.Stderr, "reading file in workspace: %s\n", filename)
					fullyFetched[filename] = struct{}{}
					budget -= size
					wg.Add(1)

					go func() {
						defer wg.Done()

						content, err := client.ReadFileInWorkspace(ctx, filename, gptscript.ReadFileInWorkspaceOptions{
							WorkspaceID: workspaceID,
						})
						if err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "failed to read file in workspace: %v\n", err)
							return
						}

						var sourceContent inputContent
						if err := json.Unmarshal(content, &sourceContent); err != nil {
							_, _ = fmt.Fprintf(os.Stderr, "failed to unmarshal content: %v\n", err)
							return
						}

						var buffer strings.Builder
						for _, sourceContentDocument := range sourceContent.Documents {
							buffer.WriteString(sourceContentDocument.Content)
						}

						if buffer.Len() > 0 {
							outDocs[index].Content = buffer.String()
						}
					}()
				} else {
					_, _ = fmt.Fprintf(os.Stderr, "file size is not within the range: %s %s %d %d\n", workspaceID, filename, size, budget)
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
