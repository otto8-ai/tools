package printers

import (
	"context"
	"fmt"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/jaytaylor/html2text"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"strings"
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
	startTZ, endTZ := EventDisplayTimeZone(event)
	sb.WriteString("  Start: " + util.Deref(event.GetStart().GetDateTime()) + startTZ + "\n")
	sb.WriteString("  End: " + util.Deref(event.GetEnd().GetDateTime()) + endTZ + "\n")
	sb.WriteString("  In calendar: " + calendarName + " (ID " + calendar.ID + ")\n")
	return sb.String()
}

func PrintEvent(event models.Eventable, detailed bool) {
	fmt.Printf("Subject: %s\n", util.Deref(event.GetSubject()))
	fmt.Printf("  ID: %s\n", util.Deref(event.GetId()))
	startTZ, endTZ := EventDisplayTimeZone(event)
	fmt.Printf("  Start: %s%s\n", util.Deref(event.GetStart().GetDateTime()), startTZ)
	fmt.Printf("  End: %s%s\n", util.Deref(event.GetEnd().GetDateTime()), endTZ)

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
		attachments := event.GetAttachments()
		if len(attachments) > 0 {
			for _, attachment := range attachments {
				attachmentType := util.Deref(attachment.GetOdataType())
				if attachmentType == "#microsoft.graph.fileAttachment" {
					fileAttachment := attachment.(*models.FileAttachment)
					fmt.Printf("File Attachment: %s, Size: %d bytes, Content Type: %s\n", *fileAttachment.GetName(), *fileAttachment.GetSize(), *fileAttachment.GetContentType())
				} else if attachmentType == "#microsoft.graph.itemAttachment" {
					itemAttachment := attachment.(*models.ItemAttachment)
					fmt.Printf("Item Attachment: %s\n", *itemAttachment.GetName())
				}
			}
		}
		fmt.Printf("You can open the event using this link: %s\n", util.Deref(event.GetWebLink()))
	}
	fmt.Println()
}

func EventDisplayTimeZone(event models.Eventable) (string, string) {
	// No TZ for all day events to avoid messing up the start/end times during conversion
	startTZ, endTZ := "", ""
	if util.Deref(event.GetIsAllDay()) {
		return startTZ, endTZ
	}
	// Assume that timestamps are UTC by default, but verify
	if util.Deref(event.GetStart().GetTimeZone()) == "UTC" {
		startTZ = "Z"
	} else {
		startTZ = " " + util.Deref(event.GetStart().GetTimeZone())
	}
	if util.Deref(event.GetEnd().GetTimeZone()) == "UTC" {
		endTZ = "Z"
		endTZ = " " + util.Deref(event.GetEnd().GetTimeZone())
	}
	return startTZ, endTZ
}
