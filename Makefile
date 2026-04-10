.PHONY: build test lint clean install release release-snapshot

BINARY   := ajolote
BUILD_DIR := dist
CMD      := ./cmd/ajolote

build:
	go build -o $(BUILD_DIR)/$(BINARY) $(CMD)

# Install the binary into $GOPATH/bin (requires Go)
install:
	go install $(CMD)

test:
	go test ./... -v

lint:
	go vet ./...

clean:
	rm -rf $(BUILD_DIR)

# Build cross-platform release archives locally (requires goreleaser)
release-snapshot:
	goreleaser release --snapshot --clean

# Publish a real release (run after pushing a git tag)
release:
	goreleaser release --clean
