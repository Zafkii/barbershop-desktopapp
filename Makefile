.PHONY: dev build run tidy fmt vet clean

dev: ## Run with hot reload (air)
	air

build: ## Build the desktop binary
	CGO_ENABLED=1 go build -o bin/barbershop-desktop .

run: build ## Build and run
	./bin/barbershop-desktop

tidy: ## Sync go.mod and go.sum
	go mod tidy

fmt: ## Format the code
	gofmt -w .

vet: ## Static checks
	go vet ./...

clean: ## Remove build artifacts
	rm -rf bin tmp
