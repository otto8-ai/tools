package defaults

import (
	"github.com/gptscript-ai/knowledge/pkg/env"
)

const (
	TopK int = 10

	TokenModel         = "llm"
	TokenEncoding      = "cl100k_base"
	ChunkSizeTokens    = 2048
	ChunkOverlapTokens = 256
)

var (
	// Default Remote Model API timeout options

	// ModelAPITimeoutSeconds is the total timeout set on the context and is the maximum time we try for, so it may span multiple retries
	ModelAPITimeoutSeconds = env.GetIntFromEnvOrDefault("KNOW_MODEL_API_TIMEOUT_SECONDS", 300)

	// ModelAPIRequestTimeoutSeconds is the timeout for each individual request to the model API
	ModelAPIRequestTimeoutSeconds = env.GetIntFromEnvOrDefault("KNOW_MODEL_API_REQUEST_TIMEOUT_SECONDS", 120)
)
