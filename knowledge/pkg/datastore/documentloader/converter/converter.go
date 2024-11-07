package converter

import (
	"context"
	"io"
)

type Converter interface {
	Convert(ctx context.Context, reader io.Reader, format string) error
}
