# InfluxDB image for Heapster

This is a minimal influxdb docker image that plays well with heapster.

## How to build:

```console
$ ARCH=${ARCH} make build
```

## How to release:

This image supports multiple architecures, which means the Makefile cross-compiles and builds docker images for all architectures automatically when pushing.
If you are releasing a new version, please bump the `VERSION` value in the `Makefile` before building the images.

How to build and push all images:

```console
# Optional: Set PREFIX if you want to push to a temporary user or another registry for testing
# This command will build images and push for all architectures
$ make push
```
