run:
	go run ./cmd/main.go

test:
	go test ./bulkhead ./timeout ./internal
