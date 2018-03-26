# Building from source

[Installation from a binary](../install.md) is recommended for normal kops operation.  However, if you want
to build from source, it is straightforward:

If you don't have a GOPATH:

```
mkdir ~/kops
cd ~/kops
export GOPATH=`pwd`
```

Check out and build the code:

```
go get -d k8s.io/kops
cd ${GOPATH}/src/k8s.io/kops/
git checkout release
make
```

* The `release` branch is where releases are taken from.  This is the stable code branch.
* The `master` branch  _should_ also be functional, but is where active development happens, so may be less stable.

## Cross Compiling

Cross compiling for things like `nodeup` are now done automatically via `make nodeup`. `make push-aws-run TARGET=admin@$TARGET` will automatically choose the linux amd64 build from your `.build` directory.

## Troubleshooting

 - Make sure `$GOPATH` is set, and your [workspace](https://golang.org/doc/code.html#Workspaces) is configured.
 - kops will not compile with symlinks in `$GOPATH`. See issue go issue [17451](https://github.com/golang/go/issues/17451) for more information
 - building kops requires go 1.8 or 1.9
 - Kops will only compile if the source is checked out in `$GOPATH/src/k8s.io/kops`. If you try to use `$GOPATH/src/github.com/kubernetes/kops` you will run into issues with package imports not working as expected.
