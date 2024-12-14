package main

import (
	"encoding/json"
	"fmt"
	"os"
)

var inputText = os.Getenv("WORKFLOW_INPUT")

const (
	webhookContext = `This workflow is being called from a webhook. The input is a JSON structure of the webhook payload and any
important headers.`
	emailContext = `This workflow is being called from an email receiver. The input is a JSON structure of the email message body and any
important email headers (from, to, subject, etc).`
)

type workflowInput struct {
	Type string `json:"type"`
}

func main() {
	var structuredInput workflowInput
	if err := json.Unmarshal([]byte(inputText), &structuredInput); err == nil {
		var context string
		switch structuredInput.Type {
		case "email":
			context = emailContext
		case "webhook":
			context = webhookContext
		}
		if context != "" {
			fmt.Printf("START WORKFLOW CONTEXT:\n%s\nEND START WORKFLOW CONTEXT\n\n", context)
		}
	}

	fmt.Printf("START WORKFLOW INPUT:\n%s\nEND WORKFLOW INPUT\n\n", inputText)
}
