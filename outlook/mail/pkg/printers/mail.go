package printers

import (
	"fmt"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func MailFolderToString(folder models.MailFolderable) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Name: %s\n", util.Deref(folder.GetDisplayName())))
	result.WriteString(fmt.Sprintf("ID: %s\n", util.Deref(folder.GetId())))
	if folder.GetParentFolderId() != nil {
		result.WriteString(fmt.Sprintf("Parent folder ID: %s\n", util.Deref(folder.GetParentFolderId())))
	}
	result.WriteString(fmt.Sprintf("Unread item count: %d\n", util.Deref(folder.GetUnreadItemCount())))
	result.WriteString(fmt.Sprintf("Total item count: %d\n", util.Deref(folder.GetTotalItemCount())))

	return result.String(), nil
}

func PrintMessage(msg models.Messageable, detailed bool) error {
	messageStr, err := MessageToString(msg, detailed)
	if err != nil {
		return fmt.Errorf("failed to convert message to string: %w", err)
	}
	fmt.Println(messageStr)
	return nil
}

func MessageToString(msg models.Messageable, detailed bool) (string, error) {
	var result strings.Builder

	result.WriteString(fmt.Sprintf("Subject: %s\n", util.Deref(msg.GetSubject())))
	result.WriteString(fmt.Sprintf("Message ID: %s\n", util.Deref(msg.GetId())))
	if !util.Deref(msg.GetIsDraft()) {
		result.WriteString(fmt.Sprintf("Sender: %s (email address: %s)\n", util.Deref(msg.GetSender().GetEmailAddress().GetName()), util.Deref(msg.GetSender().GetEmailAddress().GetAddress())))
		result.WriteString(fmt.Sprintf("Received: %s\n", msg.GetReceivedDateTime().Format(time.RFC3339)))
	} else {
		result.WriteString(fmt.Sprintf("Created: %s\n", msg.GetReceivedDateTime().Format(time.RFC3339)))
	}
	result.WriteString(fmt.Sprintf("Is unread: %t\n", !util.Deref(msg.GetIsRead())))
	result.WriteString(fmt.Sprintf("Link: %s\n", util.Deref(msg.GetWebLink())))

	if detailed {
		result.WriteString(fmt.Sprintf("To: %s\n", strings.Join(util.Map(msg.GetToRecipients(), recipientableToString), ", ")))
		result.WriteString(fmt.Sprintf("CC: %s\n", strings.Join(util.Map(msg.GetCcRecipients(), recipientableToString), ", ")))
		result.WriteString(fmt.Sprintf("Has attachments: %t\n", util.Deref(msg.GetHasAttachments())))

		converter := md.NewConverter("", true, nil)
		bodyHTML := util.Deref(msg.GetBody().GetContent())
		bodyMarkdown, err := converter.ConvertString(bodyHTML)
		if err != nil {
			return "", fmt.Errorf("failed to convert email body HTML to markdown: %w", err)
		}

		result.WriteString(fmt.Sprintf("Body: %s", strings.ReplaceAll(bodyMarkdown, "\n", "\n  ")))
	} else {
		result.WriteString(fmt.Sprintf("Body preview: %s\n", strings.ReplaceAll(util.Deref(msg.GetBodyPreview()), "\n", "\n  ")))
	}

	return result.String(), nil
}

func recipientableToString(r models.Recipientable) string {
	return fmt.Sprintf("%s (%s)", util.Deref(r.GetEmailAddress().GetName()), util.Deref(r.GetEmailAddress().GetAddress()))
}
