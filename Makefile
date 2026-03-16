.PHONY: all build clean install-lsp gen-raylib raylib check-deps

RAYLIB_VERSION := 5.5
RAYLIB_DIR := raylib
DLDIR := $(CURDIR)/.dlcache

ifeq ($(OS),Windows_NT)
SHELL := C:/PROGRA~1/Git/bin/bash.exe
.SHELLFLAGS := -c
EXE_SUFFIX := .exe
GO := "/c/Program Files/Go/bin/go.exe"
NPM := "/c/Program Files/nodejs/npm"
NPX := "/c/Program Files/nodejs/npx"
RAYLIB_ARCHIVE := raylib-$(RAYLIB_VERSION)_win64_mingw-w64.zip
RAYLIB_URL := https://github.com/raysan5/raylib/releases/download/$(RAYLIB_VERSION)/$(RAYLIB_ARCHIVE)
else ifeq ($(shell uname),Darwin)
EXE_SUFFIX :=
GO := go
NPM := npm
NPX := npx
RAYLIB_ARCHIVE := raylib-$(RAYLIB_VERSION)_macos.tar.gz
RAYLIB_URL := https://github.com/raysan5/raylib/releases/download/$(RAYLIB_VERSION)/$(RAYLIB_ARCHIVE)
else
EXE_SUFFIX :=
GO := go
NPM := npm
NPX := npx
RAYLIB_ARCHIVE := raylib-$(RAYLIB_VERSION)_linux_amd64.tar.gz
RAYLIB_URL := https://github.com/raysan5/raylib/releases/download/$(RAYLIB_VERSION)/$(RAYLIB_ARCHIVE)
endif

# ── Main target ──────────────────────────────────────────────
all: check-deps raylib gen-raylib build install-lsp
	@echo ""
	@echo "=== All done! ==="
	@echo "  kl:      bin/kl$(EXE_SUFFIX)"
	@echo "  kl-lsp:  bin/kl-lsp$(EXE_SUFFIX)"
	@echo "  vscode:  extension installed"
	@echo ""
	@echo "Restart VS Code or run 'Developer: Reload Window' to activate."

# ── Check prerequisites ─────────────────────────────────────
check-deps:
	@echo "=== Checking prerequisites ==="
	@command -v gcc >/dev/null 2>&1 || { echo "ERROR: gcc not found. Install MinGW-w64 (scoop install mingw)."; exit 1; }
	@$(GO) version >/dev/null 2>&1 || { echo "ERROR: Go not found."; exit 1; }
	@command -v node >/dev/null 2>&1 || { echo "ERROR: Node.js not found."; exit 1; }
	@command -v python3 >/dev/null 2>&1 || { echo "ERROR: python3 not found."; exit 1; }
	@command -v curl >/dev/null 2>&1 || { echo "ERROR: curl not found."; exit 1; }
	@echo "OK"

# ── Download & extract raylib ────────────────────────────────
raylib: $(RAYLIB_DIR)/lib/libraylib.a

$(RAYLIB_DIR)/lib/libraylib.a:
	@echo "=== Downloading raylib $(RAYLIB_VERSION) ==="
ifeq ($(OS),Windows_NT)
	@mkdir -p "$(DLDIR)" && curl -sL "$(RAYLIB_URL)" -o "$(DLDIR)/raylib.zip" && cd "$(DLDIR)" && unzip -qo raylib.zip
	@rm -rf $(RAYLIB_DIR)/lib $(RAYLIB_DIR)/include && mkdir -p $(RAYLIB_DIR)/lib $(RAYLIB_DIR)/include
	@cp "$(DLDIR)/raylib-$(RAYLIB_VERSION)_win64_mingw-w64/lib/"*.a $(RAYLIB_DIR)/lib/
	@cp "$(DLDIR)/raylib-$(RAYLIB_VERSION)_win64_mingw-w64/lib/"*.dll $(RAYLIB_DIR)/lib/ 2>/dev/null || true
	@cp "$(DLDIR)/raylib-$(RAYLIB_VERSION)_win64_mingw-w64/include/"*.h $(RAYLIB_DIR)/include/
else
	@mkdir -p "$(DLDIR)" && curl -sL "$(RAYLIB_URL)" -o "$(DLDIR)/raylib.tar.gz" && cd "$(DLDIR)" && tar xzf raylib.tar.gz
	@rm -rf $(RAYLIB_DIR)/lib $(RAYLIB_DIR)/include && mkdir -p $(RAYLIB_DIR)/lib $(RAYLIB_DIR)/include
	@cp "$(DLDIR)"/raylib-$(RAYLIB_VERSION)*/lib/* $(RAYLIB_DIR)/lib/
	@cp "$(DLDIR)"/raylib-$(RAYLIB_VERSION)*/include/*.h $(RAYLIB_DIR)/include/
endif
	@rm -rf "$(DLDIR)"
	@echo "OK — raylib $(RAYLIB_VERSION) installed to $(RAYLIB_DIR)/"

# ── Generate raylib bindings ─────────────────────────────────
gen-raylib:
	@echo "=== Generating raylib bindings ==="
	@mkdir -p "$(DLDIR)" && curl -sL "https://raw.githubusercontent.com/raysan5/raylib/$(RAYLIB_VERSION)/parser/output/raylib_api.json" -o "$(DLDIR)/raylib_api.json" && python3 tools/gen-raylib/gen.py "$(DLDIR)/raylib_api.json" && rm -rf "$(DLDIR)"
	@echo "OK"

# ── Build compiler + LSP ────────────────────────────────────
build:
	@echo "=== Building kl ==="
	@$(GO) build -o bin/kl$(EXE_SUFFIX) ./cmd/kl
	@echo "=== Building kl-lsp ==="
	@$(GO) build -o bin/kl-lsp$(EXE_SUFFIX) ./cmd/kl-lsp
	@echo "OK"

# ── Install VS Code extension ───────────────────────────────
install-lsp: build
	@echo "=== Installing VS Code extension ==="
	@cd editors/vscode-klang && $(NPM) install --silent
	@cd editors/vscode-klang && $(NPX) tsc -p .
	@rm -rf "$$HOME/.vscode/extensions/klang"
	@ln -sfn "$(CURDIR)/editors/vscode-klang" "$$HOME/.vscode/extensions/klang"
	@echo "OK"

# ── Clean ────────────────────────────────────────────────────
clean:
	rm -rf build/ bin/

clean-all: clean
	rm -rf $(RAYLIB_DIR)/lib $(RAYLIB_DIR)/include "$(DLDIR)"
