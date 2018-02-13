# Cluster Autoscaler Addon

We strongly recommend using Cluster Autoscaler with the kubernetes version for which it was meant. Refer to the [Cluster Autoscaler documentation compatibility matrix]( https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/README.md#releases)

Note that you likely want to change `AWS_REGION` and `GROUP_NAME`, and probably `MIN_NODES` and `MAX_NODES`. Here is an example of how you may wish to do so:

```bash
CLOUD_PROVIDER=aws
IMAGE=gcr.io/google_containers/cluster-autoscaler:v1.1.0
MIN_NODES=1
MAX_NODES=5
AWS_REGION=us-east-1
GROUP_NAME="nodes.k8s.example.com"
SSL_CERT_PATH="/etc/ssl/certs/ca-certificates.crt" # (/etc/ssl/certs for gce)

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
