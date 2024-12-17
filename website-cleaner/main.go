package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gptscript-ai/go-gptscript"
	"github.com/sirupsen/logrus"

	md "github.com/JohannesKaufmann/html-to-markdown/v2/converter"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/base"
	"github.com/JohannesKaufmann/html-to-markdown/v2/plugin/commonmark"
)

var tagsToRemove = []string{
	"script, style, noscript, meta, head",
	"header", "footer", "nav", "aside", ".header", ".top", ".navbar", "#header",
	".footer", ".bottom", "#footer", ".sidebar", ".side", ".aside", "#sidebar",
	".modal", ".popup", "#modal", ".overlay", ".ad", ".ads", ".advert", "#ad",
	".lang-selector", ".language", "#language-selector", ".social", ".social-media",
	".social-links", "#social", ".menu", ".navigation", "#nav", ".breadcrumbs",
	"#breadcrumbs", "#search-form", ".search", "#search", ".share", "#share",
	".widget", "#widget", ".cookie", "#cookie",
}

func main() {
	input := os.Getenv("INPUT")
	output := os.Getenv("OUTPUT")

	logOut := logrus.New()
	logOut.SetOutput(os.Stdout)
	logOut.SetFormatter(&logrus.JSONFormatter{})
	logErr := logrus.New()
	logErr.SetOutput(os.Stderr)

	ctx := context.Background()
	gptscriptClient, err := gptscript.NewGPTScript()
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to create gptscript client, error: %v", err)).Error()
		os.Exit(0)
	}

	if input == "" {
		logOut.WithError(fmt.Errorf("input is empty")).Error()
		os.Exit(0)
	}

	if output == "" {
		logOut.WithError(fmt.Errorf("output is empty")).Error()
		os.Exit(0)
	}

	inputFile, err := gptscriptClient.ReadFileInWorkspace(ctx, input)
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to read input file %q: %v", input, err)).Error()
		os.Exit(0)
	}

	originalSize := len(inputFile)

	doc, err := goquery.NewDocumentFromReader(bytes.NewBuffer(inputFile))
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to parse html input: %v", err)).Error()
		os.Exit(0)
	}

	// Clean HTML programmatically
	for _, tag := range tagsToRemove {
		doc.Find(tag).Remove()
	}

	// transform to Markdown
	converter := md.NewConverter(md.WithPlugins(base.NewBasePlugin(), commonmark.NewCommonmarkPlugin()))
	html, err := doc.Html()
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to get html from document: %v", err)).Error()
		os.Exit(0)
	}

	sanitizedHTMLSize := len(html)

	markdown, err := converter.ConvertString(html)
	if err != nil {
		logOut.WithError(fmt.Errorf("failed to convert html to markdown: %v", err)).Error()
		os.Exit(0)
	}

	markdownSize := len(markdown)

	if markdownSize < 100_000 {
		logErr.Info("Markdown content is less than 100,000 characters -> sending to LLM for further cleanup")
		markdown, err = llmCleaning(ctx, logOut, logErr, markdown)
		if err != nil {
			logOut.WithError(fmt.Errorf("failed to LLM-clean markdown: %v", err)).Error()
			os.Exit(0)
		}
	}

	finalSize := len(markdown)
	logErr.Infof("[%s] Original HTML size: %d, Sanitized HTML size: %d, Converted Markdown size: %d, Final Markdown size: %d", input, originalSize, sanitizedHTMLSize, markdownSize, finalSize)

	if err := gptscriptClient.WriteFileInWorkspace(ctx, output, []byte(markdown)); err != nil {
		logOut.WithError(fmt.Errorf("failed to write output file %q: %v", output, err)).Error()
		os.Exit(0)
	}

	logErr.Infof("Output written to %s", output)

}

func llmCleaning(ctx context.Context, logOut, logErr *logrus.Logger, markdown string) (string, error) {
	prompt := "The following content is a scraped webpage converted to markdown. Please remove any content that came from the website header, footer, or navigation. The output should focus on just the main content body of the page. Maintain the markdown format, including any links or images.\n\n" + markdown

	url := fmt.Sprintf("%s/chat/completions", os.Getenv("OPENAI_BASE_URL"))

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + os.Getenv("OPENAI_API_KEY"),
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	model := os.Getenv("OBOT_DEFAULT_LLM_MINI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	logErr.Infof("Sending request to %s - using model %s", url, model)

	payload := Payload{
		Model: model,
		Messages: []Message{
			{
				Role: "user",
				Content: []MessageContent{
					{Type: "text", Text: prompt},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return markdown, fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return markdown, fmt.Errorf("failed to create request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{
		Timeout: 2 * time.Minute, // per request timeout - the overall timeout is set on the context
	}

	body, err := requestWithExponentialBackoff(ctx, client, req, 5, true, logErr)
	if err != nil {
		return markdown, fmt.Errorf("error sending request(s): %v", err)
	}

	var result Response
	if err := json.Unmarshal(body, &result); err != nil {
		return markdown, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(result.Choices) == 0 {
		return markdown, fmt.Errorf("no choices in response")
	}

	return result.Choices[0].Message.Content, nil
}

type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
}

type Message struct {
	Role    string           `json:"role"`
	Content []MessageContent `json:"content"`
}

type Payload struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens,omitempty"`
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

func requestWithExponentialBackoff(ctx context.Context, client *http.Client, req *http.Request, maxRetries int, handleRateLimit bool, logger *logrus.Logger) ([]byte, error) {
	const baseDelay = time.Millisecond * 200
	var resp *http.Response
	var err error

	// Save the original request body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read request body: %v", err)
		}
	}

	var failures []string
	for i := 0; i < maxRetries; i++ {
		// Check if context was canceled (timeout) before retrying
		if ctx.Err() != nil {
			failures = append(failures, fmt.Sprintf("[!] Stopped by canceled context after try #%d/%d: %v", i, maxRetries, ctx.Err()))
			break
		}

		// Reset body to the original request body
		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		resp, err = client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				// Log the error and retry for transient error reading response body
				msg := fmt.Sprintf("#%d/%d: failed to read response body: %v", i+1, maxRetries, err)
				logger.Warn("Request failed - Retryable", "error", msg)
				failures = append(failures, msg)
				_ = resp.Body.Close()
				continue
			}

			return body, resp.Body.Close()
		}

		if resp != nil {
			var bodystr string
			if resp.Body != nil {
				body, rerr := io.ReadAll(resp.Body)
				if rerr == nil {
					bodystr = string(body)
				}
				_ = resp.Body.Close()
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
				// Non-retryable error
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

	logger.Error("request retry limit exceeded or failed with non-retryable error(s)", "request", req)

	return nil, fmt.Errorf("retry limit (%d) exceeded or failed with non-retryable error(s): %v", maxRetries, strings.Join(failures, "; "))
}
