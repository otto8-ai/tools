package vertex

import (
	"fmt"
	"strings"

	"dario.cat/mergo"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	cg "github.com/philippgille/chromem-go"
)

type EmbeddingProviderVertex struct {
	APIKey      string `koanf:"apiKey" env:"VERTEX_API_KEY" export:"false"`
	APIEndpoint string `koanf:"apiEndpoint" env:"VERTEX_API_ENDPOINT" export:"true"`
	Project     string `koanf:"project" env:"VERTEX_PROJECT" export:"true"`
	Model       string `koanf:"model" env:"VERTEX_MODEL" export:"required"`
}

const EmbeddingProviderVertexName = "vertex"

func (p *EmbeddingProviderVertex) Name() string {
	return EmbeddingProviderVertexName
}

func (p *EmbeddingProviderVertex) Configure() error {
	if err := load.FillConfigEnv(strings.ToUpper(EmbeddingProviderVertexName), &p); err != nil {
		return fmt.Errorf("failed to fill Vertex config from environment: %w", err)
	}

	if err := p.fillDefaults(); err != nil {
		return fmt.Errorf("failed to fill Vertex defaults: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderVertex) fillDefaults() error {
	defaultCfg := EmbeddingProviderVertex{
		APIEndpoint: "",
		Project:     "",
		Model:       "text-embedding-004",
	}

	if err := mergo.Merge(p, defaultCfg); err != nil {
		return fmt.Errorf("failed to merge Vertex config: %w", err)
	}

	return nil
}

func (p *EmbeddingProviderVertex) EmbeddingFunc() (cg.EmbeddingFunc, error) {
	if p.APIEndpoint != "" {
		return cg.NewEmbeddingFuncVertex(p.APIKey, p.Project, cg.EmbeddingModelVertex(p.Model), cg.WithVertexAPIEndpoint(p.APIEndpoint)), nil
	}
	return cg.NewEmbeddingFuncVertex(p.APIKey, p.Project, cg.EmbeddingModelVertex(p.Model)), nil
}

func (p *EmbeddingProviderVertex) Config() any {
	return p
}
