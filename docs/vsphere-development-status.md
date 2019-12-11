# Kops tasks and effort estimation

List of open issues and features and their effort estimation as of <time datetime="2017-06-06" class="date-past">06 Jun 2017</time>.

**TODO** Issues listed below require proper labels to be assigned, especially P0.

**Priorities**

P0: Must have fixes and features, needed to make existing vSphere support in kops work.

P1: Important fixes and features, required to give vSphere users a more useful kubernetes cluster deployment experience, with multiple masters and HA.

P2: Rarely occurring issues and features that will bring vSphere support closer to AWS and GCE support in kops.

**Notes:**

* Effort estimation includes fix for an issue or implementation for a feature, testing and generating a PR.
* There are a few issues that are related to startup and base image. If we can resolve "Use PhotonOS for vSphere node template" issue first and replace init-cloud with guestinfo, those issues **might** get resolved automatically. But further investigation is needed and fixed issues will need verifications and testings.

|Priority|Task|Type (bug, feature, test|Effort estimate(in days)|Remarks|
|--- |--- |--- |--- |--- |
|P0|Kops for vSphere is broken kubernetes/kops#2729|Bug|1||
|P0|AWS EBS is set as default volume provisioner, instead of vSphere kubernetes/kops#2732|Bug|1|Looks like the fix is available, need to be tested again before committing https://github.com/vmware/kops/pull/70/files|
|P0|Package installation through nodeup causing delay in cluster deployment kubernetes/kops#2742|Bug|2|If we get the "Use PhotonOS for vSphere node template" done first, this one may just be avoided. Verification required.|
|P0|Connection to api.clustername.skydns.local is failing  kubernetes/kops#2744|Bug|1|This might be fixed by Abrar's PR in kubernetes already. Just need to verify.|
|P0|Make update command, that includes scale up and down, work kubernetes/kops#2738. There are two possible ways to implement this- without auto scaling group (ASG) or with auto scaling group|Feature|4 days Assuming ASG is available. 4 days ASG is not available.|This effort estimation needs more analysis.|
|P0|Make end-to-end CI/CD work on vmware/kops kubernetes/kops#2730|Bug, Test|3||
|P1|Use PhotonOS for vSphere node template  kubernetes/kops#2735. Use guestinfo instead of init-cloud kubernetes/kops#2726|Feature|5|This was originally P2 issue. However several other issues might be affected by this one. So bring it to P1. Problem to solve: init-cloud on PhotonOS not working properly. Or, get rid of init-cloud and use guestinfo instead.|
|P1|Default image name extraction for vSphere needs to be fixed kubernetes/kops#2740|Bug|2||
|P1|Multi-master HA setup kubernetes/kops#2734|Feature|5||
|P1|Add unit tests for vSphere workflows kubernetes/kops#2745|Test|8|This effort estimation needs more analysis as estimator is not familiar with all the components for which tests need to be written.|
|P1|Documentation of- 1) Existing commands and usage, blog kubernetes/kops#2739, 2) Behavior for all flags for ‘kops cluster create’ command kubernetes/kops#2741||3||
|P2|vCenter running out of HTTP sessions kubernetes/kops#2747|Bug|3||
|P2|Long boot time for template VM kubernetes/kops#2746|Bug|3|This estimation needs more analysis. Same here: if Use PhotonOS for vSphere node template can be resolved first, this might not be valid. Still, verification needed.|
|P2|Improve methods that create user-data and meta-data for ISO kubernetes/kops#2748, or use guestinfo instead of cloud-init for passing in VM specific informations kubernetes/kops#2726|Feature|2|We need decide If we should use guestinfo instead of cloud-init.|
|P2|Support for rolling upgrade, normal upgrade|Feature|7|This estimation needs more analysis. Presence of ASG might simplify the implementation of this feature. Some changes might be needed in core kops code, as current rolling upgrade implementation is very much AWS specific.|
|P2|Make ETCD volumes re-attachable for vSphere (AWS and GCE already support this) kubernetes/kops#2736|Feature|7|This task needs more analysis for design and implementation. Estimate might change accordingly.|
|P2|Security and isolation- 1) Networking for master, worker nodes kubernetes/kops#2731. 2) Credentials in plain text kubernetes/kops#2743|Feature|--| Don't have enough information on this to give an estimate for effort involved.|
|P2|Explore vSphere DRS cluster’s anti-affinity rules to achieve better master HA and meaningful zone allotment for masters kubernetes/kops#2733|Feature|8|This estimation needs more analysis. DRS anti-affinity rule needs to be explored, to see if it's even fit for this problem.|
|P2|Enable user-defined dns zone name kubernetes/kops#2727|Feature|2||
||||Total 67||


# Kops commands behavior for vSphere

List of all kops commands and how they behave for vSphere cloud provider, as of <time datetime="2017-04-13" class="date-past">13 Apr 2017</time> .

# Column explanation

* Command, option and usage example are self-explanatory.
* vSphere support: whether or not the command is supported for vSphere cloud provider (Yes/No), followed by current status of that command and explanation of any failures.
* Graceful termination needed: If the command will not supported, does it need additional code to fail gracefully for vSphere provider?
* Remark: Miscellaneous comments about the command.

|Command|Option|Usage example|vSphere support|Graceful termination needed (if not fixed)|Remark|
|--- |--- |--- |--- |--- |--- |
|completion|bash|kops completion bash|Yes|No|Output shell completion code for the given shell (bash), which can easily be incorporated in a bash script to run kops commands as bash functions.|
|create|cluster||Yes. Supported/tested command flags: cloud, dns, dns-zone, image, networking, node-count, vsphere-server, vsphere-datacenter, vsphere-resource-pool, vsphere-datastore, vsphere-coredns-server, yes, zones.|Yes. Check for unsupported flags. Terminate command, if needed, with appropriate message.|Creates cluster spec and configs. If --yes is specified then creates resources as well.|
|create|instancegroup|kops create ig --name=v1c1.skydns.local --role=Node --subnet=vmw-zone nodes2|No. InstanceGroup spec gets created in object store. Command however shows this error even after setting 'image' value in spec: I0412 11:08:23.025842   80677 populate_instancegroup_spec.go:257] Cannot set default Image for CloudProvider="vsphere"|Yes. Either add a check for vSphere, or fix the issue causing the failure.||
|create|secret|kops create secret sshpublickey test_key -i ~/.ssh/git_rsa.pub|Yes|No|Creates and delete secrets can be used in combination to replace existing secrets. Justin's explanation: "k8s in theory supports multiple certificates but it was not working until 1.5 so I don't think we actually enable it in kops This will be how we do certificate rotation though - add a certificate, roll that out, roll out a new key and switch to the new key"|
|create|-f FILENAME|Three yams files are required- cluster: kops create -f ~/kops.yaml, master IG: kops create -f ~/kops.nodeig.yaml, node IG: kops create -f ~/kops.masterig.yaml |Yes|No||
|delete|cluster|kops delete cluster v2c1.skydns.local --yes|Yes|No||
|delete|instancegroup|kops delete instancegroup --name=v2c1.skydns.local  nodes.v2c1.skydns.local|No. No implementation available to list resources. Method corresponding to AWS is getting called and crashing with panic, without any useful message.|Yes||
|delete|secret||Yes|-||
|delete|-f FILENAME|kops delete -f config.yaml  --name=v2c1.skydns.local|No. Cluster deletion works. Instance group deletion is failing with error: panic: interface conversion: *vsphere.VSphereCloud is not awsup.AWSCloud: missing method AddAWSTags goroutine 1 [running]: panic(0x26fbd20, 0xc420770780) /usr/local/go/src/runtime/panic.go:500 +0x1a1 k8s.io/kops/upup/pkg/kutil.FindCloudInstanceGroups|Yes|Delete cluster, ig specified by the file.|
|describe|secrets|kops describe secrets|Yes|No|Describe secrets, based on the kubectl context.|
|edit|cluster|kops  edit cluster --name=v2c1.skydns.local nodes|Yes. Edited spec gets updated in object store.|Yes|Edit works. But it would be a bad user experience if we allow users to edit the spec, followed by a failed 'kops update' and then no way to go back to the older spec.|
|edit|ig|kops  edit ig --name=v2c1.skydns.local nodes|Yes. Edited spec gets updated in object store.|Yes|Edit works. But it would be a bad user experience if we allow users to edit the spec, followed by a failed 'kops update' and then no way to go back to the older spec.|
|edit|federation||No|Yes|Federation is a group of k8s clusters. This doesn't look an important goal for vSphere in near future. Q: "How is a federation getting created? I see update and edit methods for a federation but I am not clear how to get a federation in the first place." A: Justin's reply: I'm chatting with the federation folk about kubefed & kops and whether we should integrate them etc.  The federation stuff was very alpha and I believe is (trivially) broken right now, but I'm debating integrating with kubefed vs fixing kops federation.  kubefed worked fine when I tried it the other day.|
|export|kubecfg|kops export kubecfg v1c1.skydns.local|Yes|-|Sets kubectl context to given cluster.|
|get|clusters||Yes|-|Gets list of clusters. If yaml output is specified, this output can be modified and used for 'kops replace' command.|
|get|federations||Yes|-|Gets list of federations. For now empty list.|
|get|instancesgroups||Yes|-|Gets list of intancegroups. If yaml output is specified, this output can be modified and used for 'kops replace' command.|
|get|secrets||Yes|-|Gets list of secrets.|
|import|cluster|kops import cluster --region=us-west-2 --name=v2c1.skydns.local nodes|No. Current implementation is very aws specific. Multiple aws services are queried to construct the api.Cluster object.|Yes|Imports spec for an existing cluster into the object store. While this functionality is good for importing and managing existing k8s clusters using kops, it doesn't seem like a high priority functionality at this point of time.|
|replace||kops replace -f FILENAME|No|Yes|Output of `kops get cluster name -oyaml` or `kops get ig name -oyaml` can be updated and passed to 'kops replace' command.|
|rolling-update|cluster||No. Current implementation is aws specific.|Yes||
|secrets|create||-|-|Legacy command, points to 'kops create secrets'.|
|secrets|describe||-|-|Legacy command, points to 'kops describe secrets'.|
|secrets|expose||-|-|Legacy command, points to 'kops get secrets -oplaintext'.|
|secrets|get||-|-|Legacy command, point to 'kops get secret'.|
|toolbox|dump||No. Current implementation is aws specific.|Yes|Dumps cloud information for the given cluster. This looks like a good to have functionality. Once resource listing is available for vsphere, which will anyways get used for deletion operation as well, this command should become easier to implement.|
|toolbox|convert-imported||No. Current implementation is aws specific.|Yes|Doesn't look like a high priority functionality.|
|update|cluster|kops  update cluster --name=v2c1.skydns.local --yes|No. 1) Works for new cluster. 2) Existing cluster scale up: vSphere provisioning code tries to provision all master and node VMs from scratch. New nodes get created and registered successfully. Existing resources keep failing with 'already exists' error. 3) Existing cluster scale down: Won't work, no resource listing or deletion logic available for vSphere. On top of that all listed resources- masters and workers are attempted for creation and fail with 'already exists' error.|Yes||
|update|federation||No|Yes|Federation is a group of k8s clusters. This doesn't look an important goal for vSphere in near future.|
|upgrade|cluster|kops upgrade cluster --name=v1c1.skydns.local --yes|No. Seeing this error: W0413 11:48:52.216116   15456 upgrade_cluster.go:202] No matching images specified in channel; cannot prompt for upgrade|Yes|Find out more about 'channel' in context of kops. Note that no --channel argument is specified.|
|validate|cluster|kops validate cluster --name=v1c1.skydns.local|Yes. Not working right now. Failing with this error: cannot get nodes for "v1c1.skydns.local": Get https://api.v1c1.skydns.local/api/v1/nodes: dial tcp: lookup api.v1c1.skydns.local: no such host|-|Investigation is already going on- https://github.com/kubernetes/kops/issues/2744. This issue will most likely get fixed by a fix in cloud-provider code that is not returning appropriate internal and external IP for the node.|
|version||kops version|Yes|-|Prints client version information. Eg: Version 1.6.0-alpha.1 (git-500cb69)|
