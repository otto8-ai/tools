package commands

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/gptscript-ai/tools/excel/pkg/client"
	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/graph"
	"os"
	"strings"
)

func QueryWorksheetData(ctx context.Context, workbookID, worksheetID, query, showColumns string) error {
	c, err := client.NewClient(global.ReadOnlyScopes)
	if err != nil {
		return err
	}

	data, _, err := graph.GetWorksheetData(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}

	csvData, err := convertToCSV(data)
	if err != nil {
		return err
	}

	output, err := QueryWithPandas(ctx, csvData, query, showColumns)
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func convertToCSV(data [][]any) (string, error) {
	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	for _, row := range data {
		strRow := make([]string, len(row))
		for i, val := range row {
			strRow[i] = fmt.Sprintf("%v", val)
		}
		if err := writer.Write(strRow); err != nil {
			return "", fmt.Errorf("error writing row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", fmt.Errorf("error flushing writer: %w", err)
	}
	return builder.String(), nil
}

func QueryWithPandas(ctx context.Context, inputData, inputQuery, showColumns string) (string, error) {
	g, err := gptscript.NewGPTScript()
	if err != nil {
		return "", fmt.Errorf("failed to create GPTScript client: %w", err)
	}
	defer g.Close()
	toolDir := os.Getenv("GPTSCRIPT_TOOL_DIR")
	tooldef := gptscript.ToolDef{
		Name:         "Query Data",
		Description:  "Uses Pandas to perform SQL-like queries on data in-memory",
		Instructions: fmt.Sprintf("#!/usr/bin/env python3 %s/querySpreadsheet.py", toolDir),
		MetaData: map[string]string{
			"requirements.txt": "pandas",
		},
		Arguments: &openapi3.Schema{
			Type: nil,
			Properties: openapi3.Schemas{
				"data": {
					Value: &openapi3.Schema{
						Type:        &openapi3.Types{"string"},
						Description: "A stringified representation of CSV data",
					},
				},
				"show_columns": {
					Value: &openapi3.Schema{
						Type:        &openapi3.Types{"string"},
						Description: "A comma-separated list of the column names that will be displayed in the output",
					},
				},
				"query": {
					Value: &openapi3.Schema{
						Type:        &openapi3.Types{"string"},
						Description: "A SQL-like query to run against the imported dataset. Should be the format expected by the pandas query function (e.g. \"column1 == 'value1' and column2 > 10\")",
					},
				},
			},
		},
	}

	inputJson, err := json.Marshal(map[string]string{
		"data":         inputData,
		"query":        inputQuery,
		"show_columns": showColumns,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal input: %w", err)
	}
	inputString := string(inputJson)
	run, err := g.Evaluate(ctx, gptscript.Options{
		Input: inputString,
	}, tooldef)
	if err != nil {
		return "", err
	}

	output, err := run.Text()
	if err != nil {
		return "", err
	}
	return output, nil
}
