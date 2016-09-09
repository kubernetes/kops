# Images

Changing the image for an instance group

You can choose a different AMI for an instance group.

If you `kops edit ig nodes`, you should see an `image` member of the spec.

Various syntaxes are available:

* `ami-abcdef` specifies an AMI by id directly.
* `<owner>/<name>` specifies an AMI by its owner and Name properties

The ami spec is precise, but AMIs vary by region.  So it is often more convenient to use the `<owner>/<name>`
specifier, if equivalent images have been copied to various regions with the same name.

For example, to use Ubuntu 16.04, you could specify:

`image: 099720109477/ubuntu/images/hvm-ssd/ubuntu-xenial-16.04-amd64-server-20160830`

(Please note that ubuntu is currently undergoing validation testing with k8s - use at your own risk!)

If you are creating a new cluster you can use the `--image` flag when running `kops create cluster`,
which should be easier than editing your instance groups.
