load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "sampling.go",
        "trace.go",
    ],
    importpath = "cloud.google.com/go/trace",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/time/rate:go_default_library",
        "//vendor/google.golang.org/api/cloudtrace/v1:go_default_library",
        "//vendor/google.golang.org/api/gensupport:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/api/transport:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["trace_test.go"],
    importpath = "cloud.google.com/go/trace",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/storage:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/cloudtrace/v1:go_default_library",
        "//vendor/google.golang.org/api/compute/v1:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)
