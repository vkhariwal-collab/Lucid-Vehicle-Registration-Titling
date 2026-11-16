// Package issuer ports legacy/app/services/temp_tag_issuer.rb.
package issuer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/rules"
)

var vinRE = regexp.MustCompile(`^[A-HJ-NPR-Z0-9]{17}$`)

type Issuer struct {
	opt rules.Options
}

func New(opt rules.Options) *Issuer { return &Issuer{opt: opt} }

// Issue evaluates a request. It is pure and deterministic: no clock, no I/O.
func (i *Issuer) Issue(req model.Request) (model.Result, error) {
	if !vinRE.MatchString(req.VIN) {
		return rejected([]string{"invalid_vin"}), nil
	}
	r := rules.For(req.Jurisdiction, i.opt)

	var errs []string
	for _, d := range r.RequiredDocuments(req) {
		if !slices.Contains(req.Documents, d) {
			errs = append(errs, "missing_document:"+d)
		}
	}
	errs = append(errs, r.LienErrors(req)...)
	errs = append(errs, r.ExtraErrors(req)...)
	if len(errs) > 0 {
		return rejected(errs), nil
	}

	base, err := r.BaseDate(req)
	if err != nil {
		return model.Result{}, fmt.Errorf("issue temp tag: %w", err)
	}
	exp, adjFlags := r.AdjustExpiry(base.AddDate(0, 0, r.ValidDays()))

	tag := tagNumber(r.Code(), req)
	expStr := model.FormatDate(exp)
	cur := r.Currency()
	return model.Result{
		Status:    "issued",
		TagNumber: &tag,
		ExpiresOn: &expStr,
		ValidDays: r.ValidDays(),
		FeeCents:  r.FeeCents(),
		Currency:  &cur,
		Reasons:   []string{},
		Flags:     append(nonNil(r.Flags(req)), adjFlags...),
	}, nil
}

func tagNumber(code string, req model.Request) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", req.VIN, req.Jurisdiction, req.Attempt)))
	return code + "-" + strings.ToUpper(hex.EncodeToString(sum[:])[:8])
}

func rejected(reasons []string) model.Result {
	return model.Result{Status: "rejected", Reasons: reasons, Flags: []string{}}
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
