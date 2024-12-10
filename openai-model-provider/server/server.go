package server

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gptscript-ai/chat-completion-client"
)

func Run(apiKey, port string) error {
	mux := http.NewServeMux()

	s := &server{
		apiKey: apiKey,
		port:   port,
	}

	mux.HandleFunc("/healthz", s.healthz)
	mux.Handle("GET /v1/models", &httputil.ReverseProxy{
		Director:       s.proxy,
		ModifyResponse: s.rewriteModelsResponse,
	})
	mux.Handle("/{path...}", &httputil.ReverseProxy{
		Director: s.proxy,
	})

	httpServer := &http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: mux,
	}

	if err := httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}

type server struct {
	apiKey, port string
}

func (s *server) healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("http://127.0.0.1:" + s.port))
}

func (s *server) rewriteModelsResponse(resp *http.Response) error {
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	originalBody := resp.Body
	defer originalBody.Close()

	if resp.Header.Get("Content-Encoding") == "gzip" {
		var err error
		originalBody, err = gzip.NewReader(originalBody)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer originalBody.Close()
		resp.Header.Del("Content-Encoding")
	}

	var models openai.ModelsList
	if err := json.NewDecoder(originalBody).Decode(&models); err != nil {
		return fmt.Errorf("failed to decode models response: %w, %d, %v", err, resp.StatusCode, resp.Header)
	}

	for i, model := range models.Models {
		if model.Metadata == nil {
			model.Metadata = make(map[string]string)
		}

		switch {
		case strings.Contains(model.ID, "-realtime-"):
			// Do nothing
		case strings.HasPrefix(model.ID, "gpt-") || strings.HasPrefix(model.ID, "ft:gpt-") || strings.HasPrefix(model.ID, "o1-") || strings.HasPrefix(model.ID, "ft:o1-"):
			model.Metadata["usage"] = "llm"
		case strings.HasPrefix(model.ID, "text-embedding-") || strings.HasPrefix(model.ID, "ft:text-embedding-"):
			model.Metadata["usage"] = "text-embedding"
		case strings.HasPrefix(model.ID, "dall-e") || strings.HasPrefix(model.ID, "ft:dall-e"):
			model.Metadata["usage"] = "image-generation"
		}

		models.Models[i] = model
	}

	b, err := json.Marshal(models)
	if err != nil {
		return fmt.Errorf("failed to marshal models response: %w", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	return nil
}

func (s *server) proxy(req *http.Request) {
	req.URL.Host = "api.openai.com"
	req.URL.Scheme = "https"
	req.Host = req.URL.Host

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
}
