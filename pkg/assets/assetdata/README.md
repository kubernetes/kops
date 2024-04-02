This directory contains hashes for well-known dependencies of kubernetes / a kOps cluster.

We store the hashes here, rather than downloading them every time - it is a little more secure,
and it is more efficient.

The yaml file structure is intended to mirror the asset structure used by the k8s project,
e.g. by [kpromo](https://github.com/kubernetes-sigs/promo-tools/blob/main/docs/file-promotion.md).
However, this should be treated as an implementation detail.

Currently many hash files are manually curated.  Some of them can be automatically generated,
and we have scripts named `generate-<foo>.sh` to generate them.