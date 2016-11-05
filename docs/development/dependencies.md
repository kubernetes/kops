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

## Adding a dependency

1. Add dependency as a submodule in the `_vendor` directory.
2. Make sure you have all the git submodules populated. If not you can run `git submodule init` and `git submodule update`.
3. Run `make copydeps`.
4. The make command should move the new dependency (and all other current dependencies) into the `vendor` folder, ignoring each submodules own `vendor` directory.
5. Commit your changes.

Here is an example of us adding the `go-md2man` package and it's required dependencies:

```
git submodule add https://github.com/cpuguy83/go-md2man.git _vendor/github.com/cpuguy83/go-md2man
git submodule add https://github.com/russross/blackfriday.git _vendor/github.com/russross/blackfriday
git submodule add https://github.com/shurcooL/sanitized_anchor_name.git _vendor/github.com/shurcooL/sanitized_anchor_name
git submodule init
git submodule update
make copydeps
git add .gitmodules
git add _vendor/github.com/cpuguy83/go-md2man
git add _vendor/github.com/russross/blackfriday
git add _vendor/github.com/shurcooL/sanitized_anchor_name
git add vendor/github.com/cpuguy83/go-md2man
git add vendor/github.com/russross/blackfriday
git add vendor/github.com/shurcooL/sanitized_anchor_name
git commit -m "Add go-md2man, blackfriday and sanitized_anchor_name deps"
```

## Updating a dependency

```
pushd _vendor/github.com/aws/aws-sdk-go
git fetch
git checkout v1.5.2
popd
make copydeps
git add _vendor/github.com/aws/aws-sdk-go/
git add vendor/github.com/aws/aws-sdk-go/
git commit -m "Update aws-sdk-go to 1.5.2"
```