# Semantic Version Functions

Some version schemes are easily parseable and comparable. Sprig provides functions
for working with [SemVer 2](http://semver.org) versions.

## semver

The `semver` function parses a string into a Semantic Version:

```
$version := semver "1.2.3-alpha.1+123"
```

_If the parser fails, it will cause template execution to halt with an error._

At this point, `$version` is a pointer to a `Version` object with the following
properties:

- `$version.Major`: The major number (`1` above)
- `$version.Minor`: The minor number (`2` above)
- `$version.Patch`: The patch number (`3` above)
- `$version.Prerelease`: The prerelease (`alpha.1` above)
- `$version.Metadata`: The build metadata (`123` above)
- `$version.Original`: The original version as a string

Additionally, you can compare a `Version` to another `version` using the `Compare`
function:

```
semver "1.4.3" | (semver "1.2.3").Compare
```

The above will return `-1`.

The return values are:

- `-1` if the given semver is greater than the semver whose `Compare` method was called
- `1` if the version who's `Compare` function was called is greater.
- `0` if they are the same version

(Note that in SemVer, the `Metadata` field is not compared during version
comparison operations.)


## semverCompare

A more robust comparison function is provided as `semverCompare`. This version
supports version ranges:

- `semverCompare "1.2.3" "1.2.3"` checks for an exact match
- `semverCompare "^1.2.0" "1.2.3"` checks that the major and minor versions match, and that the patch 
  number of the second version is _greater than or equal to_ the first parameter.

The SemVer functions use the [Masterminds semver library](https://github.com/Masterminds/semver),
from the creators of Sprig.


## Basic Comparisons

There are two elements to the comparisons. First, a comparison string is a list
of comma separated and comparisons. These are then separated by || separated or
comparisons. For example, `">= 1.2, < 3.0.0 || >= 4.2.3"` is looking for a
comparison that's greater than or equal to 1.2 and less than 3.0.0 or is
greater than or equal to 4.2.3.

The basic comparisons are:

* `=`: equal (aliased to no operator)
* `!=`: not equal
* `>`: greater than
* `<`: less than
* `>=`: greater than or equal to
* `<=`: less than or equal to

_Note, according to the Semantic Version specification pre-releases may not be
API compliant with their release counterpart. It says,_

> _A pre-release version indicates that the version is unstable and might not satisfy the intended compatibility requirements as denoted by its associated normal version._

_SemVer comparisons without a pre-release value will skip pre-release versions.
For example, `>1.2.3` will skip pre-releases when looking at a list of values
while `>1.2.3-alpha.1` will evaluate pre-releases._

## Hyphen Range Comparisons

There are multiple methods to handle ranges and the first is hyphens ranges.
These look like:

* `1.2 - 1.4.5` which is equivalent to `>= 1.2, <= 1.4.5`
* `2.3.4 - 4.5` which is equivalent to `>= 2.3.4, <= 4.5`

## Wildcards In Comparisons

The `x`, `X`, and `*` characters can be used as a wildcard character. This works
for all comparison operators. When used on the `=` operator it falls
back to the pack level comparison (see tilde below). For example,

* `1.2.x` is equivalent to `>= 1.2.0, < 1.3.0`
* `>= 1.2.x` is equivalent to `>= 1.2.0`
* `<= 2.x` is equivalent to `<= 3`
* `*` is equivalent to `>= 0.0.0`

## Tilde Range Comparisons (Patch)

The tilde (`~`) comparison operator is for patch level ranges when a minor
version is specified and major level changes when the minor number is missing.
For example,

* `~1.2.3` is equivalent to `>= 1.2.3, < 1.3.0`
* `~1` is equivalent to `>= 1, < 2`
* `~2.3` is equivalent to `>= 2.3, < 2.4`
* `~1.2.x` is equivalent to `>= 1.2.0, < 1.3.0`
* `~1.x` is equivalent to `>= 1, < 2`

## Caret Range Comparisons (Major)

The caret (`^`) comparison operator is for major level changes. This is useful
when comparisons of API versions as a major change is API breaking. For example,

* `^1.2.3` is equivalent to `>= 1.2.3, < 2.0.0`
* `^1.2.x` is equivalent to `>= 1.2.0, < 2.0.0`
* `^2.3` is equivalent to `>= 2.3, < 3`
* `^2.x` is equivalent to `>= 2.0.0, < 3`

