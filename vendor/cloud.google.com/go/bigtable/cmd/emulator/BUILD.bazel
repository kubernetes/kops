load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["cbtemulator.go"],
    importpath = "cloud.google.com/go/bigtable/cmd/emulator",
    visibility = ["//visibility:private"],
    deps = ["//vendor/cloud.google.com/go/bigtable/bttest:go_default_library"],
)

go_binary(
    name = "emulator",
    importpath = "cloud.google.com/go/bigtable/cmd/emulator",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
