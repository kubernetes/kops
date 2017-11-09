load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "appengine.go",
        "default.go",
        "google.go",
        "jwt.go",
        "sdk.go",
    ],
    importpath = "golang.org/x/oauth2/google",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/compute/metadata:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/golang.org/x/oauth2/internal:go_default_library",
        "//vendor/golang.org/x/oauth2/jws:go_default_library",
        "//vendor/golang.org/x/oauth2/jwt:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "google_test.go",
        "jwt_test.go",
        "sdk_test.go",
    ],
    data = glob(["testdata/**"]),
    importpath = "golang.org/x/oauth2/google",
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/oauth2/jws:go_default_library"],
)
