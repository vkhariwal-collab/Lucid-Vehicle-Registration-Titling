package issuer

import (
	"slices"
	"testing"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/rules"
)

func TestIssue(t *testing.T) {
	ca := []string{"purchase_agreement", "odometer_disclosure"}
	tests := []struct {
		name       string
		req        model.Request
		wantStatus string
		wantExp    string
		wantReason []string
	}{
		{"CA cash 90d", model.Request{VIN: "5L1RA1AA0SA000101", Jurisdiction: "CA", PurchaseType: "cash", Documents: ca, RequestedAt: "2026-11-03"}, "issued", "2027-02-01", nil},
		{"CA lease needs ELT id", model.Request{VIN: "5L1RA1AA0SA000103", Jurisdiction: "CA", PurchaseType: "lease", Lienholder: &model.Lienholder{Name: "Bank"}, Documents: append(ca, "lien_agreement"), RequestedAt: "2026-11-03"}, "rejected", "", []string{"elt_lienholder_required"}},
		{"bad VIN short-circuits", model.Request{VIN: "SHORT", Jurisdiction: "CA"}, "rejected", "", []string{"invalid_vin"}},
		{"reasons keep legacy order", model.Request{VIN: "5L1RA1AA0SA000105", Jurisdiction: "TX", PurchaseType: "finance", RequestedAt: "2026-11-05"}, "rejected", "",
			[]string{"missing_document:purchase_agreement", "missing_document:form_130u", "missing_document:lien_agreement", "lienholder_required"}},
		{"unknown jurisdiction -> default 30d", model.Request{VIN: "5L1RA1AA0SA000401", Jurisdiction: "NY", PurchaseType: "cash", RequestedAt: "2026-11-07"}, "issued", "2026-12-07", nil},
		{"datetime requested_at uses date as written", model.Request{VIN: "5L1RA1AA0SA000101", Jurisdiction: "CA", PurchaseType: "cash", Documents: ca, RequestedAt: "2026-11-03T23:30:00-08:00", Attempt: 1}, "issued", "2027-02-01", nil},
	}
	iss := New(rules.Options{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := iss.Issue(tt.req)
			if err != nil {
				t.Fatal(err)
			}
			if got.Status != tt.wantStatus {
				t.Fatalf("status = %s, want %s (reasons %v)", got.Status, tt.wantStatus, got.Reasons)
			}
			if tt.wantExp != "" && *got.ExpiresOn != tt.wantExp {
				t.Errorf("expires = %s, want %s", *got.ExpiresOn, tt.wantExp)
			}
			if tt.wantReason != nil && !slices.Equal(got.Reasons, tt.wantReason) {
				t.Errorf("reasons = %v, want %v", got.Reasons, tt.wantReason)
			}
		})
	}
}

func TestInvalidDateIsError(t *testing.T) {
	_, err := New(rules.Options{}).Issue(model.Request{VIN: "5L1RA1AA0SA000401", Jurisdiction: "NY", RequestedAt: "11/07/2026"})
	if err == nil {
		t.Fatal("expected error for malformed date")
	}
}

// Legacy keeps an explicit attempt of 0 (registration.rb:11), so it feeds the tag hash as-is.
func TestTagNumberUsesAttemptAsGiven(t *testing.T) {
	base := model.Request{VIN: "5L1RA1AA0SA000202", Jurisdiction: "CA", PurchaseType: "cash",
		Documents: []string{"purchase_agreement", "odometer_disclosure"}, RequestedAt: "2026-11-25"}
	iss := New(rules.Options{})
	for attempt, want := range map[int]string{0: "CA-DC97146C", 1: "CA-22F0A8CA"} {
		r := base
		r.Attempt = attempt
		got, err := iss.Issue(r)
		if err != nil {
			t.Fatal(err)
		}
		if *got.TagNumber != want {
			t.Errorf("attempt %d: tag = %s, want %s", attempt, *got.TagNumber, want)
		}
	}
}
