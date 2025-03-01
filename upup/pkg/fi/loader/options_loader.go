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

package loader

import (
	"fmt"
	"reflect"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/util/pkg/reflectutils"
)

const maxIterations = 10

type OptionsLoader[T any] struct {
	Builders []OptionsBuilder[T]
}

type OptionsBuilder[T any] interface {
	BuildOptions(options T) error
}

type ClusterOptionsBuilder OptionsBuilder[*kops.Cluster]

func NewClusterOptionsLoader(builders []ClusterOptionsBuilder) *OptionsLoader[*kops.Cluster] {
	l := &OptionsLoader[*kops.Cluster]{}
	for _, b := range builders {
		l.Builders = append(l.Builders, b)
	}
	return l
}

// iterate performs a single iteration of all the templates, executing each template in order
func (l *OptionsLoader[T]) iterate(userConfig T, current T) (T, error) {
	t := reflect.TypeOf(current).Elem()

	next := reflect.New(t).Interface().(T)

	// Copy the current state before applying rules; they act as defaults
	reflectutils.JSONMergeStruct(next, current)

	for _, t := range l.Builders {
		klog.V(4).Infof("executing builder %T", t)

		err := t.BuildOptions(next)
		if err != nil {
			var defaultT T
			return defaultT, err
		}
	}

	// Also copy the user-provided values after applying rules; they act as overrides now
	reflectutils.JSONMergeStruct(next, userConfig)

	return next, nil
}

// Build executes the options configuration templates, until they converge
// It bails out after maxIterations
func (l *OptionsLoader[T]) Build(userConfig T) (T, error) {
	var defaultT T
	options := userConfig
	iteration := 0
	for {
		nextOptions, err := l.iterate(userConfig, options)
		if err != nil {
			return defaultT, err
		}

		if reflect.DeepEqual(options, nextOptions) {
			return options, nil
		}

		iteration++
		if iteration > maxIterations {
			return defaultT, fmt.Errorf("options did not converge after %d iterations", maxIterations)
		}

		options = nextOptions
	}
}
