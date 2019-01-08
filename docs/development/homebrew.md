# Installing Kops via Hombrew

Homebrew makes installing kops [very simple for MacOS.](../install.md)
```bash
brew update && brew install kops
```

Master can also be installed via Homebrew very easily:
```bash
# HEAD of master
brew update && brew install kops --HEAD
```

Previously we could also ship development updates to homebrew but their [policy has changed.](https://github.com/Homebrew/brew/pull/5060#issuecomment-428149176)

Note: if you already have kops installed, you need to substitute `upgrade` for `install`.

You can switch between installed releases with:
```bash
brew switch kops 1.9.0
brew switch kops 1.10.0
```

# Releasing kops to Brew

Submitting a new release of kops to Homebrew is very simple.

### From a homebrew machine

`brew bump-formula-pr` makes it easy to update our homebrew formula.
This will automatically update the provided fields and open a PR for you.
More details on this script are located [here.](https://github.com/Homebrew/brew/blob/master/Library/Homebrew/dev-cmd/bump-formula-pr.rb)

Example usage:
```bash
brew bump-formula-pr kops \
       --url=https://github.com/kubernetes/kops/archive/1.7.1.tar.gz \
       --sha256=044c5c7a737ed3acf53517e64bb27d3da8f7517d2914df89efeeaf84bc8a722a
```

* Update the URL variable to the tar.gz of the new release source code
* Update the sha256 variable to SHA256 checksum of the new tar.gz

**If we change how dependencies work or if we make the install require something other than a simple make, we'll need to update the commands**

```brew edit kops``` will open an editor on your machine to edit the formula.
You can use this to make more in depth changes to the formula.

### Test that Homebrew formula works
```brew uninstall kops && brew install kops``` will install the new version.
Test and make sure that the new release works.

### Audit the Homebrew formula
```brew audit --strict --online``` will output any code that doesn't meet the Homebrew standards.

### Send a commit to the Homebrew repo

Rather than repeating documentation that might change, head over to
[Homebrew documentation](https://github.com/Homebrew/brew/blob/master/docs/Formula-Cookbook.md#commit)
for directions and conventions.


The formula can be found in hacks/brew/kops.rb.
