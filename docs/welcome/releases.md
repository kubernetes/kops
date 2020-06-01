# Kops Releases & Versioning

Kops intends to be backward compatible.  It is always recommended using the
latest version of kops with whatever version of Kubernetes you are using.  We suggest
kops users run one of the [3 minor versions](https://github.com/kubernetes/community/blob/master/contributors/design-proposals/release/versioning.md#supported-releases-and-component-skew) Kubernetes is supporting however we
do our best to support previous releases for some period.

Kops does not, however, support Kubernetes releases with a greater minor release number
than that of kops. A minor version is the second digit in the
release number: kops version 1.16.0 has a minor version of 16. The numbering
follows the semantic versioning specification: MAJOR.MINOR.PATCH.

For example, kops 1.16.0 does not support Kubernetes 1.17.0, but kops 1.16.0
supports Kubernetes 1.16.5, 1.15.2, and several previous Kubernetes versions.

## Compatibility Matrix

| kops version  | k8s 1.13.x | k8s 1.14.x | k8s 1.15.x | k8s 1.16.x | k8s 1.17.x |
|---------------|------------|------------|------------|------------|------------|
| 1.17.0        | ✔          | ✔          | ✔          | ✔          | ✔          |
| 1.16.x        | ✔          | ✔          | ✔          | ✔          | ⚫         |
| 1.15.x        | ✔          | ✔          | ✔          | ⚫         | ⚫         |
| ~~1.14.x~~    | ✔          | ✔          | ⚫         | ⚫         | ⚫         |
| ~~1.13.x~~    | ✔          | ⚫         | ⚫         | ⚫         | ⚫         |


Use the latest version of kops for all releases of Kubernetes, with the caveat
that higher versions of Kubernetes are not _officially_ supported by kops.
Releases who are ~~crossed out~~ _should_ work, but we suggest they be upgraded soon.

## Release Schedule

This project does not follow the Kubernetes release schedule. Kops aims to
provide a reliable installation experience for Kubernetes, and typically
releases about a month after the corresponding Kubernetes release. This time
allows for the Kubernetes project to resolve any issues introduced by the new
version and ensures that we can support the latest features. Kops will release
alpha and beta pre-releases for people that are eager to try the latest
Kubernetes release.  Please only use pre-GA kops releases in environments that
can tolerate the quirks of new releases, and please do report any issues
encountered.
