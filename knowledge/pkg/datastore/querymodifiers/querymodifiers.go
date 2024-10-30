package querymodifiers

import (
	"fmt"
)

type QueryModifier interface {
	ModifyQueries(queries []string) ([]string, error)
	Name() string
}

var QueryModifiers = map[string]QueryModifier{
	SpellcheckQueryModifierName: SpellcheckQueryModifier{},
	EnhanceQueryModifierName:    EnhanceQueryModifier{},
	GenericQueryModifierName:    GenericQueryModifier{},
}

func GetQueryModifier(name string) (QueryModifier, error) {
	qm, ok := QueryModifiers[name]
	if !ok {
		return nil, fmt.Errorf("unknown queryModifier %q", name)
	}
	return qm, nil
}
