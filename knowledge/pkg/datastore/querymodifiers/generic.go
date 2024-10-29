package querymodifiers

import (
	"context"
	"encoding/json"
	"github.com/gptscript-ai/knowledge/pkg/llm"
)

const GenericQueryModifierName = "generic"

type GenericQueryModifier struct {
	Model  llm.LLMConfig
	Prompt string
}

func (s GenericQueryModifier) Name() string {
	return GenericQueryModifierName
}

var genericPromptTpl = `The following query will be used for a vector similarity search. Please modify it like described below:
{{.prompt}}

Here's the original query: "{{.query}}"

Reply only with the resulting modified queries in a JSON list (single list element is fine) like this: {"results": ["<modified-query-1>", "<modified-query-2"]}
Do not include anything else in your response and don't use markdown highlighting or formatting, just raw JSON.`

type genericResp struct {
	Results []string `json:"results"`
}

func (s GenericQueryModifier) ModifyQueries(queries []string) ([]string, error) {
	m, err := llm.NewFromConfig(s.Model)
	if err != nil {
		return nil, err
	}

	modifiedQueries := make([]string, len(queries))
	for _, query := range queries {
		result, err := m.Prompt(context.Background(), genericPromptTpl, map[string]interface{}{"query": query, "prompt": s.Prompt})
		if err != nil {
			return nil, err
		}
		var resp genericResp
		err = json.Unmarshal([]byte(result), &resp)
		if err != nil {
			return nil, err
		}
		modifiedQueries = resp.Results
	}
	return modifiedQueries, nil
}
