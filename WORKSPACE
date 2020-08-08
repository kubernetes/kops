load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

#=============================================================================
# Go rules

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "d9d71a5fdfcf5f5326f1ffc4bcaea6519cb4fcfe0aaee6ae68c7440ee8b46bc8",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.22.7/rules_go-v0.22.7.tar.gz",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.22.7/rules_go-v0.22.7.tar.gz",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "cdb02a887a7187ea4d5a27452311a75ed8637379a1287d8eeb952138ea485f7d",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.21.1/bazel-gazelle-v0.21.1.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies", "go_download_sdk")

go_rules_dependencies()

go_download_sdk(
    name = "go_sdk",
    sdks = {
        "darwin_amd64": ("go1.15rc2.darwin-amd64.tar.gz", "b07775d30e023c1570b1ba74892fc792834436c790fbb0dbb19ebaae9c155105"),
        "linux_amd64": ("go1.15rc2.linux-amd64.tar.gz", "f41a08f630f018bc5d9fd100bd9899516e4965356c78165157eb0eda9a17ac09"),
    },
)

go_register_toolchains()

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies", "go_repository")

gazelle_dependencies()

#=============================================================================
# Docker rules

http_archive(
    name = "io_bazel_rules_docker",
    sha256 = "6287241e033d247e9da5ff705dd6ef526bac39ae82f3d17de1b69f8cb313f9cd",
    strip_prefix = "rules_docker-0.14.3",
    urls = ["https://github.com/bazelbuild/rules_docker/releases/download/v0.14.3/rules_docker-v0.14.3.tar.gz"],
)

load(
    "@io_bazel_rules_docker//repositories:repositories.bzl",
    container_repositories = "repositories",
)

container_repositories()

load("@io_bazel_rules_docker//repositories:deps.bzl", container_deps = "deps")

container_deps()

load(
    "@io_bazel_rules_docker//repositories:go_repositories.bzl",
    docker_go_deps = "go_deps",
)

docker_go_deps()

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
)

container_pull(
    name = "debian_hyperkube_base_amd64",
    # 'tag' is also supported, but digest is encouraged for reproducibility.
    digest = "sha256:5d4ea2fb5fbe9a9a9da74f67cf2faefc881968bc39f2ac5d62d9167e575812a1",
    registry = "k8s.gcr.io",
    repository = "debian-hyperkube-base-amd64",
    tag = "0.12.1",
)

git_repository(
    name = "distroless",
    commit = "f905a6636c5106c36cc979bdcc19f0fe4fc01ede",
    remote = "https://github.com/googlecloudplatform/distroless.git",
    #shallow_since = "1570036739 -0700",
)

load(
    "@distroless//package_manager:package_manager.bzl",
    "package_manager_repositories",
)

package_manager_repositories()

load(
    "@distroless//package_manager:dpkg.bzl",
    "dpkg_list",
    "dpkg_src",
)

dpkg_src(
    name = "debian_stretch",
    arch = "amd64",
    distro = "stretch",
    sha256 = "da378b113f0b1edcf5b1f2c3074fd5476c7fd6e6df3752f824aad22e7547e699",
    snapshot = "20190520T104418Z",
    url = "http://snapshot.debian.org/archive",
)

dpkg_list(
    name = "package_bundle",
    packages = [
        "cgmanager",
        "dbus",
        "libapparmor1",
        "libcgmanager0",
        "libcryptsetup4",
        "libdbus-1-3",
        "libnih-dbus1",
        "libnih1",
        "libpam-systemd",
        "libprocps6",
        "libseccomp2",
        "procps",
        "systemd",
        "systemd-shim",
    ],
    sources = [
        "@debian_stretch//file:Packages.json",
    ],
)

# TODO(fejta): use load.bzl, repos.bzl from repo-infra
git_repository(
    name = "io_k8s_repo_infra",
    commit = "db6ceb5f992254db76af7c25db2edc5469b5ea82",
    remote = "https://github.com/kubernetes/repo-infra.git",
    shallow_since = "1570128715 -0700",
)

http_archive(
    name = "bazel_toolchains",
    sha256 = "1342f84d4324987f63307eb6a5aac2dff6d27967860a129f5cd40f8f9b6fd7dd",
    strip_prefix = "bazel-toolchains-2.2.0",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-toolchains/archive/2.2.0.tar.gz",
        "https://github.com/bazelbuild/bazel-toolchains/archive/2.2.0.tar.gz",
    ],
)

load("@bazel_toolchains//rules:rbe_repo.bzl", "rbe_autoconfig")

rbe_autoconfig(name = "rbe_default")

go_repository(
    name = "com_github_google_go_containerregistry",
    importpath = "github.com/google/go-containerregistry",
    sum = "h1:PTrxTL8TNRbZts4KqdJMsqRlrdjoiKFDq6MVitj8mPk=",
    version = "v0.0.0-20190829181151-21b2e01cec04",
)

# Start using distroless base
container_pull(
    name = "distroless_base",
    digest = "sha256:7fa7445dfbebae4f4b7ab0e6ef99276e96075ae42584af6286ba080750d6dfe5",
    registry = "gcr.io",
    repository = "distroless/base",
)

container_pull(
    name = "distroless_base_debug",
    digest = "sha256:6f78124292427599fcef84139cdc9f4ab2d1851fe129b140c92b997f8fe4d289",
    registry = "gcr.io",
    repository = "distroless/base",
)
