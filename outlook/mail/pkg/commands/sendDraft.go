package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
)

func SendDraft(ctx context.Context, draftID string) error {
	trueDraftID, err := id.GetOutlookID(ctx, draftID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.SendDraft(ctx, c, trueDraftID); err != nil {
		return fmt.Errorf("failed to send draft: %w", err)
	}

	fmt.Println("Draft sent successfully")
	return nil
}
