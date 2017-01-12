# fi - Fast Install

The `fi` package is responsible for holding the infrastructure that supports installing, provisioning and deploying resources to get kubernetes running in a cloud.

## Cloudup

The `cloudup` package within `fi` contains infrastructure that handles deploying cloud resources.

Cloudup has a concept of `models`. These are a core component of kops, and critical to understanding how the tool works. They offer flexibility for the tool, and serve as the main representation of infrastructure in the cloud. They are the glue that will map resources to kops functionality. These `models` are stored in `upup/models/cloudup`.

#### AWS

Currently `aws` is the most verbose and feature rich of the 2 primarily supported cloud providers. Cloudup will handle reading the `aws` models (found in `_aws`) and apply these resources to the cloud.

A user will notice `awstasks` and `awsup` directores within cloudup. These are the files that will actually map an `aws` resource (API Request) to go code.

By convention, every used API request in kops for `aws`, has an associated `awstask` with it.

#### GCE


## Nodeup

This tool is secondary to cloudup.

After the resources to the cloud have been deployed, we will now need to provision the resources.

Nodeup runs as a binary **on VMs in the cloud** and will handle bringing up a kubernetes cluster.

The core bits of `nodeup` infrastructure can be found here.
