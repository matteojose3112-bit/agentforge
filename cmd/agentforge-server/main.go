package main

import (
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
		server.SetKeepAlivesEnabled(false)
		if err := server.Shutdown(shutdownContext()); err != nil {
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

func shutdownContext() interface{ Done() <-chan struct{} } {
	// Keep the command dependency-free; the server's own shutdown timeout is
	// bounded by the platform process termination window.
	timer := time.NewTimer(8 * time.Second)
	ch := make(chan struct{})
	go func() {
		<-timer.C
		close(ch)
	}()
	return shutdownCtx{done: ch}
}

type shutdownCtx struct{ done <-chan struct{} }
func (c shutdownCtx) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c shutdownCtx) Done() <-chan struct{} { return c.done }
func (c shutdownCtx) Err() error { return nil }
func (c shutdownCtx) Value(any) any { return nil }
