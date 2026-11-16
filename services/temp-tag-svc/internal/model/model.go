// Package model defines the temp-tag request/result shapes shared with the
// legacy golden fixtures (fixtures/golden/temp_tag/*.json).
package model

import (
	"encoding/json"
	"fmt"
	"time"
)

const dateLayout = "2006-01-02"

type Lienholder struct {
	Name  string `json:"name,omitempty"`
	ELTID string `json:"elt_id,omitempty"`
}

type Customer struct {
	PostalCode string `json:"postal_code,omitempty"`
}

type Request struct {
	VIN                  string      `json:"vin"`
	Jurisdiction         string      `json:"jurisdiction"`
	DeliveryJurisdiction string      `json:"delivery_jurisdiction,omitempty"`
	PurchaseType         string      `json:"purchase_type"`
	Lienholder           *Lienholder `json:"lienholder,omitempty"`
	Documents            []string    `json:"documents"`
	Customer             Customer    `json:"customer"`
	RequestedAt          string      `json:"requested_at"`
	Attempt              int         `json:"attempt"`
	OriginalIssuedAt     string      `json:"original_issued_at,omitempty"`
}

// UnmarshalJSON defaults a missing or null attempt to 1 but keeps an explicit 0,
// mirroring legacy `@attempt ||= 1` (registration.rb:11).
func (r *Request) UnmarshalJSON(b []byte) error {
	type plain Request
	aux := struct {
		*plain
		Attempt *int `json:"attempt"`
	}{plain: (*plain)(r)}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	r.Attempt = 1
	if aux.Attempt != nil {
		r.Attempt = *aux.Attempt
	}
	return nil
}

func (r Request) Financed() bool { return r.PurchaseType == "finance" || r.PurchaseType == "lease" }

type Result struct {
	Status    string   `json:"status"`
	TagNumber *string  `json:"tag_number"`
	ExpiresOn *string  `json:"expires_on"`
	ValidDays int      `json:"valid_days"`
	FeeCents  int      `json:"fee_cents"`
	Currency  *string  `json:"currency"`
	Reasons   []string `json:"reasons"`
	Flags     []string `json:"flags"`
}

// ParseDate accepts an ISO date, optionally followed by a time part ("T", "t" or
// " " then anything). Like legacy Date.parse (registration.rb:25) it keeps the
// calendar date as written and ignores the time and offset. Other formats that
// Date.parse accepts (e.g. "2026/11/25", "Nov 25 2026") are not supported.
func ParseDate(s string) (time.Time, error) {
	d := s
	if len(s) > len(dateLayout) {
		switch s[len(dateLayout)] {
		case 'T', 't', ' ':
			d = s[:len(dateLayout)]
		}
	}
	t, err := time.Parse(dateLayout, d)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", s, err)
	}
	return t, nil
}

func FormatDate(t time.Time) string { return t.Format(dateLayout) }
