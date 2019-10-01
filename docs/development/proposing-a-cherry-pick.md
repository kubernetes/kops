# Cherry-picking changes back to previous branches

## Background

Anyone can propose a cherry-pick - you do not need to be the original author.

We typically have multiple release branches "open" at once - ideally one that is
a stable release branch ("GA"), one that is a beta, and one that is an alpha.
Sometimes the alpha branch will be the `master` branch.

Broad guidelines for acceptance of a cherry-pick:

* Once a version is beta or GA, we typically won't cherry-pick new features back
  to it.
* We typically will cherry-pick security-fixes and regression-fixes into the
  GA branch (and the alpha & beta branches).
* We may cherry-pick bug-fixes into the stable branch if the bug is
  sufficiently severe and outweighs the risk of introducing new regressions.
* Generally we are more accepting of cherry-picks to the beta branch (bug-fixes
  generally accepted), and very accepting to the alpha branch (features and
  bug-fixes generally accepted).  As a beta is maturing to GA, and the alpha is
  maturing into a beta, we will become more risk-averse on the cherry-picks, so
  that the branches can stabilize.

We are currently tracking cherry-picks to the various branches in a [spreadsheet](https://docs.google.com/spreadsheets/d/1zU67srtZUjuu_9UD7a-mBO6Gp5Z9k4EpYKsJw08_U9c/edit#gid=0)

## Process

The kubernetes repo has a [script for cherry-picking](https://github.com/kubernetes/kubernetes/blob/master/hack/cherry_pick_pull.sh) and you can download the raw form using something like:

```
wget -O /usr/local/bin/cherry_pick_pull.sh https://raw.githubusercontent.com/kubernetes/kubernetes/master/hack/cherry_pick_pull.sh
chmod +x /usr/local/bin/cherry_pick_pull.sh
```

(You may also have `~/bin` on your `PATH`, which is probably a better location than `/usr/local/bin`)

Then you can propose a cherry pick of PR using something like:

```
UPSTREAM_REMOTE=origin \
FORK_REMOTE=${USER} \
GITHUB_USER=${USER} \
cherry_pick_pull.sh origin/release-1.15 12345
```

Tip: If you find yourself doing this often, you may want to create a wrapper script
that sets `UPSTREAM_REMOTE`, `FORK_REMOTE` and `GITHUB_USER` to your values.
