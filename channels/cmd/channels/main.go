/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"k8s.io/klog/v2"

	"k8s.io/kops/channels/pkg/cmd"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "\n%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	klog.InitFlags(nil)
	// Opt into the new klog behavior so that -stderrthreshold is honored even
	// when -logtostderr=true (the default).
	// Ref: kubernetes/klog#212, kubernetes/klog#432
	flag.Set("legacy_stderr_threshold_behavior", "false") //nolint:errcheck
	flag.Set("stderrthreshold", "INFO")                   //nolint:errcheck

	klog.Info("CHANNELS_STARTUP_PROBE: locally built channels binary is running")

	f := cmd.NewChannelsFactory()

	if err := cmd.Execute(ctx, f, os.Stdout); err != nil {
		return err
	}
	return nil
}
