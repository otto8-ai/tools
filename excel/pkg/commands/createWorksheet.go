package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
)

func CreateWorksheet(ctx context.Context, workbookID, name string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return err
	}

	id, err := graph.CreateWorksheet(ctx, c, workbookID, name)
	if err != nil {
		return err
	}

	fmt.Printf("Worksheet created with ID: %s\n", id)
	return nil
}
