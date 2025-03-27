
## Deployment
### AWS
```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ingress-nginx/v1.6.0.yaml
```

### GCE
```
kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ingress-nginx/v1.6.0-gce.yaml
```

## Creating a simple ingress

```
kubectl create deployment echoheaders --image=registry.k8s.io/echoserver:1.4 --replicas=1 --port=8080
kubectl expose deployment echoheaders --port=80 --target-port=8080 --name=echoheaders-x
kubectl expose deployment echoheaders --port=80 --target-port=8080 --name=echoheaders-y

kubectl apply -f https://raw.githubusercontent.com/kubernetes/kops/master/addons/ingress-nginx/ingress.yaml

kubectl get services echoheaders-x -owide

NAME            CLUSTER-IP      EXTERNAL-IP                                                               PORT(S)          AGE       SELECTOR
echoheaders-x   100.71.196.44   a29c28f4b8b0811e685cb0a924c5a8a1-1593015597.us-east-1.elb.amazonaws.com   80/TCP,443/TCP   13m       app=echoheaders

curl -v -H "Host: bar.baz.com" http://a29c28f4b8b0811e685cb0a924c5a8a1-1593015597.us-east-1.elb.amazonaws.com/bar
```
