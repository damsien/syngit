/*
Copyright 2024.

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

package v1beta2

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	syngitv1beta2 "syngit.io/syngit/api/v1beta2"
)

// nolint:unused
// log is for logging in this package.
var remoteuserlog = logf.Log.WithName("remoteuser-resource")

// SetupRemoteUserWebhookWithManager registers the webhook for RemoteUser in the manager.
func SetupRemoteUserWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).For(&syngitv1beta2.RemoteUser{}).
		WithValidator(&RemoteUserCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-syngit-syngit-io-v1beta2-remoteuser,mutating=false,failurePolicy=fail,sideEffects=None,groups=syngit.syngit.io,resources=remoteusers,verbs=create;update,versions=v1beta2,name=vremoteuser-v1beta2.kb.io,admissionReviewVersions=v1
//+kubebuilder:webhook:path=/syngit-v1beta2-remoteuser-association,mutating=false,failurePolicy=fail,sideEffects=None,groups=syngit.syngit.io,resources=remoteusers,verbs=create;delete,versions=v1beta2,admissionReviewVersions=v1,name=vremoteusers-association.v1beta2.syngit.io

// NOTE: The +kubebuilder:object:generate=false marker prevents controller-gen from generating DeepCopy methods,
// as this struct is used only for temporary operations and does not need to be deeply copied.
type RemoteUserCustomValidator struct {
	//TODO(user): Add more fields as needed for validation
}

var _ webhook.CustomValidator = &RemoteUserCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type RemoteUser.
func (v *RemoteUserCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	remoteuser, ok := obj.(*syngitv1beta2.RemoteUser)
	if !ok {
		return nil, fmt.Errorf("expected a RemoteUser object but got %T", obj)
	}
	remoteuserlog.Info("Validation for RemoteUser upon creation", "name", remoteuser.GetName())

	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type RemoteUser.
func (v *RemoteUserCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	remoteuser, ok := newObj.(*syngitv1beta2.RemoteUser)
	if !ok {
		return nil, fmt.Errorf("expected a RemoteUser object for the newObj but got %T", newObj)
	}
	remoteuserlog.Info("Validation for RemoteUser upon update", "name", remoteuser.GetName())

	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type RemoteUser.
func (v *RemoteUserCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	remoteuser, ok := obj.(*syngitv1beta2.RemoteUser)
	if !ok {
		return nil, fmt.Errorf("expected a RemoteUser object but got %T", obj)
	}
	remoteuserlog.Info("Validation for RemoteUser upon deletion", "name", remoteuser.GetName())

	return nil, nil
}
