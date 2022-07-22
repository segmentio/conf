export GO111MODULE=on

lint:
	go install honnef.co/go/tools/cmd/staticcheck@latest
	staticcheck ./...
	go vet ./...

test:
	go test -race ./...

ci: lint test
