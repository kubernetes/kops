load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "client.go",
        "doc.go",
        "identity.go",
        "interface.go",
        "key.go",
        "provider.go",
        "transport.go",
        "util.go",
        "verification.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/go-oidc/http:go_default_library",
        "//vendor/github.com/coreos/go-oidc/jose:go_default_library",
        "//vendor/github.com/coreos/go-oidc/key:go_default_library",
        "//vendor/github.com/coreos/go-oidc/oauth2:go_default_library",
        "//vendor/github.com/coreos/pkg/timeutil:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "client_test.go",
        "identity_test.go",
        "provider_test.go",
        "transport_test.go",
        "util_test.go",
        "verification_test.go",
    ],
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/go-oidc/jose:go_default_library",
        "//vendor/github.com/coreos/go-oidc/key:go_default_library",
        "//vendor/github.com/coreos/go-oidc/oauth2:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
        "//vendor/github.com/kylelemons/godebug/diff:go_default_library",
        "//vendor/github.com/kylelemons/godebug/pretty:go_default_library",
    ],
)
