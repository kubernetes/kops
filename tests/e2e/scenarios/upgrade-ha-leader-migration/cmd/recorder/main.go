/*
Copyright 2022 The Kubernetes Authors.

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
	"os/signal"
	"syscall"
	"time"

	"k8s.io/klog/v2"
)

var Namespace = "kube-system"
var KCMPrefix = "kube-controller-manager"
var CCMPrefix = "cloud-controller-manager"
var CheckInterval = 30 * time.Second

func run(ctx context.Context) error {
	r, err := NewRecorder()
	if err != nil {
		return err
	}
	ticker := time.NewTicker(CheckInterval)
	defer ticker.Stop()
	for {
		err := r.Observe(ctx)
		if err != nil {
			if err == ErrConflictDetected {
				err = r.Observe(ctx)
				if err == ErrConflictDetected {
					klog.ErrorS(err, "conflicts detected twice in a row, test failed")
					os.Exit(1)
				}
			}
			if err != nil {
				klog.ErrorS(err, "fail to observe")
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func main() {
	flag.StringVar(&Namespace, "namespace", Namespace, "the namespace that contains system daemon sets.")
	flag.StringVar(&KCMPrefix, "kcm", KCMPrefix, "the prefix of KCM pods.")
	flag.StringVar(&CCMPrefix, "ccm", CCMPrefix, "the prefix of CCM pods.")
	flag.DurationVar(&CheckInterval, "interval", CheckInterval, "the interval between two checks.")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	err := run(ctx)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
