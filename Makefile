all: kops

DOCKER_REGISTRY=gcr.io/must-override/
S3_BUCKET=s3://must-override/
GOPATH_1ST=$(shell echo ${GOPATH} | cut -d : -f 1)

ifndef VERSION
	VERSION := git-$(shell git rev-parse --short HEAD)
endif

crossbuild:
	GOOS=darwin GOARCH=amd64 go build -o .build/darwin/amd64/kops -ldflags "-X main.BuildVersion=${VERSION}" -v k8s.io/kops/cmd/kops/...
	GOOS=linux GOARCH=amd64 go build -o .build/linux/amd64/kops -ldflags "-X main.BuildVersion=${VERSION}" -v k8s.io/kops/cmd/kops/...
	#GOOS=windows GOARCH=amd64 go build -o .build/windows/amd64/kops -ldflags "-X main.BuildVersion=${VERSION}" -v k8s.io/kops/cmd/kops/...

kops:
	GO15VENDOREXPERIMENT=1 go install -ldflags "-X main.BuildVersion=${VERSION}" k8s.io/kops/cmd/kops/...
	ln -sfn ${GOPATH_1ST}/src/k8s.io/kops/upup/models/ ${GOPATH_1ST}/bin/models

# Build in a docker container with golang 1.X
# Used to test we have not broken 1.X
check-builds-in-go15:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.5 make -f /go/src/k8s.io/kops/Makefile gocode

check-builds-in-go16:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.6 make -f /go/src/k8s.io/kops/Makefile gocode

check-builds-in-go17:
	docker run -v ${GOPATH_1ST}/src/k8s.io/kops:/go/src/k8s.io/kops golang:1.7 make -f /go/src/k8s.io/kops/Makefile gocode

codegen:
	GO15VENDOREXPERIMENT=1 go install k8s.io/kops/upup/tools/generators/...
	GO15VENDOREXPERIMENT=1 go generate k8s.io/kops/upup/pkg/fi/cloudup/awstasks
	GO15VENDOREXPERIMENT=1 go generate k8s.io/kops/upup/pkg/fi/cloudup/gcetasks
	GO15VENDOREXPERIMENT=1 go generate k8s.io/kops/upup/pkg/fi/fitasks

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

kops-tar: kops
	rm -rf .build/kops/tar
	mkdir -p .build/kops/tar/kops/
	cp ${GOPATH_1ST}/bin/kops .build/kops/tar/kops/kops
	cp -r upup/models/ .build/kops/tar/kops/models/
	tar czvf .build/kops.tar.gz -C .build/kops/tar/ .
	tar tvf .build/kops.tar.gz
	(sha1sum .build/kops.tar.gz | cut -d' ' -f1) > .build/kops.tar.gz.sha1

upload: nodeup-tar kops-tar
	rm -rf .build/s3
	mkdir -p .build/s3/nodeup
	cp .build/nodeup.tar.gz .build/s3/nodeup/nodeup-1.3.tar.gz
	cp .build/nodeup.tar.gz.sha1 .build/s3/nodeup/nodeup-1.3.tar.gz.sha1
	mkdir -p .build/s3/kops
	cp .build/kops.tar.gz .build/s3/kops/kops-1.3.tar.gz
	cp .build/kops.tar.gz.sha1 .build/s3/kops/kops-1.3.tar.gz.sha1
	aws s3 sync --acl public-read .build/s3/ ${S3_BUCKET}

push: nodeup-tar
	scp .build/nodeup.tar.gz ${TARGET}:/tmp/
	ssh ${TARGET} sudo tar zxf /tmp/nodeup.tar.gz -C /var/cache/kubernetes-install

push-gce-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/cache/kubernetes-install/nodeup/root/nodeup --conf=metadata://gce/config --dryrun --v=8 --model=/var/cache/kubernetes-install/nodeup/root/model

push-aws-dry: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/cache/kubernetes-install/nodeup/root/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --dryrun --v=8 --model=/var/cache/kubernetes-install/nodeup/root/model

push-gce-run: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/cache/kubernetes-install/nodeup/root/nodeup --conf=metadata://gce/config --v=8 --model=/var/cache/kubernetes-install/nodeup/root/model

push-aws-run: push
	ssh ${TARGET} sudo SKIP_PACKAGE_UPDATE=1 /var/cache/kubernetes-install/nodeup/root/nodeup --conf=/var/cache/kubernetes-install/kube_env.yaml --v=8 --model=/var/cache/kubernetes-install/nodeup/root/model



protokube-gocode:
	go install k8s.io/kops/protokube/cmd/protokube

protokube-builder-image:
	docker build -f images/protokube-builder/Dockerfile -t protokube-builder .

protokube-build-in-docker: protokube-builder-image
	docker run -it -v `pwd`:/src protokube-builder /onbuild.sh

protokube-image: protokube-build-in-docker
	docker build -t ${DOCKER_REGISTRY}/protokube:1.3  -f images/protokube/Dockerfile .

protokube-push: protokube-image
	docker push ${DOCKER_REGISTRY}/protokube:1.3



nodeup: nodeup-tar

nodeup-gocode:
	go install -ldflags "-X main.BuildVersion=${VERSION}" k8s.io/kops/cmd/nodeup

nodeup-builder-image:
	docker build -f images/nodeup-builder/Dockerfile -t nodeup-builder .

nodeup-build-in-docker: nodeup-builder-image
	docker run -it -v `pwd`:/src nodeup-builder /onbuild.sh

nodeup-tar: nodeup-build-in-docker
	rm -rf .build/nodeup/tar
	mkdir -p .build/nodeup/tar/nodeup/root
	cp .build/artifacts/nodeup .build/nodeup/tar/nodeup/root
	cp -r upup/models/nodeup/ .build/nodeup/tar/nodeup/root/model/
	tar czvf .build/nodeup.tar.gz -C .build/nodeup/tar/ .
	tar tvf .build/nodeup.tar.gz
	(sha1sum .build/nodeup.tar.gz | cut -d' ' -f1) > .build/nodeup.tar.gz.sha1



dns-controller-gocode:
	go install k8s.io/kops/dns-controller/cmd/dns-controller

dns-controller-builder-image:
	docker build -f images/dns-controller-builder/Dockerfile -t dns-controller-builder .

dns-controller-build-in-docker: dns-controller-builder-image
	docker run -it -v `pwd`:/src dns-controller-builder /onbuild.sh

dns-controller-image: dns-controller-build-in-docker
	docker build -t ${DOCKER_REGISTRY}/dns-controller:1.3  -f images/dns-controller/Dockerfile .

dns-controller-push: dns-controller-image
	docker push ${DOCKER_REGISTRY}/dns-controller:1.3



copydeps:
	rsync -avz _vendor/ vendor/ --exclude vendor/  --exclude .git
