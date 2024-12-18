package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
)

func Run(apiKey, port string) error {
	endpoint := os.Getenv("OBOT_VLLM_MODEL_PROVIDER_ENDPOINT")
	if endpoint == "" {
		return fmt.Errorf("OBOT_VLLM_MODEL_PROVIDER_ENDPOINT environment variable not set")
	}

	mux := http.NewServeMux()

	s := &server{
		apiKey:   apiKey,
		port:     port,
		endpoint: endpoint,
	}

	mux.HandleFunc("/{$}", s.healthz)
	mux.Handle("GET /v1/models", &httputil.ReverseProxy{
		Director: s.proxy,
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
	apiKey, port, endpoint string
}

func (s *server) healthz(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("http://127.0.0.1:" + s.port))
}

func (s *server) proxy(req *http.Request) {
	req.URL.Host = s.endpoint
	req.URL.Scheme = "https"
	req.Host = req.URL.Host

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
}
