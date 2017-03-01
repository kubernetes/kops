# Release Notes for the Heapster InfluxDB container.

## v1.1.1 (11.1.2016)
- Updated to version v1.1.1; bumped Godeps and modified some code in heapster to use the latest schema

## v0.13.0 (4.1.2016)
- Formalized the image name for every arch to `gcr.io/google_containers/influxdb-grafana-ARCH:VERSION`
- Now this image is released for multiple architectures, including amd64, arm, arm64, ppc64le and s390x
- The `gcr.io/google_containers/heapster-influxdb:VERSION` image is a manifest list, which means docker will pull the right image for the right arch automatically
- InfluxDB v0.13.0
- Added Makefile and README.md

## 0.7 (06-27-2016)
- Updated to v0.12.2-1

## 0.6 (12-10-2015)
- Updated to v0.9.6

## 0.5 (9-18-2015)
- Updated to v0.9.4

## 0.4 (9-17-2015)
- Updated to InfluxDB v0.8.9 to pave way for safely upgrading to influxDB v0.9.x

## 0.3 (1-19-2015)
- Updated Influxdb version number to 0.8.8, paving way for collectd support.
