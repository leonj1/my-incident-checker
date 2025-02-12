.PHONY: build build-arm build-amd32 all clean release

VERSION ?= $(shell git describe --tags --always --dirty)
RELEASE_DIR = release

# Default build for current architecture
build: build-arm build-amd32
	go build -o my-incident-checker main.go

# Build for ARM (e.g., Raspberry Pi)
build-arm:
	GOOS=linux GOARCH=arm go build -o my-incident-checker-arm main.go

# Build for 32-bit AMD/Intel
build-amd32:
	GOOS=linux GOARCH=386 go build -o my-incident-checker-386 main.go

test:
	go test ./...

# Clean build artifacts
clean:
	rm -f my-incident-checker*
	rm -rf $(RELEASE_DIR)

# Prepare release artifacts
release: clean build
	mkdir -p $(RELEASE_DIR)
	cp my-incident-checker $(RELEASE_DIR)/my-incident-checker-$(VERSION)
	cp my-incident-checker-arm $(RELEASE_DIR)/my-incident-checker-arm-$(VERSION)
	cp my-incident-checker-386 $(RELEASE_DIR)/my-incident-checker-386-$(VERSION)
	cd $(RELEASE_DIR) && \
		tar czf my-incident-checker-$(VERSION).tar.gz my-incident-checker-* && \
		sha256sum my-incident-checker-$(VERSION).tar.gz > my-incident-checker-$(VERSION).sha256

# Build all architectures
all: build build-arm build-amd32
