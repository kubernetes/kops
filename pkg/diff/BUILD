load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["diff.go"],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_golang_glog//:go_default_library",
        "@com_github_sergi_go_diff//diffmatchpatch:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["diff_test.go"],
    library = ":go_default_library",
    deps = ["@com_github_sergi_go_diff//diffmatchpatch:go_default_library"],
)
