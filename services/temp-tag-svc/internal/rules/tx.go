package rules

import (
	"strings"
	"time"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
)

type txRules struct {
	defaultRules
	opt Options
}

func (txRules) Code() string   { return "TX" }
func (txRules) ValidDays() int { return 60 }
func (txRules) FeeCents() int  { return 500 }

func (txRules) RequiredDocuments(r model.Request) []string {
	d := []string{"purchase_agreement", "form_130u"}
	if r.Financed() {
		d = append(d, "lien_agreement")
	}
	return d
}

func (txRules) LienErrors(r model.Request) []string {
	if r.Financed() && (r.Lienholder == nil || strings.TrimSpace(r.Lienholder.Name) == "") {
		return []string{"lienholder_required"}
	}
	return nil
}

func (txRules) Flags(r model.Request) []string {
	if r.Attempt > 1 {
		return []string{"resubmission"}
	}
	return nil
}

func (t txRules) BaseDate(r model.Request) (time.Time, error) {
	if t.opt.FixTXResubmissionExpiry && r.Attempt > 1 && r.OriginalIssuedAt != "" {
		return model.ParseDate(r.OriginalIssuedAt)
	}
	return model.ParseDate(r.RequestedAt)
}
