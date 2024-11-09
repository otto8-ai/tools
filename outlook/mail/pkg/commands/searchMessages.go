package commands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func SearchMessages(ctx context.Context, subject, fromAddress, fromName, folderID, start, end, limit string) error {
	var (
		limitInt = 10
		err      error
	)
	if limit != "" {
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			return fmt.Errorf("failed to parse limit: %w", err)
		}
	}

	trueFolderID, err := id.GetOutlookID(ctx, folderID)
	if err != nil {
		return fmt.Errorf("failed to get folder ID: %w", err)
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	messages, err := graph.SearchMessages(ctx, c, subject, fromAddress, fromName, trueFolderID, start, end, limitInt)
	if err != nil {
		return fmt.Errorf("failed to search messages: %w", err)
	}

	if len(messages) == 0 {
		fmt.Println("no messages found")
		return nil
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	// Translate Outlook IDs to friendly IDs before we print.
	messageIDs := util.Map(messages, func(message models.Messageable) string {
		return util.Deref(message.GetId())
	})
	translatedMessageIDs, err := id.SetOutlookIDs(ctx, messageIDs)
	if err != nil {
		return fmt.Errorf("failed to translate message IDs: %w", err)
	}

	folderIDs := util.Map(messages, func(message models.Messageable) string {
		return util.Deref(message.GetParentFolderId())
	})
	translatedFolderIDs, err := id.SetOutlookIDs(ctx, folderIDs)
	if err != nil {
		return fmt.Errorf("failed to translate folder IDs: %w", err)
	}

	var elements []gptscript.DatasetElement
	for _, message := range messages {
		message.SetId(util.Ptr(translatedMessageIDs[util.Deref(message.GetId())]))
		message.SetParentFolderId(util.Ptr(translatedFolderIDs[util.Deref(message.GetParentFolderId())]))

		messageStr, err := printers.MessageToString(message, false)
		if err != nil {
			return fmt.Errorf("failed to convert message to string: %w", err)
		}
		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        util.Deref(message.GetId()),
				Description: util.Deref(message.GetSubject()),
			},
			Contents: messageStr,
		})
	}

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name: fmt.Sprintf("outlook_mail_search_results"),
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d messages\n", datasetID, len(messages))
	return nil
}
