# ROADMAP

kops 1.N.x _officially_ supports Kubernetes 1.N.x and earlier. Sometimes you can get lucky and kops 1.N.x manages Kubernetes 1.N+1.x perfectly. We cannot guarantee that this will always be the case and we recommend waiting for an official release of kops with minor version >= the version of Kubernetes you wish to install. Please see the []compatibility matrix](README.md#Compatibility_Matrix) for further questions.

There is a natural lag between the release of the next Kubernetes version and when kops supports it. While the first patch versions of a minor Kubernetes release are burning in, we finalize support. And then when we feel that Kubernetes 1.N.x has become reasonably stable, that's when we will release the corresponding version of kops with version specific configuration and a selection of add-ons to match.


In practive, sometimes this means that kops can release months after the release of Kubernetes. We sincerely hope that is not the case, but it happens. We also realize that this is not an entirely determininstic release process, particularly with respect to networking plugins that are continually in development. We understand this is in some ways not ideal, but the community decided to continue down this path and in lieu of quickly releasing and iterating on patch versions, to release alpha and beta versions early and often. We are still working on improving this process as seen below. 

Our goal will be to have an official kops release less than a month after the corresponding Kubernetes version is released. 
A rough outline of our envisioned timeline/release cycle with respect to the Kubernetes release follows. Our release process for alphas and betas will be getting some attention so that we can get these alpha and beta releases out to the community and other developers to help close out open items. 

July 1: Kubernetes 1.W.0 is released.
July 7: kops 1.W.beta1
July 21: kops 1.W.0 released
August 15: kops 1.W+1alpha1
August 31: kops 1.W+1alpha2
etc...
September 25: Kubernetes1.W+1.RC-X
Oct 1: Kubernetes 1.W+1.0
Oct 7: kops 1.W+1beta1
Oct 21: kops 1.W+1.0


# CURRENT ROADMAP

## kops 1.11

* support for Kubernetes 1.11
* greatly improved automated release process that helps us to get kops into more people's hands, earlier
* etcd3 default
* more sophisticated networking, iam, security group options
* add ons
* cluster/machines API support?

## kops 1.12

* support for Kubernetes 1.12
* networking spec
* documentation revamp
* update recommendations and defaults for instances, disks, etc, with respect to performance tuning
* etcd3 improvements


# HISTORICAL Roadmap Items

### _kops 1.10_

* Support for kubernetes 1.10
* Full support for GCE
* Make the etcd-backup tool enabled-by-default, so everyone should have backups.
* Allow users to opt-in to the full etcd-manager.
* Make etcd3 the default for new clusters, now that we have an upgrade path.
* Beginning of separation of addon functionality
* Support for more clouds (Aliyun, DigitalOcean, OpenStack)

### _kops 1.11_

* Make the etcd-manager the default, deprecate the protokube-integrated approach
* kops-server
* Machines API support (including bare-metal)

# 1.9

## Must-have features

* Support for k8s 1.9 _done_
* etcd backup support _done_

## Other features

* Use NodeAuthorizer / bootstrap kubeconfigs [#3551](https://github.com/kubernetes/kops/issues/3551) _no progress; may be less important with machines API_

# 1.8

## Must-have features

* Support for k8s 1.8

## Other features

* Improved GCE support
* Support for API aggregation

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


