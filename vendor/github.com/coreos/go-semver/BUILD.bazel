load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["example.go"],
    importpath = "github.com/coreos/go-semver",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/coreos/go-semver/semver:go_default_library"],
)

go_binary(
    name = "go-semver",
    importpath = "github.com/coreos/go-semver",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
