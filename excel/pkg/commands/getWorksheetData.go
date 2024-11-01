package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
)

func GetWorksheetData(ctx context.Context, workbookID, worksheetID string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	data, _, err := graph.GetWorksheetData(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}

	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		return fmt.Errorf("failed to create gptscript client: %w", err)
	}

	dataset, err := gptscriptClient.CreateDataset(ctx,
		os.Getenv("GPTSCRIPT_WORKSPACE_ID"),
		fmt.Sprintf("%s_%s_worksheet_data", worksheetID, workbookID),
		fmt.Sprintf("Data from Excel worksheet %s in workbook %s", worksheetID, workbookID),
	)
	if err != nil {
		return err
	}

	var elements []gptscript.DatasetElement
	for i, row := range data {
		rowJSON, err := json.Marshal(row)
		if err != nil {
			return fmt.Errorf("failed to marshal row %d: %w", i, err)
		}

		elements = append(elements, gptscript.DatasetElement{
			DatasetElementMeta: gptscript.DatasetElementMeta{
				Name: fmt.Sprintf("row_%d", i),
			},
			Contents: rowJSON,
		})

		if i == 5000 { // Stop after 5k rows. It's just too many, at least for now.
			break
		}
	}

	if err := gptscriptClient.AddDatasetElements(ctx, os.Getenv("GPTSCRIPT_WORKSPACE_ID"), dataset.ID, elements); err != nil {
		return fmt.Errorf("failed to add elements: %w", err)
	}

	fmt.Printf("Dataset created with ID: %s\n", dataset.ID)
	return nil
}
