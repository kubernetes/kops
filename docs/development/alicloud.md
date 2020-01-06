# Development process for Alicloud

This document contains details about ongoing effort for Alicloud support in
kops. Alicloud support in kops is an experimental feature, under 
`KOPS_FEATURE_FLAGS=AlphaAllowALI` feature flag and is not production ready yet.

# Current status

Please refer to this
[issue](https://github.com/kubernetes/kops/issues/4127#issuecomment-534536277)
for the to-do list.

In order to get Alicloud support out of alpha. At least, these two
PRs([#7849|https://github.com/kubernetes/kops/pull/7849] and
[#8016|https://github.com/kubernetes/kops/pull/8016]) need to be merged.

NOTE: The following instructions don't work for `master` branch. If you start
developing with Alicloud, you will need to cherry-pick these two PRs onto your own develop branch first after you clone
`master` branch.

# Mirror docker images to Alicloud container registry

The required images are listed in `hack/alicloud/required-images.txt`. Before
you run `./hack/alicloud/mirror.sh`*, you need to:

1. Install `docker` on your laptop
2. Create a namespace in Alicloud container registry(eg: `kops-mirror`) in the
   web console.
2. run dev-build-alicloud.sh


You can use the example command as below to quickly starting developing nodeup
and kops.

```sh
export KOPS_VERSION=1.15.0-alpha.1
export CLUSTER_NAME=dev-1.k8s.local
export KOPS_STATE_STORE=oss://kops-state-bucket
export NODEUP_BUCKET=k8s-assets-bucket
export IMAGE=m-xxxxxxxxxx
export ALICLOUD_REGION=cn-shanghai
export ALIYUN_ACCESS_KEY_ID=xxxxxx
export ALIYUN_ACCESS_KEY_SECRET=xxxxxxxxxxxxxxx
export OSS_REGION=oss-cn-shanghai
export KOPS_FEATURE_FLAGS="AlphaAllowALI"
export NODEUP_URL=https://${NODEUP_BUCKET}.${OSS_REGION}.aliyuncs.com/kops/${KOPS_VERSION}/linux/amd64/nodeup
export KOPS_BASE_URL=https://${NODEUP_BUCKET}.${OSS_REGION}.aliyuncs.com/kops/${KOPS_VERSION}/
export KOPS_CREATE=no

# cd [kops_dir]
./hack/alicloud/dev-build.sh
```

# Ref

- The script `hack/alicloud/mirror.sh` is partially copied from
  https://github.com/nwcdlabs/kops-cn/blob/master/mirror/mirror-images.sh,
  thanks to [Pahud Hsieh](https://github.com/pahud).
