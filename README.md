# lucid-dmv-migration-demo

A synthetic, runnable demo of a **Strangler Fig** migration of a legacy Rails-style vehicle **registration & titling** monolith to an event-driven **Go** service on **Kubernetes**. The first slice moved is **Temporary Tag Issuance**.

> All code, rules, fees and data are **synthetic and illustrative**. They are not Lucid code and not legal guidance.

## The three steps
| Step | What | Where | Check |
|---|---|---|---|
| 1. Audit | Human-verifiable spec of the undocumented state machine (CA/TX/ON + ELT), with file:line provenance and SME questions | `docs/spec/` | review |
| 2. Lock | RSpec characterization tests + golden fixtures that pin current behaviour, quirks included | `legacy/spec/characterization/`, `fixtures/golden/` | `make legacy-test` |
| 3. Port | Go service: pure rules, idempotency keys, transactional outbox, Protobuf/AsyncAPI contracts, K8s + Istio cutover overlays; validated on the **same** fixtures plus a 500-request shadow replay | `services/temp-tag-svc/`, `contracts/`, `deploy/k8s/` | `make go-test parity k8s-validate` |

## Quick start
Requirements: Ruby ≥ 3.0 with `rspec`, Go ≥ 1.23, `kubeconform` (optional), `jq` + `curl` for `curl-demo`.
```bash
make demo        # legacy-test, go-test, parity (100%), parity-fixed, k8s-validate
make run         # service on :8080   (in another terminal: make curl-demo)
```

## Layout
```
legacy/            Rails-style monolith (framework-free Ruby so it runs anywhere)
  app/models/registration_state_machine.rb   undocumented state machine
  app/services/temp_tag_issuer.rb            slice being strangled
  app/services/jurisdictions/{ca,tx,on}_rules.rb
  app/services/elt/lien_filing_service.rb    retries without idempotency (quirk)
  bin/temp_tag_cli.rb, bin/record_golden.rb
  spec/characterization/                     Step 2 output
fixtures/golden/temp_tag/*.json              shared contract: input + legacy-recorded expected
docs/spec/                                   Step 1 output (spec, rules matrix)
docs/adr/0001-strangler-fig-seam.md          seam, SoR, outbox, promotion gate
docs/demo/                                   run of show + exact Devin prompts
contracts/                                   Protobuf, JSON Schema, AsyncAPI (NATS JetStream / Kafka)
services/temp-tag-svc/                       Step 3 output (Go 1.23, stdlib only)
  cmd/temp-tag-svc   HTTP service      cmd/parity   legacy-vs-Go replay
  internal/{rules,issuer,idempotency,outbox,service,transport,metrics,parity}
deploy/k8s/base, overlays/{shadow,canary-5,full}
reports/                                     parity reports
```

## Production notes (what the demo stubs)
- The idempotency store and outbox are in memory. In production they are Postgres tables written in one transaction.
- `WriterPublisher` stands in for a NATS JetStream or Kafka publisher. The `Publisher` interface is the swap point.
- In the real engagement the legacy side is the actual Rails app with `rspec-rails`, and fixtures come from masked production payloads.
