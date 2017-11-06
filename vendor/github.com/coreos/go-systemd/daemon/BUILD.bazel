load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "sdnotify.go",
        "watchdog.go",
    ],
    importpath = "github.com/coreos/go-systemd/daemon",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "sdnotify_test.go",
        "watchdog_test.go",
    ],
    importpath = "github.com/coreos/go-systemd/daemon",
    library = ":go_default_library",
)
