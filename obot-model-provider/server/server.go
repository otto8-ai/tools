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
	"time"

	"github.com/gptscript-ai/chat-completion-client"
)

func Run(obotHost, port string) error {
	mux := http.NewServeMux()

	s := &server{
		obotHost: obotHost,
		port:     port,
	}

	mux.HandleFunc("/{$}", s.healthz)
	mux.Handle("GET /v1/models", &httputil.ReverseProxy{
		Director:       s.proxy("/api"),
		ModifyResponse: s.rewriteModelsResponse,
	})
	mux.Handle("/{path...}", &httputil.ReverseProxy{
		Director: s.proxy("/api/llm-proxy"),
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
	obotHost, port string
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

	var models modelList
	if err := json.NewDecoder(originalBody).Decode(&models); err != nil {
		return fmt.Errorf("failed to decode models response: %w, %d, %v", err, resp.StatusCode, resp.Header)
	}

	respModels := make([]openai.Model, 0, len(models.Items))
	var createdTimestamp int64
	for _, model := range models.Items {
		createdTimestamp = 0
		if created, ok := model["created"].(string); ok {
			if createdAt, err := time.Parse(time.RFC3339, created); err == nil {
				createdTimestamp = createdAt.Unix()
			}
		}
		respModels = append(respModels, openai.Model{
			CreatedAt: createdTimestamp,
			ID:        model["id"].(string),
			Object:    "model",
			Metadata: map[string]string{
				"usage": model["usage"].(string),
			},
		})
	}

	b, err := json.Marshal(openai.ModelsList{Models: respModels})
	if err != nil {
		return fmt.Errorf("failed to marshal models response: %w", err)
	}

	resp.Body = io.NopCloser(bytes.NewReader(b))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(b)))
	return nil
}

func (s *server) proxy(prefix string) func(req *http.Request) {
	return func(req *http.Request) {
		req.URL.Host = s.obotHost
		req.URL.Scheme = "http"
		req.Host = req.URL.Host
		req.URL.Path = prefix + strings.TrimPrefix(req.URL.Path, "/v1")

		if apiKey := getAPIKey(req); apiKey != "" {
			req.Header.Set("Authorization", "Bearer "+apiKey)
		}
	}
}

func getAPIKey(req *http.Request) string {
	envHeader := req.Header.Get("X-GPTScript-Env")
	if envHeader == "" {
		return ""
	}

	for _, env := range strings.Split(envHeader, ",") {
		if strings.HasPrefix(env, "GPTSCRIPT_MODEL_PROVIDER_TOKEN=") {
			return strings.TrimSpace(strings.TrimPrefix(env, "GPTSCRIPT_MODEL_PROVIDER_TOKEN="))
		}
	}

	return ""
}

type modelList struct {
	Items []map[string]any `json:"items"`
}
