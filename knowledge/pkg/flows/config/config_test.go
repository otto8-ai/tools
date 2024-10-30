package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfigFromValidJSONFile(t *testing.T) {
	cfg, err := FromFile("testdata/valid.json")
	assert.NoError(t, err)
	require.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Flows)
	assert.Equal(t, 2, len(cfg.Flows))
	assert.Equal(t, 1, len(cfg.Flows["flow1"].Ingestion))
	assert.Equal(t, ".txt", cfg.Flows["flow1"].Ingestion[0].Filetypes[0])
	assert.Empty(t, cfg.Flows["flow1"].Retrieval)
}

func TestLoadConfigFromValidYAMLFile(t *testing.T) {
	cfg, err := FromFile("testdata/valid.yaml")
	assert.NoError(t, err)
	require.NotNil(t, cfg)
	assert.NotEmpty(t, cfg.Flows)
	assert.Equal(t, 2, len(cfg.Flows))
	assert.Equal(t, 4096.0, cfg.Flows["flow2"].Ingestion[0].TextSplitter.Options["chunkSize"])
}

func TestLoadConfigFromInvalidFile(t *testing.T) {
	cfg, err := FromFile("testdata/invalid.txt")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigFromNonexistentFile(t *testing.T) {
	cfg, err := FromFile("testdata/nonexistent.yaml")
	assert.Error(t, err)
	assert.Nil(t, cfg)
}

func TestLoadConfigInvalidDoubleDefault(t *testing.T) {
	_, err := FromFile("testdata/invalid_doubledefault.yaml")
	assert.Error(t, err)
}
