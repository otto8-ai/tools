package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/gptscript-ai/tools/outlook/common/id"
)

func CreateEvent(ctx context.Context, info graph.CreateEventInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	// If there is a calendar ID set on the info, translate it to the true Outlook ID.
	if info.ID != "" {
		trueCalendarID, err := id.GetOutlookID(ctx, info.ID)
		if err != nil {
			return fmt.Errorf("failed to get outlook ID: %w", err)
		}
		info.ID = trueCalendarID
	}

	event, err := graph.CreateEvent(ctx, c, info)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	eventID, err := id.SetOutlookID(ctx, util.Deref(event.GetId()))
	if err != nil {
		return fmt.Errorf("failed to get event ID: %w", err)
	}

	fmt.Printf("Event created with ID: %s\n", eventID)
	return nil
}
