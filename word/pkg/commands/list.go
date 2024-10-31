package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/word/pkg/client"
	"github.com/gptscript-ai/tools/word/pkg/global"
	"github.com/gptscript-ai/tools/word/pkg/graph"
)

func ListDocs(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	infos, err := graph.ListDocs(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list word docs: %w", err)
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	workspaceID := os.Getenv("GPTSCRIPT_WORKSPACE_ID")
	dataset, err := gptscriptClient.CreateDataset(ctx, workspaceID, "word_docs_list", "List of Word documents")
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	var elements []gptscript.DatasetElement
	for _, info := range infos {
		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        info.Name,
				Description: fmt.Sprintf("%s (ID: %s)", info.Name, info.ID),
			},
			Contents: info.String(),
		})
	}

	if err := gptscriptClient.AddDatasetElements(ctx, workspaceID, dataset.ID, elements); err != nil {
		return fmt.Errorf("failed to add elements to dataset: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d docs\n", dataset.ID, len(elements))
	return nil
}
