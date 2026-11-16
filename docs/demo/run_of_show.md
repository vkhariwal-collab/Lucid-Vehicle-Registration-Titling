# 15-minute run of show

Pre-flight (T-30 min):
- Run `make demo`. It must end with "All demo checks passed".
- Open the three pre-recorded Devin sessions (fallbacks).
- Start `make run` in a terminal.

| Time | Beat | Do | Say (one line) |
|---|---|---|---|
| 0:00 | Frame | Show the `legacy/` tree. Run `cd legacy && rspec registration_smoke_spec.rb`: 1 test. | "This is your situation in miniature: rules nobody wrote down, one smoke test." |
| 1:30 | **Step 1: Audit** | Live: paste Prompt 1 (devin_prompts.md). While it runs, open `docs/spec/registration_state_machine.md` (reference output). | "Before any code moves, Devin produces a spec your DMV Ops and Legal people can sign: every rule linked to a line of Rails." |
| 3:30 | Quirks | Scroll to §3 Open questions: TX resubmission, duplicate ELT liens, NY fall-through. | "These are decisions for you to make. Devin doesn't silently fix legal behaviour." |
| 5:00 | **Step 2: Lock** | Run `make legacy-test` (31 examples, quirks labelled). Show one fixture JSON. | "Current behaviour is now frozen in fixtures, including the bugs. Those same files will judge the Go code." |
| 8:00 | **Step 3: Port** | Show Devin's PR (pre-recorded) or the `services/temp-tag-svc` tree: rules/issuer (pure), idempotency, outbox, Protobuf/AsyncAPI, K8s. | "Idiomatic Go, your conventions from AGENTS.md: context everywhere, wrapped errors, no PII in logs." |
| 10:00 | Prove | Run `make go-test` (golden parity + race tests). Then `make parity`: **500/500**. | "Same fixtures, same answers. Plus 500 randomized shadow requests against the real Ruby." |
| 11:30 | Idempotency | Run `make curl-demo`: 201, then a 200 replay with the same tag, then 409 on key reuse. | "A retry can never mint a second tag or a second lien event." |
| 12:30 | Controlled change | Run `make parity-fixed`: 489/500, with all 11 diffs in TX resubmissions and only `expires_on` changed. | "When you approve the TX fix, the harness proves exactly what changed and that nothing else did." |
| 13:30 | Cutover | Show `deploy/k8s/overlays/{shadow,canary-5,full}`. | "Shadow, then 5% canary in CA, then full. Rollback is a route flip. Rails stays the system of record." |
| 14:30 | Hand back | Ask the question. | "What would your week-3 exit criteria be for a pilot on a real slice?" |
