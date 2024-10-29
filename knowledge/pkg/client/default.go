package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/acorn-io/z"
	dstypes "github.com/gptscript-ai/knowledge/pkg/datastore/types"

	"github.com/gptscript-ai/knowledge/pkg/datastore"
	"github.com/gptscript-ai/knowledge/pkg/index"
	"github.com/gptscript-ai/knowledge/pkg/server/types"
)

type DefaultClient struct {
	ServerURL string
}

func NewDefaultClient(serverURL string) *DefaultClient {
	return &DefaultClient{
		ServerURL: strings.TrimSuffix(serverURL, "/"),
	}
}

func (c *DefaultClient) FindFile(_ context.Context, searchFile index.File) (*index.File, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}

func (c *DefaultClient) DeleteFile(_ context.Context, datasetID, fileID string) error {
	// TODO: implement
	return fmt.Errorf("not implemented")
}

func (c *DefaultClient) CreateDataset(_ context.Context, datasetID string) (*index.Dataset, error) {
	ds := types.Dataset{
		ID: datasetID,
	}

	payload, err := json.Marshal(ds)
	if err != nil {
		return nil, err
	}

	resp, err := c.request(http.MethodPost, "/datasets/create", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	var dataset *index.Dataset
	err = json.Unmarshal(resp, dataset)
	if err != nil {
		return nil, err
	}

	return dataset, nil
}

func (c *DefaultClient) DeleteDataset(_ context.Context, datasetID string) error {
	_, err := c.request(http.MethodDelete, fmt.Sprintf("/datasets/%s", datasetID), nil)
	return err
}

func (c *DefaultClient) GetDataset(_ context.Context, datasetID string) (*index.Dataset, error) {
	resp, err := c.request(http.MethodGet, fmt.Sprintf("/datasets/%s", datasetID), nil)
	if err != nil {
		return nil, err
	}

	var dataset *index.Dataset
	err = json.Unmarshal(resp, dataset)
	if err != nil {
		return nil, err
	}

	return dataset, nil
}

func (c *DefaultClient) ListDatasets(_ context.Context) ([]types.Dataset, error) {
	resp, err := c.request(http.MethodGet, "/datasets", nil)
	if err != nil {
		return []types.Dataset{}, err
	}

	var datasets []types.Dataset
	err = json.Unmarshal(resp, &datasets)
	if err != nil {
		return []types.Dataset{}, err
	}

	return datasets, nil
}

func (c *DefaultClient) Ingest(_ context.Context, datasetID string, name string, data []byte, opts datastore.IngestOpts) ([]string, error) {
	payload := types.Ingest{
		Filename: z.Pointer(name),
		Content:  base64.StdEncoding.EncodeToString(data),
	}
	if opts.FileMetadata != nil {
		payload.FileMetadata = opts.FileMetadata
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	resp, err := c.request(http.MethodPost, fmt.Sprintf("/datasets/%s/ingest", datasetID), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var res types.IngestResponse
	err = json.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return res.Documents, nil
}

func (c *DefaultClient) IngestPaths(ctx context.Context, datasetID string, opts *IngestPathsOpts, paths ...string) (int, int, error) {
	_, err := getOrCreateDataset(ctx, c, datasetID, !opts.NoCreateDataset)
	if err != nil {
		return 0, 0, err
	}

	ingestFile := func(path string, extraMetadata map[string]any) error {
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Gather metadata
		finfo, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("failed to stat file %s: %w", path, err)
		}

		abspath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %s: %w", path, err)
		}

		filename := filepath.Base(path)
		payload := datastore.IngestOpts{
			FileMetadata: &index.FileMetadata{
				Name:         filepath.Base(path),
				AbsolutePath: abspath,
				Size:         finfo.Size(),
				ModifiedAt:   finfo.ModTime(),
			},
			IsDuplicateFuncName: "file_metadata",
			ExtraMetadata:       extraMetadata,
		}
		if opts != nil {
			payload.TextSplitterOpts = opts.TextSplitterOpts
		}
		_, err = c.Ingest(ctx, datasetID, filename, content, payload)
		return err
	}

	return ingestPaths(ctx, c, opts, datasetID, ingestFile, paths...)
}

func (c *DefaultClient) PrunePath(ctx context.Context, datasetID string, path string, keep []string) ([]index.File, error) {
	// TODO: implement
	return nil, fmt.Errorf("not implemented")
}

func (c *DefaultClient) DeleteDocuments(_ context.Context, datasetID string, documentIDs ...string) error {
	for _, documentID := range documentIDs {
		_, err := c.request(http.MethodDelete, fmt.Sprintf("/datasets/%s/documents/%s", datasetID, documentID), nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *DefaultClient) Retrieve(_ context.Context, datasetIDs []string, query string, opts datastore.RetrieveOpts) (*dstypes.RetrievalResponse, error) {
	q := types.Query{Prompt: query}

	if opts.TopK != 0 {
		q.TopK = &opts.TopK
	}

	data, err := json.Marshal(q)
	if err != nil {
		return nil, err
	}

	// TODO: change to allow for multiple datasets
	resp, err := c.request(http.MethodPost, fmt.Sprintf("/datasets/%s/retrieve", datasetIDs), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	var res dstypes.RetrievalResponse
	err = json.Unmarshal(resp, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (c *DefaultClient) AskDirectory(ctx context.Context, path string, query string, opts *IngestPathsOpts, ropts *datastore.RetrieveOpts) (*dstypes.RetrievalResponse, error) {
	return AskDir(ctx, c, path, query, opts, ropts)
}

func (c *DefaultClient) request(method, path string, body io.Reader) ([]byte, error) {
	url := c.ServerURL + path

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("API request failed: %s", res.Status)
	}

	if res.Body != nil {
		defer res.Body.Close()
		return io.ReadAll(res.Body)
	}

	return nil, nil
}

func (c *DefaultClient) ExportDatasets(ctx context.Context, path string, datasets ...string) error {
	// TODO: implement
	panic("not implemented")
}

func (c *DefaultClient) ImportDatasets(ctx context.Context, path string, datasets ...string) error {
	// TODO: implement
	panic("not implemented")
}

func (c *DefaultClient) UpdateDataset(ctx context.Context, dataset index.Dataset, opts *datastore.UpdateDatasetOpts) (*index.Dataset, error) {
	// TODO: implement
	panic("not implemented")
}
