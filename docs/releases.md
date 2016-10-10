## Branch strategy

We develop on the master branch.  The master branch is expected to build and generally to work,
but has not necessarily undergone the more complete validation that a release would.  The `release`
branch is expected to always be stable.

We tag releases as needed from the `release` branch and upload them to github.

We occasionally batch merge from the master branch to the `release` branch.  We don't maintain
multiple release branches as we expect most people to upgrade kops to the latest version.  We also
don't (yet) do lots of cherry-picking for that reason.

The intention is that this allows for development velocity, while also allowing for stable releases.
