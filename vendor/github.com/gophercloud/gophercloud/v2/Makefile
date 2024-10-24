undefine GOFLAGS

GOLANGCI_LINT_VERSION?=v1.57.1

ifeq ($(shell command -v podman 2> /dev/null),)
	RUNNER=docker
else
	RUNNER=podman
endif

# if the golangci-lint steps fails with the following error message:
#
#   directory prefix . does not contain main module or its selected dependencies
#
# you probably have to fix the SELinux security context for root directory plus your cache
#
#   chcon -Rt svirt_sandbox_file_t .
#   chcon -Rt svirt_sandbox_file_t ~/.cache/golangci-lint
lint:
	$(RUNNER) run -t --rm \
		-v $(shell pwd):/app \
		-v ~/.cache/golangci-lint/$(GOLANGCI_LINT_VERSION):/root/.cache \
		-w /app \
		-e GOFLAGS="-tags=acceptance" \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) golangci-lint run
.PHONY: lint

unit:
	go test ./...
.PHONY: unit

coverage:
	go test -covermode count -coverprofile cover.out -coverpkg=./... ./...
.PHONY: coverage

acceptance: acceptance-baremetal acceptance-blockstorage acceptance-compute acceptance-container acceptance-containerinfra acceptance-db acceptance-dns acceptance-identity acceptance-imageservice acceptance-keymanager acceptance-loadbalancer acceptance-messaging acceptance-networking acceptance-objectstorage acceptance-orchestration acceptance-placement acceptance-sharedfilesystems acceptance-workflow
.PHONY: acceptance

acceptance-baremetal:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/baremetal/...
.PHONY: acceptance-baremetal

acceptance-blockstorage:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/blockstorage/...
.PHONY: acceptance-blockstorage

acceptance-compute:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/compute/...
.PHONY: acceptance-compute

acceptance-container:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/container/...
.PHONY: acceptance-container

acceptance-containerinfra:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/containerinfra/...
.PHONY: acceptance-containerinfra

acceptance-db:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/db/...
.PHONY: acceptance-db

acceptance-dns:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/dns/...
.PHONY: acceptance-dns

acceptance-identity:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/identity/...
.PHONY: acceptance-identity

acceptance-image:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/imageservice/...
.PHONY: acceptance-image

acceptance-keymanager:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/keymanager/...
.PHONY: acceptance-keymanager

acceptance-loadbalancer:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/loadbalancer/...
.PHONY: acceptance-loadbalancer

acceptance-messaging:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/messaging/...
.PHONY: acceptance-messaging

acceptance-networking:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/networking/...
.PHONY: acceptance-networking

acceptance-objectstorage:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/objectstorage/...
.PHONY: acceptance-objectstorage

acceptance-orchestration:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/orchestration/...
.PHONY: acceptance-orchestration

acceptance-placement:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/placement/...
.PHONY: acceptance-placement

acceptance-sharedfilesystems:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/sharedfilesystems/...
.PHONY: acceptance-sharefilesystems

acceptance-workflow:
	go test -tags "fixtures acceptance" ./internal/acceptance/openstack/workflow/...
.PHONY: acceptance-workflow
