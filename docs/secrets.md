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

