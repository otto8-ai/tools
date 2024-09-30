package commands

import (
	"context"

	"github.com/gptscript-ai/tools/apis/excel/code/pkg/client"
	"github.com/gptscript-ai/tools/apis/excel/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/excel/code/pkg/graph"
	"github.com/gptscript-ai/tools/apis/excel/code/pkg/printers"
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
