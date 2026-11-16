// Package service wires the pure issuer to idempotency and the outbox.
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/idempotency"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/issuer"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/outbox"
)

// ErrKeyReuse means the Idempotency-Key was already used with a different payload.
var ErrKeyReuse = errors.New("idempotency key reused with different payload")

type Service struct {
	Issuer *issuer.Issuer
	Store  *idempotency.MemoryStore
	Now    func() time.Time
}

// Outcome is the result of Handle.
type Outcome struct {
	Result model.Result
	Replay bool
}

// Handle issues (or replays) a temp tag for key. Safe under concurrent duplicates.
func (s *Service) Handle(key string, req model.Request) (Outcome, error) {
	h, err := hashRequest(req)
	if err != nil {
		return Outcome{}, err
	}
	res, err := s.Issuer.Issue(req)
	if err != nil {
		return Outcome{}, err
	}
	evt, err := buildEvent(key, req, res, s.Now())
	if err != nil {
		return Outcome{}, err
	}
	rec, stored := s.Store.SaveIfAbsent(key, idempotency.Record{RequestHash: h, Result: res}, evt)
	if !stored && rec.RequestHash != h {
		return Outcome{}, ErrKeyReuse
	}
	return Outcome{Result: rec.Result, Replay: !stored}, nil
}

func hashRequest(req model.Request) (string, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("hash request: %w", err)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

type issuedPayload struct {
	SchemaVersion  int      `json:"schema_version"`
	IdempotencyKey string   `json:"idempotency_key"`
	VIN            string   `json:"vin"`
	Jurisdiction   string   `json:"jurisdiction"`
	Status         string   `json:"status"`
	TagNumber      *string  `json:"tag_number,omitempty"`
	ExpiresOn      *string  `json:"expires_on,omitempty"`
	FeeCents       int      `json:"fee_cents"`
	Currency       *string  `json:"currency,omitempty"`
	Reasons        []string `json:"reasons"`
	Flags          []string `json:"flags"`
}

func buildEvent(key string, req model.Request, res model.Result, now time.Time) (outbox.Event, error) {
	typ := "TempTagIssued"
	if res.Status != "issued" {
		typ = "TempTagRejected"
	}
	p, err := json.Marshal(issuedPayload{
		SchemaVersion: 1, IdempotencyKey: key, VIN: req.VIN, Jurisdiction: req.Jurisdiction,
		Status: res.Status, TagNumber: res.TagNumber, ExpiresOn: res.ExpiresOn,
		FeeCents: res.FeeCents, Currency: res.Currency, Reasons: res.Reasons, Flags: res.Flags,
	})
	if err != nil {
		return outbox.Event{}, fmt.Errorf("build event: %w", err)
	}
	return outbox.Event{
		EventID:      eventID(key),
		Type:         typ,
		Subject:      fmt.Sprintf("registration.temp_tag.%s.v1.%s", map[bool]string{true: "issued", false: "rejected"}[res.Status == "issued"], req.Jurisdiction),
		PartitionKey: req.VIN,
		OccurredAt:   now.UTC(),
		Payload:      p,
	}, nil
}

// eventID is derived from the idempotency key so a replayed publish carries the same ID.
func eventID(key string) string {
	s := sha256.Sum256([]byte("temp-tag:" + key))
	h := hex.EncodeToString(s[:16])
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
