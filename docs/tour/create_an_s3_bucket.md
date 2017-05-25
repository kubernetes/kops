## Create an S3 bucket

You need to create an S3 bucket that will be used to store configuration for your cluster.  Multiple
clusters can exist within the same S3 bucket.  We call the bucket the "kops state store", and so
you typically configure it by `export KOPS_STATE_STORE=s3://bucketname`.

You will create a private S3 bucket, but you can share the S3 bucket amongst your team.  Granting access to
the S3 bucket gives full control to all the clusters in it, so typically you share the bucket amongst the
cluster administrators.  And you create multiple buckets when you have some clusters that one group administers,
but another set of clusters administered by a separate group.  Smaller organization typically use
a single S3 bucket, although some peopel will isolate prod & development.

The name of the S3 bucket does not matter, but note that because S3 buckets are globally unique, you
need to choose something nobody else has chosen.   I like to use `clusters.<hosted-zone-name>`, i.e.
`clusters.example.com` or `clusters.k8s.example.com`.

`aws mb clusters.example.com`

then run `export KOPS_STATE_STORE=s3://clusters.example.com`


Next step: [Creating our first cluster](creating_simple_cluster.md)