package main

import (
	"encoding/json"
	"fmt"
	"os"

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
		output output
		out    = gptscript.GetEnv("OUTPUT", "")
	)

	if err := json.Unmarshal([]byte(out), &output); err != nil {
		fmt.Print(out)
		return
	}

	var outDocs []hit
	for _, result := range output.Results {
		for _, doc := range result.ResultDocuments {
			outDocs = append(outDocs, hit{
				URL:     doc.Metadata["url"],
				Content: doc.Content,
			})
		}
	}
	json.NewEncoder(os.Stdout).Encode(outDocs)
}
