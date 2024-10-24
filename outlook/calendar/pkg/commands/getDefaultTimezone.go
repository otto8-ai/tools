package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
)

func GetDefaultTimezone(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	settings, err := c.Me().MailboxSettings().Get(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get mailbox settings: %w", err)
	}

	if tz := settings.GetTimeZone(); tz != nil && *tz != "" {
		fmt.Println("The user's default timezone is", *tz)
	} else {
		fmt.Println("The user's default timezone not defined")
	}

	return nil
}
