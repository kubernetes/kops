all: image

.PHONY: image push

image:
	docker build -t kopeio/nvidia-bootstrap:1.6 -f image/Dockerfile image/

push: image
	docker push kopeio/nvidia-bootstrap:1.6
