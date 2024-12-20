package main

import (
	"fmt"
	"os"

	"github.com/obot-platform/tools/vllm-model-provider/server"
)

func main() {
	apiKey := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_API_KEY")
	if apiKey == "" {
		fmt.Println("OBOT_VLLM_MODEL_PROVIDER_API_KEY environment variable not set")
		os.Exit(1)
	}

	endpoint := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_ENDPOINT")
	if endpoint == "" {
		fmt.Println("OBOT_VLLM_MODEL_PROVIDER_ENDPOINT environment variable not set")
		os.Exit(1)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	if err := server.Run(apiKey, endpoint, port); err != nil {
		panic(err)
	}
}
