# Parity report: legacy Ruby vs Go temp-tag-svc

- Generated: 2026-11-16T06:39:25Z
- Requests: 500 (seed 42, synthetic)
- Go options: FixTXResubmissionExpiry=true
- Edge-case inputs: attempt=0=15, lienholder name whitespace-only=33, requested_at with time=38
- **Parity: 490/500 (98.00%)**

| Jurisdiction | Requests | Mismatches |
|---|---|---|
| CA | 212 | 0 |
| NY | 61 | 0 |
| ON | 61 | 0 |
| TX | 121 | 10 |
| WA | 45 | 0 |

## Mismatch clusters

| Cluster | Count |
|---|---|
| TX attempt>1=true fields=expires_on | 10 |

## First examples

- #77 TX VIN …GLWUM6 attempt=3 requested_at=2027-01-12 original_issued_at=2026-12-23: legacy `map[expires_on:2027-03-13]` vs go `map[expires_on:2027-02-21]` (fields: expires_on)
- #144 TX VIN …5N8ULR attempt=3 requested_at=2027-02-10 original_issued_at=2027-01-21: legacy `map[expires_on:2027-04-11]` vs go `map[expires_on:2027-03-22]` (fields: expires_on)
- #228 TX VIN …9TNTM3 attempt=3 requested_at=2027-02-06 original_issued_at=2027-01-21: legacy `map[expires_on:2027-04-07]` vs go `map[expires_on:2027-03-22]` (fields: expires_on)
- #244 TX VIN …GGLBJ9 attempt=3 requested_at=2026-12-17 original_issued_at=2026-12-04: legacy `map[expires_on:2027-02-15]` vs go `map[expires_on:2027-02-02]` (fields: expires_on)
- #246 TX VIN …WPCEEY attempt=3 requested_at=2026-12-03 original_issued_at=2026-11-04: legacy `map[expires_on:2027-02-01]` vs go `map[expires_on:2027-01-03]` (fields: expires_on)
