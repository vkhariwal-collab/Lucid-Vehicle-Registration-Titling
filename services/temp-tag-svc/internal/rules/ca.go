package rules

import (
	"time"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
)

type caRules struct{ defaultRules }

func (caRules) Code() string   { return "CA" }
func (caRules) ValidDays() int { return 90 }

func (caRules) RequiredDocuments(r model.Request) []string {
	d := []string{"purchase_agreement", "odometer_disclosure"}
	if r.Financed() {
		d = append(d, "lien_agreement")
	}
	return d
}

func (caRules) LienErrors(r model.Request) []string {
	if r.Financed() && (r.Lienholder == nil || r.Lienholder.ELTID == "") {
		return []string{"elt_lienholder_required"}
	}
	return nil
}

func (caRules) Flags(model.Request) []string { return nil }

// AdjustExpiry rolls a Sunday expiry to Monday (legacy quirk CA-SUNDAY-ROLL).
func (caRules) AdjustExpiry(d time.Time) (time.Time, []string) {
	if d.Weekday() == time.Sunday {
		return d.AddDate(0, 0, 1), []string{"expiry_rolled_from_sunday"}
	}
	return d, nil
}
