load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "dbus.go",
        "methods.go",
        "properties.go",
        "set.go",
        "subscription.go",
        "subscription_set.go",
    ],
    importpath = "github.com/coreos/go-systemd/dbus",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/godbus/dbus:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "dbus_test.go",
        "methods_test.go",
        "set_test.go",
        "subscription_set_test.go",
        "subscription_test.go",
    ],
    importpath = "github.com/coreos/go-systemd/dbus",
    library = ":go_default_library",
    deps = ["//vendor/github.com/godbus/dbus:go_default_library"],
)
