package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/common/id"
	"github.com/gptscript-ai/tools/apis/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/mail/pkg/printers"
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

	return printers.PrintMessages(messages, false)
}
