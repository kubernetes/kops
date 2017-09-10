# File Path Functions

While Sprig does not grant access to the filesystem, it does provide functions
for working with strings that follow file path conventions.

# base

Return the last element of a path.

```
base "foo/bar/baz"
```

The above prints "baz"

# dir

Return the directory, stripping the last part of the path. So `dir "foo/bar/baz"`
returns `foo/bar`

# clean

Clean up a path.

```
clean "foo/bar/../baz"
```

The above resolves the `..` and returns `foo/baz`

# ext

Return the file extension.

```
ext "foo.bar"
```

The above returns `.bar`.

# isAbs

To check whether a file path is absolute, use `isAbs`.
