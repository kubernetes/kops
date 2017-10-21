load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "key.go",
        "manager.go",
        "repo.go",
        "rotate.go",
        "sync.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/go-oidc/jose:go_default_library",
        "//vendor/github.com/coreos/pkg/health:go_default_library",
        "//vendor/github.com/coreos/pkg/timeutil:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "key_test.go",
        "manager_test.go",
        "rotate_test.go",
        "sync_test.go",
    ],
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/go-oidc/jose:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
    ],
)
