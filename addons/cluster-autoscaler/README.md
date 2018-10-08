# Cluster Autoscaler Addon

We strongly recommend using Cluster Autoscaler with the kubernetes version for which it was meant. Refer to the [Cluster Autoscaler documentation compatibility matrix]( https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/README.md#releases)

Note that you likely want to change `AWS_REGION` and `GROUP_NAME`, and probably `MIN_NODES` and `MAX_NODES`. Here is an example of how you may wish to do so:

```bash
CLOUD_PROVIDER=aws
IMAGE=k8s.gcr.io/cluster-autoscaler:v1.2.2
MIN_NODES=1
MAX_NODES=5
AWS_REGION=us-east-1
# For AWS GROUP_NAME should be the name of ASG as seen on AWS console
GROUP_NAME="nodes.k8s.example.com"
SSL_CERT_PATH="/etc/ssl/certs/ca-certificates.crt" # (/etc/ssl/certs for gce, /etc/ssl/certs/ca-bundle.crt for RHEL7.X)

addon=cluster-autoscaler.yml
wget -O ${addon} https://raw.githubusercontent.com/kubernetes/kops/master/addons/cluster-autoscaler/v1.8.0.yaml

sed -i -e "s@{{CLOUD_PROVIDER}}@${CLOUD_PROVIDER}@g" "${addon}"
sed -i -e "s@{{IMAGE}}@${IMAGE}@g" "${addon}"
sed -i -e "s@{{MIN_NODES}}@${MIN_NODES}@g" "${addon}"
sed -i -e "s@{{MAX_NODES}}@${MAX_NODES}@g" "${addon}"
sed -i -e "s@{{GROUP_NAME}}@${GROUP_NAME}@g" "${addon}"
sed -i -e "s@{{AWS_REGION}}@${AWS_REGION}@g" "${addon}"
sed -i -e "s@{{SSL_CERT_PATH}}@${SSL_CERT_PATH}@g" "${addon}"

kubectl apply -f ${addon}
```
An enhanced script which also adds the IAM policies is included here [cluster-autoscaler.sh](cluster-autoscaler.sh) 

Question: Which ASG group should be autoscaled?  
Answer: By default, kops creates a "nodes" instancegroup and a corresponding ASG group which will have a name such as "nodes.$CLUSTER_NAME", visible in the AWS Console. That ASG is a good choice to begin with. Optionally, you may also create a new instancegroup "kops create ig _newgroupname_", and configure that instead. Set the maxSize of the kops instancesgroup, and update the cluster so the maxSize propagates to the ASG.
  
Question: The cluster-autoscaler [documentation](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler/cloudprovider/aws) mentions an IAM Policy. Which IAM Role should the Policy be attached to?    
Answer: Kops creates two Roles, nodes.$CLUSTER_NAME and masters.$CLUSTER_NAME. Currently the example scripts run the autoscaler process on the k8s master node, so the IAM Policy should be assigned to masters.$CLUSTER_NAME (substituting that variable for your actual cluster name).

