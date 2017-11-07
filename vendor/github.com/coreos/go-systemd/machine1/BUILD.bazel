load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["dbus.go"],
    importpath = "github.com/coreos/go-systemd/machine1",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/godbus/dbus:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["dbus_test.go"],
    importpath = "github.com/coreos/go-systemd/machine1",
    library = ":go_default_library",
)
