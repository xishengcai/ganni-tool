

test:
	go test --cover ./...

lint:
	gofmt -w .
	golangci-lint run ./...