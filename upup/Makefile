gocode: godeps
	GO15VENDOREXPERIMENT=1 go install k8s.io/kube-deploy/upup/cmd/...
	ln -sfn ${GOPATH}/src/k8s.io/kube-deploy/upup/models/ ${GOPATH}/bin/models

codegen:
	GO15VENDOREXPERIMENT=1 go install k8s.io/kube-deploy/upup/tools/generators/...
	GO15VENDOREXPERIMENT=1 go generate k8s.io/kube-deploy/upup/pkg/fi/cloudup/awstasks
	GO15VENDOREXPERIMENT=1 go generate k8s.io/kube-deploy/upup/pkg/fi/cloudup/gcetasks
	GO15VENDOREXPERIMENT=1 go generate k8s.io/kube-deploy/upup/pkg/fi/fitasks

test:
	GO15VENDOREXPERIMENT=1 go test k8s.io/kube-deploy/upup/cmd/...
	GO15VENDOREXPERIMENT=1 go test k8s.io/kube-deploy/upup/pkg/...

godeps:
	# I think strip-vendor is the workaround for 25572
	glide install --strip-vendor --strip-vcs

gofmt:
	gofmt -w -s cmd/
	gofmt -w -s pkg/

upup-tar: gocode
	rm -rf .build/upup/tar
	mkdir -p .build/upup/tar/upup/
	cp ${GOPATH}/bin/cloudup .build/upup/tar/upup/cloudup
	cp ${GOPATH}/bin/upup .build/upup/tar/upup/upup
	cp -r models/ .build/upup/tar/upup/models/
	tar czvf .build/upup.tar.gz -C .build/upup/tar/ .
	tar tvf .build/upup.tar.gz
	(sha1sum .build/upup.tar.gz | cut -d' ' -f1) > .build/upup.tar.gz.sha1

nodeup-tar: gocode
	rm -rf .build/nodeup/tar
	mkdir -p .build/nodeup/tar/nodeup/root
	cp ${GOPATH}/bin/nodeup .build/nodeup/tar/nodeup/root
	cp -r models/nodeup/ .build/nodeup/tar/nodeup/root/model/
	tar czvf .build/nodeup.tar.gz -C .build/nodeup/tar/ .
	tar tvf .build/nodeup.tar.gz
	(sha1sum .build/nodeup.tar.gz | cut -d' ' -f1) > .build/nodeup.tar.gz.sha1 

upload: nodeup-tar upup-tar
	rm -rf .build/s3
	mkdir -p .build/s3/nodeup
	cp .build/nodeup.tar.gz .build/s3/nodeup/
	cp .build/nodeup.tar.gz.sha1 .build/s3/nodeup/
	mkdir -p .build/s3/upup
	cp .build/upup.tar.gz .build/s3/upup/
	cp .build/upup.tar.gz.sha1 .build/s3/upup/
	aws s3 sync --acl public-read .build/s3/ s3://kubeupv2/
	#aws s3api put-object-acl --bucket kubeupv2 --key nodeup/nodeup.tar.gz --acl public-read
	#aws s3api put-object-acl --bucket kubeupv2 --key nodeup/nodeup.tar.gz.sha1 --acl public-read

push: tar
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
