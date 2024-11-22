package mupdf

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	mdconv "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
	"github.com/PuerkitoBio/goquery"
	"github.com/gen2brain/go-fitz"
	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"github.com/pkoukk/tiktoken-go"
	"golang.org/x/sync/errgroup"
)

// Compile time check to ensure PDF satisfies the DocumentLoader interface.
var _ types.DocumentLoader = (*PDF)(nil)

var MuPDFLock sync.Mutex

type PDFOptions struct {
	// Password for encrypted PDF files.
	Password string

	// Page number to start loading from (default is 1).
	StartPage uint

	// Number of goroutines to load pdf documents
	NumThread int

	// EnablePageMerge
	EnablePageMerge bool

	// ChunkSize - maximum number of tokens allowed in a single document
	ChunkSize int

	ChunkOverlap int

	// TokenEncoding - encoding for Tokenizer to use for page merging
	TokenEncoding string

	// Tokenizer - target model for Tokenizer to use for page merging
	TokenModel string
}

// WithConfig sets the PDF loader configuration.
func WithConfig(config PDFOptions) func(o *PDFOptions) {
	return func(o *PDFOptions) {
		*o = config
	}
}

func WithDisablePageMerge() func(o *PDFOptions) {
	return func(o *PDFOptions) {
		o.EnablePageMerge = false
	}
}

// PDF represents a PDF Document loader that implements the DocumentLoader interface.
type PDF struct {
	Opts      PDFOptions
	Document  *fitz.Document
	Converter *mdconv.Converter
	Lock      *sync.Mutex
	Tokenizer *tiktoken.Tiktoken
}

// NewPDF creates a new PDF loader with the given options.
func NewPDF(r io.Reader, optFns ...func(o *PDFOptions)) (*PDF, error) {
	doc, err := fitz.NewFromReader(r)
	if err != nil {
		return nil, err
	}

	opts := PDFOptions{
		StartPage:       1,
		EnablePageMerge: true,
		ChunkSize:       defaults.ChunkSizeTokens,
		ChunkOverlap:    defaults.ChunkOverlapTokens,
	}

	for _, fn := range optFns {
		fn(&opts)
	}

	var tk *tiktoken.Tiktoken
	if opts.TokenEncoding != "" {
		tk, err = tiktoken.GetEncoding(opts.TokenEncoding)
	} else if opts.TokenModel != "" {
		tk, err = tiktoken.EncodingForModel(opts.TokenModel)
	} else {
		tk, err = tiktoken.GetEncoding(defaults.TokenEncoding)
	}

	if opts.StartPage == 0 {
		opts.StartPage = 1
	}

	converter := mdconv.NewConverter(mdconv.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))

	if opts.NumThread == 0 {
		opts.NumThread = 100
	}

	return &PDF{
		Opts:      opts,
		Document:  doc,
		Converter: converter,
		Tokenizer: tk,
		Lock:      &sync.Mutex{},
	}, nil
}

// Load loads the PDF Document and returns a slice of vs.Document containing the page contents and metadata.
func (l *PDF) Load(ctx context.Context) ([]vs.Document, error) {
	docs := make([]vs.Document, l.Document.NumPage())

	var docTokenCounts []int
	if l.Opts.EnablePageMerge {
		docTokenCounts = make([]int, l.Document.NumPage())
	}
	numPages := l.Document.NumPage()

	// We need a Lock here, since MuPDF is not thread-safe and there are some edge cases that can cause a CGO panic.
	// See https://github.com/gptscript-ai/knowledge/issues/135
	MuPDFLock.Lock()
	defer MuPDFLock.Unlock()
	g, childCtx := errgroup.WithContext(ctx)
	g.SetLimit(l.Opts.NumThread)
	for pageNum := 0; pageNum < numPages; pageNum++ {
		html, err := l.Document.HTML(pageNum, true)
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

				markdown, err := l.Converter.ConvertString(ret)
				if err != nil {
					return err
				}

				content := strings.TrimSpace(markdown)

				doc := vs.Document{
					Content: content,
					Metadata: map[string]any{
						"page":       pageNum + 1,
						"totalPages": numPages,
						"docIndex":   pageNum,
					},
				}

				l.Lock.Lock()
				docs[pageNum] = doc
				if l.Opts.EnablePageMerge {
					docTokenCounts[pageNum] = len(l.Tokenizer.Encode(content, []string{}, []string{"all"}))
				}
				l.Lock.Unlock()
				return nil
			}
		})
	}

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return l.mergePages(docs, docTokenCounts, numPages), nil
}

func (l *PDF) mergePages(docs []vs.Document, docTokenCounts []int, totalPages int) []vs.Document {
	if !l.Opts.EnablePageMerge {
		return docs
	}

	sizeLimit := l.Opts.ChunkSize - l.Opts.ChunkOverlap

	type pDoc struct {
		pageStart int
		pageEnd   int
		content   string
		tokens    int
	}

	var mergedDocs []vs.Document
	var currentDoc pDoc
	for i, doc := range docs {
		// If the current Document is empty, set it to the current Document and continue
		// TODO: (we just assume that it's impossible to exceed the token limit with a single page)
		if currentDoc.content == "" {
			currentDoc = pDoc{
				pageStart: i + 1,
				pageEnd:   i + 1,
				content:   doc.Content,
				tokens:    docTokenCounts[i],
			}
			continue
		}

		// Check if adding the next page will exceed the token limit
		// If it does, append the current document to the list and start over
		if currentDoc.tokens+docTokenCounts[i] > sizeLimit {
			// Append currentDoc to mergedDocs, as we reached the token limit
			mergedDocs = append(mergedDocs, vs.Document{
				Content: currentDoc.content,
				Metadata: map[string]any{
					"pages":                   fmt.Sprintf("%d-%d", currentDoc.pageStart, currentDoc.pageEnd),
					"totalPages":              totalPages,
					"tokenCount":              currentDoc.tokens,
					vs.DocMetadataKeyDocIndex: len(mergedDocs),
				},
			})
			// Start a new Document for the next pages
			currentDoc = pDoc{
				pageStart: i + 1,
				pageEnd:   i + 1,
				content:   doc.Content,
				tokens:    docTokenCounts[i],
			}
			continue
		}

		// If the token limit is not exceeded, append the content of the current page to the current Document
		currentDoc.content += "\n" + doc.Content
		currentDoc.tokens += docTokenCounts[i]
		currentDoc.pageEnd = i + 1
	}

	// Add any remaining content as a new Document
	if currentDoc.content != "" {
		mergedDocs = append(mergedDocs, vs.Document{
			Content: currentDoc.content,
			Metadata: map[string]any{
				"pages":                   fmt.Sprintf("%d-%d", currentDoc.pageStart, currentDoc.pageEnd),
				"totalPages":              totalPages,
				"tokenCount":              currentDoc.tokens,
				vs.DocMetadataKeyDocIndex: len(mergedDocs),
			},
		})
	}

	slog.Debug("Merged PDF pages", "totalPages", totalPages, "mergedPages", len(mergedDocs))

	return mergedDocs
}
