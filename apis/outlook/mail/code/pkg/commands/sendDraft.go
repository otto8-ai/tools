package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/graph"
)

func SendDraft(ctx context.Context, draftID string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.SendDraft(ctx, c, draftID); err != nil {
		return fmt.Errorf("failed to send draft: %w", err)
	}

	fmt.Println("Draft sent successfully")
	return nil
}
