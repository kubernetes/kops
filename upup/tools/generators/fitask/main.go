package main

import (
	"k8s.io/kops/upup/tools/generators/pkg/codegen"
)

func main() {
	generator := &FitaskGenerator{}
	codegen.RunGenerator("fitask", generator)
}
