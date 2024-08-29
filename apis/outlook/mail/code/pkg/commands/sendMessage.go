package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/util"
)

func SendMessage(ctx context.Context, info graph.DraftInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	message, err := graph.CreateDraft(ctx, c, info)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	if err := graph.SendDraft(ctx, c, util.Deref(message.GetId())); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	fmt.Println("Message sent successfully")
	return nil
}
