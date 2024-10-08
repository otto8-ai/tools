package graph

import (
	"context"

	"github.com/gptscript-ai/tools/outlook/calendar/pkg/util"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

func GetGroupNameFromID(ctx context.Context, client *msgraphsdkgo.GraphServiceClient, id string) (string, error) {
	resp, err := client.Groups().ByGroupId(id).Get(ctx, nil)
	if err != nil {
		return "", err
	}

	return util.Deref(resp.GetDisplayName()), nil
}
