package embeddings

import (
	"github.com/gptscript-ai/knowledge/pkg/config"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/openai"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/vertex"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestLoadConfOpenAI(t *testing.T) {
	// Unset the OPENAI_API_KEY env var so test passes even if it's set in the system env
	originalEnv := os.Getenv("OPENAI_API_KEY")
	defer os.Setenv("OPENAI_API_KEY", originalEnv)
	_ = os.Unsetenv("OPENAI_API_KEY")

	dotenv := "test_assets/openai_env"
	require.NoError(t, godotenv.Load(dotenv))

	configFile := "test_assets/testcfg.yaml"
	cfg, err := config.LoadConfig(configFile)
	require.NoError(t, err)

	p, err := GetSelectedEmbeddingsModelProvider("openai", cfg.EmbeddingsConfig)
	require.NoError(t, err)
	require.Equal(t, "openai", p.Name())

	t.Logf("Provider: %s", p.Name())
	t.Logf("Config: %#v", p.Config())

	conf := p.Config().(*openai.EmbeddingModelProviderOpenAI)

	assert.Equal(t, "https://foo.bar.spam", conf.BaseURL) // this is in config and env, so env should take precedence
	assert.Equal(t, "sk-1234567890abcdef", conf.APIKey)   // this should come from config
}

func TestLoadConfVertex(t *testing.T) {
	dotenv := "test_assets/vertex_env"
	require.NoError(t, godotenv.Load(dotenv))

	cfg, err := config.LoadConfig("")
	require.NoError(t, err)

	// Load the configuration
	p, err := GetSelectedEmbeddingsModelProvider("vertex", cfg.EmbeddingsConfig)
	require.NoError(t, err)
	require.Equal(t, "vertex", p.Name())

	t.Logf("Config: %#v", p.Config())

	conf := p.Config().(*vertex.EmbeddingProviderVertex)

	require.Equal(t, "foo-embedding-001", conf.Model)
	require.Equal(t, "foo-project", conf.Project)
}

func TestExportConfigWithValidStruct(t *testing.T) {
	type Config struct {
		Field1 string `export:"true"`
		Field2 int    `export:"true"`
		Field3 bool   `export:"false"`
	}

	input := &Config{
		Field1: "value",
		Field2: 42,
		Field3: true,
	}

	expected := &Config{
		Field1: "value",
		Field2: 42,
		Field3: false,
	}

	result, err := ExportConfig(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExportConfigWithNestedStruct(t *testing.T) {
	type Nested struct {
		InnerField string `export:"true"`
	}

	type Config struct {
		Field1 string `export:"true"`
		Nested Nested `export:"true"`
	}

	input := &Config{
		Field1: "value",
		Nested: Nested{
			InnerField: "inner value",
		},
	}

	expected := &Config{
		Field1: "value",
		Nested: Nested{
			InnerField: "inner value",
		},
	}

	result, err := ExportConfig(input)
	require.NoError(t, err)
	assert.Equal(t, expected, result)
}

func TestExportConfigWithRequiredField(t *testing.T) {
	type Config struct {
		Field1 string `export:"required"`
		Field2 int    `export:"true"`
	}

	input := &Config{
		Field2: 42,
	}

	_, err := ExportConfig(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "\"Field1\" is required")
}

func TestExportConfigWithNilPointer(t *testing.T) {
	type Config struct {
		Field1 *string `export:"true"`
	}

	var input *Config

	_, err := ExportConfig(input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "input must be a non-nil pointe")
}
