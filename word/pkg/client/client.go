package client

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/gptscript-ai/tools/word/pkg/global"
)

type StaticTokenCredential struct {
	token string
}

func (s StaticTokenCredential) GetToken(_ context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{Token: s.token}, nil
}

func NewClient(scopes []string) (*msgraphsdkgo.GraphServiceClient, error) {
	return msgraphsdkgo.NewGraphServiceClientWithCredentials(StaticTokenCredential{
		token: os.Getenv(global.CredentialEnv),
	}, scopes)
}
