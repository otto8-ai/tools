package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/printers"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/util"
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

		switch cal.Owner {
		case graph.OwnerTypeUser:
			fmt.Printf("Events for calendar %s:\n\n", util.Deref(cal.Calendar.GetName()))
			printers.PrintEvents(events, false)
		case graph.OwnerTypeGroup:
			groupName, err := graph.GetGroupNameFromID(ctx, c, cal.ID)
			if err != nil {
				return fmt.Errorf("failed to get group name: %w", err)
			}
			fmt.Printf("Events for group calendar %s:\n\n", groupName)
			printers.PrintEvents(events, false)
		}

		fmt.Println()
	}

	return nil
}
