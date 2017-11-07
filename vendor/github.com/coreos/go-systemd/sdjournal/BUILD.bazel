load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "functions.go",
        "journal.go",
        "read.go",
    ],
    cgo = True,
    importpath = "github.com/coreos/go-systemd/sdjournal",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/pkg/dlopen:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "functions_test.go",
        "journal_test.go",
    ],
    importpath = "github.com/coreos/go-systemd/sdjournal",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/go-systemd/journal:go_default_library"],
)
