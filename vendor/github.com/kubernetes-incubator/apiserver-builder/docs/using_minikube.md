# THIS DOC HAS BEEN DEPRECATED

`apiserver-boot` now supports running in minikube using `run local-minikube`
See [Running in minikube](running_in_minikube.md) instead


# Running the apiserver with delegated auth against minikube

- start [minikube](https://github.com/kubernetes/minikube)
  - `minikube start`
- copy `~/.kube/config` to `~/.kube/auth_config`
  - `kubectl config use-context minikube`
  - `cp ~/.kube/config ~/.kube/auth_config`
- add a `~/.kube/config` entry for your apiserver, using the minikube user
  - `kubectl config set-cluster mycluster --server=https://localhost:9443 --certificate-authority=/var/run/kubernetes/apiserver.crt` // Use the cluster you created and the minikube user
  - `kubectl config set-context mycluster --user=minikube --cluster=mycluster`
  - `kubectl config use-context mycluster`
- make the directory `/var/run/kubernetes` if it doesn't exist
  - `sudo mkdir /var/run/kubernetes`
  - `sudo chown $(whoami) /var/run/kubernetes`
- run the server with ` ./main --authentication-kubeconfig ~/.kube/auth_config --authorization-kubeconfig ~/.kube/auth_config --client-ca-file /var/run/kubernetes/apiserver.crt  --requestheader-client-ca-file /var/run/kubernetes/apiserver.crt --requestheader-username-headers=X-Remote-User --requestheader-group-headers=X-Remote-Group --requestheader-extra-headers-prefix=X-Remote-Extra- --etcd-servers=http://localhost:2379 --secure-port=9443 --tls-ca-file  /var/run/kubernetes/apiserver.crt  --print-bearer-token`
  - This will have the server use minikube for delegated auth