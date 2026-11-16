package transport

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/idempotency"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/issuer"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/metrics"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/outbox"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/rules"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/service"
)

const caBody = `{"vin":"5L1RA1AA0SA000101","jurisdiction":"CA","purchase_type":"cash","documents":["purchase_agreement","odometer_disclosure"],"requested_at":"2026-11-03","attempt":1}`

func newServer(t *testing.T) (*httptest.Server, *outbox.Outbox) {
	t.Helper()
	ob := &outbox.Outbox{}
	svc := &service.Service{Issuer: issuer.New(rules.Options{}), Store: idempotency.NewMemoryStore(ob),
		Now: func() time.Time { return time.Date(2026, 11, 3, 9, 0, 0, 0, time.UTC) }}
	h := &Handler{Svc: svc, Metrics: metrics.New(), Log: slog.New(slog.NewTextHandler(io.Discard, nil))}
	srv := httptest.NewServer(h.Routes())
	t.Cleanup(srv.Close)
	return srv, ob
}

func post(t *testing.T, url, key, body string) (*http.Response, string) {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, url+"/v1/temp-tags", strings.NewReader(body))
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp, string(b)
}

func TestIdempotentReplay(t *testing.T) {
	srv, ob := newServer(t)
	r1, b1 := post(t, srv.URL, "order-123:temp-tag", caBody)
	r2, b2 := post(t, srv.URL, "order-123:temp-tag", caBody)
	if r1.StatusCode != http.StatusCreated || r2.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, %d; want 201, 200", r1.StatusCode, r2.StatusCode)
	}
	if r2.Header.Get("Idempotent-Replay") != "true" || b1 != b2 {
		t.Fatalf("replay must return identical body with Idempotent-Replay header")
	}
	if n := len(ob.Pending()); n != 1 {
		t.Fatalf("outbox events = %d, want 1", n)
	}
}

func TestConcurrentDuplicatesEmitOneEvent(t *testing.T) {
	srv, ob := newServer(t)
	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() { defer wg.Done(); post(t, srv.URL, "dup-key", caBody) }()
	}
	wg.Wait()
	if n := len(ob.Pending()); n != 1 {
		t.Fatalf("outbox events = %d, want exactly 1", n)
	}
}

func TestKeyReuseWithDifferentPayloadIsConflict(t *testing.T) {
	srv, _ := newServer(t)
	post(t, srv.URL, "k1", caBody)
	r, _ := post(t, srv.URL, "k1", strings.Replace(caBody, "000101", "000999", 1))
	if r.StatusCode != http.StatusConflict {
		t.Fatalf("status = %d, want 409", r.StatusCode)
	}
}

func TestMissingKeyAndRejection(t *testing.T) {
	srv, _ := newServer(t)
	if r, _ := post(t, srv.URL, "", caBody); r.StatusCode != http.StatusBadRequest {
		t.Fatalf("missing key status = %d, want 400", r.StatusCode)
	}
	bad := strings.Replace(caBody, `"purchase_agreement",`, "", 1)
	if r, b := post(t, srv.URL, "k2", bad); r.StatusCode != http.StatusUnprocessableEntity || !strings.Contains(b, "missing_document:purchase_agreement") {
		t.Fatalf("rejection status = %d body %s", r.StatusCode, b)
	}
}

func TestMaskVIN(t *testing.T) {
	if got := maskVIN("5L1RA1AA0SA000101"); got != "***********000101" {
		t.Fatalf("maskVIN = %s", got)
	}
}
