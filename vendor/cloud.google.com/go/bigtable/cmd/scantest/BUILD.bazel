load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["scantest.go"],
    importpath = "cloud.google.com/go/bigtable/cmd/scantest",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/cloud.google.com/go/bigtable:go_default_library",
        "//vendor/cloud.google.com/go/bigtable/internal/cbtrc:go_default_library",
        "//vendor/cloud.google.com/go/bigtable/internal/stat:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_binary(
    name = "scantest",
    importpath = "cloud.google.com/go/bigtable/cmd/scantest",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
