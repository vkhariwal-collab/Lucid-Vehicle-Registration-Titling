package rules

import (
	"slices"
	"testing"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
)

func TestTXLienholderNameUsesRubyStrip(t *testing.T) {
	tests := []struct {
		name string
		want []string
	}{
		{"Lucid Financial Services", nil},
		{"", []string{"lienholder_required"}},
		{" \t\n\v\f\r ", []string{"lienholder_required"}},
		{"\x00", []string{"lienholder_required"}},
		{"\u00a0", nil},
		{"\u3000", nil},
	}
	for _, tt := range tests {
		r := model.Request{PurchaseType: "finance", Lienholder: &model.Lienholder{Name: tt.name}}
		if got := (txRules{}).LienErrors(r); !slices.Equal(got, tt.want) {
			t.Errorf("name %q: got %v, want %v", tt.name, got, tt.want)
		}
	}
}
