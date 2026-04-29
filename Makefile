BINARY      := lazy-cron
VERSION     := 0.1.0
BUILD_DIR   := dist
CMD_PATH    := ./cmd/lazy-cron

LDFLAGS := -ldflags "-s -w -X github.com/domenez-dev/lazy-cron/internal/ui.AppVersion=$(VERSION)"

.PHONY: all build install clean deb arch tidy

all: build

## Download dependencies
tidy:
	go mod tidy

## Build the binary for the current platform
build: tidy
	mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) $(CMD_PATH)
	@echo "Binary: $(BUILD_DIR)/$(BINARY)"

## Install to /usr/local/bin (may need sudo)
install: build
	install -Dm755 $(BUILD_DIR)/$(BINARY) /usr/local/bin/$(BINARY)
	@echo "Installed to /usr/local/bin/$(BINARY)"

## Cross-compile for linux/amd64
build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY)-linux-amd64 $(CMD_PATH)

## Cross-compile for linux/arm64
build-arm64:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build $(LDFLAGS) \
		-o $(BUILD_DIR)/$(BINARY)-linux-arm64 $(CMD_PATH)

## Build all Linux targets
build-all: build-linux build-arm64

## Build Debian .deb package (requires nfpm)
deb: build-linux
	nfpm package --packager deb --config packaging/nfpm.yaml --target $(BUILD_DIR)/

## Build RPM package (requires nfpm)
rpm: build-linux
	nfpm package --packager rpm --config packaging/nfpm.yaml --target $(BUILD_DIR)/

## Build for Arch Linux via PKGBUILD (requires makepkg)
arch:
	@echo "Copy packaging/arch/PKGBUILD to a clean directory and run makepkg -si"

clean:
	rm -rf $(BUILD_DIR)
