package graph

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gptscript-ai/tools/excel/pkg/global"
	"github.com/gptscript-ai/tools/excel/pkg/util"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

type WorkbookInfo struct {
	ID, Name string
}

type Table struct {
	ID, Name string
}

type HTTPErrorBody struct {
	Error struct {
		Code       string `json:"code,omitempty"`
		Message    string `json:"message,omitempty"`
		InnerError struct {
			Code            string `json:"code,omitempty"`
			Message         string `json:"message,omitempty"`
			Date            string `json:"date,omitempty"`
			RequestID       string `json:"request-id,omitempty"`
			ClientRequestID string `json:"client-request-id,omitempty"`
		} `json:"innerError,omitempty"`
	} `json:"error,omitempty"`
}

func ListWorkbooks(ctx context.Context, c *msgraphsdkgo.GraphServiceClient) ([]WorkbookInfo, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var infos []WorkbookInfo
	workbooks, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).SearchWithQ(util.Ptr("xlsx")).GetAsSearchWithQGetResponse(ctx, nil)
	if err != nil {
		return nil, err
	}
	for _, workbook := range workbooks.GetValue() {
		infos = append(infos, WorkbookInfo{
			ID:   util.Deref(workbook.GetId()),
			Name: util.Deref(workbook.GetName()),
		})
	}
	return infos, nil
}

func CreateWorksheet(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, name string) (string, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return "", err
	}

	requestBody := drives.NewItemItemsItemWorkbookWorksheetsAddPostRequestBody()
	requestBody.SetName(util.Ptr(name))
	worksheet, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().Add().Post(ctx, requestBody, nil)
	if err != nil {
		return "", err
	}
	return util.Deref(worksheet.GetId()), nil
}

type WorksheetInfo struct {
	ID, Name, WorkbookID string
}

func ListWorksheetsInWorkbook(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID string) ([]WorksheetInfo, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	sheets, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().Get(ctx, nil)
	if err != nil {
		return nil, err
	}

	var infos []WorksheetInfo
	for _, sheet := range sheets.GetValue() {
		infos = append(infos, WorksheetInfo{
			ID:         util.Deref(sheet.GetId()),
			Name:       util.Deref(sheet.GetName()),
			WorkbookID: workbookID,
		})
	}
	return infos, nil
}

func GetWorksheetData(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string) ([][]any, models.WorkbookRangeable, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	usedRange, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).UsedRange().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	result, err := serialization.SerializeToJson(usedRange.GetValues())
	if err != nil {
		return nil, nil, err
	}

	var data [][]any
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return data, usedRange, nil
}

func GetWorksheetTables(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string) ([]Table, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	result, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).Tables().Get(ctx, nil)
	if err != nil {
		return nil, err
	}
	var tables []Table
	for _, table := range result.GetValue() {
		tables = append(tables, Table{ID: util.Deref(table.GetId()), Name: util.Deref(table.GetName())})
	}

	return tables, nil
}

func GetColumnIDFromName(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID, tableID, columnName string) (string, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return "", err
	}
	result, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Tables().ByWorkbookTableId(tableID).Columns().Get(ctx, nil)
	if err != nil {
		return "", err
	}
	for _, column := range result.GetValue() {
		rangeColumnName := util.Deref(column.GetName())
		if rangeColumnName == columnName {
			return util.Deref(column.GetId()), nil
		}
	}
	return "", errors.New("column not found")
}

func FilterWorksheetData(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID, tableID, columnID string, filterValues []string) ([][]any, models.WorkbookRangeable, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	criteria := map[string]any{
		"criteria": map[string]any{
			"filterOn": "Values",
			"Values":   filterValues,
		},
	}

	bodyJSON, err := json.Marshal(criteria)
	if err != nil {
		return nil, nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/tables/%s/columns/%s/filter/apply", workbookID, worksheetID, tableID, columnID), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv(global.CredentialEnv))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		var errBody HTTPErrorBody
		if err := json.NewDecoder(resp.Body).Decode(&errBody); err != nil {
			return nil, nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
		}
		return nil, nil, fmt.Errorf("error applying filter: %s", errBody.Error.Message)
	}
	defer func(clear *drives.ItemItemsItemWorkbookWorksheetsItemTablesItemColumnsItemFilterClearRequestBuilder, ctx context.Context, requestConfiguration *drives.ItemItemsItemWorkbookWorksheetsItemTablesItemColumnsItemFilterClearRequestBuilderPostRequestConfiguration) {
		err := clear.Post(ctx, requestConfiguration)
		if err != nil {
			fmt.Printf("failed to clear filter: %s", err)
		}
	}(c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).Tables().ByWorkbookTableId(tableID).Columns().ByWorkbookTableColumnId(columnID).Filter().Clear(), ctx, nil)

	usedRange, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).UsedRange().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	result, err := serialization.SerializeToJson(usedRange.GetValues())
	if err != nil {
		return nil, nil, err
	}

	var data [][]any
	if err = json.Unmarshal(result, &data); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}
	return data, usedRange, nil
}

func AddWorksheetRow(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string, contents []string) error {
	// First, get the existing data.
	data, usedRange, err := GetWorksheetData(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}

	if len(data) == 1 && len(data[0]) == 1 && data[0][0] == "" {
		data = [][]any{}
	}

	// Append the new row.
	var maxRowLength int
	data, maxRowLength = padData(append(data, stringArrayToAnyArray(contents)))

	// Update the worksheet.
	// Unfortunately, the SDK lacks a function to do what we need to do, so we need to make a raw HTTP request.
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}
	bodyStr := fmt.Sprintf("{\"values\": %s}", dataJSON)
	_, address, _ := strings.Cut(util.Deref(usedRange.GetAddress()), "!")
	address, err = updateAddress(address, len(data), maxRowLength)
	if err != nil {
		return fmt.Errorf("failed to update address: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/range(address='%s')", workbookID, worksheetID, address), strings.NewReader(bodyStr))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer"+os.Getenv(global.CredentialEnv))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func stringArrayToAnyArray(arr []string) []any {
	var result []any
	for _, s := range arr {
		result = append(result, s)
	}
	return result
}

// padData pads the data with nil values to ensure all rows have the same length.
func padData(data [][]any) ([][]any, int) {
	// Find the longest row in the data.
	length := 0
	for _, row := range data {
		if len(row) > length {
			length = len(row)
		}
	}

	// Pad all rows to the same length.
	for i := range data {
		for len(data[i]) < length {
			data[i] = append(data[i], nil)
		}
	}

	return data, length
}

// updateAddress takes the current range address (i.e. A1:B3) and updates it to include the new row that will be added.
func updateAddress(address string, numRows, maxRowLength int) (string, error) {
	start, _, _ := strings.Cut(address, ":")
	if len(start) < 2 {
		return "", fmt.Errorf("invalid address: %s", address)
	}

	startLetter, startNum := splitPartialAddress(start)

	startNumInt, err := strconv.Atoi(startNum)
	if err != nil {
		return "", fmt.Errorf("failed to parse end row number: %w", err)
	}
	endNumInt := startNumInt + numRows - 1

	// This can sometimes happen if we are appending to a new spreadsheet.
	if endNumInt < numRows {
		endNumInt = numRows
	}

	// The column needs to be the letter corresponding to the number of values. I.e., C is 3.
	// We start from the column letter from the start of the range.
	endLetter := util.ColumnNumberToLetters(util.ColumnLettersToNumber(startLetter) + maxRowLength - 1)

	return fmt.Sprintf("%s:%s%d", start, endLetter, endNumInt), nil
}

// splitPartialAddress splits a partial address (e.g. A1) into the column letter and row number.
func splitPartialAddress(partial string) (string, string) {
	for i := 0; i < len(partial); i++ {
		if partial[i] >= '0' && partial[i] <= '9' {
			return partial[:i], partial[i:]
		}
	}
	return "", ""
}
