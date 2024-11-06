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
	"log/slog"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/acorn-io/z"
	"github.com/gen2brain/go-fitz"
	"github.com/gptscript-ai/knowledge/pkg/datastore/defaults"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/load"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/openai"
	"github.com/gptscript-ai/knowledge/pkg/log"
	vs "github.com/gptscript-ai/knowledge/pkg/vectorstore/types"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type OpenAIOCR struct {
	openai.OpenAIConfig `mapstructure:",squash"`
	Prompt              string
	MaxTokens           *int
	Concurrency         int
}

type ImagePayload struct {
	URL string `json:"url"`
}

type MessageContent struct {
	Type     string       `json:"type"`
	Text     string       `json:"text,omitempty"`
	ImageURL ImagePayload `json:"image_url,omitempty"`
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
	Message      RespMessage `json:"message"`
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
		o.Prompt = `What is in this image? If it's a pure text page, try to return it verbatim.
Don't add any additional text as the output will be used for a retrieval pipeline later on.
Leave out introductory sentences like "The image seems to contain...", etc.
For images and tabular data, try to describe the content in a way that it's useful for retrieval later on.
If you identify a specific page type, like book cover, table of contents, etc., please add that information to the beginning of the text.
`
	}

	if err := load.FillConfigEnv("OPENAI_", &o.OpenAIConfig); err != nil {
		return nil, fmt.Errorf("error filling OpenAI config: %w", err)
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

	for i, img := range images {
		pageNo := i + 1

		g.Go(func() error {
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)

			slog.Debug("Processing PDF image", "page", pageNo, "totalPages", len(images))
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

	ctx = log.ToCtx(ctx, log.FromCtx(ctx).With("tool", "openai-ocr"))

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
					{Type: "image_url", ImageURL: ImagePayload{URL: "data:image/png;base64," + base64Image}},
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

	client := &http.Client{}
	// Send the request and get the body.
	body, err := requestWithExponentialBackoff(ctx, client, req, 5, true)
	if err != nil {
		return "", fmt.Errorf("OpenAI OCR error sending request(s): %w", err)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Choices[0].Message.Content, nil
}

func requestWithExponentialBackoff(ctx context.Context, client *http.Client, req *http.Request, maxRetries int, handleRateLimit bool) ([]byte, error) {
	const baseDelay = time.Millisecond * 200
	var resp *http.Response
	var err error

	logger := log.FromCtx(ctx)

	var failures []string

	// Save the original request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		if err != nil {
			failures = append(failures, fmt.Sprintf("failed to read request body: %v", err))
			return nil, fmt.Errorf("failed to read request body: %v; failures: %v", err, strings.Join(failures, "; "))
		}
	}

	for i := 0; i < maxRetries; i++ {
		// Reset body to the original request body
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// Log the error and retry for transient error reading response body
				msg := fmt.Sprintf("#%d/%d: failed to read response body: %v", i+1, maxRetries, err)
				logger.Warn("Request failed - Retryable", "error", msg)
				failures = append(failures, msg)
				continue
			}

			return body, nil
		}

		if resp != nil {
			var bodystr string
			if resp.Body != nil {
				body, rerr := io.ReadAll(resp.Body)
				if rerr == nil {
					bodystr = string(body)
				}
				resp.Body.Close()
			}

			msg := fmt.Sprintf("#%d/%d: %d <%s> (err: %v)", i+1, maxRetries, resp.StatusCode, bodystr, err)
			failures = append(failures, msg)

			if resp.StatusCode >= 500 || (handleRateLimit && resp.StatusCode == http.StatusTooManyRequests) {
				logger.Warn("Request failed - Retryable", "error", msg)
				// Retry for 5xx (Server Errors)
				// We're also handling rate limit here (without checking the Retry-After header), if handleRateLimit is true,
				// since it's what e.g. OpenAI recommends (see https://github.com/openai/openai-cookbook/blob/457f4310700f93e7018b1822213ca99c613dbd1b/examples/How_to_handle_rate_limits.ipynb).
				delay := baseDelay * time.Duration(1<<i)
				jitter := time.Duration(rand.Int63n(int64(baseDelay)))
				time.Sleep(delay + jitter)
				continue
			} else {
				// Non-retriable error
				logger.Error("Request failed - Non-retryable", "error", msg)
				break
			}
		} else {
			// Log connection errors (client.Do error) and retry if needed
			msg := fmt.Sprintf("#%d/%d: failed to send request: %v", i+1, maxRetries, err)
			logger.Warn("Request failed - Retryable", "error", msg)
			failures = append(failures, msg)
		}
	}

	return nil, fmt.Errorf("retry limit (%d) exceeded or failed with non-retriable error(s): %v", maxRetries, strings.Join(failures, "; "))
}
