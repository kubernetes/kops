## Prerequisite

`kubectl` is required, see [here](https://kubernetes.io/docs/tasks/tools/install-kubectl/).


## OSX From Homebrew

```console
brew update && brew install kops
```

The `kops` binary is also available via our [releases](https://github.com/kubernetes/kops/releases/latest).


## Linux

```console
curl -LO https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod +x kops-linux-amd64
sudo mv kops-linux-amd64 /usr/local/bin/kops
```

## Windows

1. Get `kops-windows-amd64` from our [releases](https://github.com/kubernetes/kops/releases/latest).
2. Rename `kops-windows-amd64` to `kops.exe` and store it in a preferred path.
3. Make sure the path you chose is added to your `Path` environment variable.
