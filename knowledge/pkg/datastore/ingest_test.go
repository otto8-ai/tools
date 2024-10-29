package datastore

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/gptscript-ai/knowledge/pkg/datastore/textsplitter"
	"github.com/gptscript-ai/knowledge/pkg/datastore/transformers"
	"github.com/gptscript-ai/knowledge/pkg/flows"
	"github.com/stretchr/testify/require"
)

func TestExtractPDF(t *testing.T) {
	ctx := context.Background()
	textSplitterOpts := textsplitter.NewTextSplitterOpts()
	err := filepath.WalkDir("testdata/pdf", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatalf("filepath.WalkDir() error = %v", err)
		}
		if d.IsDir() {
			return nil
		}
		t.Logf("Processing %s", path)
		f, err := os.Open(path)
		require.NoError(t, err, "os.Open() error = %v", err)

		filetype := ".pdf"

		ingestionFlow, err := flows.NewDefaultIngestionFlow(filetype, &textSplitterOpts)
		require.NoError(t, err, "NewDefaultIngestionFlow() error = %v", err)
		require.NotNil(t, ingestionFlow.Load, "ingestionFlow.Load is nil")

		// Mandatory Transformation: Add filename to metadata
		em := &transformers.ExtraMetadata{Metadata: map[string]any{"filename": d.Name()}}
		ingestionFlow.Transformations = append(ingestionFlow.Transformations, em)

		docs, err := ingestionFlow.Run(ctx, f)
		require.NoError(t, err, "GetDocuments() error = %v", err)
		require.NotEmpty(t, docs, "GetDocuments() returned no documents")
		return nil
	})
	require.NoError(t, err, "filepath.WalkDir() error = %v", err)
}
