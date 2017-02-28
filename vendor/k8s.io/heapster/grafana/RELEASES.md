# Release Notes for Grafana container.

## 4.0.2 (4.1.2016)
- Formalized the image name for every arch to `gcr.io/google_containers/heapster-grafana-ARCH:VERSION`
- Now this image is released for multiple architectures, including amd64, arm, arm64, ppc64le and s390x
- The `gcr.io/google_containers/heapster-grafana:VERSION` image is a manifest list, which means docker will pull the right image for the right arch automatically
- Grafana v4.0.2
- Enhanced the Makefile and the README

## 3.1.1 (24-11-2016) 
- Support Grafana 3.1.1.

## 2.6.0-2 (29-02-2016)
- Handle new Influxdb format
- Updated dashboards

## 2.6.0 (29-12-2015)
- Support Grafana 2.6.0.
- Improve default dashboards
  - Accurate CPU metrics
  - Cluster network graphs
  - Fix data aggregation

## 2.5.0 (12-11-2015)
- Support Grafana 2.5.0.

## 2.1.0 (9-28-2015)
- Support Grafana 2.1.0.
- Auto populate pods and nodes using Grafana templates.
