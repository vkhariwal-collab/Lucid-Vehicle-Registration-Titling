// Package outbox implements the transactional-outbox pattern: events are
// appended in the same unit of work as the idempotency record, then relayed
// to the broker (NATS JetStream or Kafka) at-least-once. Consumers dedupe on EventID.
package outbox

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
	"time"
)

// Event is the envelope published to the broker. Payload matches
// contracts/jsonschema/temp_tag_*.v1.json.
type Event struct {
	EventID      string          `json:"event_id"`
	Type         string          `json:"event_type"`
	Subject      string          `json:"subject"`
	PartitionKey string          `json:"partition_key"`
	OccurredAt   time.Time       `json:"occurred_at"`
	Payload      json.RawMessage `json:"payload"`
}

// Publisher sends one event to the broker. Implementations must be safe to retry.
type Publisher interface {
	Publish(ctx context.Context, e Event) error
}

// Outbox is an in-memory outbox. In production this is a table written in the
// same DB transaction as the idempotency record.
type Outbox struct {
	mu      sync.Mutex
	pending []Event
	sent    int
}

// AppendLocked appends without locking; callers must hold the unit-of-work lock.
func (o *Outbox) AppendLocked(e Event) { o.pending = append(o.pending, e) }

func (o *Outbox) Lock()   { o.mu.Lock() }
func (o *Outbox) Unlock() { o.mu.Unlock() }

// Pending returns a copy of unsent events.
func (o *Outbox) Pending() []Event {
	o.mu.Lock()
	defer o.mu.Unlock()
	return append([]Event(nil), o.pending...)
}

// Sent returns the number of relayed events.
func (o *Outbox) Sent() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.sent
}

// RelayOnce publishes pending events in order; it stops at the first failure so
// per-key ordering is preserved and the event is retried next tick.
func (o *Outbox) RelayOnce(ctx context.Context, p Publisher) error {
	for _, e := range o.Pending() {
		if err := p.Publish(ctx, e); err != nil {
			return err
		}
		o.mu.Lock()
		o.pending = o.pending[1:]
		o.sent++
		o.mu.Unlock()
	}
	return nil
}

// Run relays on every tick until ctx is cancelled.
func (o *Outbox) Run(ctx context.Context, p Publisher, every time.Duration, log *slog.Logger) {
	t := time.NewTicker(every)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			if err := o.RelayOnce(ctx, p); err != nil {
				log.Warn("outbox relay failed; will retry", "err", err)
			}
		}
	}
}

// WriterPublisher writes events as JSON lines (local demo stand-in for the broker).
type WriterPublisher struct {
	mu sync.Mutex
	W  io.Writer
}

func (w *WriterPublisher) Publish(_ context.Context, e Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return json.NewEncoder(w.W).Encode(e)
}
