package printers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/util"
	"github.com/jaytaylor/html2text"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func PrintMailFolders(folders []models.MailFolderable) {
	for _, folder := range folders {
		PrintMailFolder(folder)
		fmt.Println()
	}
}

func PrintMailFolder(folder models.MailFolderable) {
	fmt.Printf("Name: %s\n", util.Deref(folder.GetDisplayName()))
	fmt.Printf("ID: %s\n", util.Deref(folder.GetId()))
	if folder.GetParentFolderId() != nil {
		fmt.Printf("Parent folder ID: %s\n", util.Deref(folder.GetParentFolderId()))
	}
	fmt.Printf("Unread item count: %d\n", util.Deref(folder.GetUnreadItemCount()))
	fmt.Printf("Total item count: %d\n", util.Deref(folder.GetTotalItemCount()))
}

func PrintMessagesForFolder(folder models.MailFolderable, messages []models.Messageable, detailed bool) {
	fmt.Printf("Messages in folder %s:\n", util.Deref(folder.GetDisplayName()))
	PrintMessages(messages, detailed)
	fmt.Println()
}

func PrintMessages(messages []models.Messageable, detailed bool) {
	for _, msg := range messages {
		if err := PrintMessage(msg, detailed); err != nil {
			fmt.Printf("failed to print message: %v\n", err)
		}
		fmt.Println()
	}
}

func PrintMessage(msg models.Messageable, detailed bool) error {
	fmt.Printf("Subject: %s\n", util.Deref(msg.GetSubject()))
	if !util.Deref(msg.GetIsDraft()) {
		fmt.Printf("Sender: %s (email address: %s)\n", util.Deref(msg.GetSender().GetEmailAddress().GetName()), util.Deref(msg.GetSender().GetEmailAddress().GetAddress()))
	}
	fmt.Printf("Message ID: %s\n", util.Deref(msg.GetId()))
	fmt.Printf("Received: %s\n", msg.GetReceivedDateTime().Format(time.RFC3339))
	fmt.Printf("Is unread: %t\n", !util.Deref(msg.GetIsRead()))

	if detailed {
		fmt.Printf("To: %s\n", strings.Join(util.Map(msg.GetToRecipients(), recipientableToString), ", "))
		fmt.Printf("CC: %s\n", strings.Join(util.Map(msg.GetCcRecipients(), recipientableToString), ", "))
		fmt.Printf("Has attachments: %t\n", util.Deref(msg.GetHasAttachments()))

		bodyText, err := html2text.FromString(util.Deref(msg.GetBody().GetContent()))
		if err != nil {
			return fmt.Errorf("failed to convert HTML to text: %w", err)
		}
		fmt.Printf("Body: %s", strings.ReplaceAll(bodyText, "\n", "\n  "))
	} else {
		fmt.Printf("Body preview: %s\n", strings.ReplaceAll(util.Deref(msg.GetBodyPreview()), "\n", "\n  "))
	}

	return nil
}

func recipientableToString(r models.Recipientable) string {
	return fmt.Sprintf("%s (%s)", util.Deref(r.GetEmailAddress().GetName()), util.Deref(r.GetEmailAddress().GetAddress()))
}
