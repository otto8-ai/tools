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

	if len(messages) > 10 {
		gptscriptClient, err := gptscript.NewGPTScript()
		if err != nil {
			return fmt.Errorf("failed to create GPTScript client: %w", err)
		}

		dataset, err := gptscriptClient.CreateDataset(ctx, os.Getenv("GPTSCRIPT_WORKSPACE_DIR"), fmt.Sprintf("%s_outlook_mail", folderID), "Outlook mail messages in folder "+folderID)
		// If we got back an error, we just print the messages. Otherwise, write them to the dataset.
		if err == nil {
			for _, message := range messages {
				messageStr, err := printers.MessageToString(message, false)
				if err != nil {
					return fmt.Errorf("failed to convert message to string: %w", err)
				}

				if _, err = gptscriptClient.AddDatasetElement(ctx, os.Getenv("GPTSCRIPT_WORKSPACE_DIR"), dataset.ID, util.Deref(message.GetId()), util.Deref(message.GetSubject()), messageStr); err != nil {
					return fmt.Errorf("failed to add element: %w", err)
				}
			}

			fmt.Printf("Created dataset with ID %s with %d messages\n", dataset.ID, len(messages))
			return nil
		}
	}

	return printers.PrintMessages(messages, false)
}
