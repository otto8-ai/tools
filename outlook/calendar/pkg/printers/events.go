package printers

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/jaytaylor/html2text"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func EventToString(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, calendar graph.CalendarInfo, event models.Eventable) string {
	var calendarName string
	if calendar.Owner == graph.OwnerTypeUser {
		calendarName = util.Deref(calendar.Calendar.GetName())
	} else {
		groupName, err := graph.GetGroupNameFromID(ctx, client, calendar.ID)
		if err != nil {
			calendarName = calendar.ID
		} else {
			calendarName = groupName
		}
	}

	var sb strings.Builder
	sb.WriteString("Subject: " + util.Deref(event.GetSubject()) + "\n")
	sb.WriteString("  ID: " + util.Deref(event.GetId()) + "\n")
	sb.WriteString("  Start: " + util.Deref(event.GetStart().GetDateTime()) + "\n")
	sb.WriteString("  End: " + util.Deref(event.GetEnd().GetDateTime()) + "\n")
	sb.WriteString("  In calendar: " + calendarName + " (ID " + calendar.ID + ")\n")
	return sb.String()
}

func PrintEvent(event models.Eventable, detailed bool) {
	fmt.Printf("Subject: %s\n", util.Deref(event.GetSubject()))
	fmt.Printf("  ID: %s\n", util.Deref(event.GetId()))
	fmt.Printf("  Start: %s\n", util.Deref(event.GetStart().GetDateTime()))
	fmt.Printf("  End: %s\n", util.Deref(event.GetEnd().GetDateTime()))

	if detailed {
		fmt.Printf("  Location: %s\n", util.Deref(event.GetLocation().GetDisplayName()))
		fmt.Printf("  Is All Day: %t\n", util.Deref(event.GetIsAllDay()))
		isRecurring := false
		if event.GetSeriesMasterId() != nil {
			isRecurring = true
		}
		fmt.Printf("  Is Recurring: %t\n", isRecurring)
		fmt.Printf("  Is Cancelled: %t\n", util.Deref(event.GetIsCancelled()))
		fmt.Printf("  Is Online Meeting: %t\n", util.Deref(event.GetIsOnlineMeeting()))
		fmt.Printf("  Response Status: %s\n", event.GetResponseStatus().GetResponse().String())
		fmt.Printf("  Attendees: %s\n", strings.Join(util.Map(event.GetAttendees(), func(a models.Attendeeable) string {
			return fmt.Sprintf("%s (%s), Response: %s", util.Deref(a.GetEmailAddress().GetName()), util.Deref(a.GetEmailAddress().GetAddress()), a.GetStatus().GetResponse().String())
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
