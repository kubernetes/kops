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
git checkout release
make
```

* The `release` branch is where releases are taken from.  This is the stable code branch.
* The `master` branch  _should_ also be functional, but is where active development happens, so may be less stable.

## Cross Compiling

Cross compiling for things like `nodeup` are now done automatically via `make nodeup`. `make push-aws-run TARGET=admin@$TARGET` will automatically choose the linux amd64 build from your `.build` directory.

