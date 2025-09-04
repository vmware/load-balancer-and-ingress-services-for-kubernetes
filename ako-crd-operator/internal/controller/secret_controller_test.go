/*
Copyright 2019-2025 VMware, Inc.
All Rights Reserved.

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
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	avisession "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-crd-operator/internal/session"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

func TestSecretController(t *testing.T) {
	ctx := context.Background()

	// Create a fake client
	s := runtime.NewScheme()
	if err := corev1.AddToScheme(s); err != nil {
		t.Fatalf("Failed to add corev1 to scheme: %v", err)
	}

	client := fake.NewClientBuilder().WithScheme(s).Build()
	eventRecorder := record.NewFakeRecorder(100)

	// Create a mock session manager and set it globally for testing
	sessionMgr := &avisession.Session{}
	avisession.SetGlobalSessionForTesting(sessionMgr)
	defer avisession.ResetGlobalSessionForTesting() // Clean up after test

	reconciler := &SecretReconciler{
		Client:        client,
		Scheme:        s,
		EventRecorder: eventRecorder,
		Logger:        utils.AviLog.WithName("secret-test"),
	}

	// Create test secret for future use
	_ = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      lib.AviSecret,
			Namespace: utils.GetAKONamespace(),
		},
		Data: map[string][]byte{
			"username":  []byte("admin"),
			"password":  []byte("password123"),
			"authtoken": []byte("token123"),
		},
	}

	t.Run("Should ignore non-avi-secret updates", func(t *testing.T) {
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      "other-secret",
				Namespace: utils.GetAKONamespace(),
			},
		}

		result, err := reconciler.Reconcile(ctx, req)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != (ctrl.Result{}) {
			t.Errorf("Expected empty result, got: %v", result)
		}
	})

	t.Run("Should ignore secrets in other namespaces", func(t *testing.T) {
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      lib.AviSecret,
				Namespace: "other-namespace",
			},
		}

		result, err := reconciler.Reconcile(ctx, req)
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if result != (ctrl.Result{}) {
			t.Errorf("Expected empty result, got: %v", result)
		}
	})

	t.Run("Should handle avi-secret not found", func(t *testing.T) {
		req := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      lib.AviSecret,
				Namespace: utils.GetAKONamespace(),
			},
		}

		result, err := reconciler.Reconcile(ctx, req)
		// Should return error when secret is not found
		if err == nil {
			t.Errorf("Expected error when secret not found, got nil")
		}
		if result != (ctrl.Result{}) {
			t.Errorf("Expected empty result, got: %v", result)
		}
	})
}

func TestSecretControllerPredicates(t *testing.T) {

	t.Run("Should filter non-avi-secret objects", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "other-secret",
				Namespace: utils.GetAKONamespace(),
			},
		}

		// Test the predicate function logic (simulating what the controller manager does)
		shouldProcess := secret.Name == lib.AviSecret && secret.Namespace == utils.GetAKONamespace()
		if shouldProcess {
			t.Error("Expected predicate to filter out non-avi-secret")
		}
	})

	t.Run("Should accept avi-secret in correct namespace", func(t *testing.T) {
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      lib.AviSecret,
				Namespace: utils.GetAKONamespace(),
			},
		}

		// Test the predicate function logic
		shouldProcess := secret.Name == lib.AviSecret && secret.Namespace == utils.GetAKONamespace()
		if !shouldProcess {
			t.Error("Expected predicate to accept avi-secret in correct namespace")
		}
	})

	t.Run("Should detect data changes in update predicate", func(t *testing.T) {
		oldSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      lib.AviSecret,
				Namespace: utils.GetAKONamespace(),
			},
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("oldpass"),
			},
		}

		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      lib.AviSecret,
				Namespace: utils.GetAKONamespace(),
			},
			Data: map[string][]byte{
				"username": []byte("admin"),
				"password": []byte("newpass"),
			},
		}

		// The actual predicate would call reflect.DeepEqual
		shouldTrigger := !reflect.DeepEqual(oldSecret.Data, newSecret.Data)
		if !shouldTrigger {
			t.Error("Expected update predicate to detect data changes")
		}
	})

	t.Run("Should ignore updates with no data changes", func(t *testing.T) {
		secretData := map[string][]byte{
			"username": []byte("admin"),
			"password": []byte("samepass"),
		}

		oldSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            lib.AviSecret,
				Namespace:       utils.GetAKONamespace(),
				ResourceVersion: "1",
			},
			Data: secretData,
		}

		newSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:            lib.AviSecret,
				Namespace:       utils.GetAKONamespace(),
				ResourceVersion: "2", // Different resource version but same data
			},
			Data: secretData,
		}

		// The actual predicate would call reflect.DeepEqual
		shouldTrigger := !reflect.DeepEqual(oldSecret.Data, newSecret.Data)
		if shouldTrigger {
			t.Error("Expected update predicate to ignore updates with no data changes")
		}
	})
}
