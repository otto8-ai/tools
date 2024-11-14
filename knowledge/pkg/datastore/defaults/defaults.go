package defaults

import (
	"github.com/gptscript-ai/knowledge/pkg/env"
)

const (
	TopK int = 10

	TokenModel         = "gpt-4"
	TokenEncoding      = "cl100k_base"
	ChunkSizeTokens    = 2048
	ChunkOverlapTokens = 256
)

var ModelAPIRequestTimeoutSeconds = env.GetIntFromEnvOrDefault("KNOW_MODEL_API_REQUEST_TIMEOUT_SECONDS", 120)
