all:
	go build
	cd machine-controller && go build

controller:
	cd machine-controller && go build && make push
	go build

fmt:
	find -path './vendor' -prune -o -name '*.go' -exec gofmt -w -l {} \;

.PHONY: all fmt
