# Makefile for building a macOS .pkg installer for securevibe.
#
# Prerequisites:
#   - macOS with Xcode Command Line Tools (pkgbuild, productbuild)
#   - Pre-built securevibe universal or platform binary at $(BINARY)
#
# Usage:
#   make BINARY=../../dist-build/securevibe-darwin-arm64 VERSION=2026.05.12
#   # produces: build/securevibe-$(VERSION).pkg

BINARY   ?= ../../securevibe
VERSION  ?= 0.0.0
INSTALL_DIR = /usr/local/bin
IDENTIFIER = com.skills-library.securevibe
BUILD_DIR  = build
PKG_ROOT   = $(BUILD_DIR)/pkg-root

.PHONY: all clean

all: $(BUILD_DIR)/securevibe-$(VERSION).pkg

$(BUILD_DIR)/securevibe-$(VERSION).pkg: $(BUILD_DIR)/component.pkg distribution.xml
	productbuild \
		--distribution distribution.xml \
		--package-path $(BUILD_DIR) \
		--version "$(VERSION)" \
		$@

$(BUILD_DIR)/component.pkg: $(PKG_ROOT)$(INSTALL_DIR)/securevibe
	pkgbuild \
		--identifier "$(IDENTIFIER)" \
		--version "$(VERSION)" \
		--root $(PKG_ROOT) \
		--install-location / \
		$@

$(PKG_ROOT)$(INSTALL_DIR)/securevibe: $(BINARY)
	mkdir -p $(dir $@)
	cp $(BINARY) $@
	chmod 755 $@

clean:
	rm -rf $(BUILD_DIR)
