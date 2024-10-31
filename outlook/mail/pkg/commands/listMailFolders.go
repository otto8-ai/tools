package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
)

func ListMailFolders(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := graph.ListMailFolders(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list mail folders: %w", err)
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	workspaceID := os.Getenv("GPTSCRIPT_WORKSPACE_ID")

	dataset, err := gptscriptClient.CreateDataset(ctx, workspaceID, "outlook_mail_folders", "Outlook mail folders")
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	var elements []gptscript.DatasetElement
	for _, folder := range result {
		folderStr, err := printers.MailFolderToString(folder)
		if err != nil {
			return fmt.Errorf("failed to convert mail folder to string: %w", err)
		}

		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        util.Deref(folder.GetId()),
				Description: util.Deref(folder.GetDisplayName()),
			},
			Contents: folderStr,
		})
	}

	if err := gptscriptClient.AddDatasetElements(ctx, workspaceID, dataset.ID, elements); err != nil {
		return fmt.Errorf("failed to add dataset elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d folders\n", dataset.ID, len(result))
	return nil
}
