# Go dependency management in kops

There is (currently) no perfect solution for dependency management in go; an "official"
solution is planned, but in the meantime we use a solution based on git submodules.

The biggest problem is the double-import problem, where a vendored dependency
has a `vendor` directory.  There is no way to "flatten" the imports.

See for example the discussion [here](https://github.com/dpw/vendetta/issues/13)

## The current solution

* We use git submodules to specify our dependencies.  This allows simple (if tedious) direct specification of dependencies without extra tooling.
* We want to ignore the `vendor` directories in our dependencies though.  There is no way to do a "filtered" git submodule.
* We therefore put the git submodules them into `_vendor`, as this is ignored by go for historical reasons.
* We then rsync the subset of files we want into `vendor` (via `make copydeps`)
* We commit the contents of the `vendor` directory to git

## Shortcomings

* We have to manually manage our dependencies (this is arguably also an advantage, in the absence of any real rules to resolve conflicts)
* `go get` will fetch the submodules, so we pull a lot more data than we need to
