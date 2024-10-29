package documentloader

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/pdf/gopdf"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/structured"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"

	golcdocloaders "github.com/hupe1980/golc/documentloader"
	"github.com/lu4p/cat"
	"github.com/mitchellh/mapstructure"
	lcgodocloaders "github.com/tmc/langchaingo/documentloaders"
)

func GetDocumentLoaderConfig(name string) (any, error) {
	switch name {
	case "plaintext":
		return nil, nil
	case "markdown":
		return nil, nil
	case "html":
		return nil, nil
	case "pdf", "gopdf":
		return gopdf.PDFOptions{}, nil
	case "ocr_openai":
		if OpenAIOCRConfig == nil {
			return nil, fmt.Errorf("OpenAI OCR is not available")
		}
		return OpenAIOCRConfig, nil
	case "mupdf":
		if MuPDFConfig == nil {
			return nil, fmt.Errorf("MuPDF is not available")
		}
		return MuPDFConfig, nil
	case "csv":
		return golcdocloaders.CSVOptions{}, nil
	case "notebook":
		return golcdocloaders.NotebookOptions{}, nil
	case "structured":
		return structured.Structured{}, nil
	default:
		return nil, fmt.Errorf("unknown document loader %q", name)
	}
}

type LoaderFunc func(ctx context.Context, reader io.Reader) ([]vs.Document, error)

var MuPDFGetter func(config any) (LoaderFunc, error) = nil
var MuPDFConfig any

var OpenAIOCRGetter func(config any) (LoaderFunc, error) = nil
var OpenAIOCRConfig any

func GetDocumentLoaderFunc(name string, config any) (LoaderFunc, error) {
	switch name {
	case "plaintext", "markdown":
		if config != nil {
			return nil, fmt.Errorf("plaintext/markdown document loader does not accept configuration")
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromLangchain(lcgodocloaders.NewText(reader)).Load(ctx)
		}, nil
	case "html":
		if config != nil {
			return nil, fmt.Errorf("html document loader does not accept configuration")
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromLangchain(lcgodocloaders.NewHTML(reader)).Load(ctx)
		}, nil
	case "ocr_openai":
		if OpenAIOCRGetter == nil {
			return nil, fmt.Errorf("OpenAI OCR is not available")
		}
		return OpenAIOCRGetter(config)
	case "mupdf":
		if MuPDFGetter == nil {
			return nil, fmt.Errorf("MuPDF is not available")
		}
		return MuPDFGetter(config)
	case "pdf", "gopdf":
		var pdfConfig gopdf.PDFOptions
		if config != nil {
			slog.Debug("PDF custom config", "config", config)
			if err := mapstructure.Decode(config, &pdfConfig); err != nil {
				return nil, fmt.Errorf("failed to decode PDF document loader configuration: %w", err)
			}
			slog.Debug("PDF custom config (decoded)", "pdfConfig", pdfConfig)
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			r, err := gopdf.NewPDFFromReader(reader, gopdf.WithConfig(pdfConfig))
			if err != nil {
				slog.Error("Failed to create PDF loader", "error", err)
				return nil, err
			}
			return r.Load(ctx)
		}, nil

	case "csv":
		var csvConfig golcdocloaders.CSVOptions
		if config != nil {
			if err := mapstructure.Decode(config, &csvConfig); err != nil {
				return nil, fmt.Errorf("failed to decode CSV document loader configuration: %w", err)
			}
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			docs, err := FromGolc(golcdocloaders.NewCSV(reader, func(o *golcdocloaders.CSVOptions) {
				*o = csvConfig
			})).Load(ctx)
			if err != nil && errors.Is(err, csv.ErrBareQuote) {
				oerr := err
				err = nil
				var nerr error
				docs, nerr = FromGolc(golcdocloaders.NewCSV(reader, func(o *golcdocloaders.CSVOptions) {
					*o = csvConfig
					o.LazyQuotes = true
				})).Load(ctx)
				if nerr != nil {
					err = errors.Join(oerr, nerr)
				}
			}
			return docs, err
		}, nil
	case "notebook":
		var nbConfig golcdocloaders.NotebookOptions
		if config != nil {
			if err := mapstructure.Decode(config, &nbConfig); err != nil {
				return nil, fmt.Errorf("failed to decode Notebook document loader configuration: %w", err)
			}
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			return FromGolc(golcdocloaders.NewNotebook(reader, func(o *golcdocloaders.NotebookOptions) {
				*o = nbConfig
			})).Load(ctx)
		}, nil
	case "document": // doc, docx, odt, rtf
		if config != nil {
			return nil, fmt.Errorf("'document' document loader does not accept configuration")
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			data, err := io.ReadAll(reader)
			if err != nil {
				return nil, fmt.Errorf("failed to read data: %w", err)
			}
			text, err := cat.FromBytes(data)
			if err != nil {
				return nil, fmt.Errorf("failed to extract text from document: %w", err)
			}
			return FromLangchain(lcgodocloaders.NewText(strings.NewReader(text))).Load(ctx)
		}, nil
	case "structured":
		var structuredCfg structured.Structured
		if config != nil {
			if err := mapstructure.Decode(config, &structuredCfg); err != nil {
				return nil, fmt.Errorf("failed to decode structured document loader configuration: %w", err)
			}
		}
		return structuredCfg.Load, nil
	default:
		return nil, fmt.Errorf("unknown document loader %q", name)
	}
}
