load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
load("@bazel_tools//tools/build_defs/repo:git.bzl", "git_repository")

#=============================================================================
# Go rules

http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.15.8/rules_go-0.15.8.tar.gz",
    sha256 = "ca79fed5b24dcc0696e1651ecdd916f7a11111283ba46ea07633a53d8e1f5199",
)

http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.15.0/bazel-gazelle-0.15.0.tar.gz",
    sha256 = "6e875ab4b6bf64a38c352887760f21203ab054676d9c1b274963907e0768740d",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains")

go_rules_dependencies()

go_register_toolchains(
    go_version = "1.10.5",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

#=============================================================================
# Docker rules

git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.5.1",
)

load(
    "@io_bazel_rules_docker//container:container.bzl",
    "container_pull",
    container_repositories = "repositories",
)

container_repositories()

container_pull(
    name = "debian_hyperkube_base_amd64",
    # 'tag' is also supported, but digest is encouraged for reproducibility.
    digest = "sha256:cc782ed16599000ca4c85d47ec6264753747ae1e77520894dca84b104a7621e2",
    registry = "k8s.gcr.io",
    repository = "debian-hyperkube-base-amd64",
    tag = "0.10",
)

git_repository(
    name = "distroless",
    remote = "https://github.com/googlecloudplatform/distroless.git",
    commit = "886114394dfed219001ec3b068b139a3456e49d4",
)

load(
    "@distroless//package_manager:package_manager.bzl",
    "package_manager_repositories",
    "dpkg_src",
    "dpkg_list",
)

package_manager_repositories()

dpkg_src(
    name = "debian_stretch",
    arch = "amd64",
    distro = "stretch",
    sha256 = "4cb2fac3e32292613b92d3162e99eb8a1ed7ce47d1b142852b0de3092b25910c",
    snapshot = "20180406T154421Z",
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
        "systemd-shim",
        "systemd",
    ],
    sources = [
        "@debian_stretch//file:Packages.json",
    ],
)
