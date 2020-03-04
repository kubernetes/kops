# ROADMAP

## VERSION SUPPORT
kops 1.N.x _officially_ supports Kubernetes 1.N.x and prior versions. We understand that those in the community run a wide selection of versions and we do our best to maintain backward compatibility as far as we can. 

However, kops 1.N.x does NOT support Kubernetes 1.N+1.x. Sometimes you get lucky and kops 1.N will technically install a later version of Kubernetes, but we cannot guarantee or support this situation. As always, we recommend waiting for the official release of kops with minor version >= the version of Kubernetes you wish to install. Please see the [compatibility matrix](README.md#Compatibility_Matrix) for further questions.

## RELEASE SCHEDULE
There is a natural lag between the release of Kubernetes and the corresponding version of kops that has full support for it. While the first patch versions of a minor Kubernetes release are burning in, the kops team races to incorporate all the updates needed to release. Once we have both some stability in the upstream version of Kubernetes AND full support in kops, we will cut a release that includes version specific configuration and a selection of add-ons to match.

In practice, sometimes this means that kops release lags the upstream release by 1 or more months. We sincerely try to avoid this scenario- we understand how important this project is and respect the need that teams have to maintain their clusters. 

Our goal is to have an official kops release no later than a month after the corresponding Kubernetes version is released. Please help us achieve this timeline and meet our goals by jumping in and giving us a hand. We always need assistance closing issues, reviewing PRs, and contributing code! Stop by office hours if you're interested. 

A rough outline of the timeline/release cycle with respect to the Kubernetes release follows. We are revising the automation around the release process so that we can get alpha and beta releases out to the community and other developers much faster for testing and to get more eyes on open issues.

Example release timeline based on Kubernetes quarterly release cycle:

July 1: Kubernetes 1.W.0 is released.  
July 7: kops 1.W.beta1  
July 21: kops 1.W.0 released  
August 15: kops 1.W+1alpha1  
August 31: kops 1.W+1alpha2  

... etc

September 25: Kubernetes1.W+1.RC-X  
Oct 1: Kubernetes 1.W+1.0  
Oct 7: kops 1.W+1beta1  
Oct 21: kops 1.W+1.0  


## UPCOMING RELEASES

### kops 1.17

* Full support for Kubernetes 1.17

### kops 1.18

* Full support for Kubernetes 1.18
* Support for Containerd as an alternate container runtime
* Surging and greater parallelism in rolling updates

## UPCOMING FEATURES
NB: These are features that are in process and may be introduced behind flags or in alpha capacity but are not explicitly targeting specific releases. 

* Documentation revamp that is closer to k8s.io: Stories and walkthroughs of common scenarios, restructure and update information
* Additional cloud provider support: spotinst, aliyun, azure...?
* Revisit recommended base cluster configurations to get them modernized. Update recommendations and defaults for instances, disks, etc, 
* Improved node bootstrapping
