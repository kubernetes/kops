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

## Copying assets into repositories

{{ kops_feature_table(kops_added_default='1.22') }}

You can copy assets into their repositories either by running `kops get assets --copy` or through an external process.

When running `kops get assets --copy`, kOps copies assets into their respective repositories if
they do not already exist there.

For file assets, kOps only supports copying to a repository that is either an S3 or GCS bucket.
An S3 bucket must be configured using the [regional naming conventions of S3](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_region).
A GCS bucket must be configured with a prefix of `https://storage.googleapis.com/`.

## Listing assets

{{ kops_feature_table(kops_added_default='1.22') }}

You can obtain a list of image and file assets used by a particular cluster by running `kops get assets`. You can get output in table, YAML, or JSON format.
You can feed this into a process, external to kOps, for copying the assets to their respective repositories.
