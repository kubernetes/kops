/srv/kubernetes  - secrets

/srv/sshproxy - not used in "normal" environments?  Contains SSH keypairs for tunnelling.  Secrets, really.

/var/etcd - the etcd data volume.  This should be a direct EBS volume
