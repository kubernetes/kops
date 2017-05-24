# ROADMAP

# 1.7

## Must-have features

* Support for k8s 1.7
 
## Other features we are working on in the 1.7 timeframe

* etcd controller to allow moving between versions
* kops server for better team scenarios
* support for bare-metal
* more gossip backends
* IAM integration
* more cloud providers
* promote GCE to stable
* RBAC policies for all components
* bringing rolling-update out of alpha

# HISTORICAL

## 1.6

### Must-have features

* Support for k8s 1.6 _done_
* RBAC enabled by default _yes, but we kept RBAC optional_

## Other features we are working on in the 1.6 timeframe

* Support for GCE _alpha_
* Support for Google's [Container Optimized OS](https://cloud.google.com/container-optimized-os) (formerly known as GCI) _alpha_
* Some support for bare-metal _private branches, not merged_
* Some support for more cloud providers _initial work on vsphere_
* Some IAM integration _discussions, but no code_
* Federation made easy _no progress_
* Authentication made easy _no progress_
* Integration with kubeadm _kops now uses kubeadm for some RBAC related functionality_
* CloudFormation integration on AWS _beta_


