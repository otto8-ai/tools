package mistral

import (
	"dario.cat/mergo"
	"fmt"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	cg "github.com/philippgille/chromem-go"
	"strings"
)

type EmbeddingProviderMistral struct {
	APIKey string `koanf:"apiKey" env:"MISTRAL_API_KEY" export:"false"`
	Model  string `koanf:"model" env:"MISTRAL_MODEL" export:"required"`
}

const EmbeddingProviderMistralName = "mistral"

func (p *EmbeddingProviderMistral) Name() string {
	return EmbeddingProviderMistralName
}

func (p *EmbeddingProviderMistral) Configure() error {
	if err := load.FillConfigEnv(strings.ToUpper(EmbeddingProviderMistralName), &p); err != nil {
		return fmt.Errorf("failed to fill Mistral config from environment: %w", err)
	}

	if err := p.fillDefaults(); err != nil {
		return fmt.Errorf("failed to fill Mistral defaults: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderMistral) fillDefaults() error {
	defaultCfg := EmbeddingProviderMistral{
		Model: "mistral-embed",
	}

	if err := mergo.Merge(p, defaultCfg); err != nil {
		return fmt.Errorf("failed to merge Mistral config: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderMistral) EmbeddingFunc() (cg.EmbeddingFunc, error) {
	return cg.NewEmbeddingFuncMistral(p.APIKey), nil
}

func (p *EmbeddingProviderMistral) Config() any {
	return p
}
