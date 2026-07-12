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

By default, the kOps binary is built with optimizations enabled and debug symbols stripped. To debug a prebuilt binary, it needs to be compiled with debugging symbols instead. Add `DEBUGGABLE=true` to the `make` invocation to set the compile flags appropriately.

For example, `DEBUGGABLE=true make` will produce a kOps binary that can be interactively debugged.

### Interactive debugging with Delve

[Delve](https://github.com/go-delve/delve) can be used to interactively debug the kOps binary.

The simplest way to start a debug session is `dlv debug`, which compiles the package with optimizations disabled and runs it under the debugger, so it does not require a special build. Run it from the root of the kOps source tree, and pass the kOps arguments after `--`, omitting the `kops` command itself:

```bash
dlv debug k8s.io/kops/cmd/kops -- update cluster --name mycluster.example.com
```

To debug a binary that was already built with `DEBUGGABLE=true make`, use `dlv exec` instead:

```bash
dlv exec ${GOPATH}/bin/kops -- update cluster --name mycluster.example.com
```

Environment variables such as `KOPS_STATE_STORE` are inherited by the debugged process, so set them the same way you would for a normal kOps invocation:

```bash
KOPS_STATE_STORE=s3://my-state-store \
dlv debug k8s.io/kops/cmd/kops -- update cluster --name mycluster.example.com
```

### Headless mode

To use Delve with an Interactive Development Environment (IDE), run it in headless mode and let the IDE connect to it:

```bash
dlv debug --headless --listen=:2345 --api-version=2 k8s.io/kops/cmd/kops -- update cluster --name mycluster.example.com
```

Then configure your IDE to connect its debugger to port 2345 on localhost.

### Debugging with VS Code

The VS Code Go extension manages Delve itself, so a headless server is not needed. Create a `.vscode/launch.json` file with the kOps arguments and environment variables for the command you want to debug:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "kops update cluster",
      "type": "go",
      "request": "launch",
      "mode": "debug",
      "program": "${workspaceFolder}/cmd/kops",
      "args": ["update", "cluster", "--name", "mycluster.example.com"],
      "env": {
        "KOPS_STATE_STORE": "s3://my-state-store"
      }
    }
  ]
}
```

Alternatively, to attach VS Code to a Delve server that was started in headless mode, use a configuration like this:

```json
    {
      "name": "Attach to Delve (kOps)",
      "type": "go",
      "request": "attach",
      "mode": "remote",
      "port": 2345,
      "host": "127.0.0.1"
    }
```

## Troubleshooting

 - Make sure `$GOPATH` is set, and your [workspace](https://golang.org/doc/code.html#Workspaces) is configured.
 - kOps will not compile with symlinks in `$GOPATH`. See issue go issue [17451](https://github.com/golang/go/issues/17451) for more information
 - building kops requires go 1.15
 - kOps will only compile if the source is checked out in `$GOPATH/src/k8s.io/kops`. If you try to use `$GOPATH/src/github.com/kubernetes/kops` you will run into issues with package imports not working as expected.
