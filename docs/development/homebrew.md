# Releasing kops to Brew

Submitting a new release of kops to Homebrew is very simple.

### From a homebrew machine

```brew edit kops``` will open an editor on your machine to edit the formula. Make the following changes:

* Update the URL variable to the tar.gz of the new release source code
* Update the sha256 variable to SHA256 checksum of the new tar.gz

**If we change how dependencies work or if we make the install require something other than a simple make, we'll need to update the commands**

### Test that Homebrew formula works
```brew uninstall kops && brew install kops``` will install the new version. Test and make sure that the new release works.

### Audit the Homebrew formula
```brew audit --strict --online``` will output any code that that doesn't meet the Homebrew standards.

### Send a commit to the Homebrew repo

Rather than repeating documentation that might change, head over to [Homebrew documentation](https://github.com/Homebrew/brew/blob/master/docs/Formula-Cookbook.md#commit) for directions and conventions.


The formula can be found in hacks/brew/kops.rb.
