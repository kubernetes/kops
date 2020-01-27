Random scribblings useful for development...


## Developing nodeup

ssh ${HOST} sudo mkdir -p /opt/nodeup/state
ssh ${HOST} sudo chown -R ${USER} /opt/nodeup

go install k8s.io/kops/upup/... && rsync ~/k8s/bin/nodeup ${HOST}:/opt/nodeup/nodeup && rsync --delete -avz trees/ ${HOST}:/opt/nodeup/trees/ \
&& rsync state/node.yaml ${HOST}:/opt/nodeup/state/node.yaml \
&& ssh ${HOST} sudo /opt/nodeup/nodeup --v=2  --template=/opt/nodeup/trees/nodeup --state=/opt/nodeup/state --tags=kubernetes_pool,debian_family,gce,systemd


# Random misc

Extract the master node config from a terraform output

cat tf/k8s.tf.json | jq -r '.resource.google_compute_instance["kubernetes-master"].metadata.config' > state/node.yaml



TODOS
======

* Implement number-of-tags prioritization
* Allow files ending in .md to be ignored.  Useful for comments.
* Better dependency tracking on systemd services?
* Automatically use different file mode if starts with #! ?
* Support .static under files to allow for files ending in .template?
* How to inherit options
* Allow customization of ordering?  Maybe prefix based.
* Cache hashes in-process (along with timestamp?) so we don't hash the kubernetes binary bundle repeatedly
* Fix the fact that we hash assets twice
* Confirm that we drop support for init.d
* Can we just use JSON custom marshaling instead of all our reflection stuff (or at least lighten the load)

* Do we officially publish https://storage.googleapis.com/kubernetes-release/release/v1.2.2/kubernetes-server-linux-amd64.tar.gz (ie just the server tar.gz)?

* Need to start docker-healthcheck once

* Can we replace some or all of nodeup config with pkg/apis/componentconfig/types.go ?

