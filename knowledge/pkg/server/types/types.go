package types

import (
	"github.com/gptscript-ai/knowledge/pkg/datastore/textsplitter"
	"github.com/gptscript-ai/knowledge/pkg/index"
)

// Dataset represents a new knowledge vector space
type Dataset struct {
	// Dataset ID - must be a valid RFC 1123 hostname
	ID string `json:"id" format:"hostname_rfc1123" binding:"required" example:"asst-12345"`
}

// Query represents an incoming user query
type Query struct {
	Prompt string `json:"prompt" binding:"required"`
	TopK   *int   `json:"topk" example:"5" swaggertype:"integer"`
}

// Ingest represents incoming content that should be ingested
type Ingest struct {
	Filename         *string                        `json:"filename" `
	Content          string                         `json:"content" binding:"required,base64"`
	FileMetadata     *index.FileMetadata            `json:"metadata"`
	TextSplitterOpts *textsplitter.TextSplitterOpts `json:"text_splitter_opts"`
}

type IngestResponse struct {
	Documents []string `json:"documents"`
}
