package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/printers"
)

func SearchMessages(ctx context.Context, query, fromAddress, fromName, folderID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	messages, err := graph.SearchMessages(ctx, c, query, fromAddress, fromName, folderID)
	if err != nil {
		return fmt.Errorf("failed to search messages: %w", err)
	}

	printers.PrintMessages(messages, false)
	return nil
}
