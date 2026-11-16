# ADR 0001: Strangler Fig seam for Temporary Tag Issuance

**Status:** Proposed (demo) · **Context:** synthetic

## Decision
1. **Seam = HTTP route at the mesh.** `/temp_tags` is routed by Istio. The stages are SHADOW (mirror 100%, Rails authoritative), then CANARY (5%, CA only), then FULL. Rollback is a route-weight flip, with no deploy (`deploy/k8s/overlays/*`).
2. **Rails stays system of record** for the registration aggregate until the whole registration flow moves. Go owns only temp-tag issuance and its idempotency table.
3. **No dual writes to external parties.** In SHADOW, Go never calls DMV or ELT systems and publishes to a shadow subject. Rails updates its record from `TempTagIssued` events once Go is authoritative.
4. **Transactional outbox + idempotency key.** `UNIQUE(idempotency_key)` and the outbox row are written in one transaction. The relay is at-least-once, events are keyed by VIN for per-vehicle ordering, and consumers dedupe on `event_id`.
5. **Parity is the promotion gate.** Promote only when `parity_mismatch_total == 0` over N days of mirrored traffic (or every mismatch is an approved, flagged divergence) and the error budget is intact.

## Consequences
- Each slice moves independently. Legacy quirks are preserved until the business explicitly approves a fix.
- Idempotency is a deliberate improvement over legacy: legacy issues a new tag per retry, and its ELT retries can duplicate liens. This divergence is documented, not hidden.
