// Command temp-tag-svc serves Temporary Tag Issuance, the first slice
// strangled out of the legacy registration monolith.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/idempotency"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/issuer"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/metrics"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/outbox"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/rules"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/service"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/transport"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(log); err != nil {
		log.Error("exit", "err", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	opt := rules.Options{FixTXResubmissionExpiry: os.Getenv("FIX_TX_RESUBMISSION_EXPIRY") == "true"}
	ob := &outbox.Outbox{}
	svc := &service.Service{Issuer: issuer.New(opt), Store: idempotency.NewMemoryStore(ob), Now: time.Now}
	h := &transport.Handler{Svc: svc, Metrics: metrics.New(), Log: log}

	go ob.Run(ctx, &outbox.WriterPublisher{W: os.Stderr}, 200*time.Millisecond, log)

	addr := envOr("ADDR", ":8080")
	srv := &http.Server{Addr: addr, Handler: h.Routes(), ReadHeaderTimeout: 5 * time.Second}
	errc := make(chan error, 1)
	go func() { errc <- srv.ListenAndServe() }()
	log.Info("listening", "addr", addr, "fix_tx_resubmission_expiry", opt.FixTXResubmissionExpiry)

	select {
	case err := <-errc:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
