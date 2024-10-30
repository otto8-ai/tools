package structured

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/knadh/koanf/maps"
)

type StructuredInputDocument struct {
	Metadata map[string]any `json:"metadata"`
	Content  string         `json:"content"`
}

type StructuredInput struct {
	Metadata  map[string]any            `json:"metadata"`
	Documents []StructuredInputDocument `json:"documents"`
}

type Structured struct{}

func (s *Structured) Load(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
	var input StructuredInput
	if err := json.NewDecoder(reader).Decode(&input); err != nil {
		return nil, fmt.Errorf("failed to decode input: %w", err)
	}

	docs := make([]vs.Document, 0, len(input.Documents))
	for _, doc := range input.Documents {
		maps.Merge(maps.Copy(input.Metadata), doc.Metadata)
		docs = append(docs, vs.Document{
			Content:  doc.Content,
			Metadata: doc.Metadata,
		})
	}

	return docs, nil
}
