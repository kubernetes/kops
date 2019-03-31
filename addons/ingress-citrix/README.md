# Deploy Citrix Ingress Controller through KOPS

[Citrix Ingress Controller](https://github.com/citrix/citrix-k8s-ingress-controller) has many robust Ingress functionalities. It can also be deployed when creating a Kubernetes cluster using KOPS addon.

# Quick Deploy using `kubectl`

## GCP

To deploy in Google Cloud Platform, use the below command 

```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ingress-citrix/v1.1.1.yaml
```

## AWS

To deploy in AWS, use the below command 

```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ingress-citrix/v1.1.1-aws.yaml
```

## AZURE
 
To deploy in Azure, use the below command 

```
kubectl create -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ingress-citrix/v1.1.1-azure.yaml
```
