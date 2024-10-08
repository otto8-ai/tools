package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
)

func DeleteEvent(ctx context.Context, eventID, calendarID string, owner graph.OwnerType) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	if err := graph.DeleteEvent(ctx, c, eventID, calendarID, owner); err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}
	fmt.Println("Event deleted successfully")
	return nil
}
