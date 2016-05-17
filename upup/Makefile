gocode: godeps
	go install k8s.io/kube-deploy/upup/cmd/...

godeps:
	# I think strip-vendor is the workaround for 25572
	glide install --strip-vendor --strip-vcs

tar: gocode
	rm -rf .build/tar
	mkdir -p .build/tar/nodeup/root
	cp ${GOPATH}/bin/nodeup .build/tar/nodeup/root
	cp -r models/nodeup/ .build/tar/nodeup/root/model/
	tar czvf .build/nodeup.tar.gz -C .build/tar/ .
	tar tvf .build/nodeup.tar.gz
	(sha1sum .build/nodeup.tar.gz | cut -d' ' -f1) > .build/nodeup.tar.gz.sha1 

upload: tar
	rm -rf .build/s3
	mkdir -p .build/s3/nodeup
	cp .build/nodeup.tar.gz .build/s3/nodeup/
	cp .build/nodeup.tar.gz.sha1 .build/s3/nodeup/
	aws s3 sync .build/s3/ s3://kubeupv2/
	aws s3api put-object-acl --bucket kubeupv2 --key nodeup/nodeup.tar.gz --acl public-read
	aws s3api put-object-acl --bucket kubeupv2 --key nodeup/nodeup.tar.gz.sha1 --acl public-read

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
