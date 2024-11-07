package commands

import (
	"context"
	"fmt"

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

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name: "word_docs_list",
	})

	fmt.Printf("Created dataset with ID %s with %d docs\n", datasetID, len(elements))
	return nil
}
