package scores

import (
	"log/slog"
	"math"

	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

func FindMinMaxScores(docs []vs.Document) (float32, float32) {
	minScore, maxScore := float32(math.MaxFloat32), float32(-math.MaxFloat32)

	// Find min and max scores
	for _, doc := range docs {
		score := doc.SimilarityScore
		if score < minScore {
			minScore = score
		}
		if score > maxScore {
			maxScore = score
		}
	}

	return minScore, maxScore
}

// NormalizeDocScores normalizes scores to a 0-1 range
func NormalizeDocScores(docs []vs.Document) []vs.Document {
	minScore, maxScore := FindMinMaxScores(docs)

	// Normalize scores
	for i, doc := range docs {
		normalizedScore := NormalizeScore(doc.SimilarityScore, minScore, maxScore)
		slog.Debug("Normalized similarity score", "score", doc.SimilarityScore, "normalized_score", normalizedScore)
		docs[i].SimilarityScore = normalizedScore
	}

	return docs
}

// NormalizeScore normalizes a single score
func NormalizeScore(score float32, minScore float32, maxScore float32) float32 {
	if maxScore == 0 {
		return 0
	}
	if maxScore-minScore == 0 {
		return 1 // Avoid division by zero - also, this happens for a single document, so we want a score of 1 here
	}
	normalizedScore := (score - minScore) / (maxScore - minScore)
	return normalizedScore
}

func SortBySimilarityScore(i, j vs.Document) int {
	if i.SimilarityScore > j.SimilarityScore {
		return -1
	}
	if i.SimilarityScore < j.SimilarityScore {
		return 1
	}
	return 0
}
