package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/gptscript-ai/tools/outlook/common/id"
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

	calendarIDs := util.Map(calendars, func(cal graph.CalendarInfo) string {
		return cal.ID
	})
	translatedCalendarIDs, err := id.SetOutlookIDs(ctx, calendarIDs)
	if err != nil {
		return fmt.Errorf("failed to set calendar IDs: %w", err)
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

		if len(events) == 0 {
			continue
		}

		// Update the ID to the translated ID
		cal.ID = translatedCalendarIDs[cal.ID]

		eventIDs := util.Map(events, func(event models.Eventable) string {
			return util.Deref(event.GetId())
		})
		translatedEventIDs, err := id.SetOutlookIDs(ctx, eventIDs)
		if err != nil {
			return fmt.Errorf("failed to set event IDs: %w", err)
		}

		for i := range events {
			events[i].SetId(util.Ptr(translatedEventIDs[util.Deref(events[i].GetId())]))
		}

		if len(events) > 0 {
			calendarEvents[cal] = events
		}
	}

	if len(calendarEvents) == 0 {
		fmt.Println("No events found")
		return nil
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	var elements []gptscript.DatasetElement
	for cal, events := range calendarEvents {
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
		Name:        "event_list",
		Description: "List of Outlook Calendar events",
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset with elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d events\n", datasetID, len(elements))
	return nil
}
