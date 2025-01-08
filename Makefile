.PHONY: build build-arm build-amd32 all

# Default build for current architecture
build:
	go build -o my-incident-checker main.go

# Build for ARM (e.g., Raspberry Pi)
build-arm:
	GOOS=linux GOARCH=arm go build -o my-incident-checker-arm main.go

# Build for 32-bit AMD/Intel
build-amd32:
	GOOS=linux GOARCH=386 go build -o my-incident-checker-386 main.go

# Build all architectures
all: build build-arm build-amd32