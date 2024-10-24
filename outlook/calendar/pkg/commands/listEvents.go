package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
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

	calendarEvents := map[graph.CalendarInfo][]models.Eventable{}
	for _, cal := range calendars {
		if cal.ID == "" {
			continue
		}

		events, err := graph.ListCalendarView(ctx, c, cal.ID, cal.Owner, &start, &end)
		if err != nil {
			return fmt.Errorf("failed to list events for calendar %s: %w", util.Deref(cal.Calendar.GetName()), err)
		}

		if len(events) > 0 {
			calendarEvents[cal] = events
		}
	}

	if len(util.Flatten(util.MapValues(calendarEvents))) > 10 {
		workspaceID := os.Getenv("GPTSCRIPT_WORKSPACE_ID")
		gptscriptClient, err := gptscript.NewGPTScript()
		if err != nil {
			return fmt.Errorf("failed to create GPTScript client: %w", err)
		}

		dataset, err := gptscriptClient.CreateDataset(ctx, workspaceID, "event list", "List of Outlook Calendar events")
		if err != nil {
			return fmt.Errorf("failed to create dataset: %w", err)
		}

		names := map[string]any{} // Store names to avoid duplicates
		var elements []gptscript.DatasetElement
		for cal, events := range calendarEvents {
			for _, event := range events {
				name := util.Deref(event.GetSubject()) + "_" + util.Deref(cal.Calendar.GetName()) + "_" + util.Deref(event.GetStart().GetDateTime())
				if _, exists := names[name]; !exists {
					elements = append(elements, gptscript.DatasetElement{
						DatasetElementMeta: gptscript.DatasetElementMeta{
							Name:        name,
							Description: util.Deref(event.GetBodyPreview()),
						},
						Contents: printers.EventToString(ctx, c, cal, event),
					})
					names[name] = struct{}{}
				}

			}
		}

		if err := gptscriptClient.AddDatasetElements(ctx, workspaceID, dataset.ID, elements); err != nil {
			return fmt.Errorf("failed to add dataset elements: %w", err)
		}

		fmt.Printf("Created dataset with ID %s with %d events\n", dataset.ID, len(util.Flatten(util.MapValues(calendarEvents))))
		return nil
	}

	for cal, events := range calendarEvents {
		if err := printers.PrintEventsForCalendar(ctx, c, cal, events, false); err != nil {
			return fmt.Errorf("failed to print events: %w", err)
		}
	}
	return nil
}
