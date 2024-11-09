package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
)

func MoveMessage(ctx context.Context, messageID, destinationFolderID string) error {
	trueMessageID, err := id.GetOutlookID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get message ID: %w", err)
	}

	trueDestinationFolderID, err := id.GetOutlookID(ctx, destinationFolderID)
	if err != nil {
		return fmt.Errorf("failed to get destination folder ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	message, err := graph.MoveMessage(ctx, c, trueMessageID, trueDestinationFolderID)
	if err != nil {
		return fmt.Errorf("failed to move message: %w", err)
	}

	// Save the new message ID
	newMessageID, err := id.SetOutlookID(ctx, util.Deref(message.GetId()))
	if err != nil {
		return fmt.Errorf("failed to save new message ID: %w", err)
	}

	fmt.Printf("Message moved successfully. New message ID: %s\n", newMessageID)
	return nil
}
