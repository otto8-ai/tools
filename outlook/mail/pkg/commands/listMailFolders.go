package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/client"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/global"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/printers"
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

	return printers.PrintMailFolders(result)
}
