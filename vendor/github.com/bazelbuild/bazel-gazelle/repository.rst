Repository rules
================

.. _http_archive.strip_prefix: https://docs.bazel.build/versions/master/be/workspace.html#http_archive.strip_prefix
.. _native git_repository rule: https://docs.bazel.build/versions/master/be/workspace.html#git_repository
.. _native http_archive rule: https://docs.bazel.build/versions/master/be/workspace.html#http_archive
.. _manifest.bzl: third_party/manifest.bzl

.. role:: param(kbd)
.. role:: type(emphasis)
.. role:: value(code)
.. |mandatory| replace:: **mandatory value**

Repository rules are Bazel rules that can be used in WORKSPACE files to import
projects in external repositories. Repository rules may download projects
and transform them by applying patches or generating build files.

The Gazelle repository provides three rules:

* `go_repository`_ downloads a Go project over HTTP or using a version control
  tool like git. It understands Go import path redirection. If build files are
  not already present, it can generate them with Gazelle.
* `git_repository`_ downloads a project with git. Unlike the native
  ``git_repository``, this rule allows you to specify an "overlay": a set of
  files to be copied into the downloaded project. This may be used to add
  pre-generated build files to a project that doesn't have them.
* `http_archive`_ downloads a project via HTTP. It also lets you specify
  overlay files.

Repository rules can be loaded and used in WORKSPACE like this:

.. code:: bzl

  load("@bazel_gazelle//:deps.bzl", "go_repository")

  go_repository(
      name = "com_github_pkg_errors",
      commit = "816c9085562cd7ee03e7f8188a1cfd942858cded",
      importpath = "github.com/pkg/errors",
  )

Gazelle can add and update some of these rules automatically using the
``update-repos`` command. For example, the rule above can be added with:

.. code::

  $ gazelle update-repos github.com/pkg/errors

go_repository
-------------

``go_repository`` downloads a Go project and generates build files with Gazelle
if they are not already present. This is the simplest way to depend on
external Go projects.

**Example**

.. code:: bzl

  load("@bazel_gazelle//:deps.bzl", "go_repository")

  # Download automatically via git
  go_repository(
      name = "com_github_pkg_errors",
      commit = "816c9085562cd7ee03e7f8188a1cfd942858cded",
      importpath = "github.com/pkg/errors",
  )

  # Download from git fork
  go_repository(
      name = "com_github_pkg_errors",
      commit = "816c9085562cd7ee03e7f8188a1cfd942858cded",
      importpath = "github.com/pkg/errors",
      remote = "https://example.com/fork/github.com/pkg/errors",
      vcs = "git",
  )

  # Download via HTTP
  go_repository(
      name = "com_github_pkg_errors",
      importpath = "github.com/pkg/errors",
      urls = ["https://codeload.github.com/pkg/errors/zip/816c9085562cd7ee03e7f8188a1cfd942858cded"],
      strip_prefix = ["errors-816c9085562cd7ee03e7f8188a1cfd942858cded"],
      type = "zip",
  )

**Attributes**

+--------------------------------+----------------------+-------------------------------------------------+
| **Name**                       | **Type**             | **Default value**                               |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`name`                  | :type:`string`       | |mandatory|                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A unique name for this rule. This should usually be the Java-package-style                              |
| name of the URL, with underscores as separators, for example,                                           |
| ``com_github_example_project``.                                                                         |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`importpath`            | :type:`string`       | |mandatory|                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| The Go import path that matches the root directory of this repository. If                               |
| neither ``urls`` nor ``remote`` are specified, ``go_repository`` will download                          |
| the repository from this location. This supports import path redirection.                               |
| If build files are generated, libraries will have ``importpath`` prefixed                               |
| with this string.                                                                                       |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`commit`                | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| If the repository is downloaded using a version control tool, this is the                               |
| commit or revision to check out. With git, this would be a sha1 commit id.                              |
| ``commit`` and ``tag`` may not both be set.                                                             |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`tag`                   | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| If the repository is downloaded using a version control tool, this is the                               |
| named revision to check out. ``commit`` and ``tag`` may not both be set.                                |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`vcs`                   | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| One of ``"git"``, ``"hg"``, ``"svn"``, ``"bzr"``.                                                       |
|                                                                                                         |
| The version control system to use. This is usually determined automatically,                            |
| but it may be necessary to set this when ``remote`` is set and the VCS cannot                           |
| be inferred. You must have the corresponding tool installed on your host.                               |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`remote`                | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| The VCS location where the repository should be downloaded from. This is                                |
| usually inferred from ``importpath``, but you can set ``remote`` to download                            |
| from a private repository or a fork.                                                                    |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`urls`                  | :type:`string list`  | :value:`[]`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A list of HTTP(S) URLs where an archive containing the project can be                                   |
| downloaded. Bazel will attempt to download from the first URL; the others                               |
| are mirrors.                                                                                            |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`strip_prefix`          | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| If the repository is downloaded via HTTP (``urls`` is set), this is a                                   |
| directory prefix to strip. See `http_archive.strip_prefix`_.                                            |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`type`                  | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| One of ``"zip"``, ``"tar.gz"``, ``"tgz"``, ``"tar.bz2"``, ``"tar.xz"``.                                 |
|                                                                                                         |
| If the repository is downloaded via HTTP (``urls`` is set), this is the                                 |
| file format of the repository archive. This is normally inferred from the                               |
| downloaded file name.                                                                                   |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`sha256`                | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| If the repository is downloaded via HTTP (``urls`` is set), this is the                                 |
| SHA-256 sum of the downloaded archive. When set, Bazel will verify the archive                          |
| against this sum before extracting it.                                                                  |
|                                                                                                         |
| **CAUTION:** Do not use this with services that prepare source archives on                              |
| demand, such as codeload.github.com. Any minor change in the server software                            |
| can cause differences in file order, alignment, and compression that break                              |
| SHA-256 sums.                                                                                           |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`build_file_generation` | :type:`string`       | :value:`"auto"`                                 |
+--------------------------------+----------------------+-------------------------------------------------+
| One of ``"auto"``, ``"on"``, ``"off"``.                                                                 |
|                                                                                                         |
| Whether Gazelle should generate build files in the repository. In ``"auto"``                            |
| mode, Gazelle will run if there is no build file in the repository root                                 |
| directory.                                                                                              |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`build_file_name`       | :type:`string`       | :value:`BUILD.bazel,BUILD`                      |
+--------------------------------+----------------------+-------------------------------------------------+
| Comma-separated list of names Gazelle will consider to be build files.                                  |
| If a repository contains files named ``build`` that aren't related to Bazel,                            |
| it may help to set this to ``"BUILD.bazel"``, especially on case-insensitive                            |
| file systems.                                                                                           |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`build_external`        | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| One of ``"external"``, ``"vendored"``.                                                                  |
|                                                                                                         |
| This sets Gazelle's ``-external`` command line flag.                                                    |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`build_tags`            | :type:`string list`  | :value:`[]`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| This sets Gazelle's ``-build_tags`` command line flag.                                                  |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`build_file_proto_mode` | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| One of ``"default"``, ``"legacy"``, ``"disable"``.                                                      |
|                                                                                                         |
| This sets Gazelle's ``-proto`` command line flag.                                                       |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`build_extra_args`      | :type:`string list`  | :value:`[]`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A list of additional command line arguments to pass to Gazelle when                                     |
| generating build files.                                                                                 |
+--------------------------------+----------------------+-------------------------------------------------+

git_repository
--------------

``git_repository`` downloads a project with git. It has the same features as the
`native git_repository rule`_, but it also allows you to copy a set of files
into the repository after download. This is particularly useful for placing
pre-generated build files.

**Example**

.. code:: bzl

  load("@bazel_gazelle//:deps.bzl", "git_repository")

  git_repository(
      name = "com_github_pkg_errors",
      remote = "https://github.com/pkg/errors",
      commit = "816c9085562cd7ee03e7f8188a1cfd942858cded",
      overlay = {
          "@my_repo//third_party:com_github_pkg_errors/BUILD.bazel.in" : "BUILD.bazel",
      },
  )

**Attributes**

+--------------------------------+----------------------+-------------------------------------------------+
| **Name**                       | **Type**             | **Default value**                               |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`name`                  | :type:`string`       | |mandatory|                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A unique name for this rule. This should usually be the Java-package-style                              |
| name of the URL, with underscores as separators, for example,                                           |
| ``com_github_example_project``.                                                                         |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`remote`                | :type:`string`       | |mandatory|                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| The remote repository to download.                                                                      |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`commit`                | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| The git commit to check out. Either ``commit`` or ``tag`` may be specified.                             |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`tag`                   | :type:`tag`          | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| The git tag to check out. Either ``commit`` or ``tag`` may be specified.                                |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`overlay`               | :type:`dict`         | :value:`{}`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A set of files to copy into the downloaded repository. The keys in this                                 |
| dictionary are Bazel labels that point to the files to copy. These must be                              |
| fully qualified labels (i.e., ``@repo//pkg:name``) because relative labels                              |
| are interpreted in the checked out repository, not the repository containing                            |
| the WORKSPACE file. The values in this dictionary are root-relative paths                               |
| where the overlay files should be written.                                                              |
|                                                                                                         |
| It's convenient to store the overlay dictionaries for all repositories in                               |
| a separate .bzl file. See Gazelle's `manifest.bzl`_ for an example.                                     |
+--------------------------------+----------------------+-------------------------------------------------+

http_archive
------------

``http_archive`` downloads a project over HTTP(S). It has the same features as
the `native http_archive rule`_, but it also allows you to copy a set of files
into the repository after download. This is particularly useful for placing
pre-generated build files.

**Example**

.. code:: bzl

  load("@bazel_gazelle//:deps.bzl", "http_archive")

  http_archive(
      name = "com_github_pkg_errors",
      urls = ["https://codeload.github.com/pkg/errors/zip/816c9085562cd7ee03e7f8188a1cfd942858cded"],
      strip_prefix = "errors-816c9085562cd7ee03e7f8188a1cfd942858cded",
      type = "zip",
      overlay = {
          "@my_repo//third_party:com_github_pkg_errors/BUILD.bazel.in" : "BUILD.bazel",
      },
  )

**Attributes**

+--------------------------------+----------------------+-------------------------------------------------+
| **Name**                       | **Type**             | **Default value**                               |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`name`                  | :type:`string`       | |mandatory|                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A unique name for this rule. This should usually be the Java-package-style                              |
| name of the URL, with underscores as separators, for example,                                           |
| ``com_github_example_project``.                                                                         |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`urls`                  | :type:`string list`  | |mandatory|                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A list of HTTP(S) URLs where the project can be downloaded. Bazel will                                  |
| attempt to download the first URL; the others are mirrors.                                              |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`sha256`                | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| The SHA-256 sum of the downloaded archive. When set, Bazel will verify the                              |
| archive against this sum before extracting it.                                                          |
|                                                                                                         |
| **CAUTION:** Do not use this with services that prepare source archives on                              |
| demand, such as codeload.github.com. Any minor change in the server software                            |
| can cause differences in file order, alignment, and compression that break                              |
| SHA-256 sums.                                                                                           |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`strip_prefix`          | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A directory prefix to strip. See `http_archive.strip_prefix`_.                                          |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`type`                  | :type:`string`       | :value:`""`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| One of ``"zip"``, ``"tar.gz"``, ``"tgz"``, ``"tar.bz2"``, ``"tar.xz"``.                                 |
|                                                                                                         |
| The file format of the repository archive. This is normally inferred from                               |
| the downloaded file name.                                                                               |
+--------------------------------+----------------------+-------------------------------------------------+
| :param:`overlay`               | :type:`dict`         | :value:`{}`                                     |
+--------------------------------+----------------------+-------------------------------------------------+
| A set of files to copy into the downloaded repository. The keys in this                                 |
| dictionary are Bazel labels that point to the files to copy. These must be                              |
| fully qualified labels (i.e., ``@repo//pkg:name``) because relative labels                              |
| are interpreted in the checked out repository, not the repository containing                            |
| the WORKSPACE file. The values in this dictionary are root-relative paths                               |
| where the overlay files should be written.                                                              |
|                                                                                                         |
| It's convenient to store the overlay dictionaries for all repositories in                               |
| a separate .bzl file. See Gazelle's `manifest.bzl`_ for an example.                                     |
+--------------------------------+----------------------+-------------------------------------------------+
