package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/common/id"
)

func ListCalendars(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	calendars, err := graph.ListCalendars(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}

	// Translate Outlook IDs to friendly numerical IDs.
	for i := range calendars {
		calendarID, err := id.SetOutlookID(calendars[i].ID)
		if err != nil {
			return fmt.Errorf("failed to set calendar ID: %w", err)
		}
		calendars[i].ID = calendarID
	}

	printers.PrintCalendars(calendars)
	return nil
}
