package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gptscript-ai/tools/outlook/mail/pkg/commands"
	"github.com/gptscript-ai/tools/outlook/mail/pkg/graph"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: mail <command>")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "listMailFolders":
		if err := commands.ListMailFolders(context.Background()); err != nil {
			fmt.Printf("failed to list mail folders: %v\n", err)
			os.Exit(1)
		}
	case "listMessages":
		if err := commands.ListMessages(context.Background(), os.Getenv("FOLDER_ID")); err != nil {
			fmt.Printf("failed to list mail: %v\n", err)
			os.Exit(1)
		}
	case "getMessageDetails":
		if err := commands.GetMessageDetails(context.Background(), os.Getenv("MESSAGE_ID")); err != nil {
			fmt.Printf("failed to get message details: %v\n", err)
			os.Exit(1)
		}
	case "searchMessages":
		if err := commands.SearchMessages(
			context.Background(),
			os.Getenv("SUBJECT"),
			os.Getenv("FROM_ADDRESS"),
			os.Getenv("FROM_NAME"),
			os.Getenv("FOLDER_ID"),
			os.Getenv("START"),
			os.Getenv("END"),
			os.Getenv("LIMIT"),
		); err != nil {
			fmt.Printf("failed to search messages: %v\n", err)
			os.Exit(1)
		}
	case "createDraft":
		if err := commands.CreateDraft(context.Background(), getDraftInfoFromEnv()); err != nil {
			fmt.Printf("failed to create draft: %v\n", err)
			os.Exit(1)
		}
	case "sendDraft":
		if err := commands.SendDraft(context.Background(), os.Getenv("DRAFT_ID")); err != nil {
			fmt.Printf("failed to send draft: %v\n", err)
			os.Exit(1)
		}
	case "sendMessage":
		if err := commands.SendMessage(context.Background(), getDraftInfoFromEnv()); err != nil {
			fmt.Printf("failed to send message: %v\n", err)
			os.Exit(1)
		}
	case "deleteMessage":
		if err := commands.DeleteMessage(context.Background(), os.Getenv("MESSAGE_ID")); err != nil {
			fmt.Printf("failed to delete message: %v\n", err)
			os.Exit(1)
		}
	case "moveMessage":
		if err := commands.MoveMessage(context.Background(), os.Getenv("MESSAGE_ID"), os.Getenv("DESTINATION_FOLDER_ID")); err != nil {
			fmt.Printf("failed to move message: %v\n", err)
			os.Exit(1)
		}
	case "getDefaultTimezone":
		if err := commands.GetDefaultTimezone(context.Background()); err != nil {
			fmt.Printf("failed to get default timezone: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func getDraftInfoFromEnv() graph.DraftInfo {
	return graph.DraftInfo{
		Subject:    os.Getenv("SUBJECT"),
		Body:       os.Getenv("BODY"),
		Recipients: strings.Split(os.Getenv("RECIPIENTS"), ","),
		CC:         strings.Split(os.Getenv("CC"), ","),
		BCC:        strings.Split(os.Getenv("BCC"), ","),
	}
}
