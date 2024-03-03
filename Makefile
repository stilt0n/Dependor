.PHONY=test

test:
	go test . && go test ./internal/tokenizer && go test ./internal/config