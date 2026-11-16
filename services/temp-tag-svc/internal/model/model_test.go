package model

import (
	"encoding/json"
	"testing"
)

func TestRequestAttemptDefault(t *testing.T) {
	tests := []struct {
		body string
		want int
	}{
		{`{"vin":"X"}`, 1},
		{`{"vin":"X","attempt":null}`, 1},
		{`{"vin":"X","attempt":0}`, 0},
		{`{"vin":"X","attempt":3}`, 3},
	}
	for _, tt := range tests {
		var r Request
		if err := json.Unmarshal([]byte(tt.body), &r); err != nil {
			t.Fatalf("%s: %v", tt.body, err)
		}
		if r.Attempt != tt.want || r.VIN != "X" {
			t.Errorf("%s: attempt = %d, vin = %q; want %d, X", tt.body, r.Attempt, r.VIN, tt.want)
		}
	}
}

func TestParseDate(t *testing.T) {
	tests := []struct {
		in      string
		want    string
		wantErr bool
	}{
		{"2026-11-25", "2026-11-25", false},
		{"2026-11-25T23:30:00Z", "2026-11-25", false},
		{"2026-11-25T23:30:00-08:00", "2026-11-25", false},
		{"2026-11-25T01:00:00+14:00", "2026-11-25", false},
		{"2026-11-25t10:00", "2026-11-25", false},
		{"2026-11-25 23:30", "2026-11-25", false},
		{"2026-02-30", "", true},
		{"2026-02-30T10:00:00Z", "", true},
		{"2026-11-25x", "", true},
		{"11/07/2026", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		got, err := ParseDate(tt.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q: err = %v, wantErr %v", tt.in, err, tt.wantErr)
			continue
		}
		if !tt.wantErr && FormatDate(got) != tt.want {
			t.Errorf("%q: got %s, want %s", tt.in, FormatDate(got), tt.want)
		}
	}
}
