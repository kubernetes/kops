# Copyright 2016 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build the influxdb image for amd64, arm, arm64, ppc64le and s390x
#
# Usage:
# 	[PREFIX=gcr.io/google_containers] [ARCH=amd64] make (build|push)

all: build

VERSION?=v4.0.2
PREFIX?=gcr.io/google_containers
ARCH?=amd64
TEMP_DIR:=$(shell mktemp -d)
LDFLAGS=-w -X main.version=$(VERSION) -X main.commit=unknown-dev -X main.timestamp=0 -extldflags '-static'
DEB_BUILD=4.0.2-1481203731
KUBE_CROSS_IMAGE=gcr.io/google_containers/kube-cross:v1.7.3-0

# s390x
ALL_ARCHITECTURES=amd64 arm arm64 ppc64le
ML_PLATFORMS=linux/amd64,linux/arm,linux/arm64,linux/ppc64le,linux/s390x

# Set default base image dynamically for each arch
ifeq ($(ARCH),amd64)
	BASEIMAGE?=busybox
	CC=gcc
endif
ifeq ($(ARCH),arm)
	BASEIMAGE?=armhf/busybox
	CC=arm-linux-gnueabi-gcc
endif
ifeq ($(ARCH),arm64)
	BASEIMAGE?=aarch64/busybox
	CC=aarch64-linux-gnu-gcc
endif
ifeq ($(ARCH),ppc64le)
	BASEIMAGE?=ppc64le/busybox
	CC=powerpc64le-linux-gnu-gcc
endif
ifeq ($(ARCH),s390x)
	BASEIMAGE?=s390x/busybox
	CC=s390x-linux-gnu-gcc
endif

build:
	# Copy the whole directory to a temporary dir and set the base image
	cp -r ./* $(TEMP_DIR)

	cd $(TEMP_DIR) && sed -i "s|BASEIMAGE|$(BASEIMAGE)|g" Dockerfile

	# This script downloads the official grafana deb package, compiles grafana for the right architecture which replaces the built-in, dynamically linked binaries
	# Then the rootfs will be compressed into a tarball again, in order to be ADDed in the Dockerfile.
	# Lastly, it compiles the go helper
	docker run --rm -it -v $(TEMP_DIR):/build -w /go/src/github.com/grafana/grafana $(KUBE_CROSS_IMAGE) /bin/bash -c "\
		curl -sSL https://github.com/grafana/grafana/archive/$(VERSION).tar.gz | tar -xz --strip-components=1 \
		&& curl -sSL https://grafanarel.s3.amazonaws.com/builds/grafana_$(DEB_BUILD)_amd64.deb > /tmp/grafana.deb \
		&& mkdir /tmp/grafanarootfs && dpkg -x /tmp/grafana.deb /tmp/grafanarootfs \
		&& CGO_ENABLED=1 GOARCH=$(ARCH) CC=$(CC) go build --ldflags=\"$(LDFLAGS)\" -o /tmp/grafanarootfs/usr/sbin/grafana-server ./pkg/cmd/grafana-server \
		&& CGO_ENABLED=1 GOARCH=$(ARCH) CC=$(CC) go build --ldflags=\"$(LDFLAGS)\" -o /tmp/grafanarootfs/usr/sbin/grafana-cli ./pkg/cmd/grafana-cli \
		&& cd /tmp/grafanarootfs && tar -cf /build/grafana.tar . \
		&& cd /build && CGO_ENABLED=0 GOARCH=$(ARCH) go build -o setup_grafana setup_grafana.go"

	docker build --pull -t $(PREFIX)/heapster-grafana-$(ARCH):$(VERSION) $(TEMP_DIR)

	rm -rf $(TEMP_DIR)

# Should depend on target: ./manifest-tool
push: gcr-login $(addprefix sub-push-,$(ALL_ARCHITECTURES))
#	./manifest-tool push from-args --platforms $(ML_PLATFORMS) --template $(PREFIX)/heapster-grafana-ARCH:$(VERSION) --target $(PREFIX)/heapster-grafana:$(VERSION)

sub-push-%:
	$(MAKE) ARCH=$* PREFIX=$(PREFIX) VERSION=$(VERSION) build
	docker push $(PREFIX)/heapster-grafana-$*:$(VERSION)

# TODO(luxas): As soon as it's working to push fat manifests to gcr.io, reenable this code
#./manifest-tool:
#	curl -sSL https://github.com/luxas/manifest-tool/releases/download/v0.3.0/manifest-tool > manifest-tool
#	chmod +x manifest-tool

gcr-login:
ifeq ($(findstring gcr.io,$(PREFIX)),gcr.io)
	@echo "If you are pushing to a gcr.io registry, you have to be logged in via 'docker login'; 'gcloud docker push' can't push manifest lists yet."
	@echo "This script is automatically logging you in now."
	gcloud docker -a
endif
