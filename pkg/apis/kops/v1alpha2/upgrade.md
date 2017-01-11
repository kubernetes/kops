If private topology

Rename your private subnet.  It will be called something like `private-us-east-1c.cluster.example.com`, rename it to 
 `us-east-1c.cluster.example.com`
 
Rename your route table  It will be called something like `main-cluster.example.com`, rename it to 
                          `cluster.example.com`

Create an instance group for the bastions.  A name of `bastion` will minimize changes.

`kops create ig bastion --role bastions`
