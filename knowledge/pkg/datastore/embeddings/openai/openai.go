package openai

import (
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"dario.cat/mergo"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	cg "github.com/philippgille/chromem-go"
)

const EmbeddingModelProviderOpenAIName string = "openai"

type EmbeddingModelProviderOpenAI struct {
	BaseURL           string            `usage:"OpenAI API base" default:"https://api.openai.com/v1" env:"OPENAI_BASE_URL" koanf:"baseURL"`
	APIKey            string            `usage:"OpenAI API key (not required if used with clicky-chats)" default:"sk-foo" env:"OPENAI_API_KEY" koanf:"apiKey" mapstructure:"apiKey" export:"false"`
	Model             string            `usage:"OpenAI model" default:"gpt-4" env:"OPENAI_MODEL" koanf:"openai-model"`
	EmbeddingModel    string            `usage:"OpenAI Embedding model" default:"text-embedding-3-small" env:"OPENAI_EMBEDDING_MODEL" koanf:"embeddingModel" export:"required"`
	EmbeddingEndpoint string            `usage:"OpenAI Embedding endpoint" default:"/embeddings" env:"OPENAI_EMBEDDING_ENDPOINT" koanf:"embeddingEndpoint"`
	APIVersion        string            `usage:"OpenAI API version (for Azure)" default:"2024-02-01" env:"OPENAI_API_VERSION" koanf:"apiVersion"`
	APIType           string            `usage:"OpenAI API type (OPEN_AI, AZURE, AZURE_AD, ...)" default:"OPEN_AI" env:"OPENAI_API_TYPE" koanf:"apiType"`
	AzureOpenAIConfig AzureOpenAIConfig `koanf:"azure"`
}

type OpenAIConfig struct {
	BaseURL           string            `usage:"OpenAI API base" default:"https://api.openai.com/v1" env:"OPENAI_BASE_URL" koanf:"baseURL"`
	APIKey            string            `usage:"OpenAI API key (not required if used with clicky-chats)" default:"sk-foo" env:"OPENAI_API_KEY" koanf:"apiKey" mapstructure:"apiKey" export:"false"`
	Model             string            `usage:"OpenAI model" default:"gpt-4" env:"OPENAI_MODEL" koanf:"openai-model"`
	EmbeddingModel    string            `usage:"OpenAI Embedding model" default:"text-embedding-3-small" env:"OPENAI_EMBEDDING_MODEL" koanf:"embeddingModel" export:"required"`
	EmbeddingEndpoint string            `usage:"OpenAI Embedding endpoint" default:"/embeddings" env:"OPENAI_EMBEDDING_ENDPOINT" koanf:"embeddingEndpoint"`
	APIVersion        string            `usage:"OpenAI API version (for Azure)" default:"2024-02-01" env:"OPENAI_API_VERSION" koanf:"apiVersion"`
	APIType           string            `usage:"OpenAI API type (OPEN_AI, AZURE, AZURE_AD, ...)" default:"OPEN_AI" env:"OPENAI_API_TYPE" koanf:"apiType"`
	AzureOpenAIConfig AzureOpenAIConfig `koanf:"azure"`
}

func (o OpenAIConfig) Name() string {
	return EmbeddingModelProviderOpenAIName
}

type AzureOpenAIConfig struct {
	Deployment string `usage:"Azure OpenAI deployment name (overrides openai-embedding-model, if set)" default:"" env:"OPENAI_AZURE_DEPLOYMENT" koanf:"deployment"`
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
		Model:             "gpt-4",
		EmbeddingModel:    "text-embedding-3-small",
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

		embeddingFunc = cg.NewEmbeddingFuncAzureOpenAI(
			p.APIKey,
			deploymentURL.String(),
			p.APIVersion,
			"",
		)
	case "open_ai":
		cfg := cg.NewOpenAICompatConfig(
			p.BaseURL,
			p.APIKey,
			p.EmbeddingModel,
		).
			WithNormalized(true).
			WithEmbeddingsEndpoint(p.EmbeddingEndpoint)
		embeddingFunc = cg.NewEmbeddingFuncOpenAICompat(cfg)
	default:
		return nil, fmt.Errorf("unknown OpenAI API type: %q", p.APIType)
	}

	return embeddingFunc, nil
}

func (p *EmbeddingModelProviderOpenAI) Config() any {
	return p
}
