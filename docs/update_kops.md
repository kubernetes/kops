# Updating kops (Binaries)

## MacOS

From Homebrew:

```bash
brew update && brew upgrade kops
```

From Github:

```bash
rm -rf /usr/local/bin/kops
wget -O kops https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-darwin-amd64
chmod +x ./kops
sudo mv ./kops /usr/local/bin/
```

You can also rerun [these steps](development/building.md) if previously built from source.

## Linux

From Github:

```bash
rm -rf /usr/local/bin/kops
wget -O kops https://github.com/kubernetes/kops/releases/download/$(curl -s https://api.github.com/repos/kubernetes/kops/releases/latest | grep tag_name | cut -d '"' -f 4)/kops-linux-amd64
chmod +x ./kops
sudo mv ./kops /usr/local/bin/
```

You can also rerun [these steps](development/building.md) if previously built from source.
