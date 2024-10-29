package retrievers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	"github.com/gptscript-ai/knowledge/pkg/datastore/store"
	"github.com/gptscript-ai/knowledge/pkg/llm"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/philippgille/chromem-go"
)

const RoutingRetrieverName = "routing"

type RoutingRetriever struct {
	Model             llm.LLMConfig
	AvailableDatasets []string
	TopK              int
}

func (r *RoutingRetriever) Name() string {
	return RoutingRetrieverName
}

func (r *RoutingRetriever) NormalizedScores() bool {
	return true
}

func (r *RoutingRetriever) DecodeConfig(cfg map[string]any) error {
	return DefaultConfigDecoder(r, cfg)
}

var routingPromptTpl = `The following query will be used for a vector similarity search.
Please route it to the appropriate dataset. Choose the one that fits best to the query based on the metadata.
Query: "{{.query}}"
Available datasets in a JSON map, where the key is the dataset ID and the value is a map of metadata fields:
{{ .datasets }}
Reply only in the following JSON format, without any styling or markdown syntax:
{"result": "<dataset-id>"}`

type routingResp struct {
	Result string `json:"result"`
}

func (r *RoutingRetriever) Retrieve(ctx context.Context, store store.Store, query string, datasetIDs []string, where map[string]string, whereDocument []chromem.WhereDocument) ([]vs.Document, error) {
	log := slog.With("component", "RoutingRetriever")

	// TODO: properly handle the datasetIDs input
	log.Debug("Ignoring input datasetIDs in routing retriever, as it chooses one by itself", "query", query, "inputDataset", datasetIDs)

	if r.TopK <= 0 {
		log.Debug("TopK not set, using default", "default", defaults.TopK)
		r.TopK = defaults.TopK
	}

	if len(r.AvailableDatasets) == 0 {
		allDatasets, err := store.ListDatasets(ctx)
		if err != nil {
			return nil, err
		}
		for _, ds := range allDatasets {
			r.AvailableDatasets = append(r.AvailableDatasets, ds.ID)
		}
	}
	slog.Debug("Available datasets", "datasets", r.AvailableDatasets)

	datasets := map[string]map[string]any{}
	for _, dsID := range r.AvailableDatasets {
		dataset, err := store.GetDataset(ctx, dsID)
		if err != nil {
			return nil, err
		}
		if dataset == nil {
			return nil, fmt.Errorf("dataset not found: %q", dsID)
		}
		datasets[dataset.ID] = dataset.Metadata
	}

	datasetsJSON, err := json.Marshal(datasets)
	if err != nil {
		return nil, err
	}

	m, err := llm.NewFromConfig(r.Model)
	if err != nil {
		return nil, err
	}

	result, err := m.Prompt(context.Background(), routingPromptTpl, map[string]interface{}{"query": query, "datasets": string(datasetsJSON)})
	if err != nil {
		return nil, err
	}
	slog.Debug("Routing result", "result", result)
	var resp routingResp
	err = json.Unmarshal([]byte(result), &resp)
	if err != nil {
		return nil, err
	}

	slog.Debug("Routing query to dataset", "query", query, "dataset", resp.Result)

	return store.SimilaritySearch(ctx, query, r.TopK, resp.Result, where, whereDocument)
}
