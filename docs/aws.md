# Bringing up a cluster on AWS

* Ensure you have kubectl installed and on your path.  (We need it to set kubecfg configuration.)

* Set up a DNS hosted zone in Route 53, e.g. `mydomain.com`, and set up the DNS nameservers as normal
  so that domains will resolve.  You can reuse an existing domain name (e.g. `mydomain.com`), or you can create
  a "child" hosted zone (e.g. `myclusters.mydomain.com`) if you want to isolate them.  Note that with AWS Route53,
  you can have subdomains in a single hosted zone, so you can have `cluster1.testclusters.mydomain.com` under
  `mydomain.com`.

* Pick a DNS name under this zone to be the name of your cluster.  kops will set up DNS so your cluster
  can be reached on this name.  For example, if your zone was `mydomain.com`, a good name would be
  `kubernetes.mydomain.com`, or `dev.k8s.mydomain.com`, or even `dev.k8s.myproject.mydomain.com`. We'll call this `NAME`.

* Set `AWS_PROFILE` (if you need to select a profile for the AWS CLI to work)

* Pick an S3 bucket that you'll use to store your cluster configuration - this is called your state store.  You
  can `export KOPS_STATE_STORE=s3://<mystatestorebucket>` and then kops will use this location by default.  We
  suggest putting this in your bash profile or similar.  A single registry can hold multiple clusters, and it
  can also be shared amongst your ops team (which is much easier than passing around kubecfg files!)

* Run "kops create cluster" to create your cluster configuration:
```
${GOPATH}/bin/kops create cluster --cloud=aws --zones=us-east-1c ${NAME}
```
(protip: the --cloud=aws argument is optional if the cloud can be inferred from the zones)

* Run "kops update cluster" to build your cluster:
```
${GOPATH}/bin/kops update cluster ${NAME} --yes
```

If you have problems, please set `--v=8` and open an issue, and ping justinsb on slack!
