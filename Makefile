.PHONY: build clean install-lsp

ifeq ($(OS),Windows_NT)
SHELL := C:/PROGRA~1/Git/bin/bash.exe
EXE_SUFFIX := .exe
GO := "/c/Program Files/Go/bin/go.exe"
NPM := "/c/Program Files/nodejs/npm"
NPX := "/c/Program Files/nodejs/npx"
else
EXE_SUFFIX :=
GO := go
NPM := npm
NPX := npx
endif

build:
	$(GO) build -o bin/kl$(EXE_SUFFIX) ./cmd/kl
	$(GO) build -o bin/kl-lsp$(EXE_SUFFIX) ./cmd/kl-lsp

install-lsp: build
	cd editors/vscode-klang && $(NPM) install --silent
	cd editors/vscode-klang && $(NPX) tsc -p .
	rm -rf "$$HOME/.vscode/extensions/klang"
	ln -sfn "$(CURDIR)/editors/vscode-klang" "$$HOME/.vscode/extensions/klang"
	@echo "Done. Restart the language server in VS Code."

clean:
	rm -rf build/ bin/
