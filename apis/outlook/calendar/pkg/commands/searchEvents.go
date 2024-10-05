package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/printers"
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

	for cal, events := range result {
		if len(events) == 0 {
			continue
		}

		if err := printers.PrintEventsForCalendar(ctx, c, cal, events, false); err != nil {
			return fmt.Errorf("failed to print events: %w", err)
		}
	}

	return nil
}
