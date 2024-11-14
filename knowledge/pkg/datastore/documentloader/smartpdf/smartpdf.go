package smartpdf

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/acorn-io/z"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/ocr/openai"
	"github.com/gptscript-ai/knowledge/pkg/datastore/documentloader/pdf/mupdf"
	"github.com/gptscript-ai/knowledge/pkg/datastore/types"
	"github.com/gptscript-ai/knowledge/pkg/log"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"golang.org/x/sync/errgroup"
)

// Compile time check to ensure PDF satisfies the DocumentLoader interface.
var _ types.DocumentLoader = (*SmartPDF)(nil)

type SmartPDFConfig struct {
	MuPDF           mupdf.PDFOptions `mapstructure:"muPDF" json:"muPDF"`
	OpenAIOCR       openai.OpenAIOCR `mapstructure:"openAIOCR" json:"openAIOCR"`
	FallbackOptions FallbackOptions  `mapstructure:"fallbackOptions" json:"fallbackOptions"`
}

type FallbackOptions struct {
	OnEmptyContent *bool `mapstructure:"onEmptyContent" json:"onEmptyContent,omitempty"`
	OnImageCount   int   `mapstructure:"onImageCount" json:"onImageCount,omitempty"`
	OnTable        *bool `mapstructure:"onTable" json:"onTable,omitempty"`
}

type SmartPDF struct {
	file  io.Reader
	mupdf *mupdf.PDF
	cfg   SmartPDFConfig
	lock  *sync.Mutex
}

func (f *FallbackOptions) SetDefaults() {
	if f.OnEmptyContent == nil {
		f.OnEmptyContent = z.Pointer(true)
	}
	if f.OnTable == nil {
		f.OnTable = z.Pointer(false)
	}
}

func NewSmartPDF(file io.Reader, cfg SmartPDFConfig) (*SmartPDF, error) {
	mpdf, err := mupdf.NewPDF(file, mupdf.WithConfig(cfg.MuPDF))
	if err != nil {
		return nil, err
	}

	cfg.FallbackOptions.SetDefaults()

	cfg.OpenAIOCR.Prompt = `Convert the content of the image into markdown format, ensuring the appropriate structure for various components including tables, lists, and other images. You will not add any of your own commentary to your response. Consider the following:

- **Tables:** If the image contains tables, convert them into markdown tables. Ensure that all columns and rows from the table are accurately captured. Do not convert tables into JSON unless every column and row, with all data, can be properly represented.
- **Lists:** If the image contains lists, convert them into markdown lists.
- **Images:** If the image contains other images, summarize each image into text and wrap it with ` + "`<image></image>`" + ` tags.

# Steps

1. **Image Analysis:** Identify the various elements in the image such as tables, lists, and other images.
   
2. **Markdown Conversion:**
- For tables, use the markdown format for tables. Make sure all columns and rows are preserved, including headers and any blank cells.
- For lists, use markdown list conventions (ordered or unordered as per the original).
- For images, write a brief descriptive summary of the image content and wrap it using ` + "`<image></image>`" + ` tags.

3. **Compile:** Assemble all converted elements into cohesive markdown-formatted text.

# Output Format

- The output should be in markdown format, accurately representing each element from the image with appropriate markdown syntax. Pay close attention to the structure of tables, ensuring that no columns or rows are omitted.

# Examples

**Input Example 1:**

An image containing a table with five columns and three rows, a list, and another image.

**Output Example 1:**

` + "```" + `
| Column 1 | Column 2 | Column 3 | Column 4 | Column 5 |
| -------- | -------- | -------- | -------- | -------- |
| Row 1    | Data 2   | Data 3   | Data 4   | Data 5   |
| Row 2    | Data 2   | Data 3   | Data 4   |          |
| Row 3    | Data 2   |          | Data 4   | Data 5   |

- List Item 1
- List Item 2
- List Item 3

<image></image>
Image description with as much detail as possible here.
</image>
` + "```" + `

# Notes

- Ensure that the markdown syntax is correct and renders well when processed.
- Preserve column and row structure for tables, ensuring no data is lost or misrepresented.
- Be attentive to the layout and order of elements as they appear in the image.
`

	if err := cfg.OpenAIOCR.Configure(); err != nil {
		return nil, fmt.Errorf("error configuring OpenAI OCR: %w", err)
	}

	return &SmartPDF{
		file:  file,
		mupdf: mpdf,
		cfg:   cfg,
		lock:  &sync.Mutex{},
	}, nil
}

func (s *SmartPDF) Name() string {
	return "smartpdf"
}

func (s *SmartPDF) Load(ctx context.Context) ([]vs.Document, error) {
	docs := make([]vs.Document, s.mupdf.Document.NumPage())

	var docTokenCounts []int
	if s.mupdf.Opts.EnablePageMerge {
		docTokenCounts = make([]int, s.mupdf.Document.NumPage())
	}
	numPages := s.mupdf.Document.NumPage()

	logger := log.FromCtx(ctx).With("loader", s.Name())
	ctx = log.ToCtx(ctx, logger)

	// We need a lock here, since MuPDF is not thread-safe and there are some edge cases that can cause a CGO panic.
	// See https://github.com/gptscript-ai/knowledge/issues/135
	mupdf.MuPDFLock.Lock()
	defer mupdf.MuPDFLock.Unlock()
	g, childCtx := errgroup.WithContext(ctx)
	g.SetLimit(s.mupdf.Opts.NumThread)
	for pageNum := 0; pageNum < numPages; pageNum++ {
		html, err := s.mupdf.Document.HTML(pageNum, true)
		if err != nil {
			return nil, err
		}
		g.Go(func() error {
			select {
			case <-childCtx.Done():
				return context.Canceled
			default:
				logger = log.FromCtx(childCtx).With("pdfPageNo", pageNum+1)

				htmlDoc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
				if err != nil {
					return err
				}
				imgs := htmlDoc.Find("img")
				imgCount := imgs.Length()
				imgs.Remove()

				ret, err := htmlDoc.First().Html()
				if err != nil {
					return err
				}

				markdown, err := s.mupdf.Converter.ConvertString(ret)
				if err != nil {
					return err
				}

				content := strings.TrimSpace(markdown)

				tableCount := htmlDoc.Find("table").Length()

				if (content == "" && z.Dereference(s.cfg.FallbackOptions.OnEmptyContent)) ||
					(s.cfg.FallbackOptions.OnImageCount > 0 && imgCount >= s.cfg.FallbackOptions.OnImageCount) ||
					(z.Dereference(s.cfg.FallbackOptions.OnTable) && tableCount > 0) {
					img, err := s.mupdf.Document.Image(pageNum)
					if err != nil {
						return fmt.Errorf("error getting image from PDF: %w", err)
					}
					logger = logger.With("smartpdf", "openaiOCR")
					gCtx := log.ToCtx(childCtx, logger)

					logger.Debug("MuPDF page did not meet conditions - falling back to VLM mode", "page", pageNum+1, "totalPages", numPages, "contentLen", len(content), "imgCount", imgCount, "tableCount", tableCount)
					base64Image, err := openai.EncodeImageToBase64(img)
					if err != nil {
						return fmt.Errorf("error encoding image to base64: %w", err)
					}

					result, err := s.cfg.OpenAIOCR.SendImageToOpenAI(gCtx, base64Image)
					if err != nil {
						return fmt.Errorf("error sending image to OpenAI: %w", err)
					}

					content = strings.TrimSpace(result)
				}

				doc := vs.Document{
					Content: content,
					Metadata: map[string]any{
						"page":       pageNum + 1,
						"totalPages": numPages,
					},
				}

				s.mupdf.Lock.Lock()
				docs[pageNum] = doc
				if s.mupdf.Opts.EnablePageMerge {
					docTokenCounts[pageNum] = len(s.mupdf.Tokenizer.Encode(content, []string{}, []string{"all"}))
				}
				s.mupdf.Lock.Unlock()
				return nil
			}
		})
	}

	err := g.Wait()
	if err != nil {
		return nil, err
	}

	return s.mergePages(docs, docTokenCounts, numPages), nil
}

func (s *SmartPDF) mergePages(docs []vs.Document, docTokenCounts []int, totalPages int) []vs.Document {
	if !s.mupdf.Opts.EnablePageMerge {
		return docs
	}

	sizeLimit := s.mupdf.Opts.ChunkSize - s.mupdf.Opts.ChunkOverlap

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
