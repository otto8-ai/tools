package converter

import (
	"context"
	"fmt"
	"io"
)

type Converter interface {
	Convert(ctx context.Context, reader io.Reader, sourceExt, outputFormat string) (io.Reader, error)
	Name() string
}

func GetConverterConfig(name string) (any, error) {
	switch name {
	case "soffice":
		return SofficeConverter{}, nil
	default:
		return nil, fmt.Errorf("unknown document converter %q", name)
	}
}

func GetConverter(name string, config any) (Converter, error) {
	switch name {
	case "soffice":
		return NewSofficeConverter()
	default:
		return nil, fmt.Errorf("unknown document converter %q", name)
	}
}
