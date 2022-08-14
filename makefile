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
	# Make sure dommandline tool works:
	go run tscriptify/main.go -package github.com/tkrajina/typescriptify-golang-structs/example/example-models -verbose -target tmp_classes.ts example/example-models/example_models.go
	go run tscriptify/main.go -package github.com/tkrajina/typescriptify-golang-structs/example/example-models -verbose -target tmp_interfaces.ts -interface example/example-models/example_models.go

.PHONY: lint
lint:
	go vet ./...
	-golangci-lint run
