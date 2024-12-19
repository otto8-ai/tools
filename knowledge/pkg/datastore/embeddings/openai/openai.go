package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"dario.cat/mergo"
	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	"github.com/gptscript-ai/knowledge/pkg/env"
	"github.com/gptscript-ai/knowledge/pkg/log"
	cg "github.com/philippgille/chromem-go"
)

var OpenAIEmbeddingAPITimeout = time.Duration(env.GetIntFromEnvOrDefault("KNOW_OPENAI_EMBEDDING_API_TIMEOUT_SECONDS", defaults.ModelAPIRequestTimeoutSeconds)) * time.Second
var OpenAIEmbeddingAPIRequestTimeout = time.Duration(env.GetIntFromEnvOrDefault("KNOW_OPENAI_EMBEDDING_API_REQUEST_TIMEOUT_SECONDS", 120)) * time.Second

const EmbeddingModelProviderOpenAIName string = "openai"

type EmbeddingModelProviderOpenAI struct {
	BaseURL           string            `usage:"OpenAI API base" default:"https://api.openai.com/v1" env:"OPENAI_BASE_URL" koanf:"baseURL"`
	APIKey            string            `usage:"OpenAI API key (not required if used with clicky-chats)" default:"sk-foo" env:"OPENAI_API_KEY" koanf:"apiKey" mapstructure:"apiKey" export:"false"`
	Model             string            `usage:"OpenAI model" default:"gpt-4o" env:"OPENAI_MODEL" koanf:"openai-model"`
	EmbeddingModel    string            `usage:"OpenAI Embedding model" default:"text-embedding-3-large" env:"OPENAI_EMBEDDING_MODEL" koanf:"embeddingModel" export:"required"`
	EmbeddingEndpoint string            `usage:"OpenAI Embedding endpoint" default:"/embeddings" env:"OPENAI_EMBEDDING_ENDPOINT" koanf:"embeddingEndpoint"`
	APIVersion        string            `usage:"OpenAI API version (for Azure)" default:"2024-02-01" env:"OPENAI_API_VERSION" koanf:"apiVersion"`
	APIType           string            `usage:"OpenAI API type (OPEN_AI, AZURE, AZURE_AD, ...)" default:"OPEN_AI" env:"OPENAI_API_TYPE" koanf:"apiType"`
	AzureOpenAIConfig AzureOpenAIConfig `koanf:"azure"`
}

type OpenAIConfig struct {
	BaseURL           string            `usage:"OpenAI API base" default:"https://api.openai.com/v1" env:"OPENAI_BASE_URL" koanf:"baseURL"`
	APIKey            string            `usage:"OpenAI API key (not required if used with clicky-chats)" default:"sk-foo" env:"OPENAI_API_KEY" koanf:"apiKey" mapstructure:"apiKey" export:"false"`
	Model             string            `usage:"OpenAI model" default:"gpt-4o" env:"OPENAI_MODEL" koanf:"openai-model"`
	EmbeddingModel    string            `usage:"OpenAI Embedding model" default:"text-embedding-3-large" env:"OPENAI_EMBEDDING_MODEL" koanf:"embeddingModel" export:"required"`
	EmbeddingEndpoint string            `usage:"OpenAI Embedding endpoint" default:"/embeddings" env:"OPENAI_EMBEDDING_ENDPOINT" koanf:"embeddingEndpoint"`
	APIVersion        string            `usage:"OpenAI API version (for Azure)" default:"2024-02-01" env:"OPENAI_API_VERSION" koanf:"apiVersion"`
	APIType           string            `usage:"OpenAI API type (OPEN_AI, AZURE, AZURE_AD, ...)" default:"OPEN_AI" env:"OPENAI_API_TYPE" koanf:"apiType"`
	AzureOpenAIConfig AzureOpenAIConfig `koanf:"azure"`
}

type OpenAIEmbeddingRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format,omitempty"`
	Dimensions     *int   `json:"dimensions,omitempty"`
}

func (o OpenAIConfig) Name() string {
	return EmbeddingModelProviderOpenAIName
}

func (o OpenAIConfig) EmbeddingModelName() string {
	return o.EmbeddingModel
}

type AzureOpenAIConfig struct {
	Deployment string `usage:"Azure OpenAI deployment name (overrides openai-embedding-model, if set)" default:"" env:"OPENAI_AZURE_DEPLOYMENT" koanf:"deployment"`
}

func (p *EmbeddingModelProviderOpenAI) UseEmbeddingModel(model string) {
	p.EmbeddingModel = model
}

func (p *EmbeddingModelProviderOpenAI) EmbeddingModelName() string {
	return p.EmbeddingModel
}

func (p *EmbeddingModelProviderOpenAI) Name() string {
	return EmbeddingModelProviderOpenAIName
}

func (p *EmbeddingModelProviderOpenAI) Configure() error {
	if err := load.FillConfigEnv("OPENAI_", &p); err != nil {
		return fmt.Errorf("failed to fill OpenAI config from environment: %w", err)
	}

	if err := p.fillDefaults(); err != nil {
		return fmt.Errorf("failed to fill OpenAI defaults: %w", err)
	}

	return nil
}

func (p *EmbeddingModelProviderOpenAI) fillDefaults() error {
	defaultAzureOpenAIConfig := AzureOpenAIConfig{
		Deployment: "",
	}

	defaultConfig := EmbeddingModelProviderOpenAI{
		BaseURL:           "https://api.openai.com/v1",
		APIKey:            "sk-foo",
		Model:             "gpt-4o",
		EmbeddingModel:    "text-embedding-3-large",
		EmbeddingEndpoint: "/embeddings",
		APIVersion:        "2024-02-01",
		APIType:           "OPEN_AI",
		AzureOpenAIConfig: defaultAzureOpenAIConfig,
	}

	err := mergo.Merge(p, defaultConfig)
	if err != nil {
		return fmt.Errorf("failed to merge OpenAI config: %w", err)
	}

	return nil
}

func (p *EmbeddingModelProviderOpenAI) EmbeddingFunc() (cg.EmbeddingFunc, error) {
	var embeddingFunc cg.EmbeddingFunc

	switch strings.ToLower(p.APIType) {
	// except for Azure, most other OpenAI API compatible providers only differ in the normalization of output vectors (apart from the obvious API endpoint, etc.)
	case "azure", "azure_ad":
		// TODO: clean this up to support inputting the full deployment URL
		deployment := p.AzureOpenAIConfig.Deployment
		if deployment == "" {
			deployment = p.EmbeddingModel
		}

		deploymentURL, err := url.Parse(p.BaseURL)
		if err != nil || deploymentURL == nil {
			return nil, fmt.Errorf("failed to parse OpenAI Base URL %q: %w", p.BaseURL, err)
		}

		deploymentURL = deploymentURL.JoinPath("openai", "deployments", deployment)

		slog.Debug("Using Azure OpenAI API", "deploymentURL", deploymentURL.String(), "APIVersion", p.APIVersion)

		embeddingFunc = NewEmbeddingFuncAzureOpenAI(
			p.APIKey,
			deploymentURL.String(),
			p.APIVersion,
			"",
		)
	case "open_ai":
		cfg := NewOpenAICompatConfig(
			p.BaseURL,
			p.APIKey,
			p.EmbeddingModel,
		).
			WithNormalized(true).
			WithEmbeddingsEndpoint(p.EmbeddingEndpoint)
		embeddingFunc = NewEmbeddingFuncOpenAICompat(cfg)
	default:
		return nil, fmt.Errorf("unknown OpenAI API type: %q", p.APIType)
	}

	return embeddingFunc, nil
}

func (p *EmbeddingModelProviderOpenAI) Config() any {
	return p
}

/*
 * NOTICE: The following was copied over from github.com/philippgille/chromem-go to lessen the changes to our fork at github.com/iwilltry42/chromem-go
 */

// NewEmbeddingFuncOpenAICompat returns a function that creates embeddings for a text
// using an OpenAI compatible API. For example:
//   - Azure OpenAI: https://azure.microsoft.com/en-us/products/ai-services/openai-service
//   - LitLLM: https://github.com/BerriAI/litellm
//   - Ollama: https://github.com/ollama/ollama/blob/main/docs/openai.md
//   - etc.
//
// It offers options to set request headers and query parameters
// e.g. to pass the `api-key` header and the `api-version` query parameter for Azure OpenAI.
//
// The `normalized` parameter indicates whether the vectors returned by the embedding
// model are already normalized, as is the case for OpenAI's and Mistral's models.
// The flag is optional. If it's nil, it will be autodetected on the first request
// (which bears a small risk that the vector just happens to have a length of 1).
func NewEmbeddingFuncOpenAICompat(config *OpenAICompatConfig) cg.EmbeddingFunc {
	if config == nil {
		panic("config must not be nil")
	}

	client := &http.Client{
		Timeout: OpenAIEmbeddingAPIRequestTimeout, // per request timeout - the overall timeout is set on the context
	}

	var checkedNormalized bool
	checkNormalized := sync.Once{}

	return func(ctx context.Context, text string) ([]float32, error) {
		// Create the OpenAI request payload
		embedReq := OpenAIEmbeddingRequest{
			Input:          text,
			Model:          config.model,
			EncodingFormat: "float",
		}

		// Only set dimensions for text-embedding-3-large
		if config.model == "text-embedding-3-large" {
			dims := 2000
			embedReq.Dimensions = &dims
		}

		// Prepare the request body
		reqBody, err := json.Marshal(embedReq)
		if err != nil {
			return nil, fmt.Errorf("couldn't marshal request body: %w", err)
		}

		fullURL, err := url.JoinPath(config.baseURL, config.embeddingsEndpoint)
		if err != nil {
			return nil, fmt.Errorf("couldn't join base URL and endpoint: %w", err)
		}

		// Create the request. Creating it with context is important for a timeout
		// to be possible, because the client is configured without a timeout.
		req, err := http.NewRequestWithContext(ctx, "POST", fullURL, bytes.NewBuffer(reqBody))
		if err != nil {
			return nil, fmt.Errorf("couldn't create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+config.apiKey)

		// Add headers
		for k, v := range config.headers {
			req.Header.Add(k, v)
		}

		// Add query parameters
		q := req.URL.Query()
		for k, v := range config.queryParams {
			q.Add(k, v)
		}
		req.URL.RawQuery = q.Encode()

		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, OpenAIEmbeddingAPITimeout)
		defer cancel()

		// Send the request and get the body.
		body, err := RequestWithExponentialBackoff(ctx, client, req, 5, true)
		if err != nil {
			return nil, fmt.Errorf("error sending request(s): %w", err)
		}

		var embeddingResponse cg.OpenAIResponse
		err = json.Unmarshal(body, &embeddingResponse)
		if err != nil {
			return nil, fmt.Errorf("couldn't unmarshal response body: %w", err)
		}

		// Check if the response contains embeddings.
		if len(embeddingResponse.Data) == 0 || len(embeddingResponse.Data[0].Embedding) == 0 {
			return nil, errors.New("no embeddings found in the response")
		}

		v := embeddingResponse.Data[0].Embedding
		if config.normalized != nil {
			if *config.normalized {
				return v, nil
			}
			return cg.NormalizeVector(v), nil
		}
		checkNormalized.Do(func() {
			if cg.IsNormalized(v) {
				checkedNormalized = true
			} else {
				checkedNormalized = false
			}
		})
		if !checkedNormalized {
			v = cg.NormalizeVector(v)
		}

		return v, nil
	}
}

func RequestWithExponentialBackoff(ctx context.Context, client *http.Client, req *http.Request, maxRetries int, handleRateLimit bool) ([]byte, error) {
	const baseDelay = time.Millisecond * 200
	var resp *http.Response
	var err error

	logger := log.FromCtx(ctx)

	var failures []string

	// Save the original request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
	}

	for i := 0; i < maxRetries; i++ {
		// Check if context was canceled (timeout) before retrying
		if ctx.Err() != nil {
			failures = append(failures, fmt.Sprintf("[!] Stopped by canceled context after try #%d/%d: %v", i, maxRetries, ctx.Err()))
			break
		}

		// Reset body to the original request body
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// Log the error and retry for transient error reading response body
				msg := fmt.Sprintf("#%d/%d: failed to read response body: %v", i+1, maxRetries, err)
				logger.Warn("Request failed - Retryable", "error", msg)
				failures = append(failures, msg)
				_ = resp.Body.Close()
				continue
			}

			return body, resp.Body.Close()
		}

		if resp != nil {
			var bodystr string
			if resp.Body != nil {
				body, rerr := io.ReadAll(resp.Body)
				if rerr == nil {
					bodystr = string(body)
				}
				_ = resp.Body.Close()
			}

			msg := fmt.Sprintf("#%d/%d: %d <%s> (err: %v)", i+1, maxRetries, resp.StatusCode, bodystr, err)
			failures = append(failures, msg)

			if resp.StatusCode >= 500 || (handleRateLimit && resp.StatusCode == http.StatusTooManyRequests) {
				logger.Warn("Request failed - Retryable", "error", msg)
				// Retry for 5xx (Server Errors)
				// We're also handling rate limit here (without checking the Retry-After header), if handleRateLimit is true,
				// since it's what e.g. OpenAI recommends (see https://github.com/openai/openai-cookbook/blob/457f4310700f93e7018b1822213ca99c613dbd1b/examples/How_to_handle_rate_limits.ipynb).
				delay := baseDelay * time.Duration(1<<i)
				jitter := time.Duration(rand.Int63n(int64(baseDelay)))
				time.Sleep(delay + jitter)
				continue
			} else {
				// Non-retryable error
				logger.Error("Request failed - Non-retryable", "error", msg)
				break
			}
		} else {
			// Log connection errors (client.Do error) and retry if needed
			msg := fmt.Sprintf("#%d/%d: failed to send request: %v", i+1, maxRetries, err)
			logger.Warn("Request failed - Retryable", "error", msg)
			failures = append(failures, msg)
		}
	}

	logger.Error("request retry limit exceeded or failed with non-retryable error(s)", "request", req, "maxTries", maxRetries, "failures", strings.Join(failures, ";"))

	return nil, fmt.Errorf("retry limit exceeded or request failed with non-retryable error: %v", failures[len(failures)-1])
}

type OpenAICompatConfig struct {
	baseURL string
	apiKey  string
	model   string

	// Optional
	normalized         *bool
	embeddingsEndpoint string
	headers            map[string]string
	queryParams        map[string]string
}

func NewOpenAICompatConfig(baseURL, apiKey, model string) *OpenAICompatConfig {
	return &OpenAICompatConfig{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,

		embeddingsEndpoint: "/embeddings",
	}
}

func (c *OpenAICompatConfig) WithEmbeddingsEndpoint(endpoint string) *OpenAICompatConfig {
	c.embeddingsEndpoint = endpoint
	return c
}

func (c *OpenAICompatConfig) WithHeaders(headers map[string]string) *OpenAICompatConfig {
	c.headers = headers
	return c
}

func (c *OpenAICompatConfig) WithQueryParams(queryParams map[string]string) *OpenAICompatConfig {
	c.queryParams = queryParams
	return c
}

func (c *OpenAICompatConfig) WithNormalized(normalized bool) *OpenAICompatConfig {
	c.normalized = &normalized
	return c
}

const (
	azureDefaultAPIVersion = "2024-02-01"
)

// NewEmbeddingFuncAzureOpenAI returns a function that creates embeddings for a text
// using the Azure OpenAI API.
// The `deploymentURL` is the URL of the deployed model, e.g. "https://YOUR_RESOURCE_NAME.openai.azure.com/openai/deployments/YOUR_DEPLOYMENT_NAME"
// See https://learn.microsoft.com/en-us/azure/ai-services/openai/how-to/embeddings?tabs=console#how-to-get-embeddings
func NewEmbeddingFuncAzureOpenAI(apiKey string, deploymentURL string, apiVersion string, model string) cg.EmbeddingFunc {
	if apiVersion == "" {
		apiVersion = azureDefaultAPIVersion
	}
	return NewEmbeddingFuncOpenAICompat(NewOpenAICompatConfig(deploymentURL, apiKey, model).WithHeaders(map[string]string{"api-key": apiKey}).WithQueryParams(map[string]string{"api-version": apiVersion}))
}
