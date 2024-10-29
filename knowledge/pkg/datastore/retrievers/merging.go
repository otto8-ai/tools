package retrievers

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"

	"github.com/acorn-io/z"
	"github.com/gptscript-ai/knowledge/pkg/datastore/lib/scores"
	"github.com/gptscript-ai/knowledge/pkg/datastore/store"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/mitchellh/mapstructure"
	"github.com/philippgille/chromem-go"
	"golang.org/x/sync/errgroup"
)

const MergingRetrieverName = "merge"

type MergingRetriever struct {
	TopK       int
	Retrievers []RetrieverToMerge `json:"retrievers" mapstructure:"retrievers" yaml:"retrievers"`
	retrievers []Retriever
}

type RetrieverToMerge struct {
	Name    string         `json:"name,omitempty" mapstructure:"name" yaml:"name"`
	Weight  *float32       `json:"weight,omitempty" mapstructure:"weight" yaml:"weight"`
	Options map[string]any `json:"options,omitempty" mapstructure:"options" yaml:"options"`
}

func (r *MergingRetriever) Name() string {
	return MergingRetrieverName
}

func (r *MergingRetriever) NormalizedScores() bool {
	return true
}

func (r *MergingRetriever) DecodeConfig(cfg map[string]any) error {
	if err := mapstructure.Decode(cfg, &r); err != nil {
		return fmt.Errorf("failed to decode merging retriever configuration: %w", err)
	}

	r.retrievers = make([]Retriever, len(r.Retrievers))
	for i, retrieverConfig := range r.Retrievers {
		retriever, err := GetRetriever(retrieverConfig.Name)
		if err != nil {
			return err
		}
		if err := retriever.DecodeConfig(retrieverConfig.Options); err != nil {
			return err
		}
		r.retrievers[i] = retriever
	}

	return nil
}

func (r *MergingRetriever) Retrieve(ctx context.Context, store store.Store, query string, datasetIDs []string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error) {
	log := slog.With("component", "MergingRetriever")

	// Set default weight to 1.0 if not provided
	for i, retriever := range r.Retrievers {
		if retriever.Weight == nil {
			r.Retrievers[i].Weight = z.Pointer(float32(1.0))
		}
	}

	slog.Debug("Retrieving documents from merging retriever", "query", query, "datasetIDs", datasetIDs, "where", where, "whereDocument", whereDocument)

	g, ctx := errgroup.WithContext(ctx)

	var mu sync.Mutex
	documentsMap := make(map[string]vs.Document)

	for ri, retriever := range r.Retrievers {
		ri := ri
		retriever := retriever
		g.Go(func() error {
			log.Debug("Retrieving documents from retriever", "retriever", retriever.Name)
			retrievedDocs, err := r.retrievers[ri].Retrieve(ctx, store, query, datasetIDs, where, whereDocument)
			if err != nil {
				log.Error("Failed to retrieve documents from retriever", "retriever", retriever.Name, "error", err)
				return err
			}

			slog.Debug("Retrieved documents from retriever", "retriever", retriever.Name, "numDocs", len(retrievedDocs))

			normalized := r.retrievers[ri].NormalizedScores()
			var minScore, maxScore float32
			if !normalized {
				minScore, maxScore = scores.FindMinMaxScores(retrievedDocs)
			}

			mu.Lock()
			defer mu.Unlock()

			for _, retrievedDoc := range retrievedDocs {
				if existingDoc, found := documentsMap[retrievedDoc.ID]; found {
					// Document already exists, update its similarity score and metadata
					existingDoc.Metadata["retriever"] = fmt.Sprintf("%s,%s", existingDoc.Metadata["retriever"], retriever.Name)
					existingDoc.Metadata["retrieverScore::"+retriever.Name] = retrievedDoc.SimilarityScore

					normalizedScore := retrievedDoc.SimilarityScore
					if !normalized {
						normalizedScore = scores.NormalizeScore(retrievedDoc.SimilarityScore, minScore, maxScore)
						slog.Debug("Normalized score", "retriever", retriever.Name, "score", retrievedDoc.SimilarityScore, "minScore", minScore, "maxScore", maxScore, "normalizedScore", normalizedScore)
					}
					existingDoc.Metadata["retrieverScoreNormalized::"+retriever.Name] = normalizedScore
					existingDoc.SimilarityScore += normalizedScore * z.Dereference(retriever.Weight)

					documentsMap[retrievedDoc.ID] = existingDoc
				} else {
					// New document, add it to the map
					normalizedScore := scores.NormalizeScore(retrievedDoc.SimilarityScore, minScore, maxScore)
					retrievedDoc.Metadata["retriever"] = retriever.Name
					retrievedDoc.Metadata["retrieverScore::"+retriever.Name] = retrievedDoc.SimilarityScore
					retrievedDoc.Metadata["retrieverScoreNormalized::"+retriever.Name] = normalizedScore
					retrievedDoc.SimilarityScore = normalizedScore * z.Dereference(retriever.Weight)

					documentsMap[retrievedDoc.ID] = retrievedDoc
				}
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Convert map to slice
	var resultDocs []vs.Document
	for _, doc := range documentsMap {
		resultDocs = append(resultDocs, doc)
	}

	// Sort the resultDocs by similarity score
	slices.SortFunc(resultDocs, scores.SortBySimilarityScore)

	topK := r.TopK
	if len(resultDocs) < topK {
		topK = len(resultDocs)
	}

	slog.Debug("MergingRetriever", "topK", topK, "numDocs", len(resultDocs))

	if len(resultDocs) == 0 {
		return nil, nil
	}

	return resultDocs[:topK], nil
}
