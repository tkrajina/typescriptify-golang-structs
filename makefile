.PHONY: build
build:
	go build -i -v -o /dev/null ./...

.PHONY: install
install:
	go install ./...

.PHONY: test
test: lint
	go test ./...
	go run example/example.go
	tsc browser_test/example_output.ts

.PHONY: lint
lint:
	go vet ./...
	golangci-lint run