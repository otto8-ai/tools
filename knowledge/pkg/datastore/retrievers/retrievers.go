package retrievers

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore/lib/scores"
	"github.com/gptscript-ai/knowledge/pkg/datastore/store"
	"github.com/gptscript-ai/knowledge/pkg/output"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/mitchellh/mapstructure"
	"github.com/philippgille/chromem-go"

	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
)

type Retriever interface {
	Retrieve(ctx context.Context, store store.Store, query string, datasetIDs []string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error)
	Name() string
	DecodeConfig(cfg map[string]any) error
	NormalizedScores() bool // whether the retriever returns normalized scores
}

func GetRetriever(name string) (Retriever, error) {
	switch name {
	case BasicRetrieverName, "default":
		return &BasicRetriever{TopK: defaults.TopK}, nil
	case SubqueryRetrieverName:
		return &SubqueryRetriever{Limit: 3, TopK: 3}, nil
	case RoutingRetrieverName:
		return &RoutingRetriever{TopK: defaults.TopK}, nil
	case MergingRetrieverName:
		return &MergingRetriever{TopK: defaults.TopK}, nil
	case BM25RetrieverName:
		return &BM25Retriever{TopN: defaults.TopK, K1: 1.2, B: 0.75}, nil
	default:
		return nil, fmt.Errorf("unknown retriever %q", name)
	}
}

func DefaultConfigDecoder(ret Retriever, cfg map[string]any) error {
	if ret == nil {
		return fmt.Errorf("retriever is nil")
	}
	if len(cfg) == 0 {
		return nil
	}
	if err := mapstructure.Decode(cfg, &ret); err != nil {
		return fmt.Errorf("failed to decode retriever configuration: %w", err)
	}
	slog.Debug("Retriever custom configuration", "name", ret.Name(), "config", output.RedactSensitive(ret))
	return nil
}

func GetDefaultRetriever() Retriever {
	return &BasicRetriever{TopK: defaults.TopK}
}

const BasicRetrieverName = "basic"

type BasicRetriever struct {
	TopK int
}

func (r *BasicRetriever) Name() string {
	return BasicRetrieverName
}

func (r *BasicRetriever) NormalizedScores() bool {
	return true
}

func (r *BasicRetriever) DecodeConfig(cfg map[string]any) error {
	return DefaultConfigDecoder(r, cfg)
}

func (r *BasicRetriever) Retrieve(ctx context.Context, store store.Store, query string, datasetIDs []string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error) {
	if len(datasetIDs) == 0 {
		datasetIDs = []string{"default"}
	}

	var results []vs.Document
	for _, dataset := range datasetIDs {
		// TODO: make configurable via RetrieveOpts
		// silently ignore non-existent datasets
		ds, err := store.GetDataset(ctx, dataset)
		if err != nil {
			if strings.HasPrefix(err.Error(), "dataset not found") {
				continue
			}
			return nil, err
		}
		if ds == nil {
			continue
		}

		docs, err := store.SimilaritySearch(ctx, query, r.TopK, dataset, where, whereDocument)
		if err != nil {
			return nil, err
		}

		results = append(results, docs...)
	}

	slices.SortFunc(results, scores.SortBySimilarityScore)

	log := slog.With("retriever", r.Name())
	if r.TopK <= 0 {
		log.Debug("[BasicRetriever] TopK not set, using default", "default", defaults.TopK)
		r.TopK = defaults.TopK
	}

	topK := r.TopK
	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}
