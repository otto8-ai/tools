package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/word/pkg/client"
	"github.com/gptscript-ai/tools/word/pkg/global"
	"github.com/gptscript-ai/tools/word/pkg/graph"
)

func ListDocs(ctx context.Context) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	infos, err := graph.ListDocs(ctx, c)
	if err != nil {
		return fmt.Errorf("failed to list word docs: %w", err)
	}

	printAll(infos...)

	return nil
}
