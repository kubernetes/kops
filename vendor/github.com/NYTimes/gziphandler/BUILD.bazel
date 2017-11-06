load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "gzip.go",
        "gzip_go18.go",
    ],
    importpath = "github.com/NYTimes/gziphandler",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "gzip_go18_test.go",
        "gzip_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "github.com/NYTimes/gziphandler",
    library = ":go_default_library",
    deps = ["//vendor/github.com/stretchr/testify/assert:go_default_library"],
)
