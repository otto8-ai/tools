package commands

import (
	"context"

	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
	"github.com/gptscript-ai/tools/excel/pkg/printers"
)

func ListWorksheets(ctx context.Context, workbookID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	infos, err := graph.ListWorksheetsInWorkbook(ctx, c, workbookID)
	if err != nil {
		return err
	}

	printers.PrintWorksheetInfos(infos)
	return nil
}
