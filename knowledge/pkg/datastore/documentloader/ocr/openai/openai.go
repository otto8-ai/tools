//go:build !(linux && arm64) && !(windows && arm64)

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
	"net/http"

	"github.com/acorn-io/z"
	"github.com/gen2brain/go-fitz"
	"github.com/gptscript-ai/knowledge/pkg/datastore/embeddings/openai"
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

func (o *OpenAIOCR) Load(ctx context.Context, reader io.Reader) ([]vs.Document, error) {
	if o.Prompt == "" {
		o.Prompt = `What is in this image? If it's a pure text page, try to return it verbatim.
Don't add any additional text as the output will be used for a retrieval pipeline later on.
Leave out introductory sentences like "The image seems to contain...", etc.
For images and tabular data, try to describe the content in a way that it's useful for retrieval later on.
If you identify a specific page type, like book cover, table of contents, etc., please add that information to the beginning of the text.
`
	}

	if o.BaseURL == "" {
		o.BaseURL = "https://api.openai.com/v1"
	}

	if o.APIKey == "" {
		return nil, fmt.Errorf("OpenAI API key is required for OpenAI OCR")
	}

	if o.Concurrency == 0 {
		o.Concurrency = 3
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
			base64Image, err := encodeImageToBase64(img)
			if err != nil {
				return fmt.Errorf("error encoding image to base64: %w", err)
			}

			result, err := o.sendImageToOpenAI(base64Image)
			if err != nil {
				return fmt.Errorf("error sending image to OpenAI: %w", err)
			}

			slog.Debug("OpenAI OCR result", "page", pageNo, "result", result)

			docs = append(docs, vs.Document{
				Metadata: map[string]interface{}{
					"page":       pageNo,
					"totalPages": len(images),
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

func encodeImageToBase64(img image.Image) (string, error) {
	var buffer bytes.Buffer
	err := png.Encode(&buffer, img)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}

func (o *OpenAIOCR) sendImageToOpenAI(base64Image string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", o.BaseURL)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + o.APIKey,
	}

	if o.MaxTokens == nil {
		o.MaxTokens = z.Pointer(300)
	}

	if o.Model == "" {
		o.Model = "gpt-4o"
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

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	return result.Choices[0].Message.Content, nil
}
