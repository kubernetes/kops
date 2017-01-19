## Create a simple cluster

We'll create a simple cluster

Make sure you have run:
`export KOPS_STATE_STORE=s3://clusters.example.com`

Then to create the configuration for your cluster:
`kops create cluster simple.example.com --zones us-east-1c`

That creates the configuration, but doesn't actually apply it to AWS.  But we can look at that later.

To preview the changes that will be made:
`kops update cluster simple.example.com`

And then when you're ready
`kops update cluster simple.example.com --yes`

That will create your cluster, and it should be ready in a few minutes (~5 minutes).

A few things to note while you're waiting:

* This creates a simple cluster, so we didn't specify many flags to `kops create cluster`.
* We specified the name of our cluster `simple.example.com`, which means we will set up DNS records for the API
  at `api.simple.example.com` along with some other internal names.  The hosted zone you created must be a suffix
  of your cluster name, so that these names will resolve.
* We specified a single `--zones`.  We will default to a single master running in that zone, and 2 nodes also
  running in that zone.
* We automatically exported a kubecfg so your machine can access the Kubernetes API.
* We didn't have to specify the `--cloud`, because kops knows that `us-east-1c` is on AWS.

Within a few minutes, your cluster should be ready, and you should be able to:

`kubectl get nodes --show-labels`

And you should see 3 nodes running in us-east-1c.

Next step: [Deleting the cluster](deleting_simple_cluster.md)