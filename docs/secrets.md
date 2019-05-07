## Managing secrets

### get secrets

### get secret <name> -oplaintext

-oplaintext exposes the raw secret value.

### describe secret

`kops describe secret`

### create secret

`kops create secret sshpublickey admin -i ~/.ssh/id_rsa.pub`

### delete secret

Syntax: `kops delete secret <type> <name>`
or `kops delete secret <type> <name> <id>`

The ID form can be used when there are multiple matching keys.

example:
`kops delete secret sshpublickey admin`

Note: it is currently not possible to delete secrets from the keystore that have the type "Secret"

### adding ssh credential from spec file
```bash
apiVersion: kops.k8s.io/v1alpha2
kind: SSHCredential
metadata:
  labels:
    kops.k8s.io/cluster: dev.k8s.example.com
spec:
  publicKey: "ssh-rsa AAAAB3NzaC1 dev@devbox"
```

## Workaround for changing secrets with type "Secret"
As it is currently not possible to modify or delete + create secrets of type "Secret" with the CLI you have to modify them directly in the kops s3 bucket.

They are stored /clustername/secrets/ and contain the secret as a base64 encoded string. To change the secret base64 encode it with:

```echo -n 'MY_SECRET' | base64```

and replace it in the "Data" field of the file. Verify your change with get secrets and perform a rolling update of the cluster.
