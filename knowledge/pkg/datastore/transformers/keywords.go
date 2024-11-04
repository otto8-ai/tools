package transformers

import (
	"context"
	"log/slog"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/llm"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

const KeywordExtractorName = "keywords"

type KeywordExtractor struct {
	NumKeywords int
	Model       llm.LLMConfig
}

func (k *KeywordExtractor) extractKeywords(ctx context.Context, doc vs.Document) ([]string, error) {
	m, err := llm.NewFromConfig(k.Model)
	if err != nil {
		return nil, err
	}
	// Implement keyword extraction here
	result, err := m.Prompt(ctx, tpl, map[string]any{"numKeywords": k.NumKeywords, "content": strings.TrimSpace(doc.Content)})
	if err != nil {
		return nil, err
	}
	keywords := strings.Split(result, ",")
	return keywords, nil
}

var tpl = `Extract {.numKeywords} keywords from the following document and return them as a comma-separated list:
{.content}
`

func (k *KeywordExtractor) Transform(ctx context.Context, docs []vs.Document) ([]vs.Document, error) {
	slog.Debug("Extracting keywords from documents")
	for i, doc := range docs {
		keywords, err := k.extractKeywords(ctx, doc)
		if err != nil {
			return nil, err
		}
		slog.Debug("Extracted keywords", "keywords", keywords)
		docs[i].Metadata["keywords"] = strings.Join(keywords, ",")
	}
	return docs, nil
}

func (k *KeywordExtractor) Name() string {
	return KeywordExtractorName
}
