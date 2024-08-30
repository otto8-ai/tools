package graph

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/apis/outlook/mail/code/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

func ListMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, folderID string) ([]models.Messageable, error) {
	result, err := client.Me().MailFolders().ByMailFolderId(folderID).Messages().Get(ctx, &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{
			Top: util.Ptr(int32(100)),
		},
	})

	// TODO - handle if there are more than 100

	if err != nil {
		return nil, fmt.Errorf("failed to list mail: %w", err)
	}

	return result.GetValue(), nil
}

func GetMessageDetails(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, messageID string) (models.Messageable, error) {
	result, err := client.Me().Messages().ByMessageId(messageID).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get message details: %w", err)
	}

	return result, nil
}

func SearchMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, subject, fromAddress, fromName, folderID string) ([]models.Messageable, error) {
	var (
		result models.MessageCollectionResponseable
		err    error
		filter []string
	)

	if subject != "" {
		filter = append(filter, fmt.Sprintf("contains(subject, '%s')", subject))
	}
	if fromAddress != "" {
		filter = append(filter, fmt.Sprintf("contains(from/emailAddress/address, '%s')", fromAddress))
	}
	if fromName != "" {
		filter = append(filter, fmt.Sprintf("contains(from/emailAddress/name, '%s')", fromName))
	}

	if len(filter) == 0 {
		return nil, fmt.Errorf("at least one of subject, from_address, or from_name must be provided")
	}

	if folderID != "" {
		result, err = client.Me().MailFolders().ByMailFolderId(folderID).Messages().Get(ctx, &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{
				Filter: util.Ptr(strings.Join(filter, " and ")),
				Top:    util.Ptr(int32(10)),
			},
		})
	} else {
		result, err = client.Me().Messages().Get(ctx, &users.ItemMessagesRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemMessagesRequestBuilderGetQueryParameters{
				Filter: util.Ptr(strings.Join(filter, " and ")),
				Top:    util.Ptr(int32(10)),
			},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	return result.GetValue(), nil
}

type DraftInfo struct {
	Subject, Content    string
	Recipients, CC, BCC []string // slice of email addresses
}

func CreateDraft(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, info DraftInfo) (models.Messageable, error) {
	requestBody := models.NewMessage()
	requestBody.SetIsDraft(util.Ptr(true))
	requestBody.SetSubject(util.Ptr(info.Subject))
	requestBody.SetToRecipients(emailAddressesToRecipientable(info.Recipients))

	if len(info.CC) > 0 {
		requestBody.SetCcRecipients(emailAddressesToRecipientable(info.CC))
	}

	if len(info.BCC) > 0 {
		requestBody.SetBccRecipients(emailAddressesToRecipientable(info.BCC))
	}

	body := models.NewItemBody()
	body.SetContentType(util.Ptr(models.HTML_BODYTYPE))
	body.SetContent(util.Ptr(info.Content))

	requestBody.SetBody(body)

	draft, err := client.Me().Messages().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft message: %w", err)
	}

	return draft, nil
}

func emailAddressesToRecipientable(addresses []string) []models.Recipientable {
	var recipients []models.Recipientable
	for _, address := range addresses {
		addr := models.NewEmailAddress()
		addr.SetAddress(util.Ptr(address))
		r := models.NewRecipient()
		r.SetEmailAddress(addr)
		recipients = append(recipients, r)
	}
	return recipients
}

func SendDraft(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, draftID string) error {
	if err := client.Me().Messages().ByMessageId(draftID).Send().Post(ctx, nil); err != nil {
		return fmt.Errorf("failed to send draft: %w", err)
	}

	return nil
}

func DeleteMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, messageID string) error {
	if err := client.Me().Messages().ByMessageId(messageID).Delete(ctx, nil); err != nil {
		return fmt.Errorf("failed to delete message: %w", err)
	}

	return nil
}

func MoveMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, messageID, destinationFolderID string) (models.Messageable, error) {
	requestBody := users.NewItemMessagesItemMovePostRequestBody()
	requestBody.SetDestinationId(util.Ptr(destinationFolderID))

	message, err := client.Me().Messages().ByMessageId(messageID).Move().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to move message: %w", err)
	}

	return message, nil
}
