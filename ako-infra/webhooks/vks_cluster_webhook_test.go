/*
Copyright 2024 VMware, Inc.
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

package webhooks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
)

func TestVKSClusterWebhook_ServeHTTP(t *testing.T) {
	tests := []struct {
		name          string
		cluster       *unstructured.Unstructured
		namespace     *corev1.Namespace
		expectLabel   bool
		expectAllowed bool
		expectPatch   bool
	}{
		{
			name: "should add label to eligible cluster",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.x-k8s.io/v1beta1",
					"kind":       "Cluster",
					"metadata": map[string]interface{}{
						"name":      "test-cluster",
						"namespace": "test-namespace",
					},
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						"vmware-system-csi/serviceenginegroup": "seg-1",
					},
				},
			},
			expectLabel:   true,
			expectAllowed: true,
			expectPatch:   true,
		},
		{
			name: "should skip cluster without SEG namespace",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.x-k8s.io/v1beta1",
					"kind":       "Cluster",
					"metadata": map[string]interface{}{
						"name":      "test-cluster",
						"namespace": "test-namespace",
					},
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
				},
			},
			expectLabel:   false,
			expectAllowed: true,
			expectPatch:   false,
		},
		{
			name: "should skip cluster with explicit opt-out",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.x-k8s.io/v1beta1",
					"kind":       "Cluster",
					"metadata": map[string]interface{}{
						"name":      "test-cluster",
						"namespace": "test-namespace",
						"labels": map[string]interface{}{
							ingestion.VKSManagedLabel: ingestion.VKSManagedLabelValueFalse,
						},
					},
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						"vmware-system-csi/serviceenginegroup": "seg-1",
					},
				},
			},
			expectLabel:   false,
			expectAllowed: true,
			expectPatch:   false,
		},
		{
			name: "should add label to cluster with existing labels",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "cluster.x-k8s.io/v1beta1",
					"kind":       "Cluster",
					"metadata": map[string]interface{}{
						"name":      "test-cluster",
						"namespace": "test-namespace",
						"labels": map[string]interface{}{
							"existing-label": "existing-value",
						},
					},
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						"vmware-system-csi/serviceenginegroup": "seg-1",
					},
				},
			},
			expectLabel:   true,
			expectAllowed: true,
			expectPatch:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake client with namespace
			fakeClient := fake.NewSimpleClientset(tt.namespace)
			webhook := NewVKSClusterWebhook(fakeClient)

			// Create admission request
			clusterBytes, _ := json.Marshal(tt.cluster)
			admissionReq := &admissionv1.AdmissionRequest{
				UID:       "test-uid",
				Operation: admissionv1.Create,
				Kind: metav1.GroupVersionKind{
					Group:   "cluster.x-k8s.io",
					Version: "v1beta1",
					Kind:    "Cluster",
				},
				Object: runtime.RawExtension{
					Raw: clusterBytes,
				},
			}

			admissionReview := &admissionv1.AdmissionReview{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "admission.k8s.io/v1",
					Kind:       "AdmissionReview",
				},
				Request: admissionReq,
			}

			// Marshal admission review
			reviewBytes, _ := json.Marshal(admissionReview)

			// Create HTTP request
			req := httptest.NewRequest("POST", "/mutate-cluster-x-k8s-io-v1beta1-cluster", bytes.NewReader(reviewBytes))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Call webhook
			webhook.ServeHTTP(w, req)

			// Check response
			if w.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", w.Code)
			}

			// Parse response
			var responseReview admissionv1.AdmissionReview
			if err := json.Unmarshal(w.Body.Bytes(), &responseReview); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			// Check allowed
			if responseReview.Response.Allowed != tt.expectAllowed {
				t.Errorf("Expected allowed=%v, got %v", tt.expectAllowed, responseReview.Response.Allowed)
			}

			// Check patch
			if tt.expectPatch {
				if responseReview.Response.Patch == nil {
					t.Error("Expected patch, but got nil")
				} else {
					// Verify patch content
					var patches []map[string]interface{}
					if err := json.Unmarshal(responseReview.Response.Patch, &patches); err != nil {
						t.Fatalf("Failed to unmarshal patch: %v", err)
					}

					found := false
					for _, patch := range patches {
						if patch["op"] == "add" &&
							(patch["path"] == "/metadata/labels/"+escapeJSONPointer(ingestion.VKSManagedLabel) ||
								(patch["path"] == "/metadata/labels" &&
									patch["value"].(map[string]interface{})[ingestion.VKSManagedLabel] == ingestion.VKSManagedLabelValueTrue)) {
							found = true
							break
						}
					}
					if !found {
						t.Error("Expected patch to contain VKS managed label")
					}
				}
			} else {
				if responseReview.Response.Patch != nil {
					t.Error("Expected no patch, but got one")
				}
			}
		})
	}
}

func TestVKSClusterWebhook_shouldManageCluster(t *testing.T) {
	tests := []struct {
		name           string
		cluster        *unstructured.Unstructured
		namespace      *corev1.Namespace
		expectedResult bool
	}{
		{
			name: "should manage cluster in SEG namespace",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-cluster",
						"namespace": "test-namespace",
					},
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						"vmware-system-csi/serviceenginegroup": "seg-1",
					},
				},
			},
			expectedResult: true,
		},
		{
			name: "should not manage cluster without SEG",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-cluster",
						"namespace": "test-namespace",
					},
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
				},
			},
			expectedResult: false,
		},
		{
			name: "should not manage cluster with opt-out label",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name":      "test-cluster",
						"namespace": "test-namespace",
						"labels": map[string]interface{}{
							ingestion.VKSManagedLabel: ingestion.VKSManagedLabelValueFalse,
						},
					},
				},
			},
			namespace: &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-namespace",
					Annotations: map[string]string{
						"vmware-system-csi/serviceenginegroup": "seg-1",
					},
				},
			},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewSimpleClientset(tt.namespace)
			webhook := NewVKSClusterWebhook(fakeClient)

			result := webhook.shouldManageCluster(tt.cluster)
			if result != tt.expectedResult {
				t.Errorf("Expected %v, got %v", tt.expectedResult, result)
			}
		})
	}
}

func TestVKSClusterWebhook_createVKSLabelPatch(t *testing.T) {
	webhook := NewVKSClusterWebhook(fake.NewSimpleClientset())

	tests := []struct {
		name            string
		cluster         *unstructured.Unstructured
		expectedPatches int
	}{
		{
			name: "cluster without labels",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-cluster",
					},
				},
			},
			expectedPatches: 1, // add /metadata/labels
		},
		{
			name: "cluster with existing labels",
			cluster: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"metadata": map[string]interface{}{
						"name": "test-cluster",
						"labels": map[string]interface{}{
							"existing": "label",
						},
					},
				},
			},
			expectedPatches: 1, // add specific label
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patches, err := webhook.createVKSLabelPatch(tt.cluster)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if len(patches) != tt.expectedPatches {
				t.Errorf("Expected %d patches, got %d", tt.expectedPatches, len(patches))
			}

			// Verify patch structure
			for _, patch := range patches {
				if patch["op"] != "add" {
					t.Errorf("Expected op=add, got %v", patch["op"])
				}
			}
		})
	}
}

func TestEscapeJSONPointer(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with~tilde", "with~0tilde"},
		{"with/slash", "with~1slash"},
		{"with~and/both", "with~0and~1both"},
		{"ako.kubernetes.vmware.com/install", "ako.kubernetes.vmware.com~1install"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := escapeJSONPointer(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
