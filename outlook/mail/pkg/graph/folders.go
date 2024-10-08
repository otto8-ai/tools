package graph

import (
	"context"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

func ListMailFolders(ctx context.Context, client *msgraphsdkgo.GraphServiceClient) ([]models.MailFolderable, error) {
	result, err := client.Me().MailFolders().Get(ctx, &users.ItemMailFoldersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMailFoldersRequestBuilderGetQueryParameters{
			Top: util.Ptr(int32(100)),
		},
	})

	// TODO - handle if there are more than 100

	if err != nil {
		return nil, err
	}

	return result.GetValue(), nil
}
