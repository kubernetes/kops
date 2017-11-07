load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["errors.go"],
    importpath = "cloud.google.com/go/errors",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/logging:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["stack_test.go"],
    importpath = "cloud.google.com/go/errors",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = ["errors_test.go"],
    importpath = "cloud.google.com/go/errors_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)
