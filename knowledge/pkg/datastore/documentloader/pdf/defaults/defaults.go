package defaults

import (
	"context"
	"io"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/pdf/gopdf"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
)

var DefaultPDFReaderFunc func(ctx context.Context, reader io.Reader) ([]vs.Document, error) = func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
	slog.Debug("Default PDF reader is GoPDF")
	r, err := gopdf.NewDefaultPDF(reader)
	if err != nil {
		return nil, err
	}
	return r.Load(ctx)
}
