package controller

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	examplev1 "github.com/yourusername/json-server-controller/api/v1"
)

func TestReconcile_CreatesDeployment(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = examplev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	jsonServer := &examplev1.JsonServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-test",
			Namespace: "default",
		},
		Spec: examplev1.JsonServerSpec{
			Replicas:   2,
			JsonConfig: `{"users": []}`,
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(jsonServer).
		WithStatusSubresource(jsonServer).
		Build()

	r := &JsonServerReconciler{
		Client: client,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "app-test",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	// Check deployment was created
	deployment := &appsv1.Deployment{}
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      "app-test",
		Namespace: "default",
	}, deployment)
	if err != nil {
		t.Fatalf("expected deployment to be created: %v", err)
	}

	if *deployment.Spec.Replicas != 2 {
		t.Errorf("expected 2 replicas, got %d", *deployment.Spec.Replicas)
	}
}

func TestReconcile_CreatesConfigMap(t *testing.T) {
	scheme := runtime.NewScheme()
	_ = examplev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	jsonServer := &examplev1.JsonServer{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-test",
			Namespace: "default",
		},
		Spec: examplev1.JsonServerSpec{
			Replicas:   1,
			JsonConfig: `{"posts": [{"id": 1}]}`,
		},
	}

	client := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(jsonServer).
		WithStatusSubresource(jsonServer).
		Build()

	r := &JsonServerReconciler{
		Client: client,
		Scheme: scheme,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "app-test",
			Namespace: "default",
		},
	}

	_, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("reconcile failed: %v", err)
	}

	// Check configmap was created
	configMap := &corev1.ConfigMap{}
	err = client.Get(context.Background(), types.NamespacedName{
		Name:      "app-test-config",
		Namespace: "default",
	}, configMap)
	if err != nil {
		t.Fatalf("expected configmap to be created: %v", err)
	}

	if configMap.Data["db.json"] != `{"posts": [{"id": 1}]}` {
		t.Errorf("configmap data mismatch: %s", configMap.Data["db.json"])
	}
}
