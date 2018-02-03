PROJECT=k8s-cluster-api
NAME=machine-controller
VERSION=0.19

staticbuild:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .

image: staticbuild
	docker build -t "gcr.io/$(PROJECT)/$(NAME):$(VERSION)" .

push: image
	gcloud docker -- push "gcr.io/$(PROJECT)/$(NAME):$(VERSION)"

fix-image-permissions:
	gsutil acl ch -u AllUsers:READ gs://artifacts.$(PROJECT).appspot.com
	gsutil -m acl ch -r -u AllUsers:READ gs://artifacts.$(PROJECT).appspot.com

.PHONY: image push staticbuild fix-image-permissions
