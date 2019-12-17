# Installing kops (Binaries)

## MacOS

From Homebrew:

```bash
brew update && brew install kops
```

Developers can also easily install [development releases](development/homebrew.md).

From Github:

```bash
curl -Lo kops https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-darwin-amd64
chmod +x ./kops
sudo mv ./kops /usr/local/bin/
```

You can also [install from source](development/building.md).

## Linux

From Github:

```bash
curl -Lo kops https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod +x ./kops
sudo mv ./kops /usr/local/bin/
```

You can also [install from source](development/building.md).

# Installing Other Dependencies

## kubectl

`kubectl` is the CLI tool to manage and operate Kubernetes clusters.  You can install it as follows.

### MacOS

From Homebrew:
```
brew install kubernetes-cli
```

From the [official kubernetes kubectl release](https://kubernetes.io/docs/tasks/tools/install-kubectl/):

```
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/darwin/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```

### Linux

From the [official kubernetes kubectl release](https://kubernetes.io/docs/tasks/tools/install-kubectl/):

```
curl -Lo kubectl https://storage.googleapis.com/kubernetes-release/release/$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)/bin/linux/amd64/kubectl
chmod +x ./kubectl
sudo mv ./kubectl /usr/local/bin/kubectl
```
# Installing AWS CLI Tools

https://aws.amazon.com/cli/

 On MacOS, Windows and Linux OS:
 
 The officially supported way of installing the tool is with `pip`:
 
```bash
pip install awscli
```

##### _OR use these alternative methods for MacOS and Windows:_

### MacOS

You can grab the tool with homebrew, although this is not officially supported by AWS.
```bash
brew update && brew install awscli
```

### Windows

You can download the MSI installer from this page and follow the steps through the installer which requires no other dependencies: 
https://docs.aws.amazon.com/cli/latest/userguide/awscli-install-windows.html
