load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "go-bindata-assetfs",
    importpath = "github.com/elazarl/go-bindata-assetfs/go-bindata-assetfs",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
