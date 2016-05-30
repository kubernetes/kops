gocode: godeps
	go install k8s.io/kube-deploy/protokube/cmd/...

godeps:
	# I think strip-vendor is the workaround for 25572
	glide install --strip-vendor --strip-vcs

tar: gocode
	rm -rf .build/tar
	mkdir -p .build/tar/protokube/root
	cp ${GOPATH}/bin/protokube .build/tar/protokube/root
	tar czvf .build/protokube.tar.gz -C .build/tar/ .
	tar tvf .build/protokube.tar.gz
	(sha1sum .build/protokube.tar.gz | cut -d' ' -f1) > .build/protokube.tar.gz.sha1

upload: tar
	rm -rf .build/s3
	mkdir -p .build/s3/protokube
	cp .build/protokube.tar.gz .build/s3/protokube/
	cp .build/protokube.tar.gz.sha1 .build/s3/protokube/
	aws s3 sync .build/s3/ s3://kubeupv2/
	aws s3api put-object-acl --bucket kubeupv2 --key protokube/protokube.tar.gz --acl public-read
	aws s3api put-object-acl --bucket kubeupv2 --key protokube/protokube.tar.gz.sha1 --acl public-read
