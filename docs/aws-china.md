# How to use kops in AWS China Region

## Getting Started

Kops used to only support Google Cloud DNS and Amazon Route53 to provision a kubernetes cluster. But since 1.6.2 `gossip` has been added which make it possible to provision a cluster without one of those DNS providers. Thanks to `gossip`, it's officially supported to provision a fully-functional kubernetes cluster in AWS China Region [which doesn't have Route53 so far][1] since [1.7][2]. Should support both `cn-north-1` and `cn-northwest-1`, but only `cn-north-1` is tested.

Most of the following procedures to provision a cluster are the same with [the guide to use kops in AWS](getting_started/aws.md). The differences will be highlighted and the similar parts will be omitted.

*NOTE: THE FOLLOWING PROCEDURES ARE ONLY TESTED WITH KOPS 1.10.0, 1.10.1 AND KUBERNETES 1.9.11, 1.10.12*

### [Install kops](getting_started/aws.md#install-kops)

### [Install kubectl](getting_started/aws.md#install-kubectl)

### [Setup your environment](getting_started/aws.md#setup-your-environment)

#### AWS

When `aws configure`, remember to set the `default region name` to the correct one, e.g. `cn-north-1`.

```console
AWS Access Key ID [None]:
AWS Secret Access Key [None]:
Default region name [None]:
Default output format [None]:
```

And export it correctly.

```console
export AWS_REGION=$(aws configure get region)
```

## [Configure DNS](getting_started/aws.md#configure-dns)

As the note kindly pointing out, a gossip-based cluster can be easily created by having the cluster name end with `.k8s.local`. We will adopt this trick below. Rest of this section can be skipped safely.

## [Testing your DNS setup](getting_started/aws.md#testing-your-dns-setup)

Thanks to `gossip`, this section can be skipped safely as well.

## [Cluster State storage](getting_started/aws.md#cluster-state-storage)

Since we are provisioning a cluster in AWS China Region, we need to create a dedicated S3 bucket in AWS China Region.

```console
aws s3api create-bucket --bucket prefix-example-com-state-store --create-bucket-configuration LocationConstraint=$AWS_REGION
```

## [Creating your first cluster](getting_started/aws.md#creating-your-first-cluster)

### Ensure you have a VPC which can access the internet NORMALLY

First of all, we have to solve the slow and unstable connection to the internet outside China, or the following processes won't work. One way to do that is to build a NAT instance which can route the traffic via some reliable connection. The details won't be discussed here.

### Prepare kops ami

We have to build our own AMI because there is [no official kops ami in AWS China Regions][3]. There're two ways to accomplish so.

#### ImageBuilder **RECOMMENDED**

First, launch an instance in a private subnet which accesses the internet fast and stably.

Because the instance launched in a private subnet, we need to ensure it can be connected by using the private ip via a VPN or a bastion.

```console
SUBNET_ID=<subnet id> # a private subnet
SECURITY_GROUP_ID=<security group id>
KEY_NAME=<key pair name on aws>

AMI_ID=$(aws ec2 describe-images --filters Name=name,Values=debian-jessie-amd64-hvm-2016-02-20-ebs --query 'Images[*].ImageId' --output text)
INSTANCE_ID=$(aws ec2 run-instances --image-id $AMI_ID --instance-type m3.medium --key-name $KEY_NAME --security-group-ids $SECURITY_GROUP_ID --subnet-id $SUBNET_ID --no-associate-public-ip-address --query 'Instances[*].InstanceId' --output text)
aws ec2 create-tags --resources ${INSTANCE_ID} --tags Key=k8s.io/role/imagebuilder,Value=1
```

Now follow the documentation of [ImageBuilder][4] in `kube-deploy` to build the image.

```console
go get k8s.io/kube-deploy/imagebuilder
cd ${GOPATH}/src/k8s.io/kube-deploy/imagebuilder

sed -i '' "s|publicIP := aws.StringValue(instance.PublicIpAddress)|publicIP := aws.StringValue(instance.PrivateIpAddress)|" pkg/imagebuilder/aws.go
make

# cloud-init is failing due to urllib3 dependency. https://github.com/aws/aws-cli/issues/3678
sed -i '' "s/'awscli'/'awscli==1.16.38'/g" templates/1.9-jessie.yml

# If the keypair specified is not `$HOME/.ssh/id_rsa`, the config yaml file need to be modified to add the full path to the private key.
echo 'SSHPrivateKey: "/absolute/path/to/the/private/key"' >> aws-1.9-jessie.yaml

${GOPATH}/bin/imagebuilder --config aws-1.9-jessie.yaml --v=8 --publish=false --replicate=false --up=false --down=false
```

#### Copy AMI from another region

Following [the comment][5] to copy the kops image from another region, e.g. `ap-southeast-1`.

#### Get the AMI id

No matter how to build the AMI, we get an AMI finally, e.g. `k8s-1.9-debian-jessie-amd64-hvm-ebs-2018-07-18`.

### [Prepare local environment](getting_started/aws.md#prepare-local-environment)

Set up a few environment variables.

```console
export NAME=example.k8s.local
export KOPS_STATE_STORE=s3://prefix-example-com-state-store
```

### [Create cluster configuration](getting_started/aws.md#create-cluster-configuration)

We will need to note which availability zones are available to us. AWS China (Beijing) Region only has two availability zones. It will have [the same problem][6], like other regions having less than three AZs, that there is no true HA support in two AZs. You can [add more master nodes](#add-more-master-nodes) to improve the reliability in one AZ.

```console
aws ec2 describe-availability-zones
```

Below is a `create cluster` command which will create a complete internal cluster [in an existing VPC](run_in_existing_vpc.md). The below command will generate a cluster configuration, but not start building it. Make sure that you have generated SSH key pair before creating the cluster.

```console
VPC_ID=<vpc id>
VPC_NETWORK_CIDR=<vpc network cidr> # e.g. 172.30.0.0/16
AMI=<owner id/ami name> # e.g. 123456890/k8s-1.9-debian-jessie-amd64-hvm-ebs-2018-07-18

kops create cluster \
    --zones ${AWS_REGION}a \
    --vpc ${VPC_ID} \
    --network-cidr ${VPC_NETWORK_CIDR} \
    --image ${AMI} \
    --associate-public-ip=false \
    --api-loadbalancer-type internal \
    --topology private \
    --networking calico \
    ${NAME}
```

### [Customize Cluster Configuration](getting_started/aws.md#prepare-local-environment)

Now we have a cluster configuration, we adjust the subnet config to reuse [shared subnets](run_in_existing_vpc.md#shared-subnets) by editing the description.

```console
kops edit cluster $NAME
```

Then change the corresponding subnets to specify the `id` and remove the `cidr`, e.g.

```yaml
spec:
  subnets:
  - id: subnet-12345678
    name: cn-north-1a
    type: Private
    zone: cn-north-1a
  - id: subnet-87654321
    name: utility-cn-north-1a
    type: Utility
    zone: cn-north-1a
```

Another tweak we can adopt here is to add a `docker` section to change the mirror to [the official registry mirror for China][7]. This will increase stability and download speed of pulling images from docker hub.

```yaml
spec:
  docker:
    registryMirrors:
    - https://registry.docker-cn.com
```

Please note that this mirror *MIGHT BE* not suitable for some cases. It's can be replaced by any other registry mirror as long as it's compatible with the docker api.

### [Build the Cluster](getting_started/aws.md#build-the-cluster)

### [Use the Cluster](getting_started/aws.md#use-the-cluster)

### [Delete the Cluster](getting_started/aws.md#delete-the-cluster)

## [What's next?](getting_started/aws.md#whats-next)

### Add more master nodes

#### In one AZ

To achieve this, we can add more parameters to `kops create cluster`.

```console
  --master-zones ${AWS_REGION}a --master-count 3 \
  --zones ${AWS_REGION}a --node-count 2 \
```

#### In two AZs

```console
  --master-zones ${AWS_REGION}a,${AWS_REGION}b --master-count 3 \
  --zones ${AWS_REGION}a,${AWS_REGION}b --node-count 2 \
```

**Please note that this will still have 50% chance to break the cluster when one of the AZs are down.**

### Offline mode

Here is a naive, uncompleted attempt to provision a cluster in a way minimizing the requirements to the internet because even with some kind of proxies or VPN it's still not that fast and it's always much more expensive than downloading from S3.

```console
## Setup vars

KUBERNETES_VERSION=$(curl -fsSL --retry 5 "https://dl.k8s.io/release/stable.txt")
KOPS_VERSION=$(curl -fsSL --retry 5 "https://api.github.com/repos/kubernetes/kops/releases/latest" | grep 'tag_name' | cut -d\" -f4)
ASSET_BUCKET="some-asset-bucket"
ASSET_PREFIX=""

# Please note that this filename of cni asset may change with kubernetes version
# Find this in https://github.com/kubernetes/kops/blob/master/upup/pkg/fi/cloudup/networking.go
CNI_FILENAME=cni-plugins-amd64-v0.6.0.tgz


export KOPS_BASE_URL=https://s3.cn-north-1.amazonaws.com.cn/$ASSET_BUCKET/kops/$KOPS_VERSION/
export CNI_VERSION_URL=https://s3.cn-north-1.amazonaws.com.cn/$ASSET_BUCKET/kubernetes/network-plugins/$CNI_FILENAME
export CNI_ASSET_HASH_STRING=d595d3ded6499a64e8dac02466e2f5f2ce257c9f

## Download assets

KUBERNETES_ASSETS=(
  network-plugins/$CNI_FILENAME
  release/$KUBERNETES_VERSION/bin/linux/amd64/kube-apiserver.tar
  release/$KUBERNETES_VERSION/bin/linux/amd64/kube-controller-manager.tar
  release/$KUBERNETES_VERSION/bin/linux/amd64/kube-proxy.tar
  release/$KUBERNETES_VERSION/bin/linux/amd64/kube-scheduler.tar
  release/$KUBERNETES_VERSION/bin/linux/amd64/kubectl
  release/$KUBERNETES_VERSION/bin/linux/amd64/kubelet
)
for asset in "${KUBERNETES_ASSETS[@]}"; do
  dir="kubernetes/$(dirname "$asset")"
  mkdir -p "$dir"
  url="https://storage.googleapis.com/kubernetes-release/$asset"
  wget -P "$dir" "$url"
  [ "${asset##*.}" != "gz" ] && wget -P "$dir" "$url.sha1"
  [ "${asset##*.}" == "tar" ] && wget -P "$dir" "${url%.tar}.docker_tag"
done

KOPS_ASSETS=(
  "images/protokube.tar.gz"
  "linux/amd64/nodeup"
  "linux/amd64/utils.tar.gz"
)
for asset in "${KOPS_ASSETS[@]}"; do
  kops_path="kops/$KOPS_VERSION/$asset"
  dir="$(dirname "$kops_path")"
  mkdir -p "$dir"
  url="https://kubeupv2.s3.amazonaws.com/kops/$KOPS_VERSION/$asset"
  wget -P "$dir" "$url"
  wget -P "$dir" "$url.sha1"
done

## Upload assets

## Get default S3 multipart_threshold

AWS_S3_DEFAULT_MULTIPART_THRESHOLD=$(aws configure get default.s3.multipart_threshold)

if [ ! -n "$AWS_S3_DEFAULT_MULTIPART_THRESHOLD" ]; then
  AWS_S3_DEFAULT_MULTIPART_THRESHOLD=8MB
fi

## Set multipart_threshold to 1024MB to prevent Etag not returns MD5 when upload multipart

aws configure set default.s3.multipart_threshold 1024MB

aws s3api create-bucket --bucket $ASSET_BUCKET --create-bucket-configuration LocationConstraint=$AWS_REGION
for dir in "kubernetes" "kops"; do
  aws s3 sync --acl public-read "$dir" "s3://$ASSET_BUCKET/$ASSET_PREFIX$dir"
done

aws configure set default.s3.multipart_threshold $AWS_S3_DEFAULT_MULTIPART_THRESHOLD

```

When create the cluster, add these parameters to the command line.

```console
  --kubernetes-version https://s3.cn-north-1.amazonaws.com.cn/$ASSET_BUCKET/kubernetes/release/$KUBERNETES_VERSION
```

Now most of the assets required to provision a cluster by `kops` and `kubernetes` will be downloaded from the specified S3 bucket except images like `pause-amd64`, `dns` related, etc. These images not mirrored by the docker hub mirror since they are hosted on `gcr.io`. There will be some [problem][8] if the connection is not good.

### Assets API

It hasn't been tested as this approach was only a PR when the author experimenting with provisioning a cluster in AWS China Region. It's the official way to achieve offline mode and should be superior to the previous naive attempt.


[1]: http://docs.amazonaws.cn/en_us/aws/latest/userguide/unsupported.html
[2]: https://github.com/kubernetes/kops/blob/master/docs/releases/1.7-NOTES.md
[3]: https://github.com/kubernetes/kops/issues/3282
[4]: https://github.com/kubernetes/kube-deploy/tree/master/imagebuilder
[5]: https://github.com/kubernetes-incubator/kube-aws/pull/390#issue-212435055
[6]: https://github.com/kubernetes/kops/issues/3088
[7]: https://docs.docker.com/registry/recipes/mirror/#use-case-the-china-registry-mirror
[8]: https://github.com/kubernetes/kops/issues/3236
