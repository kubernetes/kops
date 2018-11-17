NodeUp Tasks
============

Within a model, we recognize a few well-known task names:

* files
* packages
* services
* options

When a directory is found with one of these well-known names, the items in the subtree build tasks of the corresponding
types.

(TODO: Should we just prefer extensions everywhere?)

Directories which start with an underscore are tags: we only descend into those directories if the relevant tag is present.

All other directory names can be used for organization.

Alongside each task file, a file with the same name with a .meta extension will be recognized as well.  It contains
additional JSON options to parameterize the task.  This is useful for files or templates, which otherwise have
no place to put metadata.

files
=====

The contents of the filesystem tree will be created, mirroring what exists under the files directory.

Directories will be created as needed.  Created directories will be set to mode 0755.

Files will be created 0644 (change with meta 'fileMode')

Owner & group will be root:root

Two special extensions are recognized:

* .asset will be sourced from assets.  Assets are binaries that are made available to the installer, e.g. from a .tar.gz distributions
* .template is a go template

packages
========

Any files found will be considered packages.

The name of the file will be the package to be installed.

services
========

Any files found will be considered services.

The name of the file will be the service to be managed.

By default, the service will be restarted and set to auto-start on boot.


## Order of operations

Logically, all operations are collected before any are performed, according to the tags.

Then operations are performed in the following order:

options
packages
files
sysctls
services

Ties are broken as follows

* A task that required more tags is run after a task that required fewer tags
* Sorted by name
* Custom packages (install a deb) are run after OS provided packages
