package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
)

func ListEvents(ctx context.Context, start, end time.Time) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	calendars, err := graph.ListCalendars(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}

	for _, cal := range calendars {
		if cal.ID == "" {
			continue
		}

		events, err := graph.ListCalendarView(ctx, c, cal.ID, cal.Owner, &start, &end)
		if err != nil {
			return fmt.Errorf("failed to list events for calendar %s: %w", util.Deref(cal.Calendar.GetName()), err)
		}

		if len(events) == 0 {
			continue
		}

		if err := printers.PrintEventsForCalendar(ctx, c, cal, events, false); err != nil {
			return fmt.Errorf("failed to print events: %w", err)
		}
	}

	return nil
}
