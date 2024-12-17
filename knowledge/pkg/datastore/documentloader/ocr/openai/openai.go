package openai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/acorn-io/z"
	"github.com/gen2brain/go-fitz"
	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/openai"
	"github.com/gptscript-ai/knowledge/pkg/env"
	"github.com/gptscript-ai/knowledge/pkg/log"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

var OpenAIOCRAPITimeout = time.Duration(env.GetIntFromEnvOrDefault("KNOW_OPENAI_OCR_API_TIMEOUT_SECONDS", defaults.ModelAPITimeoutSeconds)) * time.Second
var OpenAIOCRAPIRequestTimeout = time.Duration(env.GetIntFromEnvOrDefault("KNOW_OPENAI_OCR_API_REQUEST_TIMEOUT_SECONDS", defaults.ModelAPIRequestTimeoutSeconds)) * time.Second

type OpenAIOCR struct {
	openai.OpenAIConfig `mapstructure:",squash"`
	Prompt              string
	MaxTokens           *int
	Concurrency         int
}

type ImagePayload struct {
	URL string `json:"url,omitempty"`
}

type MessageContent struct {
	Type     string        `json:"type"`
	Text     string        `json:"text,omitempty"`
	ImageURL *ImagePayload `json:"image_url,omitempty"`
}

type Message struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

type Payload struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type RespMessage struct {
	Content string `json:"content"`
}

type Choice struct {
	FinishReason string      `json:"finish_reason"`
	Message      RespMessage `json:"message"` // e.g. OpenAI provider resp
	Delta        RespMessage `json:"delta"`   // e.g. Anthropic/Claude provider StreamResponse
}

type Response struct {
	Choices []Choice `json:"choices"`
}

func (o *OpenAIOCR) Configure() error {
	if err := load.FillConfigEnv("OPENAI_", &o.OpenAIConfig); err != nil {
		return fmt.Errorf("error filling OpenAI config: %w", err)
	}

	if o.BaseURL == "" {
		o.BaseURL = "https://api.openai.com/v1"
	}

	if o.APIKey == "" {
		return fmt.Errorf("OpenAI API key is required for OpenAI OCR")
	}

	if o.Concurrency == 0 {
		o.Concurrency = 3
	}

	if o.MaxTokens == nil {
		o.MaxTokens = z.Pointer(defaults.ChunkSizeTokens - defaults.ChunkOverlapTokens)
	}

	if o.Model == "" {
		o.Model = "gpt-4o"
	}

	if o.Prompt == "" {
		o.Prompt = `What is in this image? If it's a pure text page, try to return it verbatim.
Don't add any additional text as the output will be used for a retrieval pipeline later on.
Leave out introductory sentences like "The image seems to contain...", etc.
For images and tabular data, try to describe the content in a way that it's useful for retrieval later on.
If you identify a specific page type, like book cover, table of contents, etc., please add that information to the beginning of the text.
`
	}

	return nil
}

func (o *OpenAIOCR) Load(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
	if o.Prompt == "" {
		o.Prompt = `Convert the content of the image into markdown format, ensuring the appropriate structure for various components including tables, lists, and other images. You will not add any of your own commentary to your response. Consider the following:

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
	}

	if err := o.Configure(); err != nil {
		return nil, fmt.Errorf("error configuring OpenAI OCR: %w", err)
	}

	// We don't pull this into the concurrent loop because we first want to make sure that the PDF can be converted to images completely
	// before firing off the requests to OpenAI
	images, err := convertPdfToImages(reader)
	if err != nil {
		return nil, fmt.Errorf("error converting PDF to images: %w", err)
	}

	docs := make([]vs.Document, len(images))

	sem := semaphore.NewWeighted(int64(o.Concurrency)) // limit max. concurrency

	g, ctx := errgroup.WithContext(ctx)

	logger := log.FromCtx(ctx)
	logger.Debug("Processing images", "totalPages", len(images))
	for i, img := range images {
		pageNo := i + 1

		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			base64Image, err := EncodeImageToBase64(img)
			if err != nil {
				return fmt.Errorf("error encoding image to base64: %w", err)
			}

			result, err := o.SendImageToOpenAI(ctx, base64Image)
			if err != nil {
				return fmt.Errorf("error sending image to OpenAI: %w", err)
			}

			docs = append(docs, vs.Document{
				Metadata: map[string]interface{}{
					"page":                    pageNo,
					"totalPages":              len(images),
					vs.DocMetadataKeyDocIndex: i,
				},
				Content: fmt.Sprintf("%v", result),
			})
			return nil
		})
	}
	return docs, g.Wait()
}

func convertPdfToImages(reader io.Reader) ([]image.Image, error) {
	doc, err := fitz.NewFromReader(reader)
	if err != nil {
		return nil, err
	}
	defer doc.Close()

	var images []image.Image
	for i := 0; i < doc.NumPage(); i++ {
		img, err := doc.Image(i)
		if err != nil {
			return nil, err
		}
		images = append(images, img)
	}
	return images, nil
}

func EncodeImageToBase64(img image.Image) (string, error) {
	var buffer bytes.Buffer
	err := png.Encode(&buffer, img)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func (o *OpenAIOCR) SendImageToOpenAI(ctx context.Context, base64Image string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", o.BaseURL)

	ctx = log.ToCtx(ctx, log.FromCtx(ctx).With("tool", "openai-ocr").With("ctxTimeout", OpenAIOCRAPITimeout))

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, OpenAIOCRAPITimeout)
	defer cancel()

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + o.APIKey,
	}

	payload := Payload{
		Model: o.Model,
		Messages: []Message{
			{
				Role: "user",
				Content: []MessageContent{
					{Type: "text", Text: o.Prompt},
					{Type: "image_url", ImageURL: &ImagePayload{URL: "data:image/png;base64," + base64Image}},
				},
			},
		},
		MaxTokens: *o.MaxTokens,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: OpenAIOCRAPIRequestTimeout, // per request timeout - the overall timeout is set on the context
	}
	// Send the request and get the body.
	body, err := openai.RequestWithExponentialBackoff(ctx, client, req, 5, true)
	if err != nil {
		return "", fmt.Errorf("OpenAI OCR error sending request(s): %w", err)
	}

	body = []byte(strings.TrimSpace(strings.TrimPrefix(string(body), "data: "))) // required e.g. for the anthropic/claude provider

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("error unmarshaling openai response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in OpenAI OCR response")
	}

	text := result.Choices[0].Message.Content
	if text == "" {
		text = result.Choices[0].Delta.Content
	}

	if text == "" {
		return "", fmt.Errorf("no content in OpenAI OCR response")
	}

	return text, nil
}
