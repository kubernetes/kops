gocode: godeps
	go install k8s.io/kube-deploy/protokube/cmd/protokube

godeps:
	# I think strip-vendor is the workaround for 25572
	glide install --strip-vendor --strip-vcs

gofmt:
	gofmt -w -s cmd/
	gofmt -w -s pkg/

builder-image:
	docker build -f images/builder/Dockerfile -t builder .

build-in-docker: builder-image
	docker run -it -v `pwd`:/src builder /onbuild.sh

image: build-in-docker
	docker build -t kope/protokube:1.3  -f images/protokube/Dockerfile .

push: image
	docker push kope/protokube:1.3
