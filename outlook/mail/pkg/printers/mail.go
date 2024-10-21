package printers

import (
	"fmt"
	"strings"
	"time"

	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
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
	folderStr, err := MailFolderToString(folder)
	if err != nil {
		return fmt.Errorf("failed to convert mail folder to string: %w", err)
	}

	fmt.Println(folderStr)
	return nil
}

func MailFolderToString(folder models.MailFolderable) (string, error) {
	var result strings.Builder
	if err := replaceMailFolderIDs(folder); err != nil {
		return "", fmt.Errorf("failed to fix mail folder: %w", err)
	}

	result.WriteString(fmt.Sprintf("Name: %s\n", util.Deref(folder.GetDisplayName())))
	result.WriteString(fmt.Sprintf("ID: %s\n", util.Deref(folder.GetId())))
	if folder.GetParentFolderId() != nil {
		result.WriteString(fmt.Sprintf("Parent folder ID: %s\n", util.Deref(folder.GetParentFolderId())))
	}
	result.WriteString(fmt.Sprintf("Unread item count: %d\n", util.Deref(folder.GetUnreadItemCount())))
	result.WriteString(fmt.Sprintf("Total item count: %d\n", util.Deref(folder.GetTotalItemCount())))

	return result.String(), nil
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
	messageStr, err := MessageToString(msg, detailed)
	if err != nil {
		return fmt.Errorf("failed to convert message to string: %w", err)
	}
	fmt.Println(messageStr)
	return nil
}

func MessageToString(msg models.Messageable, detailed bool) (string, error) {
	var result strings.Builder
	if err := replaceMessageIDs(msg); err != nil {
		return "", fmt.Errorf("failed to fix message: %w", err)
	}

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

		bodyText, err := html2text.FromString(util.Deref(msg.GetBody().GetContent()))
		if err != nil {
			return "", fmt.Errorf("failed to convert HTML to text: %w", err)
		}
		result.WriteString(fmt.Sprintf("Body: %s", strings.ReplaceAll(bodyText, "\n", "\n  ")))
	} else {
		result.WriteString(fmt.Sprintf("Body preview: %s\n", strings.ReplaceAll(util.Deref(msg.GetBodyPreview()), "\n", "\n  ")))
	}

	return result.String(), nil
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
