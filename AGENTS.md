# Instructions for Devin (and humans) working in this repo

## Non-negotiables
- **Never change behaviour in `legacy/`.** You may add seams (e.g. clock injection) only if every characterization test stays green.
- Golden fixtures in `fixtures/golden/` are the contract between legacy and Go. If you re-record them (`make record-golden`), the PR must show and explain every changed `expected` block.
- A known legacy quirk is ported **as-is** first. Fixes go behind an `rules.Options` flag (default off) in a separate PR, with a parity report proving that nothing else changed.
- Unknown jurisdiction/scenario: document it in `docs/spec` as `NEEDS-SME-REVIEW`. Do not guess legal rules.

## Go conventions (temp-tag-svc)
- Go 1.23, stdlib first; no new dependency without an ADR.
- Pure domain logic in `internal/issuer` and `internal/rules`: no clock, no I/O, table-driven tests.
- `context.Context` on every I/O path; wrap errors with `%w`; no panics; no package-level mutable state.
- Every externally visible side effect goes through the outbox. Publishing is at-least-once; event IDs derive from the idempotency key.
- Logs never contain full VINs, names, addresses or postal codes (`maskVIN`).
- Gates before PR: `gofmt`, `go vet`, `go test -race`, `make parity` at 100%, `make k8s-validate`.

## PR checklist
- [ ] Spec updated (`docs/spec/registration_state_machine.md`) with file:line provenance
- [ ] Characterization tests added/updated and green
- [ ] Parity report attached (`reports/parity_report.md`)
- [ ] Event contract changes are backward compatible (new optional fields only) or bump `v2`
