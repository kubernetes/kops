load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["version.go"],
    importpath = "github.com/coreos/etcd/version",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/go-semver/semver:go_default_library"],
)
