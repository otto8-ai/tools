package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func SearchEvents(ctx context.Context, query string, start, end time.Time) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	calendars, err := graph.ListCalendars(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}

	calendarEventsInSubject := make(map[graph.CalendarInfo][]models.Eventable, len(calendars))
	calendarEventsInPreview := make(map[graph.CalendarInfo][]models.Eventable, len(calendars))
	for _, cal := range calendars {
		result, err := graph.ListCalendarView(ctx, c, cal.ID, cal.Owner, &start, &end)
		if err != nil {
			return fmt.Errorf("failed to search events: %w", err)
		}

		for _, event := range result {
			if strings.Contains(strings.ToLower(util.Deref(event.GetSubject())), strings.ToLower(query)) {
				calendarEventsInSubject[cal] = append(calendarEventsInSubject[cal], event)
			} else if strings.Contains(strings.ToLower(util.Deref(event.GetBodyPreview())), strings.ToLower(query)) {
				calendarEventsInPreview[cal] = append(calendarEventsInPreview[cal], event)
			}
		}
	}

	allCalendarEvents := util.Merge(calendarEventsInSubject, calendarEventsInPreview)

	workspaceID := os.Getenv("GPTSCRIPT_WORKSPACE_ID")
	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	dataset, err := gptscriptClient.CreateDataset(ctx, workspaceID, "event_search "+query, "Search results for Outlook Calendar events")
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	var elements []gptscript.DatasetElement
	for cal, events := range allCalendarEvents {
		for _, event := range events {
			name := util.Deref(event.GetId()) + "_" + util.Deref(event.GetSubject())
			elements = append(elements, gptscript.DatasetElement{
				DatasetElementMeta: gptscript.DatasetElementMeta{
					Name:        name,
					Description: util.Deref(event.GetBodyPreview()),
				},
				Contents: printers.EventToString(ctx, c, cal, event),
			})
		}
	}

	if err := gptscriptClient.AddDatasetElements(ctx, workspaceID, dataset.ID, elements); err != nil {
		return fmt.Errorf("failed to add dataset elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d events\n", dataset.ID, len(elements))
	return nil
}
