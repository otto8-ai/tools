package graph

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/models/odataerrors"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

func ListMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, folderID, start, end string, limit int) ([]models.Messageable, error) {
	queryParams := &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{
		Orderby: []string{"receivedDateTime DESC"},
	}

	if limit > 0 {
		queryParams.Top = util.Ptr(int32(limit))
	}

	var filters []string
	if start != "" {
		filters = append(filters, fmt.Sprintf("receivedDateTime ge %s", start))
	}
	if end != "" {
		filters = append(filters, fmt.Sprintf("receivedDateTime le %s", end))
	}

	if len(filters) > 0 {
		queryParams.Filter = util.Ptr(strings.Join(filters, " and "))
	}

	result, err := client.Me().MailFolders().ByMailFolderId(folderID).Messages().Get(ctx, &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: queryParams,
	})
	// TODO - handle pagination if there are more messages than can be returned in a single call
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

func SearchMessages(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, subject, fromAddress, fromName, folderID, start, end string, limit int) ([]models.Messageable, error) {
	var (
		result models.MessageCollectionResponseable
		err    error
		filter []string
	)

	// It is important that a receivedDateTime filter is first in the list.
	// Details in the first answer on this question:
	// https://learn.microsoft.com/en-us/answers/questions/656200/graph-api-to-filter-results-on-from-and-subject-an
	if end != "" {
		filter = append(filter, fmt.Sprintf("receivedDateTime le %s", end))
	} else {
		// Using the receivedDateTime in the orderBy parameter requires us to have it in the filter as well.
		// So we ask for messages that were received prior to tomorrow, which should be all messages.
		tomorrow := time.Now().Add(time.Hour * 24).Format(time.RFC3339)
		filter = append(filter, fmt.Sprintf("receivedDateTime le %s", tomorrow))
	}
	if subject != "" {
		filter = append(filter, fmt.Sprintf("contains(subject, '%s')", subject))
	}
	if fromAddress != "" {
		filter = append(filter, fmt.Sprintf("contains(from/emailAddress/address, '%s')", fromAddress))
	}
	if fromName != "" {
		filter = append(filter, fmt.Sprintf("contains(from/emailAddress/name, '%s')", fromName))
	}
	if start != "" {
		filter = append(filter, fmt.Sprintf("receivedDateTime ge %s", start))
	}

	if len(filter) == 0 {
		return nil, fmt.Errorf("at least one of subject, from_address, or from_name must be provided")
	}

	if folderID != "" {
		result, err = client.Me().MailFolders().ByMailFolderId(folderID).Messages().Get(ctx, &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{
				Orderby: []string{"receivedDateTime DESC"},
				Filter:  util.Ptr(strings.Join(filter, " and ")),
				Top:     util.Ptr(int32(limit)),
			},
		})
	} else {
		result, err = client.Me().Messages().Get(ctx, &users.ItemMessagesRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemMessagesRequestBuilderGetQueryParameters{
				Orderby: []string{"receivedDateTime DESC"},
				Filter:  util.Ptr(strings.Join(filter, " and ")),
				Top:     util.Ptr(int32(limit)),
			},
		})
	}

	if err != nil {
		return nil, fmt.Errorf("failed to search messages: %w", err)
	}

	return result.GetValue(), nil
}

type DraftInfo struct {
	Subject, Body       string
	Recipients, CC, BCC []string // slice of email addresses
	Attachments         []string // slice of workspace file paths
}

var (
	mdParser   = parser.NewWithExtensions(parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock)
	mdRenderer = html.NewRenderer(html.RendererOptions{
		Flags: html.CompletePage,
	})
)

func CreateDraft(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, info DraftInfo) (models.Messageable, error) {
	requestBody := models.NewMessage()
	requestBody.SetIsDraft(util.Ptr(true))
	requestBody.SetSubject(util.Ptr(info.Subject))
	requestBody.SetToRecipients(emailAddressesToRecipientable(info.Recipients))

	for _, file := range info.Attachments {
		if file == "" {
			return nil, fmt.Errorf("attachment file path cannot be empty")
		}
	}

	if len(info.CC) > 0 {
		requestBody.SetCcRecipients(emailAddressesToRecipientable(info.CC))
	}

	if len(info.BCC) > 0 {
		requestBody.SetBccRecipients(emailAddressesToRecipientable(info.BCC))
	}

	body := models.NewItemBody()
	body.SetContentType(util.Ptr(models.HTML_BODYTYPE))

	bodyHTML := string(markdown.Render(mdParser.Parse([]byte(info.Body)), mdRenderer))
	body.SetContent(util.Ptr(bodyHTML))

	requestBody.SetBody(body)

	draft, err := client.Me().Messages().Post(ctx, requestBody, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft message: %w", err)
	}

	if len(info.Attachments) > 0 {
		if err := attachFiles(ctx, client, util.Deref(draft.GetId()), info.Attachments); err != nil {
			return nil, fmt.Errorf("failed to attach files to draft: %w", err)
		}
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

func attachFiles(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, draftID string, files []string) error {
	gsClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	// Set an upload deadline to prevent the function from hanging indefinitely
	uploadCtx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Minute*5))
	defer cancel()

	// Note: While it's tempting to paralleize attachment uploads, Microsoft Graph API doesn't
	// seem to support concurrent upload sessions (returns "change key" errors) or non-sequential
	// file chunk uploads (returns "invalid start offset" errors).
	var errs []error
	for _, file := range files {
		// Read the file from the workspace
		data, err := gsClient.ReadFileInWorkspace(uploadCtx, filepath.Join("files", file))
		if err != nil {
			return fmt.Errorf("failed to read attachment file %s from workspace: %v", file, err)
		}

		if len(data) < 1 {
			return fmt.Errorf("cannot attach empty file %s", file)
		}

		errs = append(errs, uploadFile(uploadCtx, client, draftID, file, data))
	}

	return errors.Join(errs...)
}

const uploadChunkSize = 1024 * 1024 // 1MB

func uploadFile(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, draftID string, file string, data []byte) error {
	// Prepare attachment info
	attachment := models.NewAttachmentItem()
	attachment.SetAttachmentType(util.Ptr(models.FILE_ATTACHMENTTYPE))
	attachment.SetName(util.Ptr(filepath.Base(file)))
	attachment.SetSize(util.Ptr(int64(len(data))))
	attachment.SetAdditionalData(map[string]any{"@microsoft.graph.conflictBehavior": "replace"})

	// Create the request body for the upload session
	requestBody := users.NewItemMessagesItemAttachmentsCreateUploadSessionPostRequestBody()
	requestBody.SetAttachmentItem(attachment)
	requestBody.SetAdditionalData(map[string]any{"@microsoft.graph.conflictBehavior": "replace"})

	// Create the upload session
	session, err := client.Me().Messages().ByMessageId(draftID).Attachments().CreateUploadSession().Post(ctx, requestBody, nil)
	if err != nil {
		return fmt.Errorf("failed to create upload session for file %s: %v", file, err)
	}

	// Determine the number of chunks to upload
	var (
		totalSize = len(data)
		numChunks = (totalSize + uploadChunkSize - 1) / uploadChunkSize
	)

	// Upload chunks sequentially
	for i := 0; i < numChunks; i++ {
		var (
			start = i * uploadChunkSize
			end   = start + uploadChunkSize
		)
		if end > totalSize {
			end = totalSize
		}

		chunk := data[start:end]
		contentRange := fmt.Sprintf("bytes %d-%d/%d", start, end-1, totalSize)

		// Create a request to upload the chunk
		requestInfo := abstractions.NewRequestInformation()
		requestInfo.UrlTemplate = *session.GetUploadUrl()
		requestInfo.Method = abstractions.PUT
		requestInfo.Headers.Add("Content-Length", fmt.Sprintf("%d", len(chunk)))
		requestInfo.Headers.Add("Content-Range", contentRange)
		requestInfo.SetStreamContentAndContentType(chunk, "application/octet-stream")
		errorMapping := abstractions.ErrorMappings{
			"4XX": odataerrors.CreateODataErrorFromDiscriminatorValue,
			"5XX": odataerrors.CreateODataErrorFromDiscriminatorValue,
		}

		// Upload the chunk
		if err := client.BaseRequestBuilder.RequestAdapter.SendNoContent(ctx, requestInfo, errorMapping); err != nil {
			return fmt.Errorf("failed to upload chunk %s for file %s: %v", contentRange, file, err)
		}
	}

	return nil
}

func SendDraft(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, draftID string) error {
	if err := client.Me().Messages().ByMessageId(draftID).Send().Post(ctx, nil); err != nil {
		return fmt.Errorf("failed to send draft: %w", err)
	}

	return nil
}

func DeleteMessage(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, messageID string) error {
	folders, err := ListMailFolders(ctx, client)
	if err != nil {
		return fmt.Errorf("failed to list mail folders: %w", err)
	}

	for _, folder := range folders {
		if util.Deref(folder.GetDisplayName()) != "Deleted Items" {
			continue
		}

		if _, err := MoveMessage(ctx, client, messageID, util.Deref(folder.GetId())); err != nil {
			return fmt.Errorf("failed to move message to Deleted Items: %w", err)
		}
		return nil
	}

	return fmt.Errorf("failed to find Deleted Items folder")
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
