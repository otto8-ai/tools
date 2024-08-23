package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/graph"
)

func InviteUserToEvent(ctx context.Context, eventID, calendarID string, owner graph.OwnerType, userEmail, message string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.InviteUserToEvent(ctx, c, eventID, calendarID, owner, userEmail, message); err != nil {
		return fmt.Errorf("failed to invite user to event: %w", err)
	}
	return nil
}
