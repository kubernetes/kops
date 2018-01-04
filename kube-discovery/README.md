## kube-discovery

Status: experimental

kube-discovery does master discovery, currently on bare-metal.  The intention is to split this functionality
out of protokube, to make it reusable and modular.

Discovery methods:

* mDNS/DNS-SD (aka bonjour / zeroconf).  Looks for services of type `_kubernetes._tcp`, with a name of clustername.


## mDNS

Example avahi configuration

`/etc/avahi/services/kubernetes.service`

```
<?xml version="1.0" standalone='no'?>
<!DOCTYPE service-group SYSTEM "avahi-service.dtd">

<service-group>
  <name replace-wildcards="yes">example.k8s.local</name>

  <service>
    <type>_kubernetes._tcp</type>
    <port>443</port>
  </service>

</service-group>
```
