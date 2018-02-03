all:
	cd tools/upgrader && go build
	cd tools/repair && go build
	cd tools/machineset && go build
	cd tools/machines-client && go build
	cd examples/machines-controller && go build
	cd examples/machines-crd-installer && go build

fmt:
	find -path './vendor' -prune -o -name '*.go' -exec gofmt -w -l {} \;

.PHONY: all fmt
