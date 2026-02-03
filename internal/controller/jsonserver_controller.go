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

package controller

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	examplecomv1 "github.com/yourusername/json-server-controller/api/v1"
)

// JsonServerReconciler reconciles a JsonServer object
type JsonServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.com,resources=jsonservers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/finalizers,verbs=update
// +kubebuilder:rbac:groups=example.com,resources=jsonservers/scale,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *JsonServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the JsonServer instance
	jsonServer := &examplecomv1.JsonServer{}
	err := r.Get(ctx, req.NamespacedName, jsonServer)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			logger.Info("JsonServer resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Failed to get JsonServer")
		return ctrl.Result{}, err
	}

	// Validate JSON config
	var js interface{}
	if err := json.Unmarshal([]byte(jsonServer.Spec.JsonConfig), &js); err != nil {
		// Update status with error
		return r.updateStatusWithError(ctx, jsonServer, "Error: spec.jsonConfig is not a valid json object")
	}

	// Create or update ConfigMap
	configMap, err := r.reconcileConfigMap(ctx, jsonServer)
	if err != nil {
		logger.Error(err, "Failed to reconcile ConfigMap")
		return r.updateStatusWithError(ctx, jsonServer, "Error: unexpected failure")
	}
	logger.Info("ConfigMap reconciled", "ConfigMap.Namespace", configMap.Namespace, "ConfigMap.Name", configMap.Name)

	// Create or update Deployment
	deployment, err := r.reconcileDeployment(ctx, jsonServer)
	if err != nil {
		logger.Error(err, "Failed to reconcile Deployment")
		return r.updateStatusWithError(ctx, jsonServer, "Error: unexpected failure")
	}
	logger.Info("Deployment reconciled", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)

	// Create or update Service
	service, err := r.reconcileService(ctx, jsonServer)
	if err != nil {
		logger.Error(err, "Failed to reconcile Service")
		return r.updateStatusWithError(ctx, jsonServer, "Error: unexpected failure")
	}
	logger.Info("Service reconciled", "Service.Namespace", service.Namespace, "Service.Name", service.Name)

	// Update status to Synced
	return r.updateStatusSuccess(ctx, jsonServer)
}

// reconcileConfigMap creates or updates the ConfigMap for the JsonServer
func (r *JsonServerReconciler) reconcileConfigMap(ctx context.Context, jsonServer *examplecomv1.JsonServer) (*corev1.ConfigMap, error) {
	configMapName := fmt.Sprintf("%s-config", jsonServer.Name)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: jsonServer.Namespace,
		},
	}

	// Create or Update the ConfigMap
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, configMap, func() error {
		// Set the owner reference
		if err := controllerutil.SetControllerReference(jsonServer, configMap, r.Scheme); err != nil {
			return err
		}

		// Set labels
		configMap.Labels = map[string]string{
			"app":                          jsonServer.Name,
			"app.kubernetes.io/name":       jsonServer.Name,
			"app.kubernetes.io/managed-by": "json-server-controller",
		}

		// Set the data
		configMap.Data = map[string]string{
			"db.json": jsonServer.Spec.JsonConfig,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.FromContext(ctx).Info("ConfigMap operation completed", "operation", op)
	return configMap, nil
}

// reconcileDeployment creates or updates the Deployment for the JsonServer
func (r *JsonServerReconciler) reconcileDeployment(ctx context.Context, jsonServer *examplecomv1.JsonServer) (*appsv1.Deployment, error) {
	configMapName := fmt.Sprintf("%s-config", jsonServer.Name)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jsonServer.Name,
			Namespace: jsonServer.Namespace,
		},
	}

	// Create or Update the Deployment
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		// Set the owner reference
		if err := controllerutil.SetControllerReference(jsonServer, deployment, r.Scheme); err != nil {
			return err
		}

		// Set labels
		labels := map[string]string{
			"app":                          jsonServer.Name,
			"app.kubernetes.io/name":       jsonServer.Name,
			"app.kubernetes.io/managed-by": "json-server-controller",
		}
		deployment.Labels = labels

		// Set the spec
		deployment.Spec = appsv1.DeploymentSpec{
			Replicas: &jsonServer.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": jsonServer.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": jsonServer.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "json-server",
							Image: "backplane/json-server",
							Args:  []string{"/data/db.json"},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 3000,
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "json-config",
									MountPath: "/data",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "json-config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: configMapName,
									},
								},
							},
						},
					},
				},
			},
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.FromContext(ctx).Info("Deployment operation completed", "operation", op)
	return deployment, nil
}

// reconcileService creates or updates the Service for the JsonServer
func (r *JsonServerReconciler) reconcileService(ctx context.Context, jsonServer *examplecomv1.JsonServer) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jsonServer.Name,
			Namespace: jsonServer.Namespace,
		},
	}

	// Create or Update the Service
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, service, func() error {
		// Set the owner reference
		if err := controllerutil.SetControllerReference(jsonServer, service, r.Scheme); err != nil {
			return err
		}

		// Set labels
		service.Labels = map[string]string{
			"app":                          jsonServer.Name,
			"app.kubernetes.io/name":       jsonServer.Name,
			"app.kubernetes.io/managed-by": "json-server-controller",
		}

		// Set the spec
		service.Spec = corev1.ServiceSpec{
			Selector: map[string]string{
				"app": jsonServer.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Type: corev1.ServiceTypeClusterIP,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	log.FromContext(ctx).Info("Service operation completed", "operation", op)
	return service, nil
}

// updateStatusWithError updates the JsonServer status with an error
func (r *JsonServerReconciler) updateStatusWithError(ctx context.Context, jsonServer *examplecomv1.JsonServer, message string) (ctrl.Result, error) {
	// Get the latest version of the JsonServer
	latest := &examplecomv1.JsonServer{}
	if err := r.Get(ctx, types.NamespacedName{Name: jsonServer.Name, Namespace: jsonServer.Namespace}, latest); err != nil {
		return ctrl.Result{}, err
	}

	latest.Status.State = "Error"
	latest.Status.Message = message
	latest.Status.Replicas = jsonServer.Spec.Replicas

	if err := r.Status().Update(ctx, latest); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update JsonServer status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// updateStatusSuccess updates the JsonServer status to Synced
func (r *JsonServerReconciler) updateStatusSuccess(ctx context.Context, jsonServer *examplecomv1.JsonServer) (ctrl.Result, error) {
	// Get the latest version of the JsonServer
	latest := &examplecomv1.JsonServer{}
	if err := r.Get(ctx, types.NamespacedName{Name: jsonServer.Name, Namespace: jsonServer.Namespace}, latest); err != nil {
		return ctrl.Result{}, err
	}

	latest.Status.State = "Synced"
	latest.Status.Message = "Synced succesfully!"
	latest.Status.Replicas = jsonServer.Spec.Replicas

	if err := r.Status().Update(ctx, latest); err != nil {
		log.FromContext(ctx).Error(err, "Failed to update JsonServer status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JsonServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplecomv1.JsonServer{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Complete(r)
}
