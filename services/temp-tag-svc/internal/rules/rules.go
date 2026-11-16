// Package rules holds jurisdiction-specific temporary-tag rules ported from
// legacy/app/services/jurisdictions/*.rb. Values are synthetic and illustrative.
package rules

import (
	"time"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
)

// Options toggles approved behaviour changes. The zero value means strict parity with legacy.
type Options struct {
	// FixTXResubmissionExpiry computes TX resubmission expiry from original_issued_at
	// (legacy quirk TX-RESUBMIT-EXPIRY computes it from the new request date).
	FixTXResubmissionExpiry bool
}

// Rules describes one jurisdiction.
type Rules interface {
	Code() string
	ValidDays() int
	FeeCents() int
	Currency() string
	RequiredDocuments(r model.Request) []string
	LienErrors(r model.Request) []string
	ExtraErrors(r model.Request) []string
	Flags(r model.Request) []string
	BaseDate(r model.Request) (time.Time, error)
	AdjustExpiry(d time.Time) (time.Time, []string)
}

// For returns the rules for a registration jurisdiction. Unknown codes fall
// through to defaultRules, mirroring legacy (quirk DEFAULT-RULES-FALLTHROUGH).
func For(code string, opt Options) Rules {
	switch code {
	case "CA":
		return caRules{}
	case "TX":
		return txRules{opt: opt}
	case "ON":
		return onRules{}
	default:
		return defaultRules{}
	}
}

type defaultRules struct{}

func (defaultRules) Code() string                             { return "XX" }
func (defaultRules) ValidDays() int                           { return 30 }
func (defaultRules) FeeCents() int                            { return 0 }
func (defaultRules) Currency() string                         { return "USD" }
func (defaultRules) RequiredDocuments(model.Request) []string { return nil }
func (defaultRules) LienErrors(model.Request) []string        { return nil }
func (defaultRules) ExtraErrors(model.Request) []string       { return nil }
func (defaultRules) Flags(model.Request) []string {
	return []string{"unsupported_jurisdiction_default_rules"}
}
func (defaultRules) BaseDate(r model.Request) (time.Time, error) {
	return model.ParseDate(r.RequestedAt)
}
func (defaultRules) AdjustExpiry(d time.Time) (time.Time, []string) { return d, nil }
