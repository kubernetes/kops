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


# RELEASE ROADMAP

## kops 1.11

* Full support for Kubernetes 1.
* Alpha support for bundles (etcd-manager is the test case)
* etcd3 will be the default for newly created clusters. 
  - Existing clusters will continue to run etcd2 but will need to be upgraded for 1.12
* Default to Debian stretch images which increase support for newer instance types
* Beginnings of automated releases that can be completed by any maintainer

## kops 1.12
* Full support for Kubernetes 1.12
* etcd3 improvements

# UPCOMING FEATURE ROADMAP 
NB: These are features that are in process and may be introduced behind flags or in alpha capacity but are not explicitly targeting specific releases. 

* Documentation revamp that is closer to k8s.io: Stories and walkthroughs of common scenarios, restructure and update information
* Additional cloud providor support: spotinst, aliyun
* Revisit recommended base cluster configurations to get them modernized. Update recommendations and defaults for instances, disks, etc, 


