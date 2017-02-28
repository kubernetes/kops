# Grafana Image For Heapster/InfluxDB

## What's in it:
 - Grafana 4
 - A Go binary that:
   - creates a datasource for InfluxDB
   - creates a couple of dashboards during startup.
     - these dashboards leverage templating and repeating of panels features in Grafana, to discover nodes, pods, and containers automatically.

## How to use it:
 - InfluxDB service URL can be passed in via the environment variable `INFLUXDB_SERVICE_URL`.
 - Otherwise, it will fall back to http://monitoring-influxdb:8086.

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
