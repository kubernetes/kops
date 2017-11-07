load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "files.go",
        "listeners.go",
        "packetconns.go",
    ],
    importpath = "github.com/coreos/go-systemd/activation",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "files_test.go",
        "listeners_test.go",
        "packetconns_test.go",
    ],
    importpath = "github.com/coreos/go-systemd/activation",
    library = ":go_default_library",
)
