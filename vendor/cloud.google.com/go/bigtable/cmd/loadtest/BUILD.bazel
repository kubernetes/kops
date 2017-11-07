load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["loadtest.go"],
    importpath = "cloud.google.com/go/bigtable/cmd/loadtest",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/cloud.google.com/go/bigtable:go_default_library",
        "//vendor/cloud.google.com/go/bigtable/internal/cbtrc:go_default_library",
        "//vendor/cloud.google.com/go/bigtable/internal/stat:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)

go_binary(
    name = "loadtest",
    importpath = "cloud.google.com/go/bigtable/cmd/loadtest",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
