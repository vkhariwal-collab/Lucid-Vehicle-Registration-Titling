package rules

import (
	"regexp"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
)

// Case-sensitive on purpose: legacy rejects lower-case postal codes (quirk ON-POSTAL-CASE).
var postalRE = regexp.MustCompile(`^[A-Z]\d[A-Z] ?\d[A-Z]\d$`)

type onRules struct{ defaultRules }

func (onRules) Code() string     { return "ON" }
func (onRules) ValidDays() int   { return 10 }
func (onRules) FeeCents() int    { return 3200 }
func (onRules) Currency() string { return "CAD" }

func (onRules) RequiredDocuments(model.Request) []string {
	return []string{"bill_of_sale", "proof_of_insurance"}
}

func (onRules) ExtraErrors(r model.Request) []string {
	if !postalRE.MatchString(r.Customer.PostalCode) {
		return []string{"invalid_postal_code"}
	}
	return nil
}

func (onRules) Flags(r model.Request) []string {
	var f []string
	if r.Financed() {
		f = append(f, "ppsa_registration_required")
	}
	if r.PurchaseType == "lease" {
		f = append(f, "lessor_owned")
	}
	return f
}
