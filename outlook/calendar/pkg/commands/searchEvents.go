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

	result, err := graph.SearchEvents(ctx, c, query, start, end)
	if err != nil {
		return fmt.Errorf("failed to search events: %w", err)
	}

	calendarEvents := map[graph.CalendarInfo][]models.Eventable{}
	for cal, events := range result {
		if len(events) > 0 {
			calendarEvents[cal] = events
		}
	}

	if len(util.Flatten(util.MapValues(calendarEvents))) > 10 {
		workspace := os.Getenv("GPTSCRIPT_WORKSPACE_DIR")
		gptscriptClient, err := gptscript.NewGPTScript()
		if err != nil {
			return fmt.Errorf("failed to create GPTScript client: %w", err)
		}

		dataset, err := gptscriptClient.CreateDataset(ctx, workspace, "event search "+query, "Search results for Outlook Calendar events")
		if err != nil {
			return fmt.Errorf("failed to create dataset: %w", err)
		}

		for cal, events := range calendarEvents {
			for _, event := range events {
				eventString := printers.EventToString(ctx, c, cal, event)
				if _, err := gptscriptClient.AddDatasetElement(ctx, workspace, dataset.ID,
					util.Deref(event.GetSubject())+"_"+util.Deref(cal.Calendar.GetName())+"_"+util.Deref(event.GetStart().GetDateTime()),
					util.Deref(event.GetBodyPreview()), eventString); err != nil {
					if strings.Contains(err.Error(), "already exists") {
						continue
					}

					return fmt.Errorf("failed to add dataset element: %w", err)
				}
			}
		}

		fmt.Printf("Created dataset with ID %s with %d events\n", dataset.ID, len(util.Flatten(util.MapValues(calendarEvents))))
		return nil
	}

	for cal, events := range calendarEvents {
		if err := printers.PrintEventsForCalendar(ctx, c, cal, events, false); err != nil {
			return fmt.Errorf("failed to print events for calendar %w", err)
		}
	}
	return nil
}
