.PHONY=test run

test:
	go test . && go test ./internal/tokenizer && go test ./internal/config

run:
	go run ./cmd