package types

type Document struct {
	ID              string         `json:"id"`
	Content         string         `json:"content"`
	Metadata        map[string]any `json:"metadata"`
	SimilarityScore float32        `json:"similarity_score"`
}
