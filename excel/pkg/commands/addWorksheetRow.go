package commands

import (
	"context"
	"fmt"
	"strings"

	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
)

func AddWorksheetRow(ctx context.Context, workbookID, worksheetID, contents string) error {
	c, err := client.NewClient(global.AllScopes)
	if err != nil {
		return err
	}

	if err := graph.AddWorksheetRow(ctx, c, workbookID, worksheetID, strings.Split(contents, "|")); err != nil {
		return err
	}
	fmt.Println("Row added successfully")
	return nil
}
