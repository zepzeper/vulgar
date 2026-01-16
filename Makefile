# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

# Go build settings
BINARY_NAME = vulgar
MAIN_PACKAGE = ./cmd/vulgar
BUILD_DIR = build

# Linker flags for version injection
LDFLAGS = -ldflags "\
	-X main.Version=$(VERSION) \
	-X main.GitCommit=$(GIT_COMMIT) \
	-X main.BuildTime=$(BUILD_TIME) \
	-s -w"

.PHONY: all
all: build

.PHONY: build
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

.PHONY: release
release: clean
	@echo "Building release $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Cross-compile for multiple platforms
.PHONY: release-all
release-all: clean
	@echo "Building release $(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux   GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux   GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	GOOS=darwin  GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin  GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Built all platforms in $(BUILD_DIR)/"

.PHONY: test
test:
	go test ./...

.PHONY: test-v
test-v:
	go test -v ./...

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

.PHONY: lint
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

.PHONY: fmt
fmt:
	go fmt ./...
	gofmt -s -w .

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: clean
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f vulgar.prof vulgar.trace

.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME) to ~/go/bin/..."
	cp $(BINARY_NAME) ~/go/bin/

.PHONY: run
run: build
	./$(BINARY_NAME) $(ARGS)

.PHONY: modules
modules: build
	./$(BINARY_NAME) --list-modules

.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(GIT_COMMIT)"
	@echo "Built:   $(BUILD_TIME)"

.PHONY: help
help:
	@echo "Vulgar Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  build        Build the binary (default)"
	@echo "  release      Build optimized release binary"
	@echo "  release-all  Build for all platforms (linux, darwin, windows)"
	@echo "  test         Run tests"
	@echo "  test-v       Run tests with verbose output"
	@echo "  coverage     Run tests with coverage report"
	@echo "  lint         Run golangci-lint"
	@echo "  fmt          Format code"
	@echo "  tidy         Tidy go.mod"
	@echo "  clean        Remove build artifacts"
	@echo "  install      Install to ~/go/bin/"
	@echo "  modules      List available modules"
	@echo "  version      Show version info"
	@echo "  help         Show this help"

