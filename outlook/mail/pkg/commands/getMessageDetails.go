package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/common/id"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/printers"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/util"
)

func GetMessageDetails(ctx context.Context, messageID string) error {
	trueMessageID, err := id.GetOutlookID(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}

	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	result, err := graph.GetMessageDetails(ctx, c, trueMessageID)
	if err != nil {
		return fmt.Errorf("failed to get message details: %w", err)
	}

	result.SetId(&trueMessageID)

	parentFolderID, err := id.GetOutlookID(ctx, util.Deref(result.GetParentFolderId()))
	if err != nil {
		return fmt.Errorf("failed to get outlook ID: %w", err)
	}
	result.SetParentFolderId(&parentFolderID)

	if err := printers.PrintMessage(result, true); err != nil {
		return fmt.Errorf("failed to print message: %w", err)
	}
	return nil
}
