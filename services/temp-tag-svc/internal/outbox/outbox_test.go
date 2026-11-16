package outbox

import (
	"context"
	"errors"
	"testing"
)

type flakyPub struct {
	fail int
	got  []string
}

func (f *flakyPub) Publish(_ context.Context, e Event) error {
	if f.fail > 0 {
		f.fail--
		return errors.New("broker unavailable")
	}
	f.got = append(f.got, e.EventID)
	return nil
}

func TestRelayPreservesOrderAndRetries(t *testing.T) {
	o := &Outbox{}
	o.Lock()
	for _, id := range []string{"a", "b", "c"} {
		o.AppendLocked(Event{EventID: id})
	}
	o.Unlock()
	p := &flakyPub{fail: 1}
	if err := o.RelayOnce(context.Background(), p); err == nil {
		t.Fatal("expected first relay to fail")
	}
	if len(o.Pending()) != 3 {
		t.Fatal("failed publish must not drop events")
	}
	if err := o.RelayOnce(context.Background(), p); err != nil {
		t.Fatal(err)
	}
	if len(p.got) != 3 || p.got[0] != "a" || p.got[2] != "c" || o.Sent() != 3 {
		t.Fatalf("got %v sent %d", p.got, o.Sent())
	}
}
