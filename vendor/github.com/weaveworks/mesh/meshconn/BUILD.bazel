load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "mesh_addr.go",
        "peer.go",
        "pkt.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/weaveworks/mesh:go_default_library"],
)
