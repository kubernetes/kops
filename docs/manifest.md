# Kops Manifests

Kops supports a `--filename` or `-f` command line flag for CRUD operations with Kubernetes clusters.

The purpose of the flag is to allow abstraction from command line flags into a **.yml** configuration file.

The **.yml** file is a representation of all available command line flags for a particular operation.

### YAML Parsing

Kops **.yml** parsing with `-f` follows an **if exists, populate** methodology.

For instance if a user defines a directive in the **.yml** that only exists for a single operation. The directive will only be honored during that operation.

Given the following supported directives :

`create - key1, key2, key3`

`update - key1, key4`

And a **.yml** file that defines :

`key1: val1`

`key4: val4`

The directive **key1** will be registered in both create and update. But the directive **key4** will only be registered in update.

It is important to note that both create and update would run without error in this example. 

### CRUD usage

`kops create cluster -f manifest_example.yml`

`kops get cluster -f manifest_example.yml`

`kops update cluster -f manifest_example.yml`

`kops delete cluster -f manifest_example.yml`

### Running hybrid with command line flags

Kops supports hybrid operations. The user will be able to pass a partial **.yml** config file via `-f` and also a traditional Kops command line flag.

`kops create cluster -f partial_manifest_no_image.yml --image ami-12345678`

In situations where **both** a command line flag and a manifest attempt to define the same value, the manifest will always take precedence and trump a command line flag.

### Running with environment variables

Currently hybrid environmental variable operations are not supported. It is recommended to **unset all Kops related environmental variables** before running a CRUD operation with a `-f` manifest.

