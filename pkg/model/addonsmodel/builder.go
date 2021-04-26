/*
Copyright 2021 The Kubernetes Authors.

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

package addonsmodel

import (
	"context"

	awsapi "k8s.io/kops/cloud-controllers/aws/api/v1alpha1"
	"k8s.io/kops/cloud-controllers/aws/controllers"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/upup/pkg/fi"
)

// AddonsBuilder builds tasks for well-known addons.
type AddonsBuilder struct {
	*model.KopsModelContext
	Lifecycle *fi.Lifecycle
}

var _ fi.ModelBuilder = &AddonsBuilder{}

func (b *AddonsBuilder) Build(c *fi.ModelBuilderContext) error {
	for _, obj := range b.Addons {
		gvk, err := obj.GroupVersionKind()
		if err != nil {
			return err
		}
		switch gvk.Group + "/" + gvk.Kind {
		case "kops.k8s.io/AWSIdentityBinding":
			r := &controllers.AWSIdentityBindingReconciler{}
			o := &awsapi.AWSIdentityBinding{}
			if err := obj.ConvertInto(o); err != nil {
				return err
			}
			if err := r.BuildTasks(context.TODO(), o, c, b.KopsModelContext, b.Lifecycle); err != nil {
				return err
			}
		}
	}
	return nil
}
