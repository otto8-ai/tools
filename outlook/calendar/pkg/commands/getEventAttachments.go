package commands

import (
	"context"
	"fmt"
	"log"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/client"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/global"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func GetEventAttachments(ctx context.Context, eventID, calendarID string, owner graph.OwnerType) error {
	trueEventID, err := id.GetOutlookID(ctx, eventID)
	if err != nil {
		return fmt.Errorf("failed to get Outlook ID: %w", err)
	}

	var trueCalendarID string
	if calendarID != "" {
		trueCalendarID, err = id.GetOutlookID(ctx, calendarID)
		if err != nil {
			return fmt.Errorf("failed to get Outlook ID: %w", err)
		}
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	event, err := graph.GetEvent(ctx, c, trueEventID, trueCalendarID, owner)
	if err != nil {
		return fmt.Errorf("failed to get event: %w", err)
	}

	event.SetId(util.Ptr(eventID))

	attachments := event.GetAttachments()
	if len(attachments) < 1 {
		return nil
	}
	gptscriptClient, _ := gptscript.NewGPTScript()
	for _, attachment := range attachments {
		attachmentType := util.Deref(attachment.GetOdataType())
		if attachmentType != "#microsoft.graph.fileAttachment" {
			fmt.Printf("Skipping non-file attachment: %s\n", *attachment.GetId())
			continue
		}

		fileAttachment := attachment.(*models.FileAttachment)
		fileName := *fileAttachment.GetName()
		contentBytes := fileAttachment.GetContentBytes()
		err = gptscriptClient.WriteFileInWorkspace(ctx, fileName, contentBytes)
		if err != nil {
			log.Fatalf("Error saving file: %v", err)
		}

		fmt.Printf("Attachment %s has been downloaded and saved.\n", fileName)
	}

	return nil
}
