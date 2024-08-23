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

		if cal.Owner == graph.OwnerTypeUser {
			fmt.Printf("Found events for calendar %s (ID %s):\n", util.Deref(cal.Calendar.GetName()), cal.ID)
		} else {
			groupName, err := graph.GetGroupNameFromID(ctx, c, cal.ID)
			if err != nil {
				return fmt.Errorf("failed to get group name: %w", err)
			}
			fmt.Printf("Found events for calendar %s (ID %s):\n", groupName, cal.ID)
		}
		printers.PrintEvents(events, false)
		fmt.Println()
	}

	return nil
}
