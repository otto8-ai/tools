package gopdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/ledongthuc/pdf"
)

/*
 * Credits to https://github.com/hupe1980/golc/blob/main/documentloader/pdf.go
 */

// Compile time check to ensure PDF satisfies the DocumentLoader interface.
var _ types.DocumentLoader = (*PDF)(nil)

type PDFOptions struct {
	// Password for encrypted PDF files.
	Password string

	// Page number to start loading from (default is 1).
	StartPage uint

	// Maximum number of pages to load (0 for all pages).
	MaxPages uint

	// Source is the name of the pdf document
	Source string

	// InterpreterConfig is the configuration for the PDF interpreter.
	InterpreterConfig *pdf.InterpreterConfig
}

// WithConfig sets the PDF loader configuration.
func WithConfig(config PDFOptions) func(o *PDFOptions) {
	return func(o *PDFOptions) {
		*o = config
	}
}

// WithInterpreterConfig sets the interpreter config for the PDF loader.
func WithInterpreterConfig(cfg pdf.InterpreterConfig) func(o *PDFOptions) {
	return func(o *PDFOptions) {
		o.InterpreterConfig = &cfg
	}
}

// WithInterpreterOpts sets the interpreter options for the PDF loader.
func WithInterpreterOpts(opts ...pdf.InterpreterOption) func(o *PDFOptions) {
	return func(o *PDFOptions) {
		if o.InterpreterConfig == nil {
			o.InterpreterConfig = &pdf.InterpreterConfig{}
		}
		for _, opt := range opts {
			opt(o.InterpreterConfig)
		}
	}
}

// PDF represents a PDF document loader that implements the DocumentLoader interface.
type PDF struct {
	f    io.ReaderAt
	size int64
	opts PDFOptions
}

func NewDefaultPDF(f io.Reader) (*PDF, error) {
	return NewPDFFromReader(f, WithInterpreterOpts(pdf.WithIgnoreDefOfNonNameVals([]string{"CMapName"})))
}

func NewPDFFromReader(f io.Reader, optFns ...func(o *PDFOptions)) (*PDF, error) {
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF data: %w", err)
	}
	return NewPDF(bytes.NewReader(data), int64(len(data)), optFns...)
}

// NewPDF creates a new PDF loader with the given options.
func NewPDF(f io.ReaderAt, size int64, optFns ...func(o *PDFOptions)) (*PDF, error) {
	opts := PDFOptions{
		StartPage: 1,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	if opts.StartPage == 0 {
		opts.StartPage = 1
	}

	return &PDF{
		f:    f,
		size: size,
		opts: opts,
	}, nil
}

// NewPDFFromFile creates a new PDF loader with the given options.
func NewPDFFromFile(f *os.File, optFns ...func(o *PDFOptions)) (*PDF, error) {
	opts := PDFOptions{
		StartPage: 1,
		Source:    f.Name(),
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	if opts.StartPage == 0 {
		opts.StartPage = 1
	}

	finfo, err := f.Stat()
	if err != nil {
		return nil, err
	}

	return NewPDF(f, finfo.Size(), func(o *PDFOptions) {
		*o = opts
	})
}

// Load loads the PDF document and returns a slice of vs.Document containing the page contents and metadata.
func (l *PDF) Load(ctx context.Context) ([]vs.Document, error) {
	var (
		reader *pdf.Reader
		err    error
	)

	if l.opts.Password != "" {
		reader, err = pdf.NewReaderEncrypted(l.f, l.size, func() string {
			return l.opts.Password
		})
		if err != nil {
			return nil, err
		}
	} else {
		reader, err = pdf.NewReader(l.f, l.size)
		if err != nil {
			return nil, err
		}
	}

	numPages := reader.NumPage()
	if l.opts.StartPage > uint(numPages) {
		return nil, fmt.Errorf("startpage out of page range: 1-%d", numPages)
	}

	maxPages := numPages - int(l.opts.StartPage) + 1
	if l.opts.MaxPages > 0 && numPages > int(l.opts.MaxPages) {
		maxPages = int(l.opts.MaxPages)
	}

	docs := make([]vs.Document, 0, numPages)

	fonts := make(map[string]*pdf.Font)

	page := 1

	for i := int(l.opts.StartPage); i < maxPages+int(l.opts.StartPage); i++ {
		p := reader.Page(i)

		for _, name := range p.Fonts() {
			if _, ok := fonts[name]; !ok {
				f := p.Font(name)
				fonts[name] = &f
			}
		}

		if l.opts.InterpreterConfig == nil {
			l.opts.InterpreterConfig = &pdf.InterpreterConfig{}
		}

		text, err := p.GetPlainText(fonts, pdf.WithInterpreterConfig(*l.opts.InterpreterConfig))
		if err != nil {
			return nil, err
		}

		// add the document to the doc list
		doc := vs.Document{
			Content: strings.TrimSpace(text),
			Metadata: map[string]any{
				"page":       page,
				"totalPages": maxPages,
			},
		}

		if l.opts.Source != "" {
			doc.Metadata["source"] = l.opts.Source
		}

		docs = append(docs, doc)

		page++
	}

	return docs, nil
}

// LoadAndSplit loads PDF documents from the provided reader and splits them using the specified text splitter.
func (l *PDF) LoadAndSplit(ctx context.Context, splitter types.TextSplitter) ([]vs.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}

	return splitter.SplitDocuments(docs)
}
