## How to update Kops - Kubernetes Ops

Update the latest source code from kubernetes/kops

```
cd ${GOPATH}/src/k8s.io/kops/
git pull && make
```

Alternatively, if you installed from Homebrew
```
brew update && brew upgrade kops
```
