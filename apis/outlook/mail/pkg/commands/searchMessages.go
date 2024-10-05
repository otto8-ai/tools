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

func SearchMessages(ctx context.Context, subject, fromAddress, fromName, folderID string) error {
	trueFolderID, err := id.GetOutlookID(folderID)
	if err != nil {
		return fmt.Errorf("failed to get folder ID: %w", err)
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	messages, err := graph.SearchMessages(ctx, c, subject, fromAddress, fromName, trueFolderID)
	if err != nil {
		return fmt.Errorf("failed to search messages: %w", err)
	}

	if len(messages) == 0 {
		fmt.Println("no messages found")
		return nil
	}

	return printers.PrintMessages(messages, false)
}
