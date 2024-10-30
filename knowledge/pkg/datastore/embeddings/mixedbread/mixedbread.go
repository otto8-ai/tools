package mixedbread

import (
	"dario.cat/mergo"
	"fmt"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	cg "github.com/philippgille/chromem-go"
	"strings"
)

type EmbeddingProviderMixedbread struct {
	APIKey string `koanf:"apiKey" env:"MIXEDBREAD_API_KEY" export:"false"`
	Model  string `koanf:"model" env:"MIXEDBREAD_MODEL" export:"required"`
}

const EmbeddingProviderMixedbreadName = "mixedbread"

func (p *EmbeddingProviderMixedbread) Name() string {
	return EmbeddingProviderMixedbreadName
}

func (p *EmbeddingProviderMixedbread) Configure() error {
	if err := load.FillConfigEnv(strings.ToUpper(EmbeddingProviderMixedbreadName), &p); err != nil {
		return fmt.Errorf("failed to fill Mixedbread config from environment: %w", err)
	}

	if err := p.fillDefaults(); err != nil {
		return fmt.Errorf("failed to fill Mixedbread defaults: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderMixedbread) fillDefaults() error {
	defaultCfg := EmbeddingProviderMixedbread{
		Model: "all-MiniLM-L6-v2",
	}

	if err := mergo.Merge(p, defaultCfg); err != nil {
		return fmt.Errorf("failed to merge Mixedbread config: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderMixedbread) EmbeddingFunc() (cg.EmbeddingFunc, error) {
	return cg.NewEmbeddingFuncMixedbread(p.APIKey, cg.EmbeddingModelMixedbread(p.Model)), nil
}

func (p *EmbeddingProviderMixedbread) Config() any {
	return p
}
