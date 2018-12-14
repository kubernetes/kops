Architecture of Gazelle
=======================

.. All external links are here.

.. Godoc links
.. _buildifier build: https://godoc.org/github.com/bazelbuild/buildtools/build
.. _config: https://godoc.org/github.com/bazelbuild/bazel-gazelle/internal/config
.. _go/build: https://godoc.org/go/build
.. _go/parser: https://godoc.org/go/parser
.. _merger: https://godoc.org/github.com/bazelbuild/bazel-gazelle/internal/merger
.. _packages: https://godoc.org/github.com/bazelbuild/bazel-gazelle/internal/packages
.. _resolve: https://godoc.org/github.com/bazelbuild/bazel-gazelle/internal/resolve
.. _rules: https://godoc.org/github.com/bazelbuild/bazel-gazelle/internal/rules
.. _CallExpr: https://godoc.org/github.com/bazelbuild/buildtools/build#CallExpr
.. _golang.org/x/tools/go/vcs: https://godoc.org/golang.org/x/tools/go/vcs

.. Other documentation links
.. _buildifier: https://github.com/bazelbuild/buildtools/tree/master/buildifier
.. _config_setting: https://docs.bazel.build/versions/master/be/general.html#config_setting
.. _Fix command transformations: README.rst#fix-command-transformations
.. _full list of directives: README.rst#Directives
.. _select: https://docs.bazel.build/versions/master/skylark/lib/globals.html#select

.. Issues
.. _#5: https://github.com/bazelbuild/bazel-gazelle/issues/5
.. _#7: https://github.com/bazelbuild/bazel-gazelle/issues/7

.. Actual content is below

Gazelle is a tool that generates and updates Bazel build files for Go projects
that follow the conventional "go build" project layout. It is intended to
simplify the maintenance of Bazel Go projects as much as possible.

This document describes how Gazelle works. It should help users understand why
Gazelle behaves as it does, and it should help developers understand
how to modify Gazelle and how to write similar tools.

.. contents::

Overview
--------

Gazelle generates and updates build files according the algorithm outlined
below. Each of the steps here is described in more detail in the sections below.

* Build a configuration from command line arguments and special comments
  in the top-level build file. See Configuration_.

* For each directory in the repository:

  * Read the build file if one is present.

  * If the build file should be updated (based on configuration):

    * Apply transformations to the build file to migrate away from deprecated
      APIs. See `Fixing build files`_.

    * Scan the source files and collect metadata needed to generate rules
      for the directory. See `Scanning source files`_.

    * Generate new rules from the build metadata collected earlier. See
      `Generating rules`_.

    * Merge the new rules into the directory's build file. Delete any rules
      which are now empty. See `Merging and deleting rules`_.

  * Add the library rules in the directory's build file to a global table,
    indexed by import path.

* For each updated build file:

  * Use the library table to map import paths to Bazel labels for rules that 
    were added or merged earlier. See `Resolving dependencies`_.

  * Merge the resolved rules back into the file.

  * Format the file using buildifier_ and emit it according to the output mode:
    write to disk, print the whole file, or print the diff.

Configuration
-------------

Godoc: config_

Gazelle stores configuration information in ``Config`` objects. These objects
contain settings that affect the behavior of most packages in the program.
For example:

* The list of directories that Gazelle should update.
* The path of the repository root directory. Bazel package names are based
  on paths relative to this location.
* The current import path prefix and the directory where it was set.
  Gazelle uses this to infer import paths for ``go_library`` rules.
* A list of build tags that Gazelle considers to be true on all platforms.

``Config`` objects apply to individual directories. Each directory inherits
the ``Config`` from its parent. Values in a ``Config`` may be modified within
a directory using *directives* written in the directory's build file. A
directive is a special comment formatted like this:

::

  # gazelle:key value

Here are a few examples. See the `full list of directives`_.

* ``# gazelle:prefix`` - sets the Go import path prefix for the current
  directory.
* ``# gazelle:build_tags`` - sets the list of build tags which Gazelle considers
  to be true on all platforms.

There are a few directives which are not applied to the ``Config`` object but
are interpreted directly in packages where they are relevant.

* ``# gazelle:ignore`` - the build file should not be updated by Gazelle.
  Gazelle may still index its contents so it can resolve dependencies in other
  build files.
* ``# gazelle:exclude path/to/file`` - the named file should not be read by
  Gazelle and should not be included in ``srcs`` lists. If this refers to
  a directory, Gazelle won't recurse into the directory. This directive may
  appear multiple times.

Fixing build files
------------------

Godoc: merger_

From time to time, APIs in rules_go are changed or updated. Gazelle helps
users stay up to date with these changes by automatically fixing deprecated
usage.

Minor fixes are applied by Gazelle automatically every time it runs. However,
some fixes may delete or rename existing rules. Users must run ``gazelle fix``
to apply these fixes. By default, Gazelle will only *warn* users that
``gazelle fix`` should be run.

Here are a few of the fixes Gazelle performs. See `Fix command transformations`_
for a full list.

* **Squash cgo libraries:** Gazelle will remove ``cgo_library`` rules and
  merge their attributes into ``go_library`` rules that reference them.
  This is a major fix and is only applied with ``gazelle fix``.
* **Migrate library attributes:** Gazelle replaces ``library`` attributes
  with ``embed`` attributes. The only difference between these is that
  ``library`` (which is now deprecated) accepts a single label, while ``embed``
  accepts a list. This is a minor fix and is always applied.

Users can prevent Gazelle from modifying rules, attributes, or individual
values by writing ``# keep`` comments above them.

Scanning source files
---------------------

Godoc: packages_

Nearly all of the information needed to build a program with the standard Go SDK
is implied by directory structure, file names, and file contents. This is why
``go build`` doesn't require any sort of build file. The `go/build`_ package in
the standard library collects this information.

Unfortunately, `go/build`_ can only collect information for one platform at
a time. Gazelle needs to generate build files that work on all platforms, so
we have our own implementation of this logic.

Information extracted from files
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Gazelle extracts build metadata from source files and contents in much the
same way that the standard `go/build`_ package does. It gets the following
information from file names:

* File extension (e.g., .go, .c, .proto). Normally, only .go, .s, and .h files
  are included in Go rules. If any cgo code is present, then C/C++ files are
  also included. .proto files are also used to build proto rules. Other files
  (e.g., .txt) are ignored.
* Test suffix. For example, if a file is named ``foo_test.go``, it will be
  included in a test target instead of a library or binary target.
* OS and architecture suffixes. For example, a file named ``foo_linux_amd64.go``
  will be listed in the ``linux_amd64`` section of the target it belongs to.

Gazelle gets the following information from file contents:

* Package name. This is syntactically the first part of every .go file. All
  files in the same directory must have the same package name (except for
  external test sources, which have a package name ending with ``_test``). If
  there are multiple packages, Gazelle will choose one that matches the
  directory name (if present) or report an error.
* Imported libraries. Go import paths are usually URLs. Imports in
  platform-specific source files are also platform-specific.
* Build tags. The Go toolchain recognizes comments beginning with ``// +build``
  before the package declaration. These tags tell the build system that a file
  should only be built for specific platforms. See `this article 
  <https://dave.cheney.net/2013/10/12/how-to-use-conditional-compilation-with-the-go-build-tool>`_
  for more information.
* Whether cgo code is present. This affects how packages are built and
  whether C/C++ files are included.
* C/C++ compile and link options (specified in ``#cgo`` directives in cgo
  comments). These may be platform-specific.

In most cases, only the top of the file is parsed. For Go files, we use the
standard `go/parser`_ package. For proto files, we use regular expressions that
match ``package``, ``go_package``, and ``import`` statements.

The ``Package`` object
~~~~~~~~~~~~~~~~~~~~~~

Gazelle stores build metadata in a ``Package`` object. Currently, we only
support one ``Package`` per directory (which is also what the Go SDK supports),
but this will be expanded in the future. ``Package`` objects contain some
top-level metadata (like the package name and directory path), along with
several target objects (``GoTarget`` and ``ProtoTarget``).

Target objects correspond directly to rules that will be generated later. They
store lists of sources, imports, and flags in ``PlatformStrings`` objects.

``PlatformStrings`` objects store strings in four sections: a generic list, an
OS-specific dictionary, an architecture-specific dictionary, and an
OS-and-architecture-specific dictionary. The keys in the dictionaries are OS
names, architecture names, or OS-and-architecture pairs; the values are lists of
strings. The same string may not appear more than once in a list and may not
appear in more than one section. This is due to a Bazel requirement: the same
label may not appear more than once in a ``deps`` list.

Generating rules
----------------

Godoc: rules_

Once build metadata has been extracted from the sources in a directory,
Gazelle generates rules for building those sources.

Generated rules are formatted as CallExpr_ objects. CallExpr_ is defined in the
`buildifier build`_ library. This is the same library used to parse and format
build files. This lets us manipulate newly generated rules and existing rules
with the same code.

We may generate the following rules:

* ``proto_library`` and ``go_proto_library`` are generated if there was at
  least one .proto source file.
* ``go_library`` is generated if there was at least one non-test source. This
  may embed the ``go_proto_library`` if there was one.
* ``go_test`` rules are generated for internal and external tests. Internal
  tests embed the ``go_library`` while external tests depend on the
  ``go_library`` as a separate package.
* ``go_binary`` is generated if the package name was ``main``. It embeds the
  ``go_library``.

Rules are named according to a pluggable naming policy, but there is currently
only one policy: libraries are named ``go_default_library``, tests are
named ``go_default_test``, and binaries are named after the directory. The
``go_default_library`` name is an historical artifact from before we had
index-based dependency resolution. We'll need to move away from this naming
scheme in the future (`#5`_) before we support multiple packages (`#7`_).

Sources, imports, and flags within each target are converted to expressions in a
straightforward fashion. The lists within ``PlatformStrings`` are converted to
list expressions. Dictionaries are converted to calls to `select`_ expressions
(when Bazel evaluates a `select`_ expression, it will choose one of several
provided lists, based on `config_setting`_ rules). Lists and select expressions
may be added together. For example:

.. code:: bzl

  go_library(
      name = "go_default_library",
      srcs = [
          "terminal.go",
      ] + select({
          "@io_bazel_rules_go//go/platform:darwin": [
              "util.go",
              "util_bsd.go",
          ],
          "@io_bazel_rules_go//go/platform:linux": [
              "util.go",
              "util_linux.go",
          ],
          "@io_bazel_rules_go//go/platform:windows": [
              "util_windows.go",
          ],
          "//conditions:default": [],
      }),
      ...
  )

At this point, Gazelle does not have enough information to generate expressions
``deps`` attributes. We only have a list of import strings extracted from source
files. These imports are stored temporarily in a special ``_gazelle_imports``
attribute in each rule. Later, the imports are converted to Bazel labels (see
`Resolving dependencies`_), and this attribute is replaced with ``deps``.

Merging and deleting rules
--------------------------

Godoc: merger_

Merging is the process of combining generated rules with the corresponding
rules in an existing build file. If no build file exists in a directory, a
new file is created with generated rules, and no merging is performed.

Merging occurs in two phases: pre-resolve, and post-resolve. This is due to an
interdependence with dependency resolution. Dependency resolution uses a table
of *merged* library rules, so it can't be performed until the pre-resolve merge
has occurred. After dependency resolution, we need to merge newly generated
``deps`` attributes; this is done in the post-resolve merge. The two phases use
the same algorithm.

During the merge process, Gazelle attempts to match generated rules with
existing rules that have the same name and same kind. Rules are only merged if
both name and kind match. If an existing rule has the same name as a generated
rule but a different kind, the generated rule will not be merged.  If no
existing rule matches a generated rule, the generated rule is simply appended to
the end of the file. Existing rules that don't match any generated rule are not
modified.

When Gazelle identifies a matching pair of rules, it combines each attribute
according to the algorithm below. If an attribute is present in the generated
rule but not in the existing rule, it is copied to the merged rule verbatim. If
an attribute is present in the existing rule but not the generated rule, Gazelle
behaves as if the generated attribute were present but empty.

* For each value in the existing rule's attribute:

  * If the value also appears in the generated rule's attribute or is marked
    with a ``# keep`` comment, preserve it. Otherwise, delete it.

* For each value in the generated rule's attribute:

  * If the value appears in the generated rule's attribute, ignore it.
    Otherwise, add it to the merged rule.

* If the merged attribute is empty, delete it.

When a value is present in both the existing and generated attributes, we use
the existing value instead of the generated value, since this preserves
comments.

Some attributes are considered *unmergeable*, for example, ``visibility`` and
``gc_goopts``. Gazelle may add these attributes to existing rules if they are
not already present, but existing values won't be modified or deleted.

Preserving customizations
~~~~~~~~~~~~~~~~~~~~~~~~~

Gazelle has several mechanisms for preserving manual modifications to build
files. Some of these mechanisms work automatically; others require explicit
comments.

* Gazelle will not modify or delete rules that don't appear to have been
  generated by Gazelle.
* As mentioned above, some attributes are considered unmergeable. Gazelle may
  set initial values for these but won't delete or replace existing values.
* ``# keep`` comments may be attached to any rule, attribute, or value
  to prevent Gazelle from modifying it.
* ``# gazelle:exclude <file>`` directives can be used to prevent Gazelle from
  adding files to source lists (for example, checked-in .pb.go files). They
  can also prevent Gazelle from recursing into directories that contain
  unbuildable code (e.g., ``testdata``).
* ``# gazelle:ignore`` directives prevent Gazelle from making any modifications
  to build files that contain them.

Deleting rules
~~~~~~~~~~~~~~

Deletion is a special case of the merging algorithm.

When Gazelle generates rules for a package (see `Generating rules`_), it
actually produces two lists of rules: a list of rules for buildable targets,
and a list of empty rules that may be deleted. The empty rules have no
attributes other than ``name``.

The empty rules are merged using the same algorithm as the other generated
rules. If, after merging, an empty rule has no attributes that would make the
rule buildable (for example, ``srcs``, or ``deps``), the rule will be deleted.

Resolving dependencies
----------------------

Godoc: resolve_

When Gazelle generates rules for a package (see `Generating
rules`_), it stores names of the libraries imported by each rule in a special
``_gazelle_imports`` attribute. During dependency resolution, Gazelle maps these
imports to Bazel labels and replaces ``_gazelle_imports`` with ``deps``.

Before dependency resolution starts, Gazelle builds a table of all known
libraries. This includes ``go_library``, ``go_proto_library``, and
``proto_library`` rules. The table is populated by scanning build files after
the pre-resolve merge, so existing and newly generated rules are included
in the table, and deleted rules are excluded. Once all library rules have been
added, Gazelle indexes the table by language-specific import path.

Gazelle resolves each import string in ``_gazelle_imports`` as follows:

* If the import is part of the standard library, it is dropped. Standard
  library dependencies are implicit.

* If the import is provided by exactly one rule in the library table, the label
  for that rule is used.

* If the import is provided by multiple libraries, we attempt to resolve
  the ambiguity.

  * For Go, we apply the vendoring algorithm. Vendored libraries aren't visible
    outside of the vendor directory's parent.

  * Go libraries that are embedded by other Go libraries are not considered.
    Embedded libraries may be incomplete.

  * When an ambiguity can't be resolved, Gazelle logs an error and skips
    the dependency.

* If the import is not provided by any rule in the import table, we attempt
  to resolve the dependency using heuristics:

  * If the import path starts with the current prefix (set with a 
    ``# gazelle:prefix`` directive or on the command line), we construct a label
    by concatenating the prefix directory and the portion of the import path
    below the prefix into a package name.

  * Otherwise, the import path is considered external and is resolved
    according to the external mode set on the command line.

    * In ``external`` mode, Gazelle determines the portion of the import path
      that corresponds to a repository using `golang.org/x/tools/go/vcs`_. This
      part of the path is converted into a repository name (for example,
      ``@org_golang_x_tools``), and the rest is converted to a package name.

    * In ``vendored`` mode, Gazelle constructs a label by prepending ``vendor/``
      to the import path.

Note that ``visibility`` attributes are not considered when resolving imports.
This was part of an initial prototype, but it was confusing in many situations.

Building and running Gazelle
----------------------------

Gazelle is a regular Go program. It can be built, installed, and run without
Bazel, using the regular Go SDK.

.. code:: bash

  $ go get -u github.com/bazelbuild/bazel-gazelle/cmd/gazelle
  $ gazelle -go_prefix example.com/project

We lightly discourage this method of running Gazelle. All developers on a
project should use the same version of Gazelle to ensure the build files
they generate are consistent. The easiest way to accomplish this is to build
and run Gazelle through Bazel. Gazelle may added to a WORKSPACE file, 
built as a normal ``go_binary``, then installed or run from the ``bazel-bin/``
directory.

.. code:: bash

  $ bazel build @bazel_gazelle//cmd/gazelle
  $ bazel-bin/external/bazel_gazelle/cmd/gazelle/gazelle -go_prefix example.com/project

It's usually better to invoke Gazelle through a wrapper script though. This
saves typing and ensures Gazelle is run with a consistent set of arguments.
We provide a Bazel rule that generates such a wrapper script. Developers may
add a snippet like the one below to a build file:

.. code:: bzl

  load("@bazel_gazelle//:def.bzl", "gazelle")

  gazelle(
      name = "gazelle",
      command = "fix",
      external = "vendored",
      prefix = "example.com/project",
  )

This script may be built and executed in a single command with ``bazel run``.

.. code:: bash

  $ bazel run //:gazelle

This is the most convenient way to run Gazelle, and it's what we recommend to
users. However, there are two issues with running Gazelle in this
fashion. First, binaries executed by ``bazel run`` are run in the Bazel
execroot, not the user's current directory. The wrapper script uses a hack
(dereferencing symlinks) to jump to the top of the workspace source tree before
running Gazelle. Second, ``bazel run`` holds a lock on the Bazel output
directory. This means Gazelle cannot invoke Bazel without deadlocking. Commands
like ``bazel query`` would be helpful for detecting generated code, but it's not
safe to use them.

To avoid these limitations, the wrapper script may be copied to the workspace
and optionally checked into version control. When the wrapper script is run
directly (without ``bazel run``), it will rebuild itself to ensure no changes
are needed. If the rebuilt script differs from the running script, it will
prompt the user to copy the rebuilt script into the workspace again.

.. code:: bash

  $ bazel build //:gazelle
  Target //:gazelle up-to-date:
    bazel-bin/gazelle.bash
  ____Elapsed time: 1.326s, Critical Path: 0.00s
  $ cp bazel-bin/gazelle.bash gazelle.bash
  $ ./gazelle.bash

Dependencies
------------

Gazelle has the following dependencies:

github.com/bazelbuild/bazel-skylib
  Skylark utility used to generate wrapper script in the ``gazelle`` rule.
github.com/bazelbuild/buildtools/build
  Used to parse and rewrite build files.
github.com/bazelbuild/rules_go
  Used to build and test Gazelle through Bazel. Gazelle can aslo be built on its
  own with the Go SDK.
github.com/pelletier/go-toml
  Used to import dependencies from dep Gopkg.lock files.
golang.org/x/tools/vcs
  Used during dependency resolution to determine the repository prefix for a
  given import path. This uses the network.
