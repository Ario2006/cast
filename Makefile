.PHONY: build test vet fmt

build:
	go build ./cmd/cast

test:
	go test ./...

vet:
	go vet ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')
