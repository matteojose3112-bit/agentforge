package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"crypto/subtle"
	"strings"
	"time"

	"github.com/matteojose3112-bit/agentforge/internal/tools"
)

const maxRequestBytes = 2 << 20

type Server struct {
	allowedOrigins map[string]struct{}
	apiToken       string
	mux            *http.ServeMux
}

type inspectRequest struct {
	CSV string `json:"csv"`
}

type reportRequest struct {
	CSV          string   `json:"csv"`
	StatusColumn string   `json:"status_column"`
	ValueColumn  string   `json:"value_column"`
	SourceColumn string   `json:"source_column"`
	WonStatuses  []string `json:"won_statuses"`
	LostStatuses []string `json:"lost_statuses"`
}

func New(allowedOrigins []string) *Server { return NewWithToken(allowedOrigins, "") }

// NewWithToken protects data-processing endpoints with an optional bearer token.
func NewWithToken(allowedOrigins []string, apiToken string) *Server {
	origins := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin = strings.TrimSpace(origin); origin != "" {
			origins[origin] = struct{}{}
		}
	}
	s := &Server{allowedOrigins: origins, apiToken: strings.TrimSpace(apiToken), mux: http.NewServeMux()}
	s.mux.HandleFunc("/healthz", s.health)
	s.mux.HandleFunc("/api/inspect", s.inspect)
	s.mux.HandleFunc("/api/report", s.report)
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	origin := r.Header.Get("Origin")
	if origin != "" {
		if _, ok := s.allowedOrigins[origin]; !ok {
			http.Error(w, "origin not allowed", http.StatusForbidden)
			return
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Vary", "Origin")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/") && s.apiToken != "" {
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if subtle.ConstantTimeCompare([]byte(provided), []byte(s.apiToken)) != 1 {
			writeError(w, http.StatusUnauthorized, errors.New("valid bearer token required"))
			return
		}
	}
	s.mux.ServeHTTP(w, r)
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "service": "agentforge"})
}

func (s *Server) inspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var request inspectRequest
	if err := decodeRequest(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := (tools.CSVInspect{}).Execute(r.Context(), mustJSON(request))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeRawJSON(w, http.StatusOK, output)
}

func (s *Server) report(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w)
		return
	}
	var request reportRequest
	if err := decodeRequest(w, r, &request); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	output, err := (tools.SalesOpsReport{}).Execute(r.Context(), mustJSON(request))
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, err)
		return
	}
	writeRawJSON(w, http.StatusOK, output)
}

func decodeRequest(w http.ResponseWriter, r *http.Request, target any) error {
	if r.Header.Get("Content-Type") != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		return errors.New("Content-Type must be application/json")
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("request must contain one JSON object")
		}
		return err
	}
	return nil
}

func mustJSON(value any) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}

func methodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, POST, OPTIONS")
	writeError(w, http.StatusMethodNotAllowed, errors.New("method not allowed"))
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeRawJSON(w http.ResponseWriter, status int, raw string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, raw)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

// NewHTTPServer applies conservative timeouts to the API listener.
func NewHTTPServer(address string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// Shutdown is exposed for command-level graceful termination.
func Shutdown(ctx context.Context, server *http.Server) error {
	return server.Shutdown(ctx)
}
