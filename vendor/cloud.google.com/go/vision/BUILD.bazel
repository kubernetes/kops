load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "annotations.go",
        "doc.go",
        "face.go",
        "geometry.go",
        "image.go",
        "latlng.go",
        "vision.go",
    ],
    importpath = "cloud.google.com/go/vision",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/cloud.google.com/go/vision/apiv1:go_default_library",
        "//vendor/github.com/golang/geo/r3:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/cloud/vision/v1:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/type/color:go_default_library",
        "//vendor/google.golang.org/genproto/googleapis/type/latlng:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["vision_test.go"],
    data = glob(["testdata/**"]),
    importpath = "cloud.google.com/go/vision",
    library = ":go_default_library",
    deps = [
        "//vendor/cloud.google.com/go/internal/testutil:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/api/option:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["examples_test.go"],
    data = glob(["testdata/**"]),
    importpath = "cloud.google.com/go/vision_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
