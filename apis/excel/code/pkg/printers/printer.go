package printers

import (
	"fmt"

	"github.com/gptscript-ai/tools/apis/excel/code/pkg/graph"
)

func PrintWorkbookInfos(infos []graph.WorkbookInfo) {
	for _, info := range infos {
		fmt.Printf("Name: %s\n", info.Name)
		fmt.Printf("  ID: %s\n", info.ID)
	}
}

func PrintWorksheetInfos(infos []graph.WorksheetInfo) {
	for _, info := range infos {
		fmt.Printf("Name: %s\n", info.Name)
		fmt.Printf("  ID: %s\n", info.ID)
		fmt.Printf("  Workbook ID: %s\n", info.WorkbookID)
	}
}
