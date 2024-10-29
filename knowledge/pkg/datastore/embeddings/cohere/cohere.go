package cohere

import (
	"dario.cat/mergo"
	"fmt"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	cg "github.com/philippgille/chromem-go"
)

const EmbeddingModelProviderCohereName string = "cohere"

type EmbeddingModelProviderCohere struct {
	APIKey string `env:"COHERE_API_KEY" koanf:"apiKey" export:"false"`
	Model  string `env:"COHERE_MODEL" koanf:"model" export:"required"`
}

func (p *EmbeddingModelProviderCohere) Name() string {
	return EmbeddingModelProviderCohereName
}

func (p *EmbeddingModelProviderCohere) Configure() error {
	if err := load.FillConfigEnv("COHERE_", p); err != nil {
		return fmt.Errorf("failed to fill Cohere config from environment: %w", err)
	}

	if err := p.fillDefaults(); err != nil {
		return fmt.Errorf("failed to fill Cohere defaults: %w", err)
	}

	return nil
}

func (p *EmbeddingModelProviderCohere) fillDefaults() error {
	defaultCfg := EmbeddingModelProviderCohere{
		Model: "embed-english-v3.0",
	}

	if err := mergo.Merge(&p, defaultCfg); err != nil {
		return fmt.Errorf("failed to merge Cohere config: %w", err)
	}

	return nil
}

func (p *EmbeddingModelProviderCohere) EmbeddingFunc() (cg.EmbeddingFunc, error) {
	return cg.NewEmbeddingFuncCohere(p.APIKey, cg.EmbeddingModelCohere(p.Model)), nil
}

func (p *EmbeddingModelProviderCohere) Config() any {
	return p
}
