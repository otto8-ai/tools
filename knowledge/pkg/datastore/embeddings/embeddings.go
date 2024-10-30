package embeddings

import (
	"errors"
	"fmt"
	"github.com/gptscript-ai/knowledge/pkg/config"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/cohere"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/jina"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/localai"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/mistral"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/mixedbread"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/ollama"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/openai"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/types"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/vertex"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

func GetSelectedEmbeddingsModelProvider(selected string, embeddingsConfig config.EmbeddingsConfig) (types.EmbeddingModelProvider, error) {
	providerCfg, err := GetProviderCfg(selected, embeddingsConfig)
	if err != nil {
		return nil, err
	}

	provider, err := ProviderFromConfig(*providerCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %w", err)
	}

	if err := provider.Configure(); err != nil {
		return nil, fmt.Errorf("failed to configure provider: %w", err)
	}

	return provider, nil
}

func ProviderFromConfig(providerConfig config.ModelProviderConfig) (types.EmbeddingModelProvider, error) {
	provider, err := GetProviderConfig(providerConfig.Type)
	if err != nil {
		return nil, err
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "koanf",
		Result:  provider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create decoder: %w", err)
	}
	if err := decoder.Decode(providerConfig.Config); err != nil {
		return nil, fmt.Errorf("failed to decode provider config: %w", err)
	}

	return provider, nil
}

func GetProviderConfig(providerType string) (types.EmbeddingModelProvider, error) {
	switch strings.ToLower(providerType) {
	case openai.EmbeddingModelProviderOpenAIName:
		return &openai.EmbeddingModelProviderOpenAI{}, nil
	case cohere.EmbeddingModelProviderCohereName:
		return &cohere.EmbeddingModelProviderCohere{}, nil
	case vertex.EmbeddingProviderVertexName:
		return &vertex.EmbeddingProviderVertex{}, nil
	case jina.EmbeddingProviderJinaName:
		return &jina.EmbeddingProviderJina{}, nil
	case mistral.EmbeddingProviderMistralName:
		return &mistral.EmbeddingProviderMistral{}, nil
	case mixedbread.EmbeddingProviderMixedbreadName:
		return &mixedbread.EmbeddingProviderMixedbread{}, nil
	case localai.EmbeddingProviderLocalAIName:
		return &localai.EmbeddingProviderLocalAI{}, nil
	case ollama.EmbeddingProviderOllamaName:
		return &ollama.EmbeddingProviderOllama{}, nil
	default:
		return nil, fmt.Errorf("unknown embedding model provider %q", providerType)
	}
}

func FindProviderConfig(name string, providers []config.ModelProviderConfig) *config.ModelProviderConfig {
	for _, p := range providers {
		if p.Name == name {
			return &p
		}
	}
	return nil
}

func GetProviderCfg(name string, embeddingsConfig config.EmbeddingsConfig) (*config.ModelProviderConfig, error) {
	providerCfg := FindProviderConfig(name, embeddingsConfig.Providers)
	if providerCfg == nil {
		// no config with that name exists, so we assume the name is the type
		providerCfg = &config.ModelProviderConfig{
			Name: name,
			Type: name,
		}
	}

	return providerCfg, nil
}

func ExportConfig(c any) (any, error) {
	v := reflect.ValueOf(c)

	// Check if input is a pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, errors.New("input must be a non-nil pointer")
		}
		v = v.Elem() // Dereference the pointer to get the actual value
	}

	// Ensure we're working with a struct
	if v.Kind() != reflect.Struct {
		return nil, errors.New("input must be a struct or a pointer to a struct")
	}

	// Create a new instance of the struct type
	result := reflect.New(v.Type()).Elem()

	// Iterate over the struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		// Get the export tag value
		exportTag := fieldType.Tag.Get("export")

		// Handle the "false" export tag
		if exportTag == "false" {
			continue
		}

		// Handle nested structs by calling ExportConfig recursively
		if field.Kind() == reflect.Struct {
			n, err := ExportConfig(field.Addr().Interface())
			if err != nil {
				return nil, err
			}
			result.Field(i).Set(reflect.ValueOf(n).Elem())
			continue
		}

		// Handle the "required" export tag
		if exportTag == "required" && field.IsZero() {
			return nil, fmt.Errorf("field %q is required", fieldType.Name)
		}

		// Copy the field value to the result
		result.Field(i).Set(field)
	}

	// Return the result as a pointer if the original input was a pointer
	if reflect.ValueOf(c).Kind() == reflect.Ptr {
		return result.Addr().Interface(), nil
	}

	return result.Interface(), nil
}

func CompareRequiredFields(a, b any) error {
	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// Ensure both inputs are pointers or structs
	if va.Kind() == reflect.Ptr {
		if va.IsNil() {
			return errors.New("first input must be a non-nil pointer or a struct")
		}
		va = va.Elem()
	}

	if vb.Kind() == reflect.Ptr {
		if vb.IsNil() {
			return errors.New("second input must be a non-nil pointer or a struct")
		}
		vb = vb.Elem()
	}

	if va.Kind() != reflect.Struct || vb.Kind() != reflect.Struct {
		return errors.New("both inputs must be structs or pointers to structs")
	}

	// Iterate over the fields of A
	for i := 0; i < va.NumField(); i++ {
		fieldA := va.Field(i)
		fieldTypeA := va.Type().Field(i)

		exportTag := fieldTypeA.Tag.Get("export")
		if exportTag == "required" {
			// Get the corresponding field in B
			fieldB := vb.FieldByName(fieldTypeA.Name)
			if !fieldB.IsValid() {
				return fmt.Errorf("field %q is missing in the second struct", fieldTypeA.Name)
			}

			x := fieldA.Interface()
			y := fieldB.Interface()

			// Check for equality
			if !reflect.DeepEqual(x, y) {
				return fmt.Errorf("field %q does not match: %v != %v", fieldTypeA.Name, x, y)
			}
		}
	}

	return nil
}

func AsEmbeddingModelProviderConfig(emp types.EmbeddingModelProvider, export bool) (config.ModelProviderConfig, error) {
	var cfg map[string]any

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName: "koanf",
		Result:  &cfg,
	})

	if err != nil {
		return config.ModelProviderConfig{}, err
	}

	conf := emp.Config()
	if export {
		conf, err = ExportConfig(conf)
		if err != nil {
			return config.ModelProviderConfig{}, err
		}
	}

	if err := decoder.Decode(conf); err != nil {
		return config.ModelProviderConfig{}, err
	}

	return config.ModelProviderConfig{
		Type:   emp.Name(),
		Config: cfg,
	}, nil
}
