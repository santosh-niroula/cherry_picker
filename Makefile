APP_NAME := cherry

MAIN_PACKAGE := ./cmd/main.go

RELEASE_DIR := release

VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -w"

PLATFORMS := linux darwin windows
ARCHITECTURES := amd64 arm64

# Build for Linux
.PHONY: linux
linux:
	@echo "Building $(APP_NAME) for Linux..."
	@mkdir -p $(RELEASE_DIR)/linux_amd64
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(RELEASE_DIR)/linux_amd64/$(APP_NAME) $(MAIN_PACKAGE)
	@mkdir -p $(RELEASE_DIR)/linux_arm64
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(RELEASE_DIR)/linux_arm64/$(APP_NAME) $(MAIN_PACKAGE)
	@echo "Done! Linux binaries available in $(RELEASE_DIR)/"

# Build for macOS
.PHONY: darwin
darwin:
	@echo "Building $(APP_NAME) for macOS..."
	@mkdir -p $(RELEASE_DIR)/darwin_amd64
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(RELEASE_DIR)/darwin_amd64/$(APP_NAME) $(MAIN_PACKAGE)
	@mkdir -p $(RELEASE_DIR)/darwin_arm64
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(RELEASE_DIR)/darwin_arm64/$(APP_NAME) $(MAIN_PACKAGE)
	@echo "Done! macOS binaries available in $(RELEASE_DIR)/"

# Build for Windows
.PHONY: windows
windows:
	@echo "Building $(APP_NAME) for Windows..."
	@mkdir -p $(RELEASE_DIR)/windows_amd64
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(RELEASE_DIR)/windows_amd64/$(APP_NAME).exe $(MAIN_PACKAGE)
	@mkdir -p $(RELEASE_DIR)/windows_arm64
	@GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(RELEASE_DIR)/windows_arm64/$(APP_NAME).exe $(MAIN_PACKAGE)
	@echo "Done! Windows binaries available in $(RELEASE_DIR)/"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(RELEASE_DIR)
	@echo "Done!"

# Create zip archives for releases
.PHONY: release
release:
	@echo "Creating release archives..."
	@for platform in $(PLATFORMS); do \
		for arch in $(ARCHITECTURES); do \
			if [ -d "$(RELEASE_DIR)/$${platform}_$${arch}" ]; then \
				(cd $(RELEASE_DIR)/$${platform}_$${arch} && zip -r ../$(APP_NAME)_$${platform}_$${arch}_$(VERSION).zip *); \
			fi \
		done \
	done
	@echo "Done! Release archives available in $(RELEASE_DIR)/"