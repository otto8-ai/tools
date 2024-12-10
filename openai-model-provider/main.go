package main

import (
	"fmt"
	"os"

	"github.com/otto8-ai/tools/openai-model-provider/server"
)

func main() {
	apiKey := os.Getenv("OTTO8_OPENAI_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OTTO8_OPENAI_MODEL_PROVIDER_API_KEY environment variable not set")
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
