# Release Process

#### 1. Maintainers quorum

* A majority of maintainers must agree on a single commit SHA for the release. The agreement should be documented in the release notes.
* The maintainers will match the commit SHA to the release `major.minor.patch` in the release notes
* The maintainers will agree on a changelog proposal that MUST be in a PR prior to the release.
* The main `README.md` must also reflect the new release. Ideally this should be in the same PR.

#### 2. Merge the release notes

* `HISTORY.md` serves as our official changelog.
* Before the release, the maintainers had opened a PR with the proposed changelog. This should now be merged.


#### 3. Release kops to github

* Darwin binary (generated)
* Linux binary (generated)
* Source code (zip)
* Source code (tar.gz)

#### 4. Release kops to homebrew

* Following the [documentation](development/homebrew.md) we must release a compatible homebrew formulae with the release.
* This should be done at the same time as the release, and we will iterate on how to improve timing of this.

#### 5. AWS support

* Build and release new AMI
* Build and release new nodeup binary to S3
* Build and release new kops binary to S3
* Build and release new protokube docker image
* Build and release new dnscontroller docker image
* Need to pull a tag for channels

#### 6. Update release branch

* Merge the `master` branch into the `release` branch. [More information](releases.md#branch-strategy)
* Create a `tag` from the newly merged `release` branch.

#### 7. Manual test and validate

* Maintainers should now give the repository a once over to validate everything looks in place.
* A majority of the maintainers must run the recently released code to verify success.

#### 8. Announce the release

* Announce the new release on twitter under the *k8sops* account
* Announce the new release on slack

#### 9. Clean up

* Validate the release milestone has been cleaned in github.
* All remaining issues need to be bumped to the backlog, or the next milestone



# Release Cadence

The core maintainers of kops are in agreement that we should try to keep an agile, and effective release cadence. Our goal should be to release as often as possible, while keeping our code rigid and reliable.

Under no circumstance should we rush a release and risk releasing immature code.

# Conventions

### Code Freeze

We should lock down the `master` branch on a given date. The lockdown is designed to give the maintainers a grasp on what is and isn't in the release, and provide a deadline for new features.

### Versioning

Kops uses [semantic versioning](http://semver.org/) to version the code base.

```
major.minor.patch
```

Our goal is to keep `major.minor` in sync (within reason) to the Kubernetes codebase. But to release our own patch versions as needed.

# Release Branch strategy

We develop on the master branch.  The master branch is expected to build and generally to work,
but has not necessarily undergone the more complete validation that a release would.  The `release`
branch is expected to always be stable.

We tag releases as needed from the `release` branch and upload them to github.

We occasionally batch merge from the master branch to the `release` branch.  We don't maintain
multiple release branches as we expect most people to upgrade kops to the latest version.  We also
don't (yet) do lots of cherry-picking for that reason.

The intention is that this allows for development velocity, while also allowing for stable releases.
