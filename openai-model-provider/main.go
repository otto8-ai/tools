package main

import (
	"fmt"
	"os"

	"github.com/acorn-io/tools/openai-model-provider/server"
)

func main() {
	apiKey := os.Getenv("ACORN_OPENAI_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("ACORN_OPENAI_MODEL_PROVIDER_API_KEY environment variable not set")
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
