package config

import (
	_ "embed"
	"fmt"
)

//go:embed blueprints/default.yaml
var BlueprintDefault []byte

//go:embed blueprints/context.yaml
var BlueprintContext []byte

//go:embed blueprints/obot.yaml
var BlueprintObot []byte

var Blueprints = map[string][]byte{
	"default": BlueprintDefault,
	"context": BlueprintContext,
	"obot":    BlueprintObot,
}

func GetBlueprint(name string) ([]byte, error) {
	bp, ok := Blueprints[name]
	if !ok {
		return nil, fmt.Errorf("blueprint %q not found", name)
	}
	return bp, nil
}
