// Package parity validates the Go port against the SAME golden fixtures the
// legacy RSpec characterization suite uses.
package parity

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/issuer"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/rules"
)

type fixture struct {
	Name     string          `json:"name"`
	Tags     []string        `json:"tags"`
	Input    model.Request   `json:"input"`
	Expected json.RawMessage `json:"expected"`
}

func loadFixtures(t *testing.T) []fixture {
	t.Helper()
	paths, err := filepath.Glob("../../../../fixtures/golden/temp_tag/*.json")
	if err != nil || len(paths) == 0 {
		t.Fatalf("no golden fixtures found: %v", err)
	}
	var out []fixture
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		var f fixture
		if err := json.Unmarshal(b, &f); err != nil {
			t.Fatalf("%s: %v", p, err)
		}
		out = append(out, f)
	}
	return out
}

func normalize(t *testing.T, v any) map[string]any {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestGoldenParity(t *testing.T) {
	iss := issuer.New(rules.Options{})
	for _, f := range loadFixtures(t) {
		t.Run(f.Name, func(t *testing.T) {
			got, err := iss.Issue(f.Input)
			if err != nil {
				t.Fatal(err)
			}
			var want map[string]any
			if err := json.Unmarshal(f.Expected, &want); err != nil {
				t.Fatal(err)
			}
			if g := normalize(t, got); !reflect.DeepEqual(g, want) {
				t.Errorf("parity mismatch\n got: %v\nwant: %v", g, want)
			}
		})
	}
}

// With the approved fix enabled, only the TX resubmission fixture may diverge.
func TestApprovedFixDivergesOnlyWhereExpected(t *testing.T) {
	iss := issuer.New(rules.Options{FixTXResubmissionExpiry: true})
	for _, f := range loadFixtures(t) {
		got, err := iss.Issue(f.Input)
		if err != nil {
			t.Fatal(err)
		}
		var want map[string]any
		_ = json.Unmarshal(f.Expected, &want)
		diverged := !reflect.DeepEqual(normalize(t, got), want)
		if exp := f.Name == "TX_resubmission_recomputes_expiry"; diverged != exp {
			t.Errorf("%s: diverged=%v, want %v", f.Name, diverged, exp)
		}
		if f.Name == "TX_resubmission_recomputes_expiry" && *got.ExpiresOn != "2027-01-04" {
			t.Errorf("fixed expiry = %s, want 2027-01-04 (original_issued_at + 60d)", *got.ExpiresOn)
		}
	}
}
