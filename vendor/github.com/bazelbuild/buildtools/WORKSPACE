workspace(name = "com_github_bazelbuild_buildtools")

# 0.5.5
http_archive(
    name = "io_bazel_rules_go",
    strip_prefix = "rules_go-71cdb6fd5f887d215bdbe0e4d1eb137278b09c39",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/rules_go/archive/71cdb6fd5f887d215bdbe0e4d1eb137278b09c39.tar.gz",
        "https://github.com/bazelbuild/rules_go/archive/71cdb6fd5f887d215bdbe0e4d1eb137278b09c39.tar.gz",
    ],
)

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_rules_dependencies",
    "go_register_toolchains",
    "go_repository",
)

go_rules_dependencies()

go_register_toolchains()

# used for build.proto
http_archive(
    name = "io_bazel",
    strip_prefix = "bazel-0.6.0",
    urls = [
        "http://mirror.bazel.build/github.com/bazelbuild/bazel/archive/0.6.0.tar.gz",
        "https://github.com/bazelbuild/bazel/archive/0.6.0.tar.gz",
    ],
)

go_repository(
    name = "org_golang_x_tools",
    commit = "3d92dd60033c312e3ae7cac319c792271cf67e37",
    importpath = "golang.org/x/tools",
)
