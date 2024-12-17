package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/openai-model-provider/server"
)

func main() {
	apiKey := os.Getenv("OBOT_OPENAI_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_OPENAI_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	if err := server.Run(apiKey, port); err != nil {
		panic(err)
	}
}
