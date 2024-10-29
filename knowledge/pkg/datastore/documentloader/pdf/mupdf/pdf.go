//go:build !(linux && arm64) && !(windows && arm64)

package mupdf

import (
	"context"
	"io"
	"strings"
	"sync"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/gen2brain/go-fitz"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"golang.org/x/sync/errgroup"
)

// Compile time check to ensure PDF satisfies the DocumentLoader interface.
var _ types.DocumentLoader = (*PDF)(nil)

var mupdfLock sync.Mutex

type PDFOptions struct {
	// Password for encrypted PDF files.
	Password string

	// Page number to start loading from (default is 1).
	StartPage uint

	// Maximum number of pages to load (0 for all pages).
	MaxPages uint

	// Source is the name of the pdf document
	Source string

	// Number of goroutines to load pdf documents
	NumThread int
}

// WithConfig sets the PDF loader configuration.
func WithConfig(config PDFOptions) func(o *PDFOptions) {
	return func(o *PDFOptions) {
		*o = config
	}
}

// PDF represents a PDF document loader that implements the DocumentLoader interface.
type PDF struct {
	opts      PDFOptions
	document  *fitz.Document
	converter *md.Converter
	lock      *sync.Mutex
}

// NewPDFFromFile creates a new PDF loader with the given options.
func NewPDF(r io.Reader, optFns ...func(o *PDFOptions)) (*PDF, error) {
	doc, err := fitz.NewFromReader(r)
	if err != nil {
		return nil, err
	}
	opts := PDFOptions{
		StartPage: 1,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	if opts.StartPage == 0 {
		opts.StartPage = 1
	}

	converter := md.NewConverter("", true, nil)

	if opts.NumThread == 0 {
		opts.NumThread = 100
	}

	return &PDF{
		opts:      opts,
		document:  doc,
		converter: converter,
		lock:      &sync.Mutex{},
	}, nil
}

// Load loads the PDF document and returns a slice of vs.Document containing the page contents and metadata.
func (l *PDF) Load(ctx context.Context) ([]vs.Document, error) {
	docs := make([]vs.Document, 0, l.document.NumPage())
	numPages := l.document.NumPage()

	// We need a lock here, since MuPDF is not thread-safe and there are some edge cases that can cause a CGO panic.
	// See https://github.com/gptscript-ai/knowledge/issues/135
	mupdfLock.Lock()
	defer mupdfLock.Unlock()
	g, childCtx := errgroup.WithContext(ctx)
	g.SetLimit(l.opts.NumThread)
	for pageNum := 0; pageNum < numPages; pageNum++ {
		html, err := l.document.HTML(pageNum, true)
		if err != nil {
			return nil, err
		}
		g.Go(func() error {
			select {
			case <-childCtx.Done():
				return context.Canceled
			default:
				htmlDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
				if err != nil {
					return err
				}
				htmlDoc.Find("img").Remove()

				ret, err := htmlDoc.First().Html()
				if err != nil {
					return err
				}

				markdown, err := l.converter.ConvertString(ret)
				if err != nil {
					return err
				}

				doc := vs.Document{
					Content: strings.TrimSpace(markdown),
					Metadata: map[string]any{
						"page":       pageNum + 1,
						"totalPages": numPages,
					},
				}
				l.lock.Lock()
				docs = append(docs, doc)
				l.lock.Unlock()
				return nil
			}
		})
	}

	return docs, g.Wait()
}

// LoadAndSplit loads PDF documents from the provided reader and splits them using the specified text splitter.
func (l *PDF) LoadAndSplit(ctx context.Context, splitter types.TextSplitter) ([]vs.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}

	return splitter.SplitDocuments(docs)
}
