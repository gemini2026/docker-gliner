BINARY := docker-gliner
PREFIX ?= $(HOME)/.local/bin

.PHONY: build test test-server fmt vet check install clean

build: ## Build the provider binary
	go build -o $(BINARY) ./cmd/docker-gliner

test: ## Run Go tests
	go test ./...

test-server: ## Run the bundled server's Python tests
	cd server && python -m pytest -q

fmt: ## Format Go sources
	gofmt -w .

vet: ## Vet Go sources
	go vet ./...

check: vet ## Lint + test everything
	gofmt -l . | tee /dev/stderr | (! read)
	go test ./...
	cd server && ruff check . && python -m pytest -q

install: build ## Install the binary + bundled server beside it onto $(PREFIX)
	install -d $(PREFIX) $(PREFIX)/server
	install -m 0755 $(BINARY) $(PREFIX)/$(BINARY)
	install -m 0644 server/serve_gliner.py $(PREFIX)/server/serve_gliner.py
	install -m 0644 server/requirements.txt $(PREFIX)/server/requirements.txt
	@echo "Installed $(BINARY) to $(PREFIX) (ensure it is on \$$PATH)"

clean: ## Remove build artifacts
	rm -f $(BINARY)
