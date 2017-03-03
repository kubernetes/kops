#
# Based on http://chrismckenzie.io/post/deploying-with-golang/
#

.PHONY: version all run dist clean
	
APP_NAME := heketi
CLIENT_PKG_NAME := heketi-client
SHA := $(shell git rev-parse --short HEAD)
BRANCH := $(subst /,-,$(shell git rev-parse --abbrev-ref HEAD))
VER := $(shell git describe)
ARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
DIR=.

ifdef APP_SUFFIX
  VERSION = $(VER)-$(subst /,-,$(APP_SUFFIX))
else
ifeq (master,$(BRANCH))
  VERSION = $(VER)
else
  VERSION = $(VER)-$(BRANCH)
endif
endif

# Go setup
GO=go

# Sources and Targets
EXECUTABLES :=$(APP_NAME)
# Build Binaries setting main.version and main.build vars
LDFLAGS :=-ldflags "-X main.HEKETI_VERSION=$(VERSION)"
# Package target
PACKAGE :=$(DIR)/dist/$(APP_NAME)-$(VERSION).$(GOOS).$(ARCH).tar.gz
CLIENT_PACKAGE :=$(DIR)/dist/$(APP_NAME)-client-$(VERSION).$(GOOS).$(ARCH).tar.gz
GOFILES=$(shell go list ./... | grep -v vendor)

.DEFAULT: all

all: server client

# print the version
version:
	@echo $(VERSION)

# print the name of the app
name:
	@echo $(APP_NAME)

# print the package path
package:
	@echo $(PACKAGE)

server:
	godep go build $(LDFLAGS) -o $(APP_NAME)

client:
	@$(MAKE) -C client/cli/go

run: server
	./$(APP_NAME)

test: 
	godep go test $(GOFILES)

clean:
	@echo Cleaning Workspace...
	rm -rf $(APP_NAME)
	rm -rf dist
	@$(MAKE) -C client/cli/go clean

$(PACKAGE): all
	@echo Packaging Binaries...
	@mkdir -p tmp/$(APP_NAME)
	@cp $(APP_NAME) tmp/$(APP_NAME)/
	@cp client/cli/go/heketi-cli tmp/$(APP_NAME)/
	@cp etc/heketi.json tmp/$(APP_NAME)/
	@mkdir -p $(DIR)/dist/
	tar -czf $@ -C tmp $(APP_NAME);
	@rm -rf tmp
	@echo
	@echo Package $@ saved in dist directory

$(CLIENT_PACKAGE): all
	@echo Packaging client Binaries...
	@mkdir -p tmp/$(CLIENT_PKG_NAME)/bin
	@mkdir -p tmp/$(CLIENT_PKG_NAME)/share/heketi/templates
	@cp client/cli/go/heketi-cli tmp/$(CLIENT_PKG_NAME)/bin
	@cp extras/openshift/templates/* tmp/$(CLIENT_PKG_NAME)/share/heketi/templates
	@mkdir -p $(DIR)/dist/
	tar -czf $@ -C tmp $(CLIENT_PKG_NAME);
	@rm -rf tmp
	@echo
	@echo Package $@ saved in dist directory

dist: $(PACKAGE) $(CLIENT_PACKAGE)

.PHONY: server client test clean name run version 
