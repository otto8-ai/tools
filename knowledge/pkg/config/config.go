package config

import (
	"fmt"
	"os"
	"path"

	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	EmbeddingsConfig EmbeddingsConfig `koanf:"embeddings" json:"embeddings,omitempty"`
}

type EmbeddingsConfig struct {
	Providers []ModelProviderConfig `koanf:"providers" json:"providers,omitempty" mapstructure:"providers"`
}

type ModelProviderConfig struct {
	Name   string         `koanf:"name" json:"name,omitempty" mapstructure:"name"`
	Type   string         `koanf:"type" json:"type,omitempty" mapstructure:"type"`
	Config map[string]any `koanf:"config" json:"config,omitempty" mapstructure:"config"`
}

type DatabaseConfig struct {
	DSN         string `name:"index-dsn" usage:"Index Database Connection string (relational DB) (default \"sqlite://$XDG_DATA_HOME/gptscript/knowledge/knowledge.db\")" default:"" env:"KNOW_INDEX_DSN"`
	AutoMigrate string `usage:"Auto migrate database" default:"true" env:"KNOW_DB_AUTO_MIGRATE"`
}

type VectorDBConfig struct {
	DSN string `name:"vector-dsn" usage:"DSN to the vector database (default \"chromem:$XDG_DATA_HOME/gptscript/knowledge/vector.db\")" default:"" env:"KNOW_VECTOR_DSN"`
}

func LoadConfig(configFile string) (*Config, error) {
	cfg := &Config{}
	if configFile == "" {
		return cfg, nil
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Expand environment variables in config
	content = []byte(os.ExpandEnv(string(content)))

	k := koanf.New(".")
	var pa koanf.Parser
	switch path.Ext(configFile) {
	case ".json":
		pa = json.Parser()
	case ".yaml", ".yml":
		pa = yaml.Parser()
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", path.Ext(configFile))
	}

	if err := k.Load(rawbytes.Provider(content), pa); err != nil {
		return nil, fmt.Errorf("error loading config file %q: %w", configFile, err)
	}

	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: "koanf"}); err != nil {
		return nil, fmt.Errorf("error unmarshalling config file %q: %w", configFile, err)
	}

	return cfg, nil
}

func (ec *EmbeddingsConfig) RemoveUnselected(selected string) {
	keep := make([]ModelProviderConfig, 1)
	for _, p := range ec.Providers {
		if p.Name == selected {
			keep[0] = p
		}
	}
	ec.Providers = keep
}
