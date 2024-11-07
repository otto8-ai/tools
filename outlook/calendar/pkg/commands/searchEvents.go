package commands

import (
	"context"
	"fmt"
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

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
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

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name:        "event_search " + query,
		Description: "Search results for Outlook Calendar events",
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset with elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d events\n", datasetID, len(elements))
	return nil
}
