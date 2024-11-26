package main

import (
	"context"
	"fmt"
	"github.com/gptscript-ai/tools/excel/pkg/commands"
	"os"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: gptscript-go-tool <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	var err error
	switch command {
	case "listWorkbooks":
		err = commands.ListWorkbooks(context.Background())
	case "listWorksheets":
		err = commands.ListWorksheets(context.Background(), os.Getenv("WORKBOOK_ID"))
	case "getWorksheetData":
		err = commands.GetWorksheetData(context.Background(), os.Getenv("WORKBOOK_ID"), os.Getenv("WORKSHEET_ID"))
	case "getWorksheetColumnHeaders":
		err = commands.GetWorksheetColumnHeaders(context.Background(), os.Getenv("WORKBOOK_ID"), os.Getenv("WORKSHEET_ID"))
	case "getWorksheetTables":
		err = commands.GetWorksheetTables(context.Background(), os.Getenv("WORKBOOK_ID"), os.Getenv("WORKSHEET_ID"))
	case "queryWorksheetData":
		err = commands.QueryWorksheetData(context.Background(), os.Getenv("WORKBOOK_ID"), os.Getenv("WORKSHEET_ID"), os.Getenv("QUERY"), os.Getenv("SHOW_COLUMNS"))
	case "addWorksheetRow":
		err = commands.AddWorksheetRow(context.Background(), os.Getenv("WORKBOOK_ID"), os.Getenv("WORKSHEET_ID"), os.Getenv("CONTENTS"))
	case "createWorksheet":
		err = commands.CreateWorksheet(context.Background(), os.Getenv("WORKBOOK_ID"), os.Getenv("NAME"))
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
