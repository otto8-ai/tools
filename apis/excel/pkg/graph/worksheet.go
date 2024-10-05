package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gptscript-ai/tools/apis/excel/code/pkg/global"
	"github.com/gptscript-ai/tools/apis/excel/code/pkg/util"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

type WorkbookInfo struct {
	ID, Name string
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

func CreateWorkbook(ctx context.Context, name string) (string, error) {
	if !strings.HasSuffix(name, ".xlsx") {
		name += ".xlsx"
	}

	// There didn't seem to be a usable method in the SDK for this, so once again we need to make a raw HTTP request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/root:/%s:/content", name), strings.NewReader(""))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer"+os.Getenv(global.CredentialEnv))
	req.Header.Set("Content-Type", "text/plain") // We're not actually sending any content, so the content-type doesn't really matter.

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ID, nil
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
