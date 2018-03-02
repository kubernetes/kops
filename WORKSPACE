#=============================================================================
# Go rules

http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.10.1/rules_go-0.10.1.tar.gz",
    sha256 = "4b14d8dd31c6dbaf3ff871adcd03f28c3274e42abc855cb8fb4d01233c0154dc",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_download_sdk")

go_rules_dependencies()

go_register_toolchains(
    go_version = "1.9.3",
)

#=============================================================================
# Docker rules

git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.4.0",
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
    digest = "sha256:fc1b461367730660ac5a40c1eb2d1b23221829acf8a892981c12361383b3742b",
    registry = "k8s.gcr.io",
    repository = "debian-hyperkube-base-amd64",
    tag = "0.8",
)

git_repository(
    name = "distroless_rules",
    remote = "https://github.com/googlecloudplatform/distroless.git",
    commit = "886114394dfed219001ec3b068b139a3456e49d4"
)

load(
    "@distroless_rules//package_manager:package_manager.bzl",
    "package_manager_repositories",
    "dpkg_src",
    "dpkg_list",
)

package_manager_repositories()

dpkg_src(
    name = "debian_stretch",
    arch = "amd64",
    distro = "stretch",
    sha256 = "9aea0e4c9ce210991c6edcb5370cb9b11e9e554a0f563e7754a4028a8fd0cb73",
    snapshot = "20171101T160520Z",
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
