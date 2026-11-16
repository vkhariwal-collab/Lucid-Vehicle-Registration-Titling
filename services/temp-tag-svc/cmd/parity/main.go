// Command parity replays synthetic registration requests through the legacy
// Ruby issuer and the Go issuer and reports every behavioural difference.
// It is the offline equivalent of shadow traffic.
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/issuer"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/model"
	"github.com/example/lucid-dmv-migration-demo/temp-tag-svc/internal/rules"
)

func main() {
	n := flag.Int("n", 500, "number of synthetic requests")
	seed := flag.Int64("seed", 42, "random seed")
	legacy := flag.String("legacy", "../../legacy/bin/temp_tag_cli.rb", "legacy CLI")
	out := flag.String("out", "../../reports/parity_report.md", "report path")
	fixTX := flag.Bool("fix-tx", false, "enable approved fix TX-RESUBMIT-EXPIRY in the Go port")
	minParity := flag.Float64("min", 1.0, "minimum parity ratio required to exit 0")
	flag.Parse()

	reqs := generate(*n, *seed)
	legacyOut, err := runLegacy(*legacy, reqs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "legacy:", err)
		os.Exit(2)
	}
	iss := issuer.New(rules.Options{FixTXResubmissionExpiry: *fixTX})

	type diff struct {
		i      int
		req    model.Request
		fields []string
		legacy map[string]any
		goRes  map[string]any
	}
	var diffs []diff
	byJur := map[string][2]int{}
	for i, r := range reqs {
		g, err := iss.Issue(r)
		if err != nil {
			fmt.Fprintln(os.Stderr, "go issuer:", err)
			os.Exit(2)
		}
		gm := toMap(g)
		c := byJur[r.Jurisdiction]
		c[0]++
		if !reflect.DeepEqual(gm, legacyOut[i]) {
			c[1]++
			diffs = append(diffs, diff{i, r, changedFields(legacyOut[i], gm), legacyOut[i], gm})
		}
		byJur[r.Jurisdiction] = c
	}

	ratio := float64(len(reqs)-len(diffs)) / float64(len(reqs))
	var b strings.Builder
	fmt.Fprintf(&b, "# Parity report: legacy Ruby vs Go temp-tag-svc\n\n")
	fmt.Fprintf(&b, "- Generated: %s\n- Requests: %d (seed %d, synthetic)\n- Go options: FixTXResubmissionExpiry=%v\n", time.Now().UTC().Format(time.RFC3339), len(reqs), *seed, *fixTX)
	fmt.Fprintf(&b, "- **Parity: %d/%d (%.2f%%)**\n\n", len(reqs)-len(diffs), len(reqs), ratio*100)
	fmt.Fprintf(&b, "| Jurisdiction | Requests | Mismatches |\n|---|---|---|\n")
	jurs := make([]string, 0, len(byJur))
	for j := range byJur {
		jurs = append(jurs, j)
	}
	sort.Strings(jurs)
	for _, j := range jurs {
		fmt.Fprintf(&b, "| %s | %d | %d |\n", j, byJur[j][0], byJur[j][1])
	}
	if len(diffs) > 0 {
		cause := map[string]int{}
		for _, d := range diffs {
			cause[fmt.Sprintf("%s attempt>1=%v fields=%s", d.req.Jurisdiction, d.req.Attempt > 1, strings.Join(d.fields, ","))]++
		}
		fmt.Fprintf(&b, "\n## Mismatch clusters\n\n| Cluster | Count |\n|---|---|\n")
		for k, v := range cause {
			fmt.Fprintf(&b, "| %s | %d |\n", k, v)
		}
		fmt.Fprintf(&b, "\n## First examples\n\n")
		for _, d := range diffs[:min(5, len(diffs))] {
			fmt.Fprintf(&b, "- #%d %s VIN …%s attempt=%d requested_at=%s original_issued_at=%s: legacy `%v` vs go `%v` (fields: %s)\n",
				d.i, d.req.Jurisdiction, d.req.VIN[len(d.req.VIN)-6:], d.req.Attempt, d.req.RequestedAt, d.req.OriginalIssuedAt,
				pick(d.legacy, d.fields), pick(d.goRes, d.fields), strings.Join(d.fields, ","))
		}
	}
	if err := os.WriteFile(*out, []byte(b.String()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("parity %d/%d (%.2f%%), report: %s\n", len(reqs)-len(diffs), len(reqs), ratio*100, *out)
	if ratio < *minParity {
		os.Exit(1)
	}
}

func generate(n int, seed int64) []model.Request {
	rng := rand.New(rand.NewSource(seed))
	jurs := []string{"CA", "CA", "CA", "TX", "TX", "ON", "NY", "WA"}
	types := []string{"cash", "finance", "lease"}
	docs := map[string][]string{
		"CA": {"purchase_agreement", "odometer_disclosure", "lien_agreement"},
		"TX": {"purchase_agreement", "form_130u", "lien_agreement"},
		"ON": {"bill_of_sale", "proof_of_insurance"},
		"NY": {"purchase_agreement"}, "WA": {"purchase_agreement"},
	}
	postal := map[string][]string{"ON": {"M5V 3L9", "K1A0B1", "m5v 3l9", "12345"}}
	vinChars := "ABCDEFGHJKLMNPRSTUVWXYZ0123456789"
	start := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	out := make([]model.Request, n)
	for i := range out {
		j := jurs[rng.Intn(len(jurs))]
		vin := []byte("5L1RA1AA0SA")
		for len(vin) < 17 {
			vin = append(vin, vinChars[rng.Intn(len(vinChars))])
		}
		if rng.Float64() < 0.03 {
			vin[12] = 'O'
		}
		var d []string
		for _, doc := range docs[j] {
			if rng.Float64() < 0.92 {
				d = append(d, doc)
			}
		}
		r := model.Request{VIN: string(vin), Jurisdiction: j, DeliveryJurisdiction: j, PurchaseType: types[rng.Intn(3)],
			Documents: d, RequestedAt: start.AddDate(0, 0, rng.Intn(120)).Format("2006-01-02"), Attempt: 1}
		if r.Financed() {
			lh := &model.Lienholder{Name: "Lucid Financial Services"}
			if rng.Float64() < 0.85 {
				lh.ELTID = "ELT-0042"
			}
			if rng.Float64() < 0.05 {
				lh.Name = ""
			}
			r.Lienholder = lh
		}
		if p, ok := postal[j]; ok {
			r.Customer.PostalCode = p[rng.Intn(len(p))]
		} else {
			r.Customer.PostalCode = fmt.Sprintf("%05d", rng.Intn(99999))
		}
		if rng.Float64() < 0.12 {
			orig, _ := time.Parse("2006-01-02", r.RequestedAt)
			r.OriginalIssuedAt = orig.Format("2006-01-02")
			r.RequestedAt = orig.AddDate(0, 0, 5+rng.Intn(30)).Format("2006-01-02")
			r.Attempt = 2 + rng.Intn(2)
		}
		out[i] = r
	}
	return out
}

func runLegacy(path string, reqs []model.Request) ([]map[string]any, error) {
	var in bytes.Buffer
	enc := json.NewEncoder(&in)
	for _, r := range reqs {
		if err := enc.Encode(r); err != nil {
			return nil, err
		}
	}
	cmd := exec.Command("ruby", path)
	cmd.Stdin = &in
	cmd.Stderr = os.Stderr
	outB, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var res []map[string]any
	sc := bufio.NewScanner(bytes.NewReader(outB))
	for sc.Scan() {
		var m map[string]any
		if err := json.Unmarshal(sc.Bytes(), &m); err != nil {
			return nil, err
		}
		res = append(res, m)
	}
	if len(res) != len(reqs) {
		return nil, fmt.Errorf("legacy returned %d results for %d requests", len(res), len(reqs))
	}
	return res, nil
}

func toMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	return m
}

func changedFields(a, b map[string]any) []string {
	var f []string
	for k := range a {
		if !reflect.DeepEqual(a[k], b[k]) {
			f = append(f, k)
		}
	}
	sort.Strings(f)
	return f
}

func pick(m map[string]any, keys []string) map[string]any {
	o := map[string]any{}
	for _, k := range keys {
		o[k] = m[k]
	}
	return o
}
