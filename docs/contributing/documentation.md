# Documentation guidelines

## CLI commands

`kops` uses cobra for its CLI implementation. Each command should have the following help fields defined where possible:

* `Short`: A short statement describing the command in the third person present, starting with a capital letter and ending without a period.
  * Example: "Edits a cluster"

* `Long`: An expanded description and usage of the command in the third person present tense, starting with a capital letter and ending with a period. The text from the `Short` field should be the first sentence in the `Long` field.
Example:
```
Edits a cluster.

This command changes the cloud specification in the registry.

It does not update the cloud resources, to apply the changes use "kops update cluster".
```

* `Example`: Example(s) of how to use the command. This field is formatted as a code snippet in the docs, so make sure if you have comments that these are written as a bash comment (e.g. `# this is a comment`).

## mdBook

`make live-docs` runs a local mdBook server on port 3000 for previewing docs.

`make build-docs` builds the static site into `site/`.

The site navigation is hand-maintained in [docs/SUMMARY.md](https://github.com/kubernetes/kops/tree/master/docs/SUMMARY.md);
new pages must be linked there to be rendered.