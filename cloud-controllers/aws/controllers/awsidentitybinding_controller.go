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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	api "k8s.io/kops/cloud-controllers/aws/api/v1alpha1"
	"k8s.io/kops/pkg/model"
	"k8s.io/kops/pkg/model/iam"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awstasks"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// AWSIdentityBindingReconciler reconciles a AWSIdentityBinding object
type AWSIdentityBindingReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kops.k8s.io,resources=awsidentitybindings,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kops.k8s.io,resources=awsidentitybindings/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kops.k8s.io,resources=awsidentitybindings/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AWSIdentityBinding object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.2/pkg/reconcile
func (r *AWSIdentityBindingReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("awsidentitybinding", req.NamespacedName)

	subject := &api.AWSIdentityBinding{}

	if err := r.Get(ctx, req.NamespacedName, subject); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.reconcile(ctx, subject)
}

type genericServiceAccount struct {
	NamespacedName types.NamespacedName
	Policy         *iam.Policy
}

func (g *genericServiceAccount) ServiceAccount() (types.NamespacedName, bool) {
	return g.NamespacedName, true
}

func (g *genericServiceAccount) BuildAWSPolicy(*iam.PolicyBuilder) (*iam.Policy, error) {
	return g.Policy, nil
}

func (r *AWSIdentityBindingReconciler) reconcile(ctx context.Context, subject *api.AWSIdentityBinding) error {
	return fmt.Errorf("controller execution not yet implemented")
}

func (r *AWSIdentityBindingReconciler) BuildTasks(ctx context.Context, subject *api.AWSIdentityBinding, c *fi.ModelBuilderContext, b *model.KopsModelContext, lifecycle *fi.Lifecycle) error {
	iamBuilder := &model.IAMModelBuilder{
		KopsModelContext: b,
		Lifecycle:        lifecycle,
	}

	var p *iam.Policy
	if subject.Spec.InlinePolicy != "" {
		bp, err := iamBuilder.ParsePolicy(subject.Spec.InlinePolicy)
		p = bp
		if err != nil {
			return fmt.Errorf("error parsing inline policy: %w", err)
		}
	}

	serviceAccount := &genericServiceAccount{
		NamespacedName: types.NamespacedName{
			Name:      subject.Spec.Subject.Name,
			Namespace: subject.Spec.Subject.Namespace,
		},
		Policy: p,
	}

	iamRole, err := iamBuilder.BuildServiceAccountRoleTasks(serviceAccount, c)
	if err != nil {
		return fmt.Errorf("error building service account role tasks: %w", err)
	}

	if len(subject.Spec.IAMPolicyARNs) > 0 {
		name := "external-" + fi.StringValue(iamRole.Name)
		externalPolicies := subject.Spec.IAMPolicyARNs
		c.AddTask(&awstasks.IAMRolePolicy{
			Name:             fi.String(name),
			ExternalPolicies: &externalPolicies,
			Managed:          true,
			Role:             iamRole,
			Lifecycle:        iamBuilder.Lifecycle,
		})
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AWSIdentityBindingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.AWSIdentityBinding{}).
		Complete(r)
}
