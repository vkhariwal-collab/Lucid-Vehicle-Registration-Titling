// Package transport exposes the service over HTTP.
package transport

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/metrics"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/service"
)

const maxBody = 64 << 10

type Handler struct {
	Svc     *service.Service
	Metrics *metrics.Registry
	Log     *slog.Logger
}

func (h *Handler) Routes() *http.ServeMux {
	m := http.NewServeMux()
	m.HandleFunc("POST /v1/temp-tags", h.issue)
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	m.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	m.HandleFunc("GET /metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		h.Metrics.Write(w)
	})
	return m
}

func (h *Handler) issue(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		h.fail(w, http.StatusBadRequest, "missing Idempotency-Key header", "")
		return
	}
	var req model.Request
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBody)).Decode(&req); err != nil {
		h.fail(w, http.StatusBadRequest, "invalid JSON body", "")
		return
	}
	out, err := h.Svc.Handle(key, req)
	switch {
	case errors.Is(err, service.ErrKeyReuse):
		h.fail(w, http.StatusConflict, err.Error(), req.Jurisdiction)
		return
	case err != nil:
		h.Log.ErrorContext(r.Context(), "issue failed", "err", err, "vin", maskVIN(req.VIN))
		h.fail(w, http.StatusInternalServerError, "internal error", req.Jurisdiction)
		return
	}

	status := http.StatusCreated
	if out.Result.Status != "issued" {
		status = http.StatusUnprocessableEntity
	}
	if out.Replay {
		w.Header().Set("Idempotent-Replay", "true")
		status = http.StatusOK
	}
	h.Metrics.Inc("temp_tag_requests_total", "jurisdiction", req.Jurisdiction, "status", out.Result.Status, "replay", boolStr(out.Replay))
	h.Log.InfoContext(r.Context(), "temp tag", "vin", maskVIN(req.VIN), "jurisdiction", req.Jurisdiction, "status", out.Result.Status, "replay", out.Replay)
	writeJSON(w, status, out.Result)
}

func (h *Handler) fail(w http.ResponseWriter, code int, msg, jur string) {
	h.Metrics.Inc("temp_tag_errors_total", "code", http.StatusText(code), "jurisdiction", jur)
	writeJSON(w, code, map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// maskVIN keeps only the last 6 characters; VIN plus owner data is PII.
func maskVIN(v string) string {
	if len(v) <= 6 {
		return "***"
	}
	return "***********" + v[len(v)-6:]
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
