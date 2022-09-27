
# kOps Releases & Versioning

kOps supports the latest minor version and latest-1. E.g if kOps' latest version is 1.25, also 1.24 is supported and will receive bugfixes and minor feature additions. kOps intends to be backward compatible.  It is always recommended using the
latest version of kOps that supports the Kubernetes version you are using.  

The latest Kubernetes minor version supported by a kOps release is the one matching the kOps version. E.g for kOps 1.25, the highest supported Kubernetes version is 1.25. From that version, kOps additionally support Kubernetes two additional minor versions. In this case 1.24 and 1.23. To ease migration, kOps also supports two more minor versions that are considered deprecated. Bugs isolated to deprecated Kubernetes versions will not be fixed unless they prohibit upgrades to supported versions. kOps users are advised to run one of the [3 minor versions](https://kubernetes.io/releases/version-skew-policy/#supported-versions) Kubernetes supports. 

## Release Schedule

This project does not follow the Kubernetes release schedule. kOps aims to
provide a reliable installation experience for Kubernetes, and typically
releases about a month after the corresponding Kubernetes release. This time
allows for the Kubernetes project to resolve any issues introduced by the new
version and ensures that we can support the latest features. kOps will release
alpha and beta pre-releases for people that are eager to try the latest
Kubernetes release.  Please only use pre-GA kOps releases in environments that
can tolerate the quirks of new releases, and please do report any issues
encountered.
