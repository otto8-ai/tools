package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"

	"github.com/obot-platform/obot/apiclient"
	"github.com/obot-platform/obot/apiclient/types"
)

var (
	url      = os.Getenv("OBOT_SERVER_URL")
	token    = os.Getenv("OBOT_TOKEN")
	id       = os.Getenv("ID")
	threadID = os.Getenv("OBOT_THREAD_ID")
	args     = os.Getenv("ARGS")
)

func main() {
	ctx := context.Background()
	if err := mainErr(ctx); err != nil {
		slog.Error("error", "err", err)
	}
}

type workflowInfo struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Parameters  []Param `json:"params,omitempty"`
}

type Param struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type runInfo struct {
	ID        string      `json:"id"`
	TaskID    string      `json:"taskID"`
	StartTime types.Time  `json:"startTime"`
	EndTime   *types.Time `json:"endTime,omitempty"`
	Input     string      `json:"input"`
	Error     string      `json:"error,omitempty"`
}

func list(ctx context.Context, c *apiclient.Client) error {
	result, err := c.ListWorkflows(ctx, apiclient.ListWorkflowsOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return fmt.Errorf("list tasks: %v", err)
	}

	var output []workflowInfo
	for _, workflow := range result.Items {
		info := workflowInfo{
			ID:          workflow.ID,
			Name:        workflow.Name,
			Description: workflow.Description,
		}
		for _, k := range slices.Sorted(maps.Keys(workflow.Params)) {
			info.Parameters = append(info.Parameters, Param{
				Name:        k,
				Description: workflow.Params[k],
			})
		}
		output = append(output, info)
	}

	if len(result.Items) == 0 {
		fmt.Printf("no tasks found\n")
		return nil
	}

	return json.NewEncoder(os.Stdout).Encode(output)
}

func runs(ctx context.Context, c *apiclient.Client, workflowID string) error {
	if workflowID == "" {
		return fmt.Errorf("missing task ID")
	}
	result, err := c.ListWorkflowExecutions(ctx, workflowID, apiclient.ListWorkflowExecutionsOptions{
		ThreadID: threadID,
	})
	if err != nil {
		return fmt.Errorf("list runs: %v", err)
	}

	if len(result.Items) == 0 {
		fmt.Printf("no runs found\n")
		return nil
	}

	var tasks []runInfo
	for _, task := range result.Items {
		tasks = append(tasks, runInfo{
			ID:        task.ID,
			TaskID:    workflowID,
			StartTime: task.StartTime,
			EndTime:   task.EndTime,
			Input:     task.Input,
			Error:     task.Error,
		})
	}

	return json.NewEncoder(os.Stdout).Encode(tasks)
}

func run(ctx context.Context, c *apiclient.Client) error {
	if id == "" {
		return fmt.Errorf("missing ID")
	}

	resp, err := c.Invoke(ctx, id, args, apiclient.InvokeOptions{
		Async: true,
	})
	if err != nil {
		return err
	}

	fmt.Printf("task started: %s\n", resp.ThreadID)
	return nil
}

func mainErr(ctx context.Context) error {
	if len(os.Args) == 1 {
		fmt.Printf("incorrect usage: %s [list|run]\n", os.Args[0])
		return nil
	}

	if url == "" {
		url = "http://localhost:8080/api"
	} else {
		url += "/api"
	}

	client := &apiclient.Client{
		BaseURL: url,
		Token:   token,
	}

	switch os.Args[1] {
	case "list":
		return list(ctx, client)
	case "run":
		return run(ctx, client)
	case "list-runs":
		return runs(ctx, client, id)
	}

	return nil
}
