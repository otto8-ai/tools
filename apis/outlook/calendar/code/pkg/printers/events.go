package printers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/outlook/calendar/code/pkg/util"
	"github.com/jaytaylor/html2text"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func PrintEventsForCalendar(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, calendar graph.CalendarInfo, events []models.Eventable, detailed bool) error {
	if calendar.Owner == graph.OwnerTypeUser {
		fmt.Printf("Found events for calendar %s (ID %s):\n", util.Deref(calendar.Calendar.GetName()), calendar.ID)
	} else {
		groupName, err := graph.GetGroupNameFromID(ctx, client, calendar.ID)
		if err != nil {
			return fmt.Errorf("failed to get group name: %w", err)
		}
		fmt.Printf("Found events for calendar %s (ID %s):\n", groupName, calendar.ID)
	}

	PrintEvents(events, detailed)
	fmt.Println()
	return nil
}

func PrintEvents(events []models.Eventable, detailed bool) {
	for _, event := range events {
		PrintEvent(event, detailed)
	}
}

func PrintEvent(event models.Eventable, detailed bool) {
	fmt.Printf("Subject: %s\n", util.Deref(event.GetSubject()))
	fmt.Printf("  ID: %s\n", util.Deref(event.GetId()))
	fmt.Printf("  Start: %s\n", util.Deref(event.GetStart().GetDateTime()))
	fmt.Printf("  End: %s\n", util.Deref(event.GetEnd().GetDateTime()))

	if detailed {
		fmt.Printf("  Location: %s\n", util.Deref(event.GetLocation().GetDisplayName()))
		fmt.Printf("  Is All Day: %t\n", util.Deref(event.GetIsAllDay()))
		fmt.Printf("  Is Cancelled: %t\n", util.Deref(event.GetIsCancelled()))
		fmt.Printf("  Is Online Meeting: %t\n", util.Deref(event.GetIsOnlineMeeting()))
		fmt.Printf("  Response Status: %s\n", event.GetResponseStatus().GetResponse().String())
		fmt.Printf("  Attendees: %s\n", strings.Join(util.Map(event.GetAttendees(), func(a models.Attendeeable) string {
			return fmt.Sprintf("%s (%s)", util.Deref(a.GetEmailAddress().GetName()), util.Deref(a.GetEmailAddress().GetAddress()))
		}), ", "))
		body, err := html2text.FromString(util.Deref(event.GetBody().GetContent()), html2text.Options{
			PrettyTables: true,
		})
		if err == nil {
			fmt.Printf("  Body: %s\n", strings.ReplaceAll(body, "\n", "\n  "))
			fmt.Printf("  (End Body)\n")
		}
	}
	fmt.Println()
}
