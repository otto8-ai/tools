package commands

import (
	"context"
	"fmt"

	"github.com/gptscript-ai/tools/apis/excel/code/pkg/graph"
)

func CreateWorkbook(ctx context.Context, name string) error {
	if name == "" {
		return fmt.Errorf("name cannot be empty")
	}

	id, err := graph.CreateWorkbook(ctx, name)
	if err != nil {
		return err
	}

	fmt.Printf("Workbook created with ID: %s\n", id)
	return nil
}
