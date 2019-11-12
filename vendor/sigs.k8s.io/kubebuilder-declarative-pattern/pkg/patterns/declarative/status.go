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

// Status provides health and readiness information for a given DeclarativeObject
type Status interface {
	Reconciled
	Preflight
}

type Reconciled interface {
	// Reconciled is triggered when Reconciliation has occured.
	// The caller is encouraged to determine and surface the health of the reconcilation
	// on the DeclarativeObject.
	Reconciled(context.Context, DeclarativeObject, *manifest.Objects) error
}

type Preflight interface {
	// Preflight validates if the current state of the world is ready for reconciling.
	// Returning a non-nil error on this object will prevent Reconcile from running.
	// The caller is encouraged to surface the error status on the DeclarativeObject.
	Preflight(context.Context, DeclarativeObject) error
}

// StatusBuilder provides a pluggable implementation of Status
type StatusBuilder struct {
	ReconciledImpl Reconciled
	PreflightImpl  Preflight
}

func (s *StatusBuilder) Reconciled(ctx context.Context, src DeclarativeObject, objs *manifest.Objects) error {
	if s.ReconciledImpl != nil {
		return s.ReconciledImpl.Reconciled(ctx, src, objs)
	}
	return nil
}

func (s *StatusBuilder) Preflight(ctx context.Context, src DeclarativeObject) error {
	if s.PreflightImpl != nil {
		return s.PreflightImpl.Preflight(ctx, src)
	}
	return nil
}

var _ Status = &StatusBuilder{}
