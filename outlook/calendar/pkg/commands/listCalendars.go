package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
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

	calendarIDs := util.Map(calendars, func(c graph.CalendarInfo) string {
		return c.ID
	})
	translatedCalendarIDs, err := id.SetOutlookIDs(ctx, calendarIDs)
	if err != nil {
		return fmt.Errorf("failed to set calendar IDs: %w", err)
	}

	for i := range calendars {
		calendars[i].ID = translatedCalendarIDs[calendars[i].ID]
	}

	printers.PrintCalendars(calendars)
	return nil
}
