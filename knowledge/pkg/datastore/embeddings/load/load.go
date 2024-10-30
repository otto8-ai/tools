package load

import (
	"fmt"
	"github.com/knadh/koanf/providers/env"
	koanf "github.com/knadh/koanf/v2"
)

var k = koanf.New(".")

// FillConfigEnv fills the given struct with values from a config file and environment variables.
// The envPrefix parameter is used to prefix environment variables.
func FillConfigEnv(envPrefix string, cfg interface{}) error {
	// Load environment variables and override config file values
	if err := k.Load(env.Provider(envPrefix, ".", func(s string) string { return s }), nil); err != nil {
		return fmt.Errorf("error loading environment variables: %w", err)
	}

	// Unmarshal environment variables into the struct, thus overriding file config values if present
	if err := k.UnmarshalWithConf("", cfg, koanf.UnmarshalConf{Tag: "env", FlatPaths: true}); err != nil {
		return fmt.Errorf("error unmarshalling environment variables: %w", err)
	}

	return nil
}
