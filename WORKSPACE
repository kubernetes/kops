#=============================================================================
# Go rules

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "4d8d6244320dd751590f9100cf39fd7a4b75cd901e1f3ffdfd6f048328883695",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.9.0/rules_go-0.9.0.tar.gz",
)

load("@io_bazel_rules_go//go:def.bzl", "go_rules_dependencies", "go_register_toolchains", "go_download_sdk")

go_rules_dependencies()

# 1.9.3 is not available in rules 0.9.0, so register it manually (once 0.9.1 of rules-go is released, we can likely remove this)
go_download_sdk(
    name = "go_sdk",
    sdks = {
        "darwin_amd64":      ("go1.9.3.darwin-amd64.tar.gz", "f84b39c2ed7df0c2f1648e2b90b2198a6783db56b53700dabfa58afd6335d324"),
        "linux_386":         ("go1.9.3.linux-386.tar.gz", "bc0782ac8116b2244dfe2a04972bbbcd7f1c2da455a768ab47b32864bcd0d49d"),
        "linux_amd64":       ("go1.9.3.linux-amd64.tar.gz", "a4da5f4c07dfda8194c4621611aeb7ceaab98af0b38bfb29e1be2ebb04c3556c"),
        "linux_armv6l":      ("go1.9.3.linux-armv6l.tar.gz", "926d6cd6c21ef3419dca2e5da8d4b74b99592ab1feb5a62a4da244e6333189d2"),
        "windows_386":       ("go1.9.3.windows-386.zip", "cab7d4e008adefed322d36dee87a4c1775ab60b25ce587a2b55d90c75d0bafbc"),
        "windows_amd64":     ("go1.9.3.windows-amd64.zip", "4eee59bb5b70abc357aebd0c54f75e46322eb8b58bbdabc026fdd35834d65e1e"),
        "freebsd_386":       ("go1.9.3.freebsd-386.tar.gz", "a755739e3be0415344d62ea3b168bdcc9a54f7862ac15832684ff2d3e8127a03"),
        "freebsd_amd64":     ("go1.9.3.freebsd-amd64.tar.gz", "f95066089a88749c45fae798422d04e254fe3b622ff030d12bdf333402b186ec"),
        "linux_arm64":       ("go1.9.3.linux-arm64.tar.gz", "065d79964023ccb996e9dbfbf94fc6969d2483fbdeeae6d813f514c5afcd98d9"),
        "linux_ppc64le":     ("go1.9.3.linux-ppc64le.tar.gz", "c802194b1af0cd904689923d6d32f3ed68f9d5f81a3e4a82406d9ce9be163681"),
        "linux_s390x":       ("go1.9.3.linux-s390x.tar.gz", "85e9a257664f84154e583e0877240822bb2fe4308209f5ff57d80d16e2fb95c5"),
    },
)


go_register_toolchains(
#    go_version = "1.9.3",
)

#=============================================================================
# Docker rules

git_repository(
    name = "io_bazel_rules_docker",
    remote = "https://github.com/bazelbuild/rules_docker.git",
    tag = "v0.3.0",
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
