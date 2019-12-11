## Kernel Update required for "Spectre/Meltdown" issue

| | |
|-------------|--------|
| NAME         	| Meltdown and Spectre Hardware Issues |
| Description  	| Systems with microprocessors utilizing speculative execution and branch prediction may allow unauthorized disclosure of information to an attacker with local user access via a side-channel analysis. 	|
| Related CVE(s) | [CVE-2017-5715](https://nvd.nist.gov/vuln/detail/CVE-2017-5715) [CVE-2017-5753](https://nvd.nist.gov/vuln/detail/CVE-2017-5753) [CVE-2017-5754](https://nvd.nist.gov/vuln/detail/CVE-2017-5754)|
| NVD Severity 	| medium (attack range: local) |
| Document Last Updated  | January 07,2018 |

## Summary

* All unpatched versions of linux are vulnerable when running on affected hardware, across all platforms (AWS, GCE, etc)
* Patches are included in Linux 4.4.110 for 4.4, 4.9.75 for 4.9, 4.14.12 for 4.14.
* kops can run an image of your choice, so we can only provide detailed advice for the default image.
* By default, kops runs an image that includes the 4.4 kernel. An updated image is available with the patched version (4.4.110).  Users running the default image are strongly encouraged to upgrade.
* If running another image please see your distro for updated images.

## CVEs

Three CVEs have been made public, representing different ways to exploit the same underlying
speculative-execution hardware issue:

- Variant 1: bounds check bypass (CVE-2017-5753)
- Variant 2: branch target injection (CVE-2017-5715)
- Variant 3: rogue data cache load (CVE-2017-5754)

The kernel updates that are the subject of this advisory are primarily intended to mitigate CVE-2017-5753 and CVE-2017-5754.

## Detecting vulnerable software

If you do not see "Kernel/User page tables isolation: enabled" in `dmesg`, you are vulnerable.

```console
dmesg -H | grep 'page tables isolation'
      [  +0.000000] Kernel/User page tables isolation: enabled
```

## Impacted Maintained Component(s)

* Patches were released for the linux kernel 2018-01-05.  All images prior to this date likely need updates.
* The kubernetes/kops maintained AMI is the maintained component that is vulnerable, although this likely affects all users.

### Fixed Versions

For the kops-maintained AMIs, the following AMIs contain an updated kernel:

- kope.io/k8s-1.5-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.7-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.8-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.8-debian-stretch-amd64-hvm-ebs-2018-01-05

These are the images that are maintained by the kubernetes/kops project; please refer to
other vendors for the appropriate AMI version.

### Update Process

For all examples please replace `$CLUSTER` with the appropriate kops cluster
name.

#### List instance groups

`kops get ig --name $CLUSTER`

#### Update the image for each instance group

Update the instance group with the appropriate image version via a `kops 
edit` command or `kops replace -f mycluster.yaml`.

#### Preview changes

Perform a dry-run update, verifying that all instance groups are updated.

`kops update cluster --name $CLUSTER` 

#### Apply changes

Update the cluster configuration, so that new instances will start with the updated image.

`kops update cluster --name $CLUSTER --yes`

#### Preview rolling update

Perform a dry-run rolling-update, to verify that all instance groups will be rolled.

`kops rolling-update cluster --name $CLUSTER`

#### Roll the cluster

Performing a rolling-update of the cluster ensures that all old instances and replaced with new instances,
running the updated image.

`kops rolling-update cluster --name $CLUSTER --yes`

## Resources / Notes

- https://aws.amazon.com/de/security/security-bulletins/AWS-2018-013/
- https://security.googleblog.com/2018/01/todays-cpu-vulnerability-what-you-need.html
- https://coreos.com/blog/container-linux-meltdown-patch
- https://spectreattack.com/
- https://xenbits.xen.org/xsa/advisory-254.html
- https://googleprojectzero.blogspot.co.uk/2018/01/reading-privileged-memory-with-side.html
- Paper: https://spectreattack.com/spectre.pdf
- https://01.org/security/advisories/intel-oss-10002
- https://meltdownattack.com/
- http://blog.cyberus-technology.de/posts/2018-01-03-meltdown.html
- Paper: https://meltdownattack.com/meltdown.pdf
- https://01.org/security/advisories/intel-oss-10003




