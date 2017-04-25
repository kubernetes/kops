# Release Process
## Prerequisites

* Up-to-date versions of `git`, `go` etc
* Install GitHub release tool `go get github.com/weaveworks/github-release`
* Create a [github token for
  github-release](https://help.github.com/articles/creating-an-access-token-for-command-line-use/); select the `repo` OAuth scope; set and export `$GITHUB_TOKEN` with this value

## Release Types

The release script behaves differently depending on the kind of
release you are doing. There are three types:

* **Mainline** - a release (typically from master) with version tag
  `vX.Y.Z` where Z is zero (e.g. `v2.1.0`)
* **Branch** - a bugfix release (typically from a branch) with version tag
  `vX.Y.Z` where Z is non-zero (e.g `v2.1.1`)
* **Prerelease** - a release from an arbitrary branch with an arbitrary
  version tag (e.g. `feature-preview-20150904`)

N.B. the release script _infers the release type from the format of the
version tag_. Ensure your tag is in the correct format and the desired
behaviours for that type of release will be obtained from the script.

## Build Phase
### Update CHANGELOG.md

* Checkout the branch from which you wish to release
* Choose a version tag (see above) henceforth referred to as `$TAG`.
* Add a changelog entry for the new tag at the top of `CHANGELOG.md`.
  The first line must be a markdown header of the form `## Release
  $TAG` for **Prerelease** builds, `## Release ${TAG#v}` otherwise.

Commit the changelog update:

    git commit -m "Add release $TAG" CHANGELOG.md

### Create Version Tag

Next you must tag the changelog commit with `$TAG`

    git tag -a -m "Release $TAG" $TAG

### Execute Build

You are now ready to perform the build. If you have skipped the
previous steps (e.g. because you're doing a rebuild), you must ensure
that `HEAD` points to the tagged commit. You may then execute

    bin/release build

This has the following effects:

* `git tag --points-at HEAD` is used to determine `$TAG` (hence the
  `HEAD` requirement)
* Your *local* repository is cloned into `releases/$TAG`
* `CHANGELOG.md` is checked to ensure it has an entry for `$TAG`
* Distributables injected with `$TAG` are built
* Tests are executed

## Draft Phase
### Push Version Tag Upstream

First you must push your branch and version tag upstream, so that an
associated GitHub release may be created:

    git push git@github.com:weaveworks/weave
    git push git@github.com:weaveworks/weave $TAG

N.B. if you're testing the release process, push to your fork
instead!

### Create Draft Release

You're now ready to draft your release notes:

    bin/release draft

This has the following effects:

* A [release](https://help.github.com/articles/about-releases) is
  created in GitHub for `$TAG`. This release is in the draft state, so
  it is only visible to contributors; for **Prerelease** builds the
  pre-release attribute will also be set
* The `weave` script is uploaded as an attachment to the release

Navigate to https://github.com/weaveworks/weave/releases, 'Edit' the
draft and input the release notes. When you are done make sure you
'Save draft' (and not 'Publish release'!).

Once the release notes have passed review, proceed to the publish
phase.

## Publish Phase
### Move/Force Push `latest_release` Tag

This step must only be performed for **Mainline** and **Branch**
releases:

    git tag -af -m "Release $TAG" latest_release $TAG
    git push -f git@github.com:weaveworks/weave latest_release

The `latest_release` tag *must* point at `$TAG`, *not* at `HEAD` -
the build script will complain otherwise.

N.B. if you're testing the release process, push to your fork
instead!

### Publish Release & Distributable Artefacts

You can now publish the release and upload the remaining
distributables to DockerHub:

    bin/release publish

The effects of this step depend on the inferred release type. The
following occurs for all types:

* Docker images are tagged `$TAG` and pushed to DockerHub
* GitHub release moves from draft to published state

Additionally, for **Mainline** and **Branch** types:

* Release named `latest_release` is updated on GitHub

Finally, for **Mainline** releases only:

* Images tagged `latest` are updated on DockerHub

### Finish up

* If not on master, merge branch into master and push to GitHub.
* Close the [milestone](https://github.com/weaveworks/weave/milestones) in GitHub and create the next milestone
* Update the `#weavenetwork` topic heading on freenode (requires 'chanops' permission)
* For a mainline release vX.Y.0, create a release branch X.Y from the
  tag and push to GitHub - this will result in X.Y.0 site docs being
  published to https://www.weave.works
* Add the new version of `weave-net` to the checkpoint system at
  https://checkpoint-api.weave.works/admin
* File a PR to update the version of the daemonset at
  https://github.com/kubernetes/kops/tree/master/upup/models/cloudup/resources/addons/networking.weave

## Troubleshooting

There's a few things that can go wrong.

 * If the build is wonky, e.g., the tests don't pass, you can delete
   the directory in `./releases/`, fix whatever it is, move the
   version tag (which should still be only local) and have another go.
 * If the DockerHub pushes fail (which sadly seems to happen a lot),
   you can just run `./bin/release publish` again.
 * If you need to overwrite a release you can do so by manually
   deleting the GitHub version release and re-running the process.
   Please note that the DockerHub `latest` images, GitHub
   `latest_release` and download links may be in an inconsistent state
   until the overwrite is completed.
