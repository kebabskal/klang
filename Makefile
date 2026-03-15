.PHONY: build clean

build:
	go build -o bin/kl ./cmd/kl
	go build -o bin/kl-lsp ./cmd/kl-lsp

clean:
	rm -rf build/ bin/
