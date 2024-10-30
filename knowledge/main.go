package main

import (
	"fmt"
	"os"

	acmd "github.com/acorn-io/cmd"
	"github.com/gptscript-ai/knowledge/pkg/cmd"
)

func main() {
	if os.Getenv("GPTSCRIPT_GATEWAY_API_KEY") != "" {
		if err := os.Setenv("OPENAI_API_KEY", os.Getenv("GPTSCRIPT_GATEWAY_API_KEY")); err != nil {
			panic(fmt.Errorf("failed to set OPENAI_API_KEY: %w", err))
		}

		gatewayURL := os.Getenv("GPTSCRIPT_GATEWAY_URL")
		if gatewayURL == "" {
			gatewayURL = "https://gateway-api.gptscript.ai"
		}

		if err := os.Setenv("OPENAI_BASE_URL", gatewayURL+"/llm"); err != nil {
			panic(fmt.Errorf("failed to set OPENAI_BASE_URL: %w", err))
		}
	}

	acmd.Main(cmd.New())
}
