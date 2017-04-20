.DEFAULT: all
.PHONY: all exes testrunner update tests lint publish-one-arch $(PUBLISH) clean clean-bin clean-work-dir prerequisites build run-smoketests

# If you can use docker without being root, you can do "make SUDO="
SUDO=$(shell docker info >/dev/null 2>&1 || echo "sudo -E")
BUILD_IN_CONTAINER=true
RM=--rm
RUN_FLAGS=-ti
COVERAGE=

# This specifies the architecture we're building for
ARCH?=amd64

# We're using QEMU to be able to build Docker images for other architectures
# from amd64 machines.
QEMU_VERSION=v2.7.0

# A list of all supported architectures here. Should be named as Go is naming platforms
# All supported architectures must have an "ifeq" block below that customizes the parameters
ALL_ARCHITECTURES=amd64 arm arm64
ML_PLATFORMS=linux/amd64,linux/arm,linux/arm64

ifeq ($(ARCH),amd64)
# The architecture to use when downloading the docker binary
	WEAVEEXEC_DOCKER_ARCH?=x86_64

# The name of the alpine baseimage to use as the base for weave images
	ALPINE_BASEIMAGE?=alpine:3.4

# The extension for the made images
# Specifying none means for example weaveworks/weave:latest
	ARCH_EXT?=

# The name of the gcc we're using for compiling C code
	CC=gcc

# Optional parameters that can be passed when linking C code into the Go binary
	CGO_LDFLAGS=
endif
ifeq ($(ARCH),arm)
# The architecture to use when downloading the docker binary
	WEAVEEXEC_DOCKER_ARCH?=armel

# Using the (semi-)official alpine image
	ALPINE_BASEIMAGE?=armhf/alpine:3.4

# arm images have the -arm suffix, for instance weaveworks/weave-arm:latest
	ARCH_EXT?=-arm

# The name of the gcc binary
	CC=arm-linux-gnueabihf-gcc

# The architecture name to use when downloading a prebuilt QEMU binary
	QEMUARCH=arm

# In the weaveworks/build image; libpcap libraries for arm are placed here
# Tell the gcc linker to search for libpcap here
	CGO_LDFLAGS="-L/usr/local/lib/$(CC)"
endif
ifeq ($(ARCH),arm64)
# Use the arm 32-bit docker client variant
	WEAVEEXEC_DOCKER_ARCH?=armel

# Using the (semi-)official alpine image
	ALPINE_BASEIMAGE?=aarch64/alpine:3.5

# arm64 images have the -arm64 suffix, for instance weaveworks/weave-arm64:latest
	ARCH_EXT?=-arm64

# The name of the gcc binary
	CC=aarch64-linux-gnu-gcc

# The architecture name to use when downloading a prebuilt QEMU binary
	QEMUARCH=aarch64

# In the weaveworks/build image; libpcap libraries for arm64 are placed here
# Tell the gcc linker to search for libpcap here
	CGO_LDFLAGS="-L/usr/local/lib/$(CC)"
endif

# The name of the user that this Makefile should produce image artifacts for. Can/should be overridden
DOCKERHUB_USER?=weaveworks
# The default version that's chosen when pushing the images. Can/should be overridden
WEAVE_VERSION?=git-$(shell git rev-parse --short=12 HEAD)

# Paths to all relevant binaries that should be compiled
WEAVER_EXE=prog/weaver/weaver
WEAVEPROXY_EXE=prog/weaveproxy/weaveproxy
SIGPROXY_EXE=prog/sigproxy/sigproxy
KUBEPEERS_EXE=prog/kube-peers/kube-peers
WEAVENPC_EXE=prog/weave-npc/weave-npc
WEAVEWAIT_EXE=prog/weavewait/weavewait
WEAVEWAIT_NOOP_EXE=prog/weavewait/weavewait_noop
WEAVEWAIT_NOMCAST_EXE=prog/weavewait/weavewait_nomcast
WEAVEUTIL_EXE=prog/weaveutil/weaveutil
RUNNER_EXE=tools/runner/runner
MANIFEST_TOOL_DIR=vendor/github.com/estesp/manifest-tool
MANIFEST_TOOL_EXE=$(MANIFEST_TOOL_DIR)/manifest-tool
TEST_TLS_EXE=test/tls/tls

# All binaries together in a list
EXES=$(WEAVER_EXE) $(SIGPROXY_EXE) $(KUBEPEERS_EXE) $(WEAVENPC_EXE) $(WEAVEPROXY_EXE) $(WEAVEWAIT_EXE) $(WEAVEWAIT_NOOP_EXE) $(WEAVEWAIT_NOMCAST_EXE) $(WEAVEUTIL_EXE) $(RUNNER_EXE) $(TEST_TLS_EXE) $(MANIFEST_TOOL_EXE)

# These stamp files are used to mark the current state of the build; whether an image has been built or not
BUILD_UPTODATE=.build.uptodate
WEAVER_UPTODATE=.weaver$(ARCH_EXT).uptodate
WEAVEEXEC_UPTODATE=.weaveexec$(ARCH_EXT).uptodate
PLUGIN_UPTODATE=.net-plugin$(ARCH_EXT).uptodate
WEAVEKUBE_UPTODATE=.weavekube$(ARCH_EXT).uptodate
WEAVENPC_UPTODATE=.weavenpc$(ARCH_EXT).uptodate
WEAVEDB_UPTODATE=.weavedb.uptodate

IMAGES_UPTODATE=$(WEAVER_UPTODATE) $(WEAVEEXEC_UPTODATE) $(WEAVEKUBE_UPTODATE) $(WEAVENPC_UPTODATE)

# The names of the images. Note that the images for other architectures than amd64 have a suffix in the image name.
WEAVER_IMAGE=$(DOCKERHUB_USER)/weave$(ARCH_EXT)
WEAVEEXEC_IMAGE=$(DOCKERHUB_USER)/weaveexec$(ARCH_EXT)
WEAVEKUBE_IMAGE=$(DOCKERHUB_USER)/weave-kube$(ARCH_EXT)
WEAVENPC_IMAGE=$(DOCKERHUB_USER)/weave-npc$(ARCH_EXT)
BUILD_IMAGE=weaveworks/weavebuild
WEAVEDB_IMAGE=$(DOCKERHUB_USER)/weavedb
PLUGIN_IMAGE=$(DOCKERHUB_USER)/net-plugin

IMAGES=$(WEAVER_IMAGE) $(WEAVEEXEC_IMAGE) $(WEAVEKUBE_IMAGE) $(WEAVENPC_IMAGE) $(WEAVEDB_IMAGE)

PLUGIN_WORK_DIR="prog/net-plugin/rootfs"
PLUGIN_BUILD_IMG="plugin-builder"

PUBLISH=publish_weave publish_weaveexec publish_weave-kube publish_weave-npc
PUSH_ML=push_ml_weave push_ml_weaveexec push_ml_weave-kube push_ml_weave-npc

WEAVE_EXPORT=weave$(ARCH_EXT).tar.gz

DOCKER_VERSION=1.10.3
DOCKER_DISTRIB=prog/weaveexec/docker-$(DOCKER_VERSION).tgz
DOCKER_DISTRIB_URL=https://get.docker.com/builds/Linux/$(WEAVEEXEC_DOCKER_ARCH)/docker-$(DOCKER_VERSION).tgz
NETGO_CHECK=@strings $@ | grep cgo_stub\\\.go >/dev/null || { \
	rm $@; \
	echo "\nYour go standard library was built without the 'netgo' build tag."; \
	echo "To fix that, run"; \
	echo "    sudo go clean -i net"; \
	echo "    sudo go install -tags netgo std"; \
	false; \
}
# The flags we are passing to go build. -extldflags -static for making a static binary, 
# -linkmode external for linking external C libraries into the binary, -X main.version for telling the
# Go binary which version it is, -tags netgo for enforcing the native Go DNS resolver
BUILD_FLAGS=-i -ldflags "-linkmode external -extldflags -static -X main.version=$(WEAVE_VERSION)" -tags netgo

PACKAGE_BASE=$(shell go list -e ./)

all: $(WEAVE_EXPORT)
testrunner: $(RUNNER_EXE) $(TEST_TLS_EXE)

$(WEAVER_EXE) $(WEAVEPROXY_EXE) $(WEAVEUTIL_EXE): common/*.go common/*/*.go net/*.go net/*/*.go
$(WEAVER_EXE): router/*.go ipam/*.go ipam/*/*.go db/*.go nameserver/*.go prog/weaver/*.go
$(WEAVER_EXE): api/*.go plugin/*.go plugin/*/*
$(WEAVEPROXY_EXE): proxy/*.go prog/weaveproxy/*.go
$(WEAVEUTIL_EXE): prog/weaveutil/*.go net/*.go plugin/net/*.go plugin/ipam/*.go db/*.go
$(SIGPROXY_EXE): prog/sigproxy/*.go
$(KUBEPEERS_EXE): prog/kube-peers/*.go
$(WEAVENPC_EXE): prog/weave-npc/*.go npc/*.go npc/*/*.go
$(TEST_TLS_EXE): test/tls/*.go
$(RUNNER_EXE): tools/runner/*.go
$(MANIFEST_TOOL_EXE): $(MANIFEST_TOOL_DIR)/*.go
$(WEAVEWAIT_NOOP_EXE): prog/weavewait/*.go
$(WEAVEWAIT_EXE): prog/weavewait/*.go net/*.go
$(WEAVEWAIT_NOMCAST_EXE): prog/weavewait/*.go net/*.go
tests: tools/.git
lint: tools/.git

ifeq ($(BUILD_IN_CONTAINER),true)

# This make target compiles all binaries inside of the weaveworks/build container
# It bind-mounts the source into the container and passes all important variables
exes $(EXES) tests lint: $(BUILD_UPTODATE)
	git submodule update --init
# Containernetworking has another copy of vishvananda/netlink which leads to duplicate definitions
	-@rm -r vendor/github.com/containernetworking/cni/vendor
	@mkdir -p $(shell pwd)/.pkg
	$(SUDO) docker run $(RM) $(RUN_FLAGS) \
	    -v $(shell pwd):/go/src/github.com/weaveworks/weave \
		-v $(shell pwd)/.pkg:/go/pkg \
		-e GOARCH=$(ARCH) -e CGO_ENABLED=1 -e GOOS=linux -e CIRCLECI -e CIRCLE_BUILD_NUM -e CIRCLE_NODE_TOTAL -e CIRCLE_NODE_INDEX -e COVERDIR -e SLOW -e DEBUG \
		$(BUILD_IMAGE) COVERAGE=$(COVERAGE) WEAVE_VERSION=$(WEAVE_VERSION) CC=$(CC) QEMUARCH=$(QEMUARCH) CGO_LDFLAGS=$(CGO_LDFLAGS) $@
	touch $@

else

exes: $(EXES)

$(WEAVER_EXE) $(WEAVEPROXY_EXE):
ifeq ($(COVERAGE),true)
	$(eval COVERAGE_MODULES := $(shell (go list ./$(@D); go list -f '{{join .Deps "\n"}}' ./$(@D) | grep "^$(PACKAGE_BASE)/") | grep -v "^$(PACKAGE_BASE)/vendor/" | paste -s -d,))
	go test -c -o ./$@ $(BUILD_FLAGS) -v -covermode=atomic -coverpkg $(COVERAGE_MODULES) ./$(@D)/
else
	go build $(BUILD_FLAGS) -o $@ ./$(@D)
endif
	$(NETGO_CHECK)

$(WEAVEUTIL_EXE) $(KUBEPEERS_EXE) $(WEAVENPC_EXE):
	go build $(BUILD_FLAGS) -o $@ ./$(@D)
	$(NETGO_CHECK)

$(WEAVEWAIT_EXE):
	go build $(BUILD_FLAGS) -tags "netgo iface mcast" -o $@ ./$(@D)
	$(NETGO_CHECK)

$(WEAVEWAIT_NOMCAST_EXE):
	go build $(BUILD_FLAGS) -tags "netgo iface" -o $@ ./$(@D)
	$(NETGO_CHECK)

# These programs need a separate rule as they fail the netgo check in
# the main build stanza due to not importing net package
$(SIGPROXY_EXE) $(TEST_TLS_EXE) $(WEAVEWAIT_NOOP_EXE) $(RUNNER_EXE) $(MANIFEST_TOOL_EXE):
	go build $(BUILD_FLAGS) -o $@ ./$(@D)

tests:
	./tools/test -no-go-get -netgo -timeout 8m

lint:
	./tools/lint -nocomment -notestpackage

endif

# This rule makes sure the build image is up-to-date.
# It also makes sure the multiarch hooks are reqistered in the kernel so the QEMU emulation works
$(BUILD_UPTODATE): build/*
	$(SUDO) docker build -t $(BUILD_IMAGE) build/
	$(SUDO) docker run --rm --privileged multiarch/qemu-user-static:register --reset
	touch $@

# Creates the Dockerfile.your-user-here file from the template
# Also replaces all placeholders with real values
# If the architecture is amd64, it deletes all CROSS_BUILD lines
# but otherwise, it only removes the "CROSS_BUILD_" placeholder and downloads QEMU
%/Dockerfile.$(DOCKERHUB_USER): %/Dockerfile.template
	echo "DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|ARCH_EXT|$(ARCH_EXT)|g;s|ALPINE_BASEIMAGE|$(ALPINE_BASEIMAGE)|g;s|QEMUARCH|$(QEMUARCH)"
	sed -e "s|DOCKERHUB_USER|$(DOCKERHUB_USER)|g;s|ARCH_EXT|$(ARCH_EXT)|g;s|ALPINE_BASEIMAGE|$(ALPINE_BASEIMAGE)|g;s|QEMUARCH|$(QEMUARCH)|g" $^ > $@
ifeq ($(ARCH),amd64)
# When building "normally" for amd64, remove the whole line, it has no part in the amd64 image
	sed -i "/CROSS_BUILD_/d" $@
else
# When cross-building, only the placeholder "CROSS_BUILD_" should be removed
# Register /usr/bin/qemu-ARCH-static as the handler for ARM binaries in the kernel
	curl -sSL https://github.com/multiarch/qemu-user-static/releases/download/$(QEMU_VERSION)/x86_64_qemu-$(QEMUARCH)-static.tar.gz | tar -xz -C $(shell dirname $@)
	cd $(shell dirname $@) && sha256sum -c $(shell pwd)/build/shasums/qemu-$(QEMUARCH)-static.sha256sum
	sed -i "s/CROSS_BUILD_//g" $@
endif

# The targets below builds the weave images
$(WEAVER_UPTODATE): prog/weaver/Dockerfile.$(DOCKERHUB_USER) $(WEAVER_EXE) $(WEAVEEXEC_UPTODATE)
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker build -f prog/weaver/Dockerfile.$(DOCKERHUB_USER) -t $(WEAVER_IMAGE) prog/weaver
	touch $@

$(WEAVEEXEC_UPTODATE): prog/weaveexec/Dockerfile.$(DOCKERHUB_USER) prog/weaveexec/symlink $(DOCKER_DISTRIB) weave $(SIGPROXY_EXE) $(WEAVEPROXY_EXE) $(WEAVEWAIT_EXE) $(WEAVEWAIT_NOOP_EXE) $(WEAVEWAIT_NOMCAST_EXE) $(WEAVEUTIL_EXE)
	cp weave prog/weaveexec/weave
	cp $(SIGPROXY_EXE) prog/weaveexec/sigproxy
	cp $(WEAVEPROXY_EXE) prog/weaveexec/weaveproxy
	cp $(WEAVEWAIT_EXE) prog/weaveexec/weavewait
	cp $(WEAVEWAIT_NOOP_EXE) prog/weaveexec/weavewait_noop
	cp $(WEAVEWAIT_NOMCAST_EXE) prog/weaveexec/weavewait_nomcast
	cp $(WEAVEUTIL_EXE) prog/weaveexec/weaveutil
	tar -xf $(DOCKER_DISTRIB) usr/local/bin/docker -O > prog/weaveexec/docker
	chmod +x prog/weaveexec/docker
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker build -f prog/weaveexec/Dockerfile.$(DOCKERHUB_USER) -t $(WEAVEEXEC_IMAGE) prog/weaveexec
	touch $@

# Builds Docker plugin.
$(PLUGIN_UPTODATE): prog/net-plugin/launch.sh prog/net-plugin/config.json $(WEAVER_UPTODATE)
	-$(SUDO) docker rm -f $(PLUGIN_BUILD_IMG) 2>/dev/null
	$(SUDO) docker create --name=$(PLUGIN_BUILD_IMG) $(WEAVER_IMAGE) true
	rm -rf $(PLUGIN_WORK_DIR)
	mkdir $(PLUGIN_WORK_DIR)
	$(SUDO) docker export $(PLUGIN_BUILD_IMG) | tar -x -C $(PLUGIN_WORK_DIR)
	$(SUDO) docker rm -f $(PLUGIN_BUILD_IMG)
	cp prog/net-plugin/launch.sh $(PLUGIN_WORK_DIR)/home/weave/launch.sh
	-$(SUDO) docker plugin disable $(PLUGIN_IMAGE):$(WEAVE_VERSION) 2>/dev/null
	-$(SUDO) docker plugin rm $(PLUGIN_IMAGE):$(WEAVE_VERSION) 2>/dev/null
	$(SUDO) docker plugin create $(PLUGIN_IMAGE):$(WEAVE_VERSION) prog/net-plugin
	-$(SUDO) docker plugin disable $(PLUGIN_IMAGE):latest 2>/dev/null
	-$(SUDO) docker plugin rm $(PLUGIN_IMAGE):latest 2>/dev/null
	$(SUDO) docker plugin create $(PLUGIN_IMAGE):latest prog/net-plugin
	touch $@

$(WEAVEKUBE_UPTODATE): prog/weave-kube/Dockerfile.$(DOCKERHUB_USER) prog/weave-kube/launch.sh $(KUBEPEERS_EXE) $(WEAVER_UPTODATE)
	cp $(KUBEPEERS_EXE) prog/weave-kube/
	$(SUDO) docker build -f prog/weave-kube/Dockerfile.$(DOCKERHUB_USER) -t $(WEAVEKUBE_IMAGE) prog/weave-kube
	touch $@

$(WEAVENPC_UPTODATE): prog/weave-npc/Dockerfile.$(DOCKERHUB_USER) $(WEAVENPC_EXE) prog/weave-npc/ulogd.conf
	$(SUDO) docker build -f prog/weave-npc/Dockerfile.$(DOCKERHUB_USER) -t $(WEAVENPC_IMAGE) prog/weave-npc
	touch $@

$(WEAVEDB_UPTODATE): prog/weavedb/Dockerfile
	$(SUDO) docker build -t $(WEAVEDB_IMAGE) prog/weavedb
	touch $@

$(WEAVE_EXPORT): $(IMAGES_UPTODATE) $(WEAVEDB_UPTODATE)
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker save $(addsuffix :latest,$(IMAGES)) | gzip > $@

$(DOCKER_DISTRIB):
	curl -o $(DOCKER_DISTRIB) $(DOCKER_DISTRIB_URL)
	cd $(shell dirname $@) && sha256sum -c $(shell pwd)/build/shasums/docker-tgz-$(WEAVEEXEC_DOCKER_ARCH).sha256sum

tools/.git $(MANIFEST_TOOL_DIR)/.git:
	git submodule update --init

# CODE FOR PUBLISHING THE IMAGES

# Push plugin
plugin_publish: $(PLUGIN_UPTODATE)
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker plugin push $(PLUGIN_IMAGE):$(WEAVE_VERSION)
# "latest" means "stable release" here, so only push that if explicitly told to
ifeq ($(UPDATE_LATEST),true)
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker plugin push $(PLUGIN_IMAGE):latest
endif

# This target first runs "make publish" for each architecture
# Then it pushes the manifest lists
publish: $(addprefix sub-publish-,$(ALL_ARCHITECTURES)) $(WEAVEDB_UPTODATE)
	$(MAKE) DOCKER_HOST=$(DOCKER_HOST) DOCKERHUB_USER=$(DOCKERHUB_USER) WEAVE_VERSION=$(WEAVE_VERSION) UPDATE_LATEST=$(UPDATE_LATEST) $(PUSH_ML)
ifeq ($(PUBLISH_WEAVEDB),true)
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker push $(DOCKERHUB_USER)/weavedb:latest
endif

sub-publish-%:
	$(MAKE) ARCH=$* DOCKER_HOST=$(DOCKER_HOST) clean-bin
	$(MAKE) ARCH=$* DOCKER_HOST=$(DOCKER_HOST) DOCKERHUB_USER=$(DOCKERHUB_USER) WEAVE_VERSION=$(WEAVE_VERSION) UPDATE_LATEST=$(UPDATE_LATEST) $(IMAGES_UPTODATE)
	$(MAKE) ARCH=$* DOCKER_HOST=$(DOCKER_HOST) DOCKERHUB_USER=$(DOCKERHUB_USER) WEAVE_VERSION=$(WEAVE_VERSION) UPDATE_LATEST=$(UPDATE_LATEST) publish-one-arch

publish-one-arch: $(PUBLISH)

# This rule handles the pushing of the built images
$(PUBLISH): publish_%: $(IMAGES_UPTODATE)
# Tag :latest with the real version
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker tag  $(DOCKERHUB_USER)/$*$(ARCH_EXT) $(DOCKERHUB_USER)/$*$(ARCH_EXT):$(WEAVE_VERSION)
# Push the image with the arch suffix for arm and arm64, and without suffix for amd64
	$(MAKE) DOCKER_HOST=$(DOCKER_HOST) DOCKERHUB_USER=$(DOCKERHUB_USER) WEAVE_VERSION=$(WEAVE_VERSION) UPDATE_LATEST=$(UPDATE_LATEST) push_$*$(ARCH_EXT)

ifeq ($(ARCH),amd64)
# If the architecture is amd64, add the -amd64 suffix.
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker tag  $(DOCKERHUB_USER)/$*:$(WEAVE_VERSION) $(DOCKERHUB_USER)/$*-amd64:$(WEAVE_VERSION)
# If the version is latest, tag the -amd64 with latest as well
ifneq ($(UPDATE_LATEST),false)
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker tag  $(DOCKERHUB_USER)/$*-amd64:$(WEAVE_VERSION) $(DOCKERHUB_USER)/$*-amd64:latest
endif
# Push the image with the -amd64-suffix so BINARY-ARCH-named images exist for all arches, so manifest lists may be made and ARCH is replaceable in scripts
	$(MAKE) DOCKER_HOST=$(DOCKER_HOST) DOCKERHUB_USER=$(DOCKERHUB_USER) WEAVE_VERSION=$(WEAVE_VERSION) UPDATE_LATEST=$(UPDATE_LATEST) push_$*-amd64
endif

# This target pushes an image, and if UPDATE_LATEST is anything but false, also updates the latest tag
# It takes one parameter, the image name
push_%:
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker push $(DOCKERHUB_USER)/$*:$(WEAVE_VERSION)
ifneq ($(UPDATE_LATEST),false)
	$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker push $(DOCKERHUB_USER)/$*:latest
endif

# This target pushes a manifest list; it takes one parameter, the image name.
# The variable UPDATE_LATEST controls whether it updates the versioned tag and/or the latest tag
$(PUSH_ML): push_ml_%: $(MANIFEST_TOOL_EXE)
ifneq ($(UPDATE_LATEST),latest-only)
	$(MANIFEST_TOOL_EXE) push from-args --platforms $(ML_PLATFORMS) --template $(DOCKERHUB_USER)/$*-ARCH:$(WEAVE_VERSION) --target $(DOCKERHUB_USER)/$*:$(WEAVE_VERSION)
endif
ifneq ($(UPDATE_LATEST),false)
# Push the manifest list to :latest as well
	$(MANIFEST_TOOL_EXE) push from-args --platforms $(ML_PLATFORMS) --template $(DOCKERHUB_USER)/$*-ARCH:latest --target $(DOCKERHUB_USER)/$*:latest
endif

clean-work-dir:
	rm -rf $(PLUGIN_WORK_DIR)

clean-bin:
	-$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker rmi $(IMAGES)
	find prog -type f -name "Dockerfile.*" -not -name "Dockerfile.template" -print | xargs rm -f
	find prog -type f -name "*qemu-*" -print | xargs rm -f
	rm -rf $(EXES) $(IMAGES_UPTODATE) $(WEAVEDB_UPTODATE) weave*.tar.gz $(DOCKER_DISTRIB) prog/weaveexec/docker .pkg

clean: clean-bin clean-work-dir
	-$(SUDO) DOCKER_HOST=$(DOCKER_HOST) docker rmi $(BUILD_IMAGE)
	rm -rf test/tls/*.pem test/coverage.* test/coverage $(BUILD_UPTODATE) $(MANIFEST_TOOL_EXE)

build:
	$(SUDO) go clean -i net
	$(SUDO) go install -tags netgo std
	$(MAKE)

run-smoketests: all testrunner
	cd test && ./setup.sh && ./run_all.sh

integration-tests: all testrunner
# Usage:
#   $ make \
#     NAME="<prefix used to name VMs and other resources>" \
#     PROVIDER="<provider among {vagrant|gcp|aws|do}>" \
#     NUM_HOSTS="<# test machines>" \
#     PLAYBOOK="<filename>" \
#     RUNNER_ARGS="<...>" \
#     TESTS="<...>" \  # Can be set to only run one or a few tests instead of the full test suite.
#     DOCKER_VERSION=<...> \
#     KUBERNETES_VERSION=<...> \
#     KUBERNETES_CNI_VERSION=<...> \
#     <...> # See also run-integration-test.sh for all variables and individual functions.
#     integration-tests
#
	RUNNER_ARGS="-parallel" ./test/run-integration-tests.sh
