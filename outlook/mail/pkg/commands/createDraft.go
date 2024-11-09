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

func CreateDraft(ctx context.Context, info graph.DraftInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	draft, err := graph.CreateDraft(ctx, c, info)
	if err != nil {
		return fmt.Errorf("failed to create draft: %w", err)
	}

	// Get numerical ID for the draft
	draftID, err := id.SetOutlookID(ctx, util.Deref(draft.GetId()))
	if err != nil {
		return fmt.Errorf("failed to set draft ID: %w", err)
	}

	fmt.Printf("Draft created successfully. Draft ID: %s\n", draftID)
	return nil
}
