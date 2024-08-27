package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/printers"
)

func ListMessages(ctx context.Context, folderID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	messages, err := graph.ListMessages(ctx, c, folderID)
	if err != nil {
		return fmt.Errorf("failed to list mail: %w", err)
	}

	printers.PrintMessages(messages, false)

	return nil
}
