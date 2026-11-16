// Package idempotency stores the first result for each Idempotency-Key so
// retries return the same tag instead of issuing a second one.
package idempotency

import (
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/outbox"
)

// Record is what is persisted per key. RequestHash detects key reuse with a different payload.
type Record struct {
	RequestHash string
	Result      model.Result
}

// MemoryStore is an in-memory store sharing a unit-of-work lock with the outbox.
// Production equivalent: a table with UNIQUE(idempotency_key) written in the
// same transaction as the outbox row.
type MemoryStore struct {
	ob   *outbox.Outbox
	recs map[string]Record
}

func NewMemoryStore(ob *outbox.Outbox) *MemoryStore {
	return &MemoryStore{ob: ob, recs: map[string]Record{}}
}

// SaveIfAbsent atomically stores rec and appends evt, unless key already exists,
// in which case it returns the existing record and stored=false.
func (s *MemoryStore) SaveIfAbsent(key string, rec Record, evt outbox.Event) (existing Record, stored bool) {
	s.ob.Lock()
	defer s.ob.Unlock()
	if r, ok := s.recs[key]; ok {
		return r, false
	}
	s.recs[key] = rec
	s.ob.AppendLocked(evt)
	return rec, true
}
