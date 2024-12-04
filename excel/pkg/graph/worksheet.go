package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
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

type UpdateBody struct {
	Values   [][]any `json:"values,omitempty"`
	Formulas [][]any `json:"formulas,omitempty"`
}

func (u *UpdateBody) AppendColumnCellToValues(newColumnCell []any) {
	u.Values = append(u.Values, newColumnCell)
	u.Formulas = append(u.Formulas, []any{nil})
}

func (u *UpdateBody) AppendColumnCellToFormulas(newColumnCell []any) {
	u.Formulas = append(u.Formulas, newColumnCell)
	u.Values = append(u.Values, []any{nil})
}

func (u *UpdateBody) AppendRowToValues(newRow []any) {
	u.Values = append(u.Values, newRow)
}

func (u *UpdateBody) AppendRowToFormulas(newRow []any) {
	u.Formulas = append(u.Formulas, newRow)
}

func numberToColumnLetter(n int) string {
	if n <= 0 {
		return ""
	}

	column := ""
	for n > 0 {
		n--
		column = string(rune('A'+(n%26))) + column
		n /= 26
	}

	return column
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

func GetWorksheetColumnHeaders(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string) ([][]any, models.WorkbookRangeable, error) {
	drive, err := c.Me().Drive().Get(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	address := "1:3"
	usedRange, err := c.Drives().ByDriveId(util.Deref(drive.GetId())).Items().ByDriveItemId(workbookID).Workbook().Worksheets().ByWorkbookWorksheetId(worksheetID).RangeWithAddress(&address).UsedRange().Get(ctx, nil)
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

func AddWorksheetRow(ctx context.Context, c *msgraphsdkgo.GraphServiceClient, workbookID, worksheetID string, contents []string) error {
	_, usedRange, err := GetWorksheetData(ctx, c, workbookID, worksheetID)
	if err != nil {
		return err
	}
	rowCount := util.Deref(usedRange.GetRowCount())
	newRowNumber := int(rowCount + 1)

	endColumnLetter := numberToColumnLetter(len(contents))
	address := fmt.Sprintf("A%d:%s%d", newRowNumber, endColumnLetter, newRowNumber)

	// Update the worksheet.
	// Unfortunately, the SDK lacks a function to do what we need to do, so we need to make a raw HTTP request.
	var values, formulas []any
	for _, v := range contents {
		if strings.HasPrefix(v, "=") {
			formulas = append(formulas, v)
			values = append(values, nil)
		} else {
			values = append(values, v)
			formulas = append(formulas, nil)
		}
	}
	body := &UpdateBody{
		Values:   [][]any{values},
		Formulas: [][]any{formulas},
	}
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/range(address='%s')", workbookID, worksheetID, address), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv(global.CredentialEnv))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Println(string(body))
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func AddWorksheetColumn(ctx context.Context, workbookID, worksheetID, columnID string, contents []string) error {
	re := regexp.MustCompile(`\d+$`)
	startRow, err := strconv.Atoi(re.FindString(columnID))
	if err != nil {
		return fmt.Errorf("failed to parse starting row.")
	}
	endRow := startRow + len(contents) - 1

	endColumnID := re.ReplaceAllString(columnID, fmt.Sprintf("%d", endRow))
	address := fmt.Sprintf("%s:%s", columnID, endColumnID)

	body := new(UpdateBody)
	for _, v := range contents {
		if strings.HasPrefix(v, "=") {
			body.AppendColumnCellToFormulas([]any{v})
		} else {
			body.AppendColumnCellToValues([]any{v})
		}
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, fmt.Sprintf("https://graph.microsoft.com/v1.0/me/drive/items/%s/workbook/worksheets/%s/range(address='%s')", workbookID, worksheetID, address), strings.NewReader(string(bodyJSON)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv(global.CredentialEnv))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Printf("failed to close response body: %s", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err == nil {
			fmt.Println(string(body))
		}
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
