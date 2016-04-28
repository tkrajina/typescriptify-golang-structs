build:
	ftmpl jsonconv
	go run q.go
	go install ./...
test:
	go test ./...
	go run example/example.go
	tsc browser_test/example_output.ts
