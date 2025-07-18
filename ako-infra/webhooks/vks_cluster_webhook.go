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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

// VKSClusterWebhook handles admission requests for Cluster objects
// It automatically adds VKS managed labels to new clusters in eligible namespaces
type VKSClusterWebhook struct {
	Client kubernetes.Interface
	scheme *runtime.Scheme
	codecs serializer.CodecFactory
}

// NewVKSClusterWebhook creates a new VKS cluster admission webhook
func NewVKSClusterWebhook(client kubernetes.Interface) *VKSClusterWebhook {
	scheme := runtime.NewScheme()
	codecs := serializer.NewCodecFactory(scheme)

	return &VKSClusterWebhook{
		Client: client,
		scheme: scheme,
		codecs: codecs,
	}
}

// ServeHTTP handles incoming admission webhook requests
func (w *VKSClusterWebhook) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to read request body: %v", err)
		http.Error(rw, "failed to read request body", http.StatusBadRequest)
		return
	}

	// Parse the AdmissionReview request
	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to unmarshal admission review: %v", err)
		http.Error(rw, "failed to unmarshal admission review", http.StatusBadRequest)
		return
	}

	// Process the admission request
	admissionResponse := w.processAdmissionRequest(admissionReview.Request)

	// Create the response
	responseAdmissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: admissionResponse,
	}
	responseAdmissionReview.Response.UID = admissionReview.Request.UID

	// Marshal and send response
	responseBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to marshal admission response: %v", err)
		http.Error(rw, "failed to marshal admission response", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	rw.Write(responseBytes)
}

// processAdmissionRequest processes an admission request and returns a response
func (w *VKSClusterWebhook) processAdmissionRequest(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
	// Only process CREATE operations on Cluster objects
	if req.Operation != admissionv1.Create {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
			Result:  &metav1.Status{Message: "not a create operation"},
		}
	}

	if req.Kind.Group != "cluster.x-k8s.io" || req.Kind.Kind != "Cluster" {
		return &admissionv1.AdmissionResponse{
			Allowed: true,
			Result:  &metav1.Status{Message: "not a cluster object"},
		}
	}

	// Parse the cluster object
	cluster := &unstructured.Unstructured{}
	if err := json.Unmarshal(req.Object.Raw, cluster); err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to unmarshal cluster object: %v", err)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("failed to unmarshal cluster object: %v", err),
			},
		}
	}

	utils.AviLog.Infof("VKS webhook processing cluster: %s/%s", cluster.GetNamespace(), cluster.GetName())

	// Check if cluster should be managed by VKS
	if !w.shouldManageCluster(cluster) {
		utils.AviLog.Infof("VKS webhook skipping cluster %s/%s - not eligible for VKS management",
			cluster.GetNamespace(), cluster.GetName())
		return &admissionv1.AdmissionResponse{
			Allowed: true,
			Result:  &metav1.Status{Message: "cluster not eligible for VKS management"},
		}
	}

	// Create patch to add VKS managed label
	patches, err := w.createVKSLabelPatch(cluster)
	if err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to create patch: %v", err)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("failed to create patch: %v", err),
			},
		}
	}

	patchBytes, err := json.Marshal(patches)
	if err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to marshal patches: %v", err)
		return &admissionv1.AdmissionResponse{
			Allowed: false,
			Result: &metav1.Status{
				Message: fmt.Sprintf("failed to marshal patches: %v", err),
			},
		}
	}

	patchType := admissionv1.PatchTypeJSONPatch
	utils.AviLog.Infof("VKS webhook added label %s=%s to cluster %s/%s",
		ingestion.VKSManagedLabel, ingestion.VKSManagedLabelValueTrue,
		cluster.GetNamespace(), cluster.GetName())

	return &admissionv1.AdmissionResponse{
		Allowed:   true,
		Patch:     patchBytes,
		PatchType: &patchType,
		Result:    &metav1.Status{Message: "added VKS managed label"},
	}
}

// shouldManageCluster determines if a cluster should be managed by VKS
func (w *VKSClusterWebhook) shouldManageCluster(cluster *unstructured.Unstructured) bool {
	// Check for explicit opt-out
	labels := cluster.GetLabels()
	if labels != nil {
		if value, exists := labels[ingestion.VKSManagedLabel]; exists && value == ingestion.VKSManagedLabelValueFalse {
			utils.AviLog.Infof("VKS webhook: cluster %s/%s explicitly opted out with label %s=%s",
				cluster.GetNamespace(), cluster.GetName(), ingestion.VKSManagedLabel, value)
			return false
		}
	}

	// Check if namespace has service engine group configuration
	hasSEG, err := w.namespaceHasSEG(cluster.GetNamespace())
	if err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to check SEG configuration for namespace %s: %v",
			cluster.GetNamespace(), err)
		return false
	}

	if !hasSEG {
		utils.AviLog.Debugf("VKS webhook: namespace %s does not have SEG configuration", cluster.GetNamespace())
		return false
	}

	utils.AviLog.Infof("VKS webhook: cluster %s/%s is eligible for VKS management",
		cluster.GetNamespace(), cluster.GetName())
	return true
}

// namespaceHasSEG checks if a namespace has service engine group configuration
func (w *VKSClusterWebhook) namespaceHasSEG(namespaceName string) (bool, error) {
	namespace, err := w.Client.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get namespace %s: %w", namespaceName, err)
	}

	// Check for service engine group annotation
	if namespace.Annotations != nil {
		if _, exists := namespace.Annotations["vmware-system-csi/serviceenginegroup"]; exists {
			return true, nil
		}
	}

	return false, nil
}

// createVKSLabelPatch creates a JSON patch to add the VKS managed label
func (w *VKSClusterWebhook) createVKSLabelPatch(cluster *unstructured.Unstructured) ([]map[string]interface{}, error) {
	var patches []map[string]interface{}

	// Check if labels already exist
	if cluster.GetLabels() == nil {
		// No labels exist, create the labels map with our label
		patches = append(patches, map[string]interface{}{
			"op":   "add",
			"path": "/metadata/labels",
			"value": map[string]string{
				ingestion.VKSManagedLabel: ingestion.VKSManagedLabelValueTrue,
			},
		})
	} else {
		// Labels exist, add our label to them
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + escapeJSONPointer(ingestion.VKSManagedLabel),
			"value": ingestion.VKSManagedLabelValueTrue,
		})
	}

	return patches, nil
}

// escapeJSONPointer escapes special characters for JSON Pointer paths
func escapeJSONPointer(s string) string {
	// Replace ~ with ~0 and / with ~1 as per RFC 6901
	result := ""
	for _, char := range s {
		switch char {
		case '~':
			result += "~0"
		case '/':
			result += "~1"
		default:
			result += string(char)
		}
	}
	return result
}
