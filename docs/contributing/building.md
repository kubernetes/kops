# Building from source

[Installation from a binary](../install.md) is recommended for normal kOps operation.  However, if you want
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

Cross compiling for things like `nodeup` are now done automatically via `make nodeup`. `make push-aws-run-amd64 TARGET=admin@$TARGET` will automatically choose the linux amd64 build from your `.build` directory.

## Debugging

To enable interactive debugging, the kOps binary needs to be specially compiled to include debugging symbols.
Add `DEBUGGING=true` to the `make` invocation to set the compile flags appropriately.

For example, `DEBUGGING=true make` will produce a kOps binary that can be interactively debugged.

### Interactive Debugging with Delve in Headless Mode

[Delve](https://github.com/go-delve/delve) is a debugger for Go programs. You can use it either directly via CLI or integrate it with an IDE like VS Code.

To run `kOps` under Delve in headless mode:

```bash
dlv --listen=:2345 --headless=true --api-version=2 exec ${GOPATH}/bin/kops -- <kops-arguments>
```

**Note:** Replace `<kops-arguments>` with the actual arguments you would normally pass to kOps CLI. Omit the `kops` keyword itself.

#### Example with Environment Variables

To pass environment variables to Delve, prepend them to the command as shown in the following example:

```bash
ENV_KEY=xxxxxxxxxxxxxxx \ 
ENV_KEY2=yyyyyyyyyyyyyyy \
dlv --listen=:2345 --headless=true --api-version=2 exec ${GOPATH}/bin/kops -- <kops-arguments>
```

### Configuring Delve in VS Code environment

To use Delve in VS Code, create a `.vscode/launch.json` file with the following configuration:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Attach to Delve (kOps)",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "port": 2345,
      "host": "127.0.0.1",
      "apiVersion": 2,
      "showLog": true,
      "trace": "verbose"
    }
  ]
}
```

Also, update your VS Code `settings.json` to specify your Delve binary path:

```json
"go.delveConfig": {
  "dlvPath": "/Users/username/go/bin/dlv"
}
```

Run the command `dlv --listen=:2345 --headless=true --a ...` and then you can launch the debugger by selecting the "Attach to Delve (kOps)" configuration in the Run and Debug view.

## Troubleshooting

 - Make sure `$GOPATH` is set, and your [workspace](https://golang.org/doc/code.html#Workspaces) is configured.
 - kOps will not compile with symlinks in `$GOPATH`. See issue go issue [17451](https://github.com/golang/go/issues/17451) for more information
 - building kops requires go 1.15
 - kOps will only compile if the source is checked out in `$GOPATH/src/k8s.io/kops`. If you try to use `$GOPATH/src/github.com/kubernetes/kops` you will run into issues with package imports not working as expected.
