.PHONY: build format clean-test lint test run tidy

format:
	go fmt ./...

lint:
	golangci-lint run
	go vet ./pkg/...
	go vet ./integration/...

clean-test:
	go clean -testcache

test:
	$(MAKE) clean-test && go test -parallel 4 ./pkg/...

test-v:
	$(MAKE) clean-test && go test -v -cover ./pkg/...

test-race:
	$(MAKE) clean-test && go test -race ./pkg/...

test-integration:
	$(MAKE) clean-test && go test -v ./integration/...

clean:
	rm -f ./bin/

run:
	go run cmd/main.go

build:
	go build -o bin/opera ./cmd/main.go

tidy:
	go mod tidy
	go mod vendor
