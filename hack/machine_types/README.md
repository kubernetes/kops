Machine Types Generator
=======================

To prevent errors or lagging updates, we use this generator to update the known aws machine types
that are [hard coded in kops](https://github.com/kubernetes/kops/blob/7d7112c1e9a52d4f677db6bd98943d308ec9f581/upup/pkg/fi/cloudup/awsup/machine_types.go#L76).

This generator uses the AWS Pricing API to get most of it's info on what instance types are supported.

Usage
-----
```
make update-machine-types
git add .
git commit -am "Updated machine types"
```

TODO:
-----
* Cross reference other regions besides us-east-1.  Currently we just look at one region to determine instance types.
