//go:build !(linux && arm64) && !(windows && arm64)

package mupdf

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	mdconv "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
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

	// EnablePageMerge
	EnablePageMerge bool

	// PageMergeTokenLimit - maximum number of tokens allowed in a single document
	PageMergeTokenLimit int

	// TokenEncoding - encoding for tokenizer to use for page merging
	TokenEncoding string

	// Tokenizer - target model for tokenizer to use for page merging
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

// PDF represents a PDF document loader that implements the DocumentLoader interface.
type PDF struct {
	opts                PDFOptions
	document            *fitz.Document
	converter           *mdconv.Converter
	lock                *sync.Mutex
	tokenizer           *tiktoken.Tiktoken
	contentTransformers []func(string) string
}

// NewPDF creates a new PDF loader with the given options.
func NewPDF(r io.Reader, optFns ...func(o *PDFOptions)) (*PDF, error) {
	doc, err := fitz.NewFromReader(r)
	if err != nil {
		return nil, err
	}

	opts := PDFOptions{
		StartPage:           1,
		EnablePageMerge:     true,
		PageMergeTokenLimit: defaults.TextSplitterChunkSize,
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
		tk, err = tiktoken.GetEncoding(defaults.TextSplitterTokenEncoding)
	}

	if opts.StartPage == 0 {
		opts.StartPage = 1
	}

	converter := mdconv.NewConverter(mdconv.WithPlugins(commonmark.NewCommonmarkPlugin()))

	if opts.NumThread == 0 {
		opts.NumThread = 100
	}

	contentTransformers := []func(string) string{
		func(s string) string {
			return strings.TrimSpace(strings.ReplaceAll(s, "\n\n", "\n"))
		},
	}

	return &PDF{
		opts:                opts,
		document:            doc,
		converter:           converter,
		tokenizer:           tk,
		lock:                &sync.Mutex{},
		contentTransformers: contentTransformers,
	}, nil
}

// Load loads the PDF document and returns a slice of vs.Document containing the page contents and metadata.
func (l *PDF) Load(ctx context.Context) ([]vs.Document, error) {
	docs := make([]vs.Document, l.document.NumPage())

	var docTokenCounts []int
	if l.opts.EnablePageMerge {
		docTokenCounts = make([]int, l.document.NumPage())
	}
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

				content := strings.TrimSpace(markdown)

				for _, transformer := range l.contentTransformers {
					content = transformer(content)
				}

				doc := vs.Document{
					Content: content,
					Metadata: map[string]any{
						"page":       pageNum + 1,
						"totalPages": numPages,
					},
				}

				l.lock.Lock()
				docs[pageNum] = doc
				if l.opts.EnablePageMerge {
					docTokenCounts[pageNum] = len(l.tokenizer.Encode(content, []string{}, []string{"all"}))
				}
				l.lock.Unlock()
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
	if !l.opts.EnablePageMerge {
		return docs
	}

	mergedDocs := make([]vs.Document, 0, len(docs))
	type pDoc struct {
		pageStart int
		pageEnd   int
		content   string
		tokens    int
	}
	var currentDoc pDoc
	for i, doc := range docs {
		// If the current document is empty, set it to the current document and continue
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
		if currentDoc.tokens+docTokenCounts[i] > l.opts.PageMergeTokenLimit {
			mergedDocs = append(mergedDocs, vs.Document{
				Content: currentDoc.content,
				Metadata: map[string]any{
					"pages":      fmt.Sprintf("%d-%d", currentDoc.pageStart, currentDoc.pageEnd),
					"totalPages": totalPages,
					"tokenCount": currentDoc.tokens,
				},
			})
			currentDoc = pDoc{
				pageStart: i + 1,
				pageEnd:   i + 1,
				content:   doc.Content,
				tokens:    docTokenCounts[i],
			}
			continue
		}

		// If the token limit is not exceeded, append the content of the current page to the current document
		currentDoc.content += "\n" + doc.Content
		currentDoc.tokens += docTokenCounts[i]
		currentDoc.pageEnd = i + 1

		// If this is the last page, append the current document to the list
		if i == len(docs)-1 {
			mergedDocs = append(mergedDocs, vs.Document{
				Content: currentDoc.content,
				Metadata: map[string]any{
					"pages":      fmt.Sprintf("%d-%d", currentDoc.pageStart, currentDoc.pageEnd),
					"totalPages": totalPages,
					"tokenCount": currentDoc.tokens,
				},
			})
		}
	}

	slog.Debug("Merged PDF pages", "totalPages", totalPages, "mergedPages", len(mergedDocs))

	return mergedDocs
}

// LoadAndSplit loads PDF documents from the provided reader and splits them using the specified text splitter.
func (l *PDF) LoadAndSplit(ctx context.Context, splitter types.TextSplitter) ([]vs.Document, error) {
	docs, err := l.Load(ctx)
	if err != nil {
		return nil, err
	}

	return splitter.SplitDocuments(docs)
}
