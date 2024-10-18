package graph

import (
	"bytes"
	"context"
	"fmt"

	"code.sajari.com/docconv/v2"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
)

type DocInfo struct {
	ID, Name string
}

func (d DocInfo) String() string {
	return fmt.Sprintf("Name: %s\nID: %s", d.Name, d.ID)
}

func ListDocs(ctx context.Context, c *msgraphsdkgo.GraphServiceClient) ([]DocInfo, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	docs, err := c.Drives().ByDriveId(deref(drive.GetId())).SearchWithQ(ptr("docx")).GetAsSearchWithQGetResponse(ctx, nil)
	if err != nil {
		return nil, err
	}

	var infos []DocInfo
	for _, info := range docs.GetValue() {
		infos = append(infos, DocInfo{
			ID:   deref(info.GetId()),
			Name: deref(info.GetName()),
		})
	}

	return infos, nil
}

func GetDoc(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, docID string) (string, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	doc, err := c.Drives().ByDriveId(deref(drive.GetId())).Items().ByDriveItemId(docID).Content().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	content, err := docconv.Convert(bytes.NewReader(doc), "application/vnd.ms-word", true)
	if err != nil {
		return "", err
	}

	return content.Body, nil
}

func ptr[T any](v T) *T {
	return &v
}

func deref[T any](v *T) (r T) {
	if v != nil {
		return *v
	}
	return
}
