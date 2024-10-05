package printers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gptscript-ai/tools/apis/outlook/common/id"
	"github.com/gptscript-ai/tools/apis/outlook/mail/pkg/util"
	"github.com/jaytaylor/html2text"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func PrintMailFolders(folders []models.MailFolderable) error {
	for _, folder := range folders {
		if err := PrintMailFolder(folder); err != nil {
			return err
		}
		fmt.Println()
	}
	return nil
}

func PrintMailFolder(folder models.MailFolderable) error {
	if err := replaceMailFolderIDs(folder); err != nil {
		return fmt.Errorf("failed to fix mail folder: %w", err)
	}

	fmt.Printf("Name: %s\n", util.Deref(folder.GetDisplayName()))
	fmt.Printf("ID: %s\n", util.Deref(folder.GetId()))
	if folder.GetParentFolderId() != nil {
		fmt.Printf("Parent folder ID: %s\n", util.Deref(folder.GetParentFolderId()))
	}
	fmt.Printf("Unread item count: %d\n", util.Deref(folder.GetUnreadItemCount()))
	fmt.Printf("Total item count: %d\n", util.Deref(folder.GetTotalItemCount()))
	return nil
}

func PrintMessages(messages []models.Messageable, detailed bool) error {
	for _, msg := range messages {
		if err := PrintMessage(msg, detailed); err != nil {
			return fmt.Errorf("failed to print message: %w", err)
		}
		fmt.Println()
	}
	return nil
}

func PrintMessage(msg models.Messageable, detailed bool) error {
	if err := replaceMessageIDs(msg); err != nil {
		return fmt.Errorf("failed to fix message: %w", err)
	}

	fmt.Printf("Subject: %s\n", util.Deref(msg.GetSubject()))
	fmt.Printf("Message ID: %s\n", util.Deref(msg.GetId()))
	if !util.Deref(msg.GetIsDraft()) {
		fmt.Printf("Sender: %s (email address: %s)\n", util.Deref(msg.GetSender().GetEmailAddress().GetName()), util.Deref(msg.GetSender().GetEmailAddress().GetAddress()))
		fmt.Printf("Received: %s\n", msg.GetReceivedDateTime().Format(time.RFC3339))
	} else {
		fmt.Printf("Created: %s\n", msg.GetReceivedDateTime().Format(time.RFC3339))
	}
	fmt.Printf("Is unread: %t\n", !util.Deref(msg.GetIsRead()))
	fmt.Printf("Link: %s\n", util.Deref(msg.GetWebLink()))

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

// replaceMailFolderIDs replaces the ID values of the mail folder itself and its parent
// with the corresponding numerical ID that we generate in the database.
// This is necessary to do prior to printing it for the LLM.
func replaceMailFolderIDs(folder models.MailFolderable) error {
	newFolderID, err := id.SetOutlookID(util.Deref(folder.GetId()))
	if err != nil {
		return fmt.Errorf("failed to set folder ID: %w", err)
	}

	folder.SetId(util.Ptr(newFolderID))

	if folder.GetParentFolderId() != nil {
		newParentFolderID, err := id.SetOutlookID(util.Deref(folder.GetParentFolderId()))
		if err != nil {
			return fmt.Errorf("failed to set parent folder ID: %w", err)
		}

		folder.SetParentFolderId(util.Ptr(newParentFolderID))
	}
	return nil
}

// replaceMessageIDs replaces the ID values of the message itself and its parent folder
// with the corresponding numerical ID that we generate in the database.
// This is necessary to do prior to printing it for the LLM.
func replaceMessageIDs(msg models.Messageable) error {
	newMessageID, err := id.SetOutlookID(util.Deref(msg.GetId()))
	if err != nil {
		return fmt.Errorf("failed to set message ID: %w", err)
	}

	msg.SetId(util.Ptr(newMessageID))

	newFolderID, err := id.SetOutlookID(util.Deref(msg.GetParentFolderId()))
	if err != nil {
		return fmt.Errorf("failed to set folder ID: %w", err)
	}

	msg.SetParentFolderId(util.Ptr(newFolderID))
	return nil
}
