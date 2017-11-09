load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "merge.go",
        "patch.go",
    ],
    importpath = "github.com/evanphx/json-patch",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "merge_test.go",
        "patch_test.go",
    ],
    importpath = "github.com/evanphx/json-patch",
    library = ":go_default_library",
)
