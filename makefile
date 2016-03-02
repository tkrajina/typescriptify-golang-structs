test:
	go test ./...
	go run example/example.go
	tsc browser_test/example_output.ts
