Adding this temp file to facilitate discussion WIP of this feature.


So at a glance I appears that lives under nodeup and you already have a structure in place for different os families.

```bash
./upup/models/nodeup/_automatic_upgrades/_debian_family
./upup/models/nodeup/_kubernetes_master/_kube-addons/_debian_family
./upup/models/nodeup/docker/_systemd/_debian_family
./upup/models/nodeup/ntp/_aws/_debian_family
./upup/models/nodeup/top/_debian_family
...
```

My initials thoughts

1. docker/rkt are bundled by default with CoreOS so its a matter of what configs we need to apply to satisfy k8s. Patterns for handling this the "CoreOS way" are already documented. Typically handled as systemd drop-ins via cloud-init
2. Items that are installed via apt-get (or into the /usr partition) would require work arounds to deal with a read only partition and no installer (i.e. bundled via container)
3. Leveraging user-data will maximize configuration and immutability of the deployment
4. Size limits of cloud-init would influence approach
    - 16K limit - http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-instance-metadata.html
    - Embedding all the k8s manifests, certs, and add-ons directly in the user-data may not feasible if they exceed this limit
    - They might need to be externalized and retrieved during boot time (would that go in KOPS_STATE_STORE?)
5. Systemd units hopefully would be portable (need to confirm)
6. How much would you want to handle with nodeup versus straight cloud-init?
7. Networking - what type of networking would you want to tackle first?
8. Preference between cloud-init or ignition?

If you feel these are valid comments/observations and in the right direction, I can start the PR and reverse my way out of the upup package.

