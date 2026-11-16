// Package model defines the temp-tag request/result shapes shared with the
// legacy golden fixtures (fixtures/golden/temp_tag/*.json).
package model

import (
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

func ParseDate(s string) (time.Time, error) {
	t, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse date %q: %w", s, err)
	}
	return t, nil
}

func FormatDate(t time.Time) string { return t.Format(dateLayout) }
