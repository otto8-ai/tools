package types

import (
	"slices"
)

type Document struct {
	ID              string         `json:"id"`
	Content         string         `json:"content"`
	Metadata        map[string]any `json:"metadata"`
	SimilarityScore float32        `json:"similarity_score"`
}

const (
	DocMetadataKeyDocIndex  = "docIndex"
	DocMetadataKeyDocsTotal = "docsTotal"
)

func mustInt(value any) int {
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		panic("Unsupported type")
	}
}

func SortDocumentsByMetadata(documents []Document, metadataKey string) {
	// Sort the documents by the metadata field, if present - else we have to assume the order is correct
	slices.SortFunc(documents, func(i, j Document) int {
		iVal, ok := i.Metadata[metadataKey]
		if !ok {
			return 0
		}
		jVal, ok := j.Metadata[metadataKey]
		if !ok {
			return 0
		}

		// Can be either int or float64 (if read from json)
		return mustInt(iVal) - mustInt(jVal)
	})
}

func SortDocumentsByDocIndex(documents []Document) {
	SortDocumentsByMetadata(documents, DocMetadataKeyDocIndex)
}

func SortAndEnsureDocIndex(documents []Document) {
	SortDocumentsByDocIndex(documents)
	l := len(documents)
	for i, doc := range documents {
		doc.Metadata[DocMetadataKeyDocIndex] = i
		doc.Metadata[DocMetadataKeyDocsTotal] = l
	}
}
