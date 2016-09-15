all: kops

DOCKER_REGISTRY=gcr.io/must-override/
S3_BUCKET=s3://must-override/
GOPATH_1ST=$(shell echo ${GOPATH} | cut -d : -f 1)

TAG=1.3

ifndef VERSION
  VERSION := git-$(shell git rev-parse --short HEAD)
endif

kops: gobindata
	GO15VENDOREXPERIMENT=1 go install -ldflags "-X main.BuildVersion=${VERSION}" k8s.io/kops/cmd/kops/...

gobindata:
	GO15VENDOREXPERIMENT=1 go build -o ${GOPATH_1ST}/bin/go-bindata k8s.io/kops/vendor/github.com/jteeuwen/go-bindata/go-bindata
	cd ${GOPATH_1ST}/src/k8s.io/kops; ${GOPATH_1ST}/bin/go-bindata -o upup/models/bindata.go -pkg models -prefix upup/models/ upup/models/cloudup/... upup/models/config/... upup/models/nodeup/... upup/models/proto/...

# Build in a docker container with golang 1.X
# Used to test we have not broken 1.X
check-builds-in-go15:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.5 make -f /go/src/k8s.io/kops/Makefile kops

check-builds-in-go16:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.6 make -f /go/src/k8s.io/kops/Makefile kops

check-builds-in-go17:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.7 make -f /go/src/k8s.io/kops/Makefile kops

codegen: gobindata
	GO15VENDOREXPERIMENT=1 go install k8s.io/kops/upup/tools/generators/...
	GO15VENDOREXPERIMENT=1 PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/cloudup/awstasks
	GO15VENDOREXPERIMENT=1 PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/cloudup/gcetasks
	GO15VENDOREXPERIMENT=1 PATH=${GOPATH_1ST}/bin:${PATH} go generate k8s.io/kops/upup/pkg/fi/fitasks

test:
	GO15VENDOREXPERIMENT=1 go test k8s.io/kops/upup/pkg/... -args -v=1 -logtostderr

godeps:
	# I think strip-vendor is the workaround for 25572
	glide install --strip-vendor --strip-vcs

gofmt:
	gofmt -w -s cmd/
	gofmt -w -s upup/pkg/
	gofmt -w -s protokube/cmd
	gofmt -w -s protokube/pkg
	gofmt -w -s dns-controller/cmd
	gofmt -w -s dns-controller/pkg

crossbuild:
	mkdir -p .build/dist/
	GOOS=darwin GOARCH=amd64 go build -o .build/dist/darwin/amd64/kops -ldflags "-X main.BuildVersion=${VERSION}" -v k8s.io/kops/cmd/kops/...
	GOOS=linux GOARCH=amd64 go build -o .build/dist/linux/amd64/kops -ldflags "-X main.BuildVersion=${VERSION}" -v k8s.io/kops/cmd/kops/...
	#GOOS=windows GOARCH=amd64 go build -o .build/dist/windows/amd64/kops -ldflags "-X main.BuildVersion=${VERSION}" -v k8s.io/kops/cmd/kops/...

kops-dist: crossbuild
	mkdir -p .build/dist/
	(sha1sum .build/dist/darwin/amd64/kops | cut -d' ' -f1) > .build/dist/darwin/amd64/kops.sha1
	(sha1sum .build/dist/linux/amd64/kops | cut -d' ' -f1) > .build/dist/linux/amd64/kops.sha1

upload: nodeup-dist kops-dist
	rm -rf .build/s3
	mkdir -p .build/s3/kops/${VERSION}/linux/amd64/
	mkdir -p .build/s3/kops/${VERSION}/darwin/amd64/
	cp .build/dist/nodeup .build/s3/kops/${VERSION}/linux/amd64/nodeup
	cp .build/dist/nodeup.sha1 .build/s3/kops/${VERSION}/linux/amd64/nodeup.sha1
	cp .build/dist/linux/amd64/kops .build/s3/kops/${VERSION}/linux/amd64/kops
	cp .build/dist/linux/amd64/kops.sha1 .build/s3/kops/${VERSION}/linux/amd64/kops.sha1
	cp .build/dist/darwin/amd64/kops .build/s3/kops/${VERSION}/darwin/amd64/kops
	cp .build/dist/darwin/amd64/kops.sha1 .build/s3/kops/${VERSION}/darwin/amd64/kops.sha1
	aws s3 sync --acl public-read .build/s3/ ${S3_BUCKET}

push: nodeup-dist
	scp -C .build/dist/nodeup  ${TARGET}:/tmp/

push-gce-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/cache/kubernetes-install/nodeup --conf=metadata://gce/config --dryrun --v=8

push-aws-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/cache/kubernetes-install/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --dryrun --v=8

push-gce-run: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/cache/kubernetes-install/nodeup --conf=metadata://gce/config --v=8

push-aws-run: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /tmp/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8



protokube-gocode:
	go install k8s.io/kops/protokube/cmd/protokube

protokube-builder-image:
	docker build -t protokube-builder images/protokube-builder

protokube-build-in-docker: protokube-builder-image
	docker run -it -e VERSION=${VERSION} -v `pwd`:/src protokube-builder /onbuild.sh

protokube-image: protokube-build-in-docker
	docker build -t ${DOCKER_REGISTRY}/protokube:${TAG} -f images/protokube/Dockerfile .

protokube-push: protokube-image
	docker push ${DOCKER_REGISTRY}/protokube:${TAG}



nodeup: nodeup-dist

nodeup-gocode: gobindata
	go install -ldflags "-X main.BuildVersion=${VERSION}" k8s.io/kops/cmd/nodeup

nodeup-builder-image:
	docker build -t nodeup-builder images/nodeup-builder

nodeup-build-in-docker: nodeup-builder-image
	docker run -it -e VERSION=${VERSION} -v `pwd`:/src nodeup-builder /onbuild.sh

nodeup-dist: nodeup-build-in-docker
	mkdir -p .build/dist
	cp .build/artifacts/nodeup .build/dist/
	(sha1sum .build/dist/nodeup | cut -d' ' -f1) > .build/dist/nodeup.sha1



dns-controller-gocode:
	go install k8s.io/kops/dns-controller/cmd/dns-controller

dns-controller-builder-image:
	docker build -t dns-controller-builder images/dns-controller-builder

dns-controller-build-in-docker: dns-controller-builder-image
	docker run -it -e VERSION=${VERSION} -v `pwd`:/src dns-controller-builder /onbuild.sh

dns-controller-image: dns-controller-build-in-docker
	docker build -t ${DOCKER_REGISTRY}/dns-controller:${TAG}  -f images/dns-controller/Dockerfile .

dns-controller-push: dns-controller-image
	docker push ${DOCKER_REGISTRY}/dns-controller:${TAG}



copydeps:
	rsync -avz _vendor/ vendor/ --exclude vendor/  --exclude .git
