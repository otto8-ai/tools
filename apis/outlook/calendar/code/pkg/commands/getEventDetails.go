package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/printers"
)

func GetEventDetails(ctx context.Context, eventID, calendarID string, owner graph.OwnerType) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	event, err := graph.GetEvent(ctx, c, eventID, calendarID, owner)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	printers.PrintEvent(event, true)
	return nil
}
