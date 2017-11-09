load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "cbt.go",
        "cbtdoc.go",
    ],
    importpath = "cloud.google.com/go/bigtable/cmd/cbt",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/cloud.google.com/go/bigtable:go_default_library",
        "//vendor/cloud.google.com/go/bigtable/internal/cbtrc:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_binary(
    name = "cbt",
    importpath = "cloud.google.com/go/bigtable/cmd/cbt",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["cbt_test.go"],
    importpath = "cloud.google.com/go/bigtable/cmd/cbt",
    library = ":go_default_library",
)
