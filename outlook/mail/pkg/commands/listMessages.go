package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
)

func ListMessages(ctx context.Context, folderID string) error {
	var (
		trueFolderID string
		err          error
	)

	if folderID != "" {
		trueFolderID, err = id.GetOutlookID(folderID)
		if err != nil {
			return fmt.Errorf("failed to get folder ID: %w", err)
		}
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	messages, err := graph.ListMessages(ctx, c, trueFolderID)
	if err != nil {
		return fmt.Errorf("failed to list mail: %w", err)
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	workspaceID := os.Getenv("GPTSCRIPT_WORKSPACE_ID")

	dataset, err := gptscriptClient.CreateDataset(ctx, workspaceID, fmt.Sprintf("%s_outlook_mail", folderID), "Outlook mail messages in folder "+folderID)
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}
	var elements []gptscript.DatasetElement
	for _, message := range messages {
		messageStr, err := printers.MessageToString(message, false)
		if err != nil {
			return fmt.Errorf("failed to convert message to string: %w", err)
		}
		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        util.Deref(message.GetId()),
				Description: util.Deref(message.GetSubject()),
			},
			Contents: []byte(messageStr),
		})
	}
	if err := gptscriptClient.AddDatasetElements(ctx, workspaceID, dataset.ID, elements); err != nil {
		return fmt.Errorf("failed to add dataset elements: %w", err)
	}
	fmt.Printf("Created dataset with ID %s with %d messages\n", dataset.ID, len(messages))
	return nil
}
