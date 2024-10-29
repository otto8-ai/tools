package llm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/openai"

	golcmodel "github.com/hupe1980/golc/model"
	"github.com/hupe1980/golc/model/chatmodel"
	"github.com/hupe1980/golc/prompt"
	"github.com/hupe1980/golc/schema"
)

type LLM struct {
	model schema.Model
}

type LLMConfig struct {
	OpenAI openai.OpenAIConfig
}

func NewFromConfig(cfg LLMConfig) (*LLM, error) {
	if cfg.OpenAI.APIKey != "" {
		return NewOpenAI(cfg.OpenAI)
	}
	return nil, fmt.Errorf("no LLM configuration found")
}

func NewOpenAI(cfg openai.OpenAIConfig) (*LLM, error) {
	m, err := chatmodel.NewOpenAI(cfg.APIKey, func(o *chatmodel.OpenAIOptions) {
		o.BaseURL = cfg.BaseURL
		o.ModelName = cfg.Model
	})
	if err != nil {
		return nil, err
	}

	return &LLM{model: m}, nil
}

func (llm *LLM) Prompt(ctx context.Context, promptTpl string, values map[string]any) (string, error) {
	p, err := prompt.NewTemplate(promptTpl).FormatPrompt(values)
	if err != nil {
		return "", err
	}
	slog.Debug("Prompting LLM", "prompt", p)

	res, err := golcmodel.GeneratePrompt(ctx, llm.model, p)
	if err != nil {
		return "", err
	}

	return res.Generations[0].Text, nil
}
