## Meltdown and Spectre Advisory

| | |
|-------------|--------|
| NAME         	| Meltdown and Spectre Hardware Issues |
| Description  	| Systems with microprocessors utilizing speculative execution and branch prediction may allow unauthorized disclosure of information to an attacker with local user access via a side-channel analysis. 	|
| CVE(s)       	| [CVE-2017-5753](https://nvd.nist.gov/vuln/detail/CVE-2017-5753)  [CVE-2017-5754](https://nvd.nist.gov/vuln/detail/CVE-2017-5753) |
| NVD Severity 	| medium (attack range: local) |
| Last Updated  | Jan 01 2018 |

##  Details

This hardware exploit requires the installation of the Linux Kernel >= 4.4.110.
kops is currently running the 4.4.x kernel.  A rolling-update or replacement of
the kops ami image, with kops hosted on AWS is recommended. All platforms are
affected not just AWS.

Three CVEs have been released with spectre and meltdown.

- Variant 1: bounds check bypass (CVE-2017-5753)
- Variant 2: branch target injection (CVE-2017-5715)
- Variant 3: rogue data cache load (CVE-2017-5754)

Currently, Variant 1 and Variant 3 are solved with this advisory.


### Impacted kops / kubernetes Components

- kops maintained AMI
- All AMIs without a patched kernel are impacted
- All platforms are affected, not just AWS
- Linux kernel versions needed: 4.4: >= 4.4.110
- By default, kops runs an image that includes the 4.4 kernel. An updated image is available with 4.4.110
- If running another image please update to a fixed image, which must be provided by your distro

### Fixed Versions

The following AMIs contain an updated kernel.

- kope.io/k8s-1.5-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.7-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.8-debian-jessie-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.8-debian-stretch-amd64-hvm-ebs-2018-01-05
- kope.io/k8s-1.8-debian-stretch-amd64-hvm-ebs-2018-01-05

These are the images that are maintained by the kops project; please refer to
other vendors for the appropriate AMI version.

### Update Process

For all examples please replace `$CLUSTER` with the appropriate kops cluster
name.

#### Determine which instance groups exist

`kops get ig --name $CLUSTER`

#### Edit the kops instance groups 

Update the instance group With the appropriate image version via a `kops 
edit` command or `kops replace -f mycluster.yaml`.

#### Perform dry-run update, verifying that all instance groups are updated.

`kops update $CLUSTER` 

#### Update the cluster.

`kops update $CLUSTER --yes`

#### Perform a dry-run rolling-update

Verify that all instance groups will be rolled.

`kops rolling-update cluster --name $CLUSTER`

#### Roll the cluster

`kops rolling-update cluster --name $CLUSTER --yes`

## Tools / Diagnosis

If you do not see "Kernel/User page tables isolation: enabled", you are vulnerable.

```console
dmesg -H | grep 'page tables isolation'
      [  +0.000000] Kernel/User page tables isolation: enabled
```

## Notes
- https://coreos.com/blog/container-linux-meltdown-patch
- https://aws.amazon.com/de/security/security-bulletins/AWS-2018-013/
- https://security.googleblog.com/2018/01/todays-cpu-vulnerability-what-you-need.html
- https://spectreattack.com/
- https://xenbits.xen.org/xsa/advisory-254.html
- https://googleprojectzero.blogspot.co.uk/2018/01/reading-privileged-memory-with-side.html
- Paper: https://spectreattack.com/spectre.pdf
- https://01.org/security/advisories/intel-oss-10002
- https://meltdownattack.com/
- http://blog.cyberus-technology.de/posts/2018-01-03-meltdown.html
- Paper: https://meltdownattack.com/meltdown.pdf
- https://01.org/security/advisories/intel-oss-10003


