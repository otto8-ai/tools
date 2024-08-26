package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/printers"
)

func ListCalendars(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	calendars, err := graph.ListCalendars(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list calendars: %w", err)
	}

	printers.PrintCalendars(calendars)
	return nil
}