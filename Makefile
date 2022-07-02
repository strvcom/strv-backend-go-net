GO = $(shell which go)

.PHONY:
	test \
	fmt \
	vet

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

test:
	set -eo pipefail
	$(GO) test ./... -cover
