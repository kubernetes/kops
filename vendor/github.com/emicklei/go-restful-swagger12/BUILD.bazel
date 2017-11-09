load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "api_declaration_list.go",
        "config.go",
        "model_builder.go",
        "model_list.go",
        "model_property_ext.go",
        "model_property_list.go",
        "ordered_route_map.go",
        "swagger.go",
        "swagger_builder.go",
        "swagger_webservice.go",
    ],
    importpath = "github.com/emicklei/go-restful-swagger12",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/emicklei/go-restful:go_default_library",
        "//vendor/github.com/emicklei/go-restful/log:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "model_builder_test.go",
        "model_list_test.go",
        "model_property_ext_test.go",
        "model_property_list_test.go",
        "ordered_route_map_test.go",
        "postbuild_model_test.go",
        "swagger_test.go",
        "utils_test.go",
    ],
    importpath = "github.com/emicklei/go-restful-swagger12",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/emicklei/go-restful:go_default_library",
        "//vendor/github.com/emicklei/go-restful-swagger12/test_package:go_default_library",
    ],
)
