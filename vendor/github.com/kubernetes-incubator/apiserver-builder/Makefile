# Copyright 2017 The Kubernetes Authors.
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

# from cmd/
#  bash -c "find vendor/github.com/kubernetes-incubator/apiserver-builder -name BUILD.bazel| xargs sed -i='' s'|//pkg|//vendor/github.com/kubernetes-incubator/apiserver-builder/pkg|g'"

# from /

gazelle:
	find vendor -name BUILD | xargs rm
	find vendor -name BUILD.bazel | xargs rm
	gazelle fix -go_prefix github.com/kubernetes-incubator/apiserver-builder -external vendored .
	bash -c "find vendor/ -name BUILD.bazel |  xargs sed -i '' s'|//k8s.io/|//vendor/k8s.io/|g'"
	bash -c "find vendor/ -name BUILD |  xargs sed -i '' s'|//k8s.io/|//vendor/k8s.io/|g'"
	bash -c "find vendor/ -name BUILD.bazel |  xargs sed -i '' s'|cgo = True,|cgo = False,|g'"
	bash -c "find vendor/ -name BUILD |  xargs sed -i '' s'|cgo = True,|cgo = False,|g'"

NAME=apiserver-builder
VENDOR=kubernetes-incubator
VERSION=$(shell cat VERSION)
COMMIT=$(shell git rev-parse --verify HEAD)
DESCRIPTION=apiserver-builder implements libraries and tools to quickly and easily build Kubernetes apiservers to support custom resource types.
MAINTAINER=The Kubernetes Authors
URL=https://github.com/$(VENDOR)/$(NAME)
LICENSE=Apache-2.0

.PHONY: default
default: install

.PHONY: test
test:
	go test ./pkg/... ./cmd/...

.PHONY: install
install:
	go install -v ./pkg/... ./cmd/...

.PHONY: clean
clean:
	rm -rf *.deb *.rpm *.tar.gz ./release

.PHONY: build
build: clean ## Create release artefacts for darwin:amd64, linux:amd64 and windows:amd64. Requires etcd, glide, hg.
	go run ./cmd/apiserver-builder-release/main.go vendor --version $(VERSION) --commit $(COMMIT)
	go run ./cmd/apiserver-builder-release/main.go build --version $(VERSION)

.PHONY: package
package: package-linux-amd64

.PHONY: package-linux-amd64
package-linux-amd64: package-linux-amd64-deb package-linux-amd64-rpm

.PHONY: package-linux-amd64-deb
package-linux-amd64-deb: ## Create a Debian package. Requires jordansissel/fpm.
	fpm --force --name '$(NAME)' --version '$(VERSION)' \
	  --input-type tar \
	  --output-type deb \
	  --vendor '$(VENDOR)' \
	  --description '$(DESCRIPTION)' \
	  --url '$(URL)' \
	  --maintainer '$(MAINTAINER)' \
	  --license '$(LICENSE)' \
	  --package $(NAME)-$(VERSION)-amd64.deb \
	  --prefix /usr/local/apiserver-builder \
	  $(NAME)-$(VERSION)-linux-amd64.tar.gz

.PHONY: package-linux-amd64-rpm
package-linux-amd64-rpm: ## Create an RPM package. Requires jordansissel/fpm, rpm.
	fpm --force --name '$(NAME)' --version '$(VERSION)' \
	  --input-type tar \
	  --output-type rpm \
	  --vendor '$(VENDOR)' \
	  --description '$(DESCRIPTION)' \
	  --url '$(URL)' \
	  --maintainer '$(MAINTAINER)' \
	  --license '$(LICENSE)' \
	  --rpm-os linux \
	  --package $(NAME)-$(VERSION)-amd64.rpm \
	  --prefix /usr/local/apiserver-builder \
	  $(NAME)-$(VERSION)-linux-amd64.tar.gz
