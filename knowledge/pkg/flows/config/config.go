package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/output"

	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader"
	"github.com/gptscript-ai/knowledge/pkg/datastore/postprocessors"
	"github.com/gptscript-ai/knowledge/pkg/datastore/querymodifiers"
	"github.com/gptscript-ai/knowledge/pkg/datastore/retrievers"
	"github.com/gptscript-ai/knowledge/pkg/datastore/textsplitter"
	"github.com/gptscript-ai/knowledge/pkg/datastore/transformers"
	"github.com/gptscript-ai/knowledge/pkg/flows"
	"github.com/mitchellh/mapstructure"
	"sigs.k8s.io/yaml"
)

type GenericBaseConfig struct {
	Name    string         `json:"name" yaml:"name" mapstructure:"name"`
	Options map[string]any `json:"options,omitempty" yaml:"options" mapstructure:"options"`
}

type FlowConfig struct {
	Flows    map[string]FlowConfigEntry `json:"flows" yaml:"flows" mapstructure:"flows"`
	Datasets map[string]string          `json:"datasets,omitempty" yaml:"datasets" mapstructure:"datasets"`
}

type FlowConfigEntry struct {
	Default   bool                      `json:"default,omitempty" yaml:"default" mapstructure:"default"`
	Globals   FlowConfigEntryGlobalOpts `json:"globals,omitempty" yaml:"globals" mapstructure:"globals"`
	Ingestion []IngestionFlowConfig     `json:"ingestion,omitempty" yaml:"ingestion" mapstructure:"ingestion"`
	Retrieval *RetrievalFlowConfig      `json:"retrieval,omitempty" yaml:"retrieval" mapstructure:"retrieval"`
}

type FlowConfigEntryGlobalOpts struct {
	Ingestion FlowConfigGlobalsIngestion `json:"ingestion,omitempty" yaml:"ingestion" mapstructure:"ingestion"`
}

type FlowConfigGlobalsIngestion struct {
	Textsplitter map[string]any `json:"textsplitter,omitempty" yaml:"textsplitter" mapstructure:"textsplitter"`
}

type IngestionFlowConfig struct {
	Filetypes      []string             `json:"filetypes" yaml:"filetypes" mapstructure:"filetypes"`
	DocumentLoader DocumentLoaderConfig `json:"documentLoader,omitempty" yaml:"documentLoader" mapstructure:"documentLoader"`
	TextSplitter   TextSplitterConfig   `json:"textSplitter,omitempty" yaml:"textSplitter" mapstructure:"textSplitter"`
	Transformers   []TransformerConfig  `json:"transformers,omitempty" yaml:"transformers" mapstructure:"transformers"`
}

type RetrievalFlowConfig struct {
	// QueryModifiers allows to modify the input query before it is passed to the retriever. (Query-Rewriting)
	QueryModifiers []QueryModifierConfig `json:"queryModifiers,omitempty" yaml:"queryModifiers" mapstructure:"queryModifiers"`

	// Retriever is the configuration for the retriever to be used. E.g. instead of using a naive retriever, you can use a recursive or refining retriever.
	Retriever *RetrieverConfig `json:"retriever,omitempty" yaml:"retriever" mapstructure:"retriever"`

	// Postprocessors are used to process the retrieved documents before they are returned. This may include stripping metadata or re-ranking.
	Postprocessors []TransformerConfig `json:"postprocessors,omitempty" yaml:"postprocessors" mapstructure:"postprocessors"`
}

type QueryModifierConfig struct {
	GenericBaseConfig
}

type RetrieverConfig struct {
	GenericBaseConfig
}

type DocumentLoaderConfig struct {
	GenericBaseConfig
}

type TextSplitterConfig struct {
	GenericBaseConfig
}

type TransformerConfig struct {
	GenericBaseConfig
}

func FromBlueprint(name string) (*FlowConfig, error) {
	bp, err := GetBlueprint(name)
	if err != nil {
		return nil, err
	}
	return FromBytes(bp)
}

func Load(reference string) (*FlowConfig, error) {
	if strings.HasPrefix(reference, "blueprint:") {
		return FromBlueprint(strings.TrimPrefix(reference, "blueprint:"))
	}
	return FromFile(reference)
}

// FromFile reads a configuration file and returns a FlowConfig.
func FromFile(filename string) (*FlowConfig, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return FromBytes(content)
}

func FromBytes(content []byte) (*FlowConfig, error) {
	// Expand environment variables in config
	content = []byte(os.ExpandEnv(string(content)))

	var err error

	var config FlowConfig
	jsondata := content
	if !json.Valid(content) {
		jsondata, err = yaml.YAMLToJSON(content)
		if err != nil {
			return nil, err
		}
	}

	err = json.Unmarshal(jsondata, &config)
	if err != nil {
		return nil, err
	}

	return &config, config.Validate()
}

func (f *FlowConfig) Validate() error {
	hasDefault := false
	for name, flow := range f.Flows {
		// Only one default flow is allowed
		if flow.Default {
			if hasDefault {
				return fmt.Errorf("multiple flows are marked as default")
			}
			hasDefault = true
		}

		// Each flow must have either ingestion or retrieval
		if len(flow.Ingestion) == 0 && flow.Retrieval == nil {
			return fmt.Errorf("flow %q has neither ingestion nor retrieval specified", name)
		}

		for idx, ingestion := range flow.Ingestion {
			// Each ingestion flow must have some filetypes specified
			if len(ingestion.Filetypes) == 0 {
				return fmt.Errorf("flow %q.ingestion.[%d] has no filetypes specified", name, idx)
			}
		}
	}
	return nil
}

func (f *FlowConfig) GetDefaultFlowConfigEntry() (*FlowConfigEntry, error) {
	for _, flow := range f.Flows {
		if flow.Default {
			return &flow, nil
		}
	}
	return nil, fmt.Errorf("default flow not found")
}

func (f *FlowConfig) GetFlow(name string) (*FlowConfigEntry, error) {
	flow, ok := f.Flows[name]
	if !ok {
		return nil, fmt.Errorf("flow %q not found", name)
	}
	return &flow, nil
}

// AsIngestionFlow converts an IngestionFlowConfig to an actual flows.IngestionFlow.
func (i *IngestionFlowConfig) AsIngestionFlow(globals *FlowConfigGlobalsIngestion) (*flows.IngestionFlow, error) {
	flow := &flows.IngestionFlow{
		Filetypes: i.Filetypes,
		Globals: flows.IngestionFlowGlobals{
			SplitterOpts: globals.Textsplitter,
		},
	}

	if i.DocumentLoader.Name != "" {
		name := strings.ToLower(strings.Trim(i.DocumentLoader.Name, " "))
		cfg, err := documentloader.GetDocumentLoaderConfig(name)
		if err != nil {
			return nil, err
		}
		if len(i.DocumentLoader.Options) > 0 {
			jsondata, err := json.Marshal(i.DocumentLoader.Options)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal(jsondata, &cfg)
			if err != nil {
				return nil, err
			}
		}
		loaderFunc, err := documentloader.GetDocumentLoaderFunc(name, cfg)
		if err != nil {
			return nil, err
		}
		flow.Load = loaderFunc
	}

	if i.TextSplitter.Name != "" {
		name := strings.ToLower(strings.Trim(i.TextSplitter.Name, " "))
		cfg, err := textsplitter.GetTextSplitterConfig(name)
		if err != nil {
			return nil, err
		}

		if len(globals.Textsplitter) > 0 {
			if err := mapstructure.Decode(globals.Textsplitter, &cfg); err != nil {
				return nil, fmt.Errorf("failed to decode text splitter global configuration: %w", err)
			}
		}

		if len(i.TextSplitter.Options) > 0 {
			if err := mapstructure.Decode(i.TextSplitter.Options, &cfg); err != nil {
				return nil, fmt.Errorf("failed to decode text splitter %q configuration: %w", i.TextSplitter.Name, err)
			}
		}

		splitterFunc, err := textsplitter.GetTextSplitter(name, cfg)
		if err != nil {
			return nil, err
		}
		flow.Splitter = splitterFunc
	}

	if len(i.Transformers) > 0 {
		for _, tf := range i.Transformers {
			transformer, err := transformers.GetTransformer(tf.Name)
			if err != nil {
				return nil, err
			}
			if len(tf.Options) > 0 {
				if err := mapstructure.Decode(tf.Options, &transformer); err != nil {
					return nil, fmt.Errorf("failed to decode transformer configuration: %w", err)
				}
				slog.Debug("Transformer custom configuration", "name", tf.Name, "config", output.RedactSensitive(transformer))
			}
			flow.Transformations = append(flow.Transformations, transformer)
		}
	}

	return flow, nil
}

func (f *FlowConfig) ForDataset(name string) (*FlowConfigEntry, error) {
	flowref, ok := f.Datasets[name]
	if ok {
		slog.Debug("Flow assigned to dataset", "dataset", name, "flow", flowref)
		return f.GetFlow(flowref)
	}
	slog.Debug("No flow found for dataset - using default", "dataset", name)
	return f.GetDefaultFlowConfigEntry()
}

func (r *RetrievalFlowConfig) AsRetrievalFlow() (*flows.RetrievalFlow, error) {
	flow := &flows.RetrievalFlow{}

	if len(r.QueryModifiers) > 0 {
		for _, qm := range r.QueryModifiers {
			modifier, err := querymodifiers.GetQueryModifier(qm.Name)
			if err != nil {
				return nil, err
			}
			if len(qm.Options) > 0 {
				if err := mapstructure.Decode(qm.Options, &modifier); err != nil {
					return nil, fmt.Errorf("failed to decode query modifier configuration: %w", err)
				}
				slog.Debug("Query Modifier custom configuration", "name", qm.Name, "config", output.RedactSensitive(modifier))
			}
			flow.QueryModifiers = append(flow.QueryModifiers, modifier)
		}
	}

	if r.Retriever != nil {
		ret, err := retrievers.GetRetriever(r.Retriever.Name)
		if err != nil {
			return nil, err
		}
		if len(r.Retriever.Options) > 0 {
			if err := ret.DecodeConfig(r.Retriever.Options); err != nil {
				return nil, fmt.Errorf("failed to decode retriever configuration: %w", err)
			}
			slog.Debug("Retriever custom configuration", "name", r.Retriever.Name, "config", output.RedactSensitive(ret))
		}
		flow.Retriever = ret
	}

	if len(r.Postprocessors) > 0 {
		for _, pp := range r.Postprocessors {
			postprocessor, err := postprocessors.GetPostprocessor(pp.Name)
			if err != nil {
				return nil, err
			}

			if len(pp.Options) > 0 {
				// if it's a transformer wrapper, call decode
				if transformerWrapper, ok := postprocessor.(*postprocessors.TransformerWrapper); ok {
					if err := transformerWrapper.Decode(pp.Options); err != nil {
						return nil, fmt.Errorf("failed to decode postprocessor configuration: %w", err)
					}
					postprocessor = transformerWrapper
				} else {
					if err := mapstructure.Decode(pp.Options, &postprocessor); err != nil {
						return nil, fmt.Errorf("failed to decode postprocessor configuration: %w", err)
					}
				}
				slog.Debug("Postprocessor custom configuration", "name", pp.Name, "config", output.RedactSensitive(postprocessor))
			}
			flow.Postprocessors = append(flow.Postprocessors, postprocessor)
		}
	}

	return flow, nil
}
