package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func ListMailFolders(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := graph.ListMailFolders(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list mail folders: %w", err)
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create GPTScript client: %w", err)
	}

	folderIDs := util.Map(result, func(folder models.MailFolderable) string {
		return util.Deref(folder.GetId())
	})
	translatedFolderIDs, err := id.SetOutlookIDs(ctx, folderIDs)
	if err != nil {
		return fmt.Errorf("failed to set Outlook IDs: %w", err)
	}

	parentFolderIDs := util.Filter(util.Map(result, func(folder models.MailFolderable) string {
		if folder.GetParentFolderId() != nil {
			return util.Deref(folder.GetParentFolderId())
		}
		return ""
	}), func(id string) bool {
		return id != ""
	})
	translatedParentFolderIDs, err := id.SetOutlookIDs(ctx, parentFolderIDs)
	if err != nil {
		return fmt.Errorf("failed to set Outlook IDs: %w", err)
	}

	var elements []gptscript.DatasetElement
	for _, folder := range result {
		folder.SetId(util.Ptr(translatedFolderIDs[util.Deref(folder.GetId())]))
		if folder.GetParentFolderId() != nil {
			folder.SetParentFolderId(util.Ptr(translatedParentFolderIDs[util.Deref(folder.GetParentFolderId())]))
		}

		folderStr, err := printers.MailFolderToString(folder)
		if err != nil {
			return fmt.Errorf("failed to convert mail folder to string: %w", err)
		}

		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name:        util.Deref(folder.GetId()),
				Description: util.Deref(folder.GetDisplayName()),
			},
			Contents: folderStr,
		})
	}

	datasetID, err := gptscriptClient.CreateDatasetWithElements(ctx, elements, gptscript.DatasetOptions{
		Name: "outlook_mail_folders",
	})
	if err != nil {
		return fmt.Errorf("failed to create dataset with elements: %w", err)
	}

	fmt.Printf("Created dataset with ID %s with %d folders\n", datasetID, len(result))
	return nil
}
