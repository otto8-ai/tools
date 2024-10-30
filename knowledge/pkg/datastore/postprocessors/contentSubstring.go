package postprocessors

import (
	"context"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

const ContentSubstringFilterPostprocessorName = "content_substring_filter"

type ContentSubstringFilterPostprocessor struct {
	Contains    []string
	NotContains []string
}

func (c *ContentSubstringFilterPostprocessor) Transform(ctx context.Context, response *types.RetrievalResponse) error {
	for i, resp := range response.Responses {
		var filteredDocs []vs.Document
		for _, doc := range resp.ResultDocuments {
			containsOK := true
			for _, contains := range c.Contains {
				if !strings.Contains(doc.Content, contains) {
					containsOK = false
					break
				}
			}

			notContainsOK := true
			for _, notContains := range c.NotContains {
				if strings.Contains(doc.Content, notContains) {
					notContainsOK = false
					break
				}
			}

			if containsOK && notContainsOK {
				filteredDocs = append(filteredDocs, doc)
			}
		}
		response.Responses[i].ResultDocuments = filteredDocs
	}
	return nil
}

func (c *ContentSubstringFilterPostprocessor) Name() string {
	return ContentSubstringFilterPostprocessorName
}
