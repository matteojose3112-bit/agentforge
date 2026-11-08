package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/matteojose3112-bit/agentforge/internal/httpapi"
)

func main() {
	address := os.Getenv("AGENTFORGE_ADDR")
	if address == "" {
		address = ":" + envOr("PORT", "8080")
	}
	origins := strings.Split(os.Getenv("AGENTFORGE_ALLOWED_ORIGINS"), ",")
	api := httpapi.New(origins)
	server := httpapi.NewHTTPServer(address, api)

	stopped := make(chan os.Signal, 1)
	signal.Notify(stopped, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stopped
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown: %v", err)
		}
	}()

	log.Printf("AgentForge API listening on %s", address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
