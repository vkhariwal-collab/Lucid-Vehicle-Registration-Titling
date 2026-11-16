#!/usr/bin/env bash
# Shows idempotent replay and key-reuse protection against a running temp-tag-svc.
set -euo pipefail
URL=${URL:-http://localhost:8080}
BODY=$(jq -c .input fixtures/golden/temp_tag/CA_finance_elt_lien.json)
hdr() { echo; echo "== $1"; }
hdr "1st request (Idempotency-Key: order-8812:temp-tag)"
curl -s -i -X POST "$URL/v1/temp-tags" -H 'Idempotency-Key: order-8812:temp-tag' -d "$BODY" | sed -n '1p;/Idempotent/p;$p'
hdr "retry with same key (network blip) -> same tag, 200 + Idempotent-Replay"
curl -s -i -X POST "$URL/v1/temp-tags" -H 'Idempotency-Key: order-8812:temp-tag' -d "$BODY" | sed -n '1p;/Idempotent/p;$p'
hdr "same key, different VIN -> 409"
curl -s -i -X POST "$URL/v1/temp-tags" -H 'Idempotency-Key: order-8812:temp-tag' -d "${BODY/000102/000999}" | sed -n '1p;$p'
hdr "metrics"
curl -s "$URL/metrics"
