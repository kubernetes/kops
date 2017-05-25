## Deleting our simple cluster

Previously we created a simple cluster.  Now we'll shut it down, so that we don't
continue to be charged for it.

Make sure you have run:
`export KOPS_STATE_STORE=s3://clusters.example.com`

You can list your clusters using:

```
> kops get cluster
NAME                    CLOUD           ZONES
simple.example.com      aws             us-east-1c
```

Then to preview what will be deleted when you delete the cluster:
`kops delete cluster simple.example.com`

It's always a good idea to double-check what is going to be deleted, particularly in production AWS accounts.

Finally, when you have validated what is going to be deleted:
`kops delete cluster simple.example.com --yes`

AWS resource deletion is eventually consistent in a few places, so you will likely see a few retry loops happening,
but it should delete everything within a few minutes.