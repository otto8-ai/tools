package documentloader

import (
	"context"
	"fmt"
	"io"
	"log/slog"

	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/ocr/openai"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/pdf/defaults"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/pdf/mupdf"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/smartpdf"
	"github.com/gptscript-ai/knowledge/pkg/output"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/mitchellh/mapstructure"
)

func init() {
	defaults.DefaultPDFReaderFunc = func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
		slog.Debug("Default PDF Reader is MuPDF")
		r, err := mupdf.NewPDF(reader)
		if err != nil {
			slog.Error("Failed to create MuPDF loader", "error", err)
			return nil, err
		}
		return r.Load(ctx)
	}

	MuPDFGetter = func(config any) (LoaderFunc, error) {
		var opts []func(*mupdf.PDFOptions)
		if config != nil {
			var pdfConfig mupdf.PDFOptions
			slog.Debug("PDF custom config", "config", config)
			if err := mapstructure.Decode(config, &pdfConfig); err != nil {
				return nil, fmt.Errorf("failed to decode PDF document loader configuration: %w", err)
			}
			slog.Debug("PDF custom config (decoded)", "pdfConfig", pdfConfig)
			opts = append(opts, mupdf.WithConfig(pdfConfig))
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			r, err := mupdf.NewPDF(reader, opts...)
			if err != nil {
				slog.Error("Failed to create PDF loader", "error", err)
				return nil, err
			}
			return r.Load(ctx)
		}, nil
	}

	MuPDFConfig = mupdf.PDFOptions{}

	// OpenAI OCR (depends on MuPDF)
	OpenAIOCRGetter = func(config any) (LoaderFunc, error) {
		var openAIOCR openai.OpenAIOCR
		if config != nil {
			if err := mapstructure.Decode(config, &openAIOCR); err != nil {
				return nil, fmt.Errorf("failed to decode OpenAI OCR configuration: %w", err)
			}
			slog.Debug("OpenAI OCR custom config (decoded)", "openAIOCR", output.RedactSensitive(openAIOCR))
		}
		return openAIOCR.Load, nil
	}

	OpenAIOCRConfig = openai.OpenAIOCR{}

	// SmartPDF (depends on MuPDF and OpenAI OCR)
	SmartPDFGetter = func(config any) (LoaderFunc, error) {
		var smartPDFConfig smartpdf.SmartPDFConfig
		if config != nil {
			if err := mapstructure.Decode(config, &smartPDFConfig); err != nil {
				return nil, fmt.Errorf("failed to decode SmartPDF configuration: %w", err)
			}
			slog.Debug("SmartPDF custom config (decoded)", "smartPDFConfig", output.RedactSensitive(smartPDFConfig))
		}
		return func(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
			smartPDF, err := smartpdf.NewSmartPDF(reader, smartPDFConfig)
			if err != nil {
				slog.Error("Failed to create SmartPDF loader", "error", err)
				return nil, err
			}
			return smartPDF.Load(ctx)
		}, nil
	}

	SmartPDFConfig = smartpdf.SmartPDFConfig{}
}
