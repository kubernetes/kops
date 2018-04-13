# ROADMAP

The kops 1.N.x version officially supports kubernetes 1.N.x and earlier.  While kubernetes 1.99 will likely run with kops 1.98,
the configuration will probably not be correct (for example docker versions, CNI versions etc).

kops 1.N.0 is released when it is believed that kubernetes 1.N.x is stable, along with all the core addons (e.g. networking).
This can mean that kops can release months after the release of kubernetes.  It's also not a deterministic release criteria,
particularly with some networking plugins that are supported by kops but themselves still under development.  We discussed
this challenge in kops office hours in March 2018, and the consensus was that we want to keep this, but that we should release
alphas & betas much earlier so that users can try out new kubernetes versions on release day.

For the next few releases this means that:

* 1.9.0 release target April 7th
* 1.10 alpha.1 with release of kops 1.9.0 (April 7th)
* 1.10 release target April 28th
* 1.11 alpha.1 at release of kops 1.10
* 1.11 beta.1 at release of k8s 1.11
* 1.12 alpha.1 at release of kops 1.11 etc


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

# HISTORICAL

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


