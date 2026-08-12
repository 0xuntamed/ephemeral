package httpapi

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/example/ephemeral-platform/internal/controlplane"
	"github.com/example/ephemeral-platform/internal/environment"
	"github.com/example/ephemeral-platform/internal/store"
)

type Server struct {
	service *controlplane.Service
	apiKey  string
	logger  *slog.Logger
}

var demoPreviewPage = template.Must(template.New("preview").Parse(`<!doctype html><html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>{{.ID}} preview</title><style>body{font-family:system-ui;margin:0;background:#07110e;color:#e7f9f0;display:grid;min-height:100vh;place-items:center}.card{width:min(680px,calc(100% - 48px));background:#10221c;border:1px solid #2b4b40;border-radius:24px;padding:36px;box-shadow:0 30px 80px #0008}.eyebrow{color:#7be0b6;text-transform:uppercase;letter-spacing:.14em;font-weight:700;font-size:12px}h1{font-size:clamp(34px,7vw,64px);line-height:1;margin:18px 0}.pill{display:inline-flex;padding:8px 12px;border-radius:999px;background:#18372d;color:#9ff4cf}dl{display:grid;grid-template-columns:max-content 1fr;gap:12px 24px;margin-top:30px}dt{color:#90aa9f}dd{margin:0;font-family:ui-monospace,monospace;overflow-wrap:anywhere}</style></head><body><main class="card"><div class="eyebrow">Local lifecycle simulator</div><h1>Preview is ready.</h1><span class="pill">Healthy · isolated control-plane record</span><dl><dt>Environment</dt><dd>{{.ID}}</dd><dt>Repository</dt><dd>{{.Repository}}</dd><dt>Branch</dt><dd>{{.Branch}}</dd><dt>Database schema</dt><dd>{{.DatabaseSchema}}</dd></dl></main></body></html>`))

type createRequest struct {
	Repository  string                `json:"repository"`
	Branch      string                `json:"branch"`
	CommitSHA   string                `json:"commit_sha"`
	PullRequest int                   `json:"pull_request"`
	Owner       string                `json:"owner"`
	Image       string                `json:"image"`
	TTL         string                `json:"ttl"`
	Resources   environment.Resources `json:"resources"`
}

type extendRequest struct {
	Duration string `json:"duration"`
}

type errorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func New(service *controlplane.Service, apiKey string, logger *slog.Logger) http.Handler {
	server := &Server{service: service, apiKey: strings.TrimSpace(apiKey), logger: logger}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", server.health)
	mux.HandleFunc("GET /readyz", server.ready)
	mux.HandleFunc("GET /metrics", server.metrics)
	mux.HandleFunc("GET /api/v1/environments", server.listEnvironments)
	mux.HandleFunc("POST /api/v1/environments", server.createEnvironment)
	mux.HandleFunc("GET /api/v1/environments/{id}", server.getEnvironment)
	mux.HandleFunc("DELETE /api/v1/environments/{id}", server.destroyEnvironment)
	mux.HandleFunc("POST /api/v1/environments/{id}/extend", server.extendEnvironment)
	mux.HandleFunc("POST /api/v1/environments/{id}/retry", server.retryEnvironment)
	mux.HandleFunc("POST /api/v1/reconcile", server.reconcile)
	mux.HandleFunc("GET /preview/{id}", server.demoPreview)
	mux.HandleFunc("GET /preview/{id}/healthz", server.demoPreviewHealth)
	mux.HandleFunc("GET /", func(writer http.ResponseWriter, request *http.Request) {
		writeJSON(writer, http.StatusOK, map[string]any{
			"name":   "ephemeral environment control plane",
			"api":    "/api/v1/environments",
			"health": "/healthz",
		})
	})
	return server.recover(server.logRequests(server.cors(server.authenticate(mux))))
}

func (s *Server) demoPreview(writer http.ResponseWriter, request *http.Request) {
	item, err := s.service.Get(request.Context(), request.PathValue("id"))
	if err != nil || item.Status != environment.StatusReady {
		http.NotFound(writer, request)
		return
	}
	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := demoPreviewPage.Execute(writer, item); err != nil {
		s.logger.Error("render demo preview", "error", err)
	}
}

func (s *Server) demoPreviewHealth(writer http.ResponseWriter, request *http.Request) {
	item, err := s.service.Get(request.Context(), request.PathValue("id"))
	if err != nil || item.Status != environment.StatusReady {
		writeError(writer, http.StatusNotFound, "not_ready", "preview environment is not ready")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok", "environment_id": item.ID})
}

func (s *Server) health(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(writer http.ResponseWriter, request *http.Request) {
	if _, err := s.service.List(request.Context(), store.Filter{}); err != nil {
		writeError(writer, http.StatusServiceUnavailable, "store_unavailable", "environment store is unavailable")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *Server) metrics(writer http.ResponseWriter, request *http.Request) {
	metrics, err := s.service.Metrics(request.Context())
	if err != nil {
		writeError(writer, http.StatusInternalServerError, "metrics_failed", "could not collect metrics")
		return
	}
	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	writer.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(writer, metrics)
}

func (s *Server) listEnvironments(writer http.ResponseWriter, request *http.Request) {
	filter := store.Filter{
		Status:     environment.Status(strings.ToUpper(strings.TrimSpace(request.URL.Query().Get("status")))),
		Repository: strings.TrimSpace(request.URL.Query().Get("repository")),
		Owner:      strings.TrimSpace(request.URL.Query().Get("owner")),
	}
	items, err := s.service.List(request.Context(), filter)
	if err != nil {
		s.logger.Error("list environments", "error", err)
		writeError(writer, http.StatusInternalServerError, "list_failed", "could not list environments")
		return
	}
	writeJSON(writer, http.StatusOK, map[string]any{"items": items, "total": len(items)})
}

func (s *Server) createEnvironment(writer http.ResponseWriter, request *http.Request) {
	var body createRequest
	if err := decodeJSON(request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	var ttl time.Duration
	var err error
	if strings.TrimSpace(body.TTL) != "" {
		ttl, err = time.ParseDuration(body.TTL)
		if err != nil {
			writeError(writer, http.StatusBadRequest, "invalid_ttl", "ttl must be a duration such as 6h or 30m")
			return
		}
	}
	item, created, err := s.service.Create(request.Context(), controlplane.CreateInput{
		Repository: body.Repository, Branch: body.Branch, CommitSHA: body.CommitSHA,
		PullRequest: body.PullRequest, Owner: body.Owner, Image: body.Image,
		TTL: ttl, Resources: body.Resources,
		IdempotencyKey: strings.TrimSpace(request.Header.Get("Idempotency-Key")),
	})
	if err != nil {
		s.writeServiceError(writer, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writer.Header().Set("Location", "/api/v1/environments/"+item.ID)
	writer.Header().Set("Idempotency-Replayed", strconv.FormatBool(!created))
	writeJSON(writer, status, item)
}

func (s *Server) getEnvironment(writer http.ResponseWriter, request *http.Request) {
	item, err := s.service.Get(request.Context(), request.PathValue("id"))
	if err != nil {
		s.writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, item)
}

func (s *Server) destroyEnvironment(writer http.ResponseWriter, request *http.Request) {
	item, err := s.service.RequestDestroy(request.Context(), request.PathValue("id"), "Destruction requested through the API")
	if err != nil {
		s.writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, item)
}

func (s *Server) extendEnvironment(writer http.ResponseWriter, request *http.Request) {
	var body extendRequest
	if err := decodeJSON(request, &body); err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	extension, err := time.ParseDuration(body.Duration)
	if err != nil {
		writeError(writer, http.StatusBadRequest, "invalid_duration", "duration must look like 1h or 30m")
		return
	}
	item, err := s.service.Extend(request.Context(), request.PathValue("id"), extension)
	if err != nil {
		s.writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusOK, item)
}

func (s *Server) retryEnvironment(writer http.ResponseWriter, request *http.Request) {
	item, err := s.service.Retry(request.Context(), request.PathValue("id"))
	if err != nil {
		s.writeServiceError(writer, err)
		return
	}
	writeJSON(writer, http.StatusAccepted, item)
}

func (s *Server) reconcile(writer http.ResponseWriter, request *http.Request) {
	s.service.TriggerReconcile()
	writeJSON(writer, http.StatusAccepted, map[string]string{"status": "reconciliation_scheduled"})
}

func (s *Server) writeServiceError(writer http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(writer, http.StatusNotFound, "not_found", "environment not found")
	case errors.Is(err, controlplane.ErrInvalidInput):
		writeError(writer, http.StatusBadRequest, "invalid_input", err.Error())
	case errors.Is(err, controlplane.ErrInvalidState):
		writeError(writer, http.StatusConflict, "invalid_state", err.Error())
	default:
		s.logger.Error("api operation", "error", err)
		writeError(writer, http.StatusInternalServerError, "internal_error", "the operation could not be completed")
	}
}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if s.apiKey == "" || request.URL.Path == "/healthz" || request.URL.Path == "/readyz" || request.URL.Path == "/" || strings.HasPrefix(request.URL.Path, "/preview/") {
			next.ServeHTTP(writer, request)
			return
		}
		provided := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
		if len(provided) != len(s.apiKey) || subtle.ConstantTimeCompare([]byte(provided), []byte(s.apiKey)) != 1 {
			writer.Header().Set("WWW-Authenticate", "Bearer")
			writeError(writer, http.StatusUnauthorized, "unauthorized", "a valid bearer token is required")
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func (s *Server) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Idempotency-Key")
		writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		started := time.Now()
		next.ServeHTTP(writer, request)
		s.logger.Info("http request", "method", request.Method, "path", request.URL.Path, "duration", time.Since(started))
	})
}

func (s *Server) recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				s.logger.Error("http panic", "error", fmt.Sprint(recovered))
				writeError(writer, http.StatusInternalServerError, "internal_error", "the request could not be completed")
			}
		}()
		next.ServeHTTP(writer, request)
	})
}

func decodeJSON(request *http.Request, destination any) error {
	request.Body = http.MaxBytesReader(nil, request.Body, 1<<20)
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("invalid request body: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("request body must contain one JSON object")
	}
	return nil
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}

func writeError(writer http.ResponseWriter, status int, code, message string) {
	response := errorResponse{}
	response.Error.Code = code
	response.Error.Message = message
	writeJSON(writer, status, response)
}
