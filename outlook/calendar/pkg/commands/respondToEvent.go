package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/common/id"
)

func RespondToEvent(ctx context.Context, eventID, calendarID string, owner graph.OwnerType, response string) error {
	trueEventID, err := id.GetOutlookID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get Outlook ID: %w", err)
	}

	var trueCalendarID string
	if calendarID != "" {
		trueCalendarID, err = id.GetOutlookID(ctx, calendarID)
		if err != nil {
			return fmt.Errorf("failed to get Outlook Calendar ID: %w", err)
		}
	}

	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	switch response {
	case "accept":
		if err := graph.AcceptEvent(ctx, c, trueEventID, trueCalendarID, owner); err != nil {
			return fmt.Errorf("failed to accept event: %w", err)
		}
		fmt.Println("Event accepted successfully")
	case "tentative":
		if err := graph.TentativelyAcceptEvent(ctx, c, trueEventID, trueCalendarID, owner); err != nil {
			return fmt.Errorf("failed to tentatively accept event: %w", err)
		}
		fmt.Println("Event tentatively accepted successfully")
	case "decline":
		if err := graph.DeclineEvent(ctx, c, trueEventID, trueCalendarID, owner); err != nil {
			return fmt.Errorf("failed to decline event: %w", err)
		}
		fmt.Println("Event declined successfully")
	default:
		return fmt.Errorf("invalid response: %s", response)
	}

	return nil
}
