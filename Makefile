GO = $(shell which go)

.PHONY: fmt
fmt:
	$(GO) fmt ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	set -eo pipefail
	$(GO) test ./... -cover
