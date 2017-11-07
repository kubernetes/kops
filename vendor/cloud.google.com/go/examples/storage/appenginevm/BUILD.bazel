load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["app.go"],
    importpath = "cloud.google.com/go/examples/storage/appenginevm",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/cloud.google.com/go/storage:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/appengine:go_default_library",
        "//vendor/google.golang.org/appengine/file:go_default_library",
        "//vendor/google.golang.org/appengine/log:go_default_library",
    ],
)

go_binary(
    name = "appenginevm",
    importpath = "cloud.google.com/go/examples/storage/appenginevm",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
