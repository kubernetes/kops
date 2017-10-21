load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "claims.go",
        "doc.go",
        "jose.go",
        "jwk.go",
        "jws.go",
        "jwt.go",
        "sig.go",
        "sig_rsa.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "claims_test.go",
        "jwk_test.go",
        "jws_test.go",
        "jwt_test.go",
    ],
    library = ":go_default_library",
)
