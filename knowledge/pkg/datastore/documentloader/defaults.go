package documentloader

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"code.sajari.com/docconv/v2"
	pdfdefaults "github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/pdf/defaults"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	golcdocloaders "github.com/hupe1980/golc/documentloader"
	"github.com/lu4p/cat/rtftxt"
	lcgodocloaders "github.com/tmc/langchaingo/documentloaders"
)

// UnsupportedFileTypeError is returned when a file type is not supported
type UnsupportedFileTypeError struct {
	FileType string
}

func (e *UnsupportedFileTypeError) Error() string {
	return fmt.Sprintf("unsupported file type %q", e.FileType)
}

func (e *UnsupportedFileTypeError) Is(err error) bool {
	var unsupportedFileTypeError *UnsupportedFileTypeError
	ok := errors.As(err, &unsupportedFileTypeError)
	return ok
}

type DefaultDocLoaderFuncOpts struct {
	Archive ArchiveOpts
}

type ArchiveOpts struct {
	ErrOnUnsupportedFiletype bool
	ErrOnFailedFile          bool
}

func DefaultDocLoaderFunc(filetype string, opts DefaultDocLoaderFuncOpts) LoaderFunc {
	switch filetype {
	case ".pdf", "application/pdf":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return pdfdefaults.DefaultPDFReaderFunc(ctx, reader)
		}
	case ".html", "text/html":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromLangchain(lcgodocloaders.NewHTML(reader)).Load(ctx)
		}
	case ".md", "text/markdown":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromLangchain(lcgodocloaders.NewText(reader)).Load(ctx)
		}
	case ".txt", "text/plain":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromLangchain(lcgodocloaders.NewText(reader)).Load(ctx)
		}
	case ".csv", "text/csv":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			docs, err := FromGolc(golcdocloaders.NewCSV(reader)).Load(ctx)
			if err != nil && errors.Is(err, csv.ErrBareQuote) {
				oerr := err
				err = nil
				var nerr error
				docs, nerr = FromGolc(golcdocloaders.NewCSV(reader, func(o *golcdocloaders.CSVOptions) {
					o.LazyQuotes = true
				})).Load(ctx)
				if nerr != nil {
					err = errors.Join(oerr, nerr)
				}
			}
			return docs, err
		}
	case ".json", "application/json":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromLangchain(lcgodocloaders.NewText(reader)).Load(ctx)
		}
	case ".ipynb":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromGolc(golcdocloaders.NewNotebook(reader)).Load(ctx)
		}
	case ".docx", ".odt", ".rtf", "text/rtf", "application/vnd.oasis.opendocument.text", "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			var text string
			var metadata map[string]string
			var err error
			switch filetype {
			case ".docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
				text, metadata, err = docconv.ConvertDocx(reader)
			case ".rtf", ".rtfd", "text/rtf":
				buf, err := rtftxt.Text(reader)
				if err != nil {
					return nil, err
				}
				text = buf.String()
			case ".odt", "application/vnd.oasis.opendocument.text":
				text, metadata, err = docconv.ConvertODT(reader)
			}

			if err != nil {
				return nil, err
			}

			docs, err := FromLangchain(lcgodocloaders.NewText(strings.NewReader(text))).Load(ctx)
			if err != nil {
				return nil, err
			}

			for _, doc := range docs {
				m := map[string]any{}
				for k, v := range metadata {
					m[k] = v
				}
				doc.Metadata = m
			}

			return docs, nil
		}
	default:
		slog.Debug("Unsupported file type", "type", filetype)
		return nil
	}
}
