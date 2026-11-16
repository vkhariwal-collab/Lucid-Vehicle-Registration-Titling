SVC := services/temp-tag-svc
.PHONY: help legacy-test record-golden go-test parity parity-fixed k8s-validate run curl-demo demo clean

help:            ## list targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*##/\t/'

legacy-test:     ## Step 2: RSpec characterization suite against legacy Ruby
	cd legacy && rspec --format documentation

record-golden:   ## re-record golden fixtures from legacy behaviour (review the diff!)
	ruby legacy/bin/record_golden.rb

go-test:         ## Step 3: vet + race tests, incl. golden-fixture parity
	cd $(SVC) && gofmt -l . | (! grep .) && go vet ./... && go test -race -count=1 ./...

parity:          ## shadow replay: 500 synthetic requests, legacy vs Go, must be 100%
	cd $(SVC) && go run ./cmd/parity -n 500 -min 1.0 -out ../../reports/parity_report.md

parity-fixed:    ## same replay with the approved TX fix enabled: shows exactly what changes
	cd $(SVC) && go run ./cmd/parity -n 500 -fix-tx -min 0 -out ../../reports/parity_report_fix_tx.md

k8s-validate:    ## schema-validate manifests (CRDs skipped)
	kubeconform -strict -summary -ignore-missing-schemas -skip Kustomization deploy/k8s/base/*.yaml deploy/k8s/overlays/*/virtualservice.yaml

run:             ## run the Go service on :8080
	cd $(SVC) && go run ./cmd/temp-tag-svc

curl-demo:       ## idempotency demo against a running service
	./scripts/curl_demo.sh

demo: legacy-test go-test parity parity-fixed k8s-validate  ## everything, in demo order
	@echo "\nAll demo checks passed. Reports in reports/."

clean:
	rm -f reports/*.md
