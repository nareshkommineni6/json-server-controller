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

package v1

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var jsonserverlog = logf.Log.WithName("jsonserver-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *JsonServer) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-example-com-v1-jsonserver,mutating=true,failurePolicy=fail,sideEffects=None,groups=example.com,resources=jsonservers,verbs=create;update,versions=v1,name=mjsonserver.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &JsonServer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *JsonServer) Default() {
	jsonserverlog.Info("default", "name", r.Name)

	// Set default replicas to 1 if not specified
	if r.Spec.Replicas == 0 {
		r.Spec.Replicas = 1
	}
}

// +kubebuilder:webhook:path=/validate-example-com-v1-jsonserver,mutating=false,failurePolicy=fail,sideEffects=None,groups=example.com,resources=jsonservers,verbs=create;update,versions=v1,name=vjsonserver.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &JsonServer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateCreate() (admission.Warnings, error) {
	jsonserverlog.Info("validate create", "name", r.Name)

	return r.validateJsonServer()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	jsonserverlog.Info("validate update", "name", r.Name)

	return r.validateJsonServer()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *JsonServer) ValidateDelete() (admission.Warnings, error) {
	jsonserverlog.Info("validate delete", "name", r.Name)

	// No validation needed for delete
	return nil, nil
}

// validateJsonServer validates the JsonServer resource
func (r *JsonServer) validateJsonServer() (admission.Warnings, error) {
	var warnings admission.Warnings

	// Validate naming convention: must start with "app-"
	if !strings.HasPrefix(r.Name, "app-") {
		return warnings, fmt.Errorf("metadata.name must follow the naming convention 'app-${name}': got %q", r.Name)
	}

	// Validate that the name after "app-" is not empty
	nameAfterPrefix := strings.TrimPrefix(r.Name, "app-")
	if nameAfterPrefix == "" {
		return warnings, fmt.Errorf("metadata.name must follow the naming convention 'app-${name}': name after 'app-' cannot be empty")
	}

	// Validate JSON config
	if r.Spec.JsonConfig == "" {
		return warnings, fmt.Errorf("spec.jsonConfig is required")
	}

	// Validate that jsonConfig is valid JSON
	var js interface{}
	if err := json.Unmarshal([]byte(r.Spec.JsonConfig), &js); err != nil {
		return warnings, fmt.Errorf("spec.jsonConfig is not a valid json object")
	}

	// Validate replicas
	if r.Spec.Replicas < 1 {
		return warnings, fmt.Errorf("spec.replicas must be at least 1")
	}

	return warnings, nil
}
