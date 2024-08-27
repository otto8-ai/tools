package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/util"
)

func MoveMessage(ctx context.Context, messageID, destinationFolderID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	message, err := graph.MoveMessage(ctx, c, messageID, destinationFolderID)
	if err != nil {
		return fmt.Errorf("failed to move message: %w", err)
	}

	fmt.Printf("Message moved successfully. New message ID: %s\n", util.Deref(message.GetId()))
	return nil
}
