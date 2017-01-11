# Documentation guidelines

## CLI commands

`kops` uses cobra for it's CLI implementation. Each command should have the following help fields defined where possible:

* `Short`: single sentence description of command.
* `Long`: expanded description and usage of the command. The text from the `Short` field should be the first sentence in the `Long` field.
* `Example`: example(s) of how to use the command. This field is formatted as a code snippet in the docs, so make sure if you have comments
that these are written as a bash comment (e.g. `# this is a comment`).
