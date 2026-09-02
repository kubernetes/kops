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

#### HTTP or HTTPS

```yaml
spec:
  assets:
    fileRepository: https://example.com/files
```

For an `http://` or `https://` repository, nodes must be able to read without credentials.
The repository can be public or allow access through network connectivity, such as a
particular cloud endpoint.

#### Google Cloud Storage

{{ kops_feature_table(kops_added_default='1.37') }}

On GCE, the repository can also be a `gs://` URL. Nodes then read it with the credentials of
their service account, which allows the bucket to be private. The service accounts of the
instance groups have to be granted `roles/storage.objectViewer` on that bucket.

```yaml
spec:
  assets:
    fileRepository: gs://example-bucket/files
```

#### Amazon S3

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

#### Azure Blob Storage

{{ kops_feature_table(kops_added_default='1.37') }}

On Azure, the repository can also be an `azureblob://<account>/<container>/<prefix>` URL.
Nodes then read it with their system-assigned managed identity, which allows the storage
account to be private. The managed identities of the instance groups have to be granted
`Storage Blob Data Reader` on the assets container. Do not grant access to the state-store
storage account, as that would let nodes read the cluster PKI.

```yaml
spec:
  assets:
    fileRepository: azureblob://exampleaccount/assets/files
```

#### OCI registry

{{ kops_feature_table(kops_added_default='1.37') }}

An OCI file repository URL contains a registry and an optional repository prefix:

```yaml
spec:
  assets:
    fileRepository: oci://registry.example.com/optional-prefix
```

kOps stores each asset family at
`<registry>/<optional-prefix>/<asset-family>:<tag>`. For example, containerd 2.2.4 for `amd64` is
stored as `registry.example.com/optional-prefix/containerd:v2.2.4-amd64`.

Tags contain the asset version and, for architecture-specific files, the architecture. They do not
include `linux`. If no meaningful version is available, kOps derives a deterministic tag from the
file SHA-256. The OCI blob contains the exact source file, and its SHA-256 is also the integrity
value passed to nodes.

Nodes download blobs by digest without reading tags or manifests, so the repositories must allow
anonymous pulls, including the standard OCI Distribution Bearer token flow. Staging uses the
operator's local container-registry credentials and requires HTTPS registry and token endpoints.
Stage assets before updating the cluster. Staging is a no-op if a tag already contains the expected
blobs and fails if the tag contains different content. These checks are best-effort, not atomic:
do not run multiple staging processes that publish different content for the same tags.

## Copying assets into repositories

{{ kops_feature_table(kops_added_default='1.22') }}

Run `kops get assets --copy` before updating the cluster, or stage the assets with an external
process. The command copies assets that are not already present and supports the S3, GCS, Azure
Blob Storage, and OCI repository forms described above. It also supports GCS URLs beginning with
`https://storage.googleapis.com/`; other HTTP or HTTPS repositories require external staging.

## Listing assets

{{ kops_feature_table(kops_added_default='1.22') }}

You can obtain a list of image and file assets used by a particular cluster by running `kops get assets`. You can get output in table, YAML, or JSON format.
You can feed this into a process, external to kOps, for copying the assets to their respective repositories.
