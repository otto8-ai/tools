package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/printers"
)

func GetMessageDetails(ctx context.Context, folderID, messageID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := graph.GetMessageDetails(ctx, c, folderID, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message details: %w", err)
	}

	if err := printers.PrintMessage(result, true); err != nil {
		return fmt.Errorf("failed to print message: %w", err)
	}
	return nil
}
