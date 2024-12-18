package main

import (
	"os"

	"github.com/obot-platform/tools/obot-model-provider/server"
)

func main() {
	obotHost := os.Getenv("OBOT_URL")
	if obotHost == "" {
		obotHost = "localhost:8080"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	if err := server.Run(obotHost, port); err != nil {
		panic(err)
	}
}
