package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
	"strings"
)

func FilterWorksheetData(ctx context.Context, workbookID, worksheetID, filterColumn, filterValues string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	tables, err := graph.GetWorksheetTables(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}
	tableID := tables[0].ID

	columnID, err := graph.GetColumnIDFromName(ctx, c, workbookID, worksheetID, tableID, filterColumn)
	if err != nil {
		return err
	}

	filterSlice := strings.Split(filterValues, ",")
	for i, s := range filterSlice {
		filterSlice[i] = strings.TrimSpace(s)
	}

	data, _, err := graph.FilterWorksheetData(ctx, c, workbookID, worksheetID, tableID, columnID, filterSlice)
	if err != nil {
		return err
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	fmt.Println(string(dataBytes))

	return nil
}
