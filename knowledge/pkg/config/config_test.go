package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestEmbeddingsConfig_ClearUnselected(t *testing.T) {
	ec := &EmbeddingsConfig{
		Providers: []ModelProviderConfig{
			{
				Name: "openai",
				Type: "openai",
				Config: map[string]any{
					"baseURL": "foo.bar.com",
					"apiKey":  "123456",
				},
			},
			{
				Name: "cohere",
				Type: "cohere",
				Config: map[string]any{
					"model": "some-model",
				},
			},
		},
	}

	require.Len(t, ec.Providers, 2)

	ec.RemoveUnselected("openai")

	require.Len(t, ec.Providers, 1)
	require.Equal(t, "openai", ec.Providers[0].Name)

	ec.Providers = append(ec.Providers, ModelProviderConfig{
		Name: "cohere",
		Type: "cohere",
		Config: map[string]any{
			"model": "some-model",
		},
	})

	require.Len(t, ec.Providers, 2)

	ec.RemoveUnselected("cohere")

	require.Len(t, ec.Providers, 1)
	require.Equal(t, "cohere", ec.Providers[0].Name)
}
