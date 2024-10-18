package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/word/pkg/client"
	"github.com/gptscript-ai/tools/word/pkg/global"
	"github.com/gptscript-ai/tools/word/pkg/graph"
)

func GetDoc(ctx context.Context, docID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	content, err := graph.GetDoc(ctx, c, docID)
	if err != nil {
		return fmt.Errorf("failed to list word docs: %w", err)
	}

	fmt.Println(content)

	return nil
}
