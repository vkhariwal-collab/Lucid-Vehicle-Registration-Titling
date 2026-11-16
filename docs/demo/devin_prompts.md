# Prompts used in the demo (paste into Devin)

**Prompt 1: Audit and spec**
> Read `legacy/app/models/registration_state_machine.rb`, `legacy/app/services/temp_tag_issuer.rb`, `legacy/app/services/jurisdictions/*` and `legacy/app/services/elt/*`. Produce `docs/spec/registration_state_machine.md`:
> - a Mermaid state diagram;
> - a transition table with guards and side effects;
> - a per-jurisdiction rules table for CA, TX and ON.
>
> Cite file:line for every rule. List anything that looks like a bug or a legal risk as NEEDS-SME-REVIEW. Do not modify `legacy/`. Follow AGENTS.md.

**Prompt 2: Characterization tests**
> Using the spec, create golden fixtures under `fixtures/golden/temp_tag/` covering every rule row and every NEEDS-SME-REVIEW item. Record expected outputs from legacy with `legacy/bin/record_golden.rb`. Add RSpec characterization specs that:
> - assert legacy matches every fixture;
> - cover every state-machine transition, including illegal ones and the ELT retry behaviour.
>
> Tag quirks as `legacy_quirk`. All tests must pass without changing `legacy/app`.

**Prompt 3: Port the Temporary Tag slice**
> Port temp-tag issuance to Go 1.23 in `services/temp-tag-svc`, following AGENTS.md:
> - pure `internal/rules` and `internal/issuer`;
> - HTTP `POST /v1/temp-tags` requiring an `Idempotency-Key`;
> - an idempotency store plus a transactional outbox publishing `TempTagIssued`/`TempTagRejected` (contracts in `contracts/`).
>
> Preserve every legacy quirk. Put the TX resubmission fix behind a default-off option. Add:
> - a Go test that runs the SAME golden fixtures;
> - a parity replay tool against `legacy/bin/temp_tag_cli.rb`;
> - K8s manifests (Deployment, PDB, HPA, NetworkPolicy, ServiceMonitor) and Istio overlays for shadow, canary-5 and full.
>
> Gates: gofmt, go vet, go test -race, make parity at 100%, make k8s-validate. Open a PR with the parity report.
