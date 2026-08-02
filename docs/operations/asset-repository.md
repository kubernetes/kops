# Using local asset repositories

You can configure kOps to provision a cluster to download assets (images and files) from local repositories.
This is useful when downloading assets from the internet is undersirable, for example:

* To deploy where the network is offline or internet-restricted.
* To avoid rate limits or network transfer costs.
* To limit exposure to watering-hole attacks.
* To comply with other security requirements, such as the need to scan for vulnerabilities.

There can be one repository for images and another for files.

## Configuring

### Configuring a local image repository

To configure a local image repository, set either `assets.containerRegistry` or `assets.containerProxy` in the cluster spec.
They both do essentially the same thing, but `containerRegistry` avoids using `/` characters in the local image names.

```yaml
spec:
  assets:
    containerRegistry: example.com/registry
```

or

```yaml
spec:
  assets:
    containerProxy: example.com/proxy
```

### Configuring a local file repository

To configure a local file repository, set `assets.fileRepository` in the cluster spec.

```yaml
spec:
  assets:
    fileRepository: https://example.com/files
```

For an `http://` or `https://` repository, nodes must be able to read without credentials.
The repository can be public or allow access through network connectivity, such as a
particular cloud endpoint.

On GCE, the repository can also be a `gs://` URL. Nodes then read it with the credentials of
their service account, which allows the bucket to be private. The service accounts of the
instance groups have to be granted `roles/storage.objectViewer` on that bucket.

```yaml
spec:
  assets:
    fileRepository: gs://example-bucket/files
```

{{ kops_feature_table(kops_added_default='1.37') }}

On AWS, the repository can also be an `s3://` URL. Nodes then read it with the credentials of
their instance profile, which allows the bucket to be private. The instance profiles of the
instance groups have to be granted `s3:GetObject` on that bucket. This is supported in the
commercial AWS and AWS GovCloud partitions. Downloading nodeup itself from an `s3://`
`KOPS_BASE_URL` additionally requires curl 8.0 or newer on the node image.

```yaml
spec:
  assets:
    fileRepository: s3://example-bucket/files
```

## Copying assets into repositories

{{ kops_feature_table(kops_added_default='1.22') }}

You can copy assets into their repositories either by running `kops get assets --copy` or through an external process.

When running `kops get assets --copy`, kOps copies assets into their respective repositories if
they do not already exist there.

For file assets, kOps only supports copying to a repository that is either an S3 or GCS bucket.
An S3 bucket must be configured with a prefix of `s3://` or using the [regional naming conventions of S3](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region).
A GCS bucket must be configured with a prefix of `https://storage.googleapis.com/` or `gs://`.

## Listing assets

{{ kops_feature_table(kops_added_default='1.22') }}

You can obtain a list of image and file assets used by a particular cluster by running `kops get assets`. You can get output in table, YAML, or JSON format.
You can feed this into a process, external to kOps, for copying the assets to their respective repositories.
