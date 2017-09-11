# Installing Kops (Binaries)

## Darwin (MacOS)

From Homebrew:

```bash
brew update && brew install --HEAD kops
```

# Installing Kops (Source)

```
go get -d k8s.io/kops
cd ${GOPATH}/src/k8s.io/kops/
git checkout release-1.7
make
```

* The `release-1.7` branch is where the latest release is from.  This is the stable code branch.
* The `master` branch  _should_ also be functional, but is where active development happens, so may be less stable.

## Cross Compiling

Cross compiling for things like `nodeup` are now done automatically via `make nodeup`. `make push-aws-run TARGET=admin@$TARGET` will automatically choose the linux amd64 build from your `.build` directory.

## Troubleshooting 

 - Make sure `$GOPATH` is set, and your [workspace](https://golang.org/doc/code.html#Workspaces) is configured.
 - kops will not compile with symlinks in `$GOPATH`. See issue go issue [17451](https://github.com/golang/go/issues/17451) for more information
 - kops uses the relatively new Go vendoring, so building requires Go 1.6 or later, or you must export GO15VENDOREXPERIMENT=1 when building with Go 1.5.  The makefile sets GO15VENDOREXPERIMENT for you.  Go code generation does not honor the env var in 1.5, so for development you should use Go 1.6 or later
 - Kops will only compile if the source is checked out in `$GOPATH/src/k8s.io/kops`. If you try to use `$GOPATH/src/github.com/kubernetes/kops` you will run into issues with package imports not working as expected.

# Installing Other Dependencies

## Installing Kubectl

`kubectl` is the CLI tool to manage and operate Kubernetes clusters.  You can install it as follows.

### Darwin (MacOS)

```
brew install kubernetes-cli
```

### Other Platforms

* [Kubernetes Latest Release](https://github.com/kubernetes/kubernetes/releases/latest)
* [Installation Guide](http://kubernetes.io/docs/user-guide/prereqs/)


## Installing AWS CLI Tools

### Darwin (MacOS)

The officially supported way of installing the tool is with `pip` as in

```bash
pip install awscli
```

You can also grab the tool with homebrew, although this is not officially supported by AWS.

```bash
brew update && brew install awscli
```
