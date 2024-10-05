package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/pkg/util"
)

func CreateEvent(ctx context.Context, info graph.CreateEventInfo) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	event, err := graph.CreateEvent(ctx, c, info)
	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	fmt.Printf("Event created with ID: %s\n", util.Deref(event.GetId()))
	return nil
}
