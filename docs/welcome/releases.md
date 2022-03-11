
# kOps Releases & Versioning

kOps intends to be backward compatible.  It is always recommended using the
latest version of kOps with whatever version of Kubernetes you are using.  We suggest
kOps users run one of the [3 minor versions](https://kubernetes.io/releases/version-skew-policy/#supported-versions) Kubernetes is supporting however we
do our best to support previous releases for some period.

kOps does not, however, support Kubernetes releases that have either a greater major
release number or greater minor release number than it.
(The numbers before the first and second dots are the major and minor release numbers, respectively.)
For example, kOps 1.20.0 does not support Kubernetes 1.21.0, but does
support Kubernetes 1.20.5, 1.19.2, and several previous Kubernetes versions.

## Compatibility Matrix

| kOps version  | k8s 1.19.x | k8s 1.20.x | k8s 1.21.x | k8s 1.22.x | k8s 1.23.x |
|---------------|------------|------------|------------|------------|------------|
| 1.23.x        | ✔          | ✔          | ✔          | ✔          | ✔          |
| 1.22.x        | ✔          | ✔          | ✔          | ✔          | ⚫         |
| ~~1.21.x~~    | ✔          | ✔          | ✔          | ⚫         | ⚫         |
| ~~1.20.x~~    | ✔          | ✔          | ⚫         | ⚫         | ⚫         |
| ~~1.19.x~~    | ✔          | ⚫         | ⚫         | ⚫         | ⚫         |


Use the latest version of kOps for all releases of Kubernetes, with the caveat
that higher versions of Kubernetes are not _officially_ supported by kOps.
Releases which are ~~crossed out~~ _should_ work, but are unlikely to get security or other fixes.
We suggest they be upgraded soon.

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
