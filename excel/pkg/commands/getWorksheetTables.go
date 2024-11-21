package commands

import (
	"context"
	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
	"github.com/gptscript-ai/tools/excel/pkg/printers"
)

func GetWorksheetTables(ctx context.Context, workbookID, worksheetID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	tables, err := graph.GetWorksheetTables(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}
	printers.PrintWorksheetTableInfos(tables)
	return nil
}
