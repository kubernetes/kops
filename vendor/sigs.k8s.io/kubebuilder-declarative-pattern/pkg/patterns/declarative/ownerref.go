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

package declarative

import (
	"context"

	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative/pkg/manifest"
)

// SourceAsOwner is a OwnerSelector that selects the source DeclarativeObject as the owner
func SourceAsOwner(ctx context.Context, src DeclarativeObject, obj manifest.Object, objs manifest.Objects) (DeclarativeObject, error) {
	return src, nil
}

var _ OwnerSelector = SourceAsOwner
