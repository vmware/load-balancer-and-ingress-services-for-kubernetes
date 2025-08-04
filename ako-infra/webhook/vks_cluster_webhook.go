/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
 * All Rights Reserved.
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*   http://www.apache.org/licenses/LICENSE-2.0
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package webhook

// +kubebuilder:rbac:groups=admissionregistration.k8s.io,resources=mutatingwebhookconfigurations,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates;issuers,verbs=get;list;watch;create;update;patch

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes"

	internalLib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const (
	VKSManagedLabel           = "ako.kubernetes.vmware.com/install"
	VKSManagedLabelValueTrue  = "true"
	VKSManagedLabelValueFalse = "false"
)

var vksWebhookOnce sync.Once

func StartVKSWebhook(kubeClient kubernetes.Interface, stopCh <-chan struct{}) {
	vksWebhookOnce.Do(func() {
		utils.AviLog.Infof("VKS webhook: capability activated, starting webhook")
		// Create webhook configuration
		if err := CreateWebhookConfiguration(kubeClient); err != nil {
			utils.AviLog.Fatalf("VKS webhook: failed to create configuration: %v", err)
		}

		// Start webhook server
		vksWebhook := NewVKSClusterWebhook(kubeClient)
		if err := StartWebhookServer(vksWebhook, stopCh); err != nil {
			utils.AviLog.Fatalf("VKS webhook: server failed: %v", err)
		}

		utils.AviLog.Infof("VKS webhook: startup initiated successfully")
	})
}

// VKSClusterWebhook handles admission requests for Cluster objects
// It automatically adds VKS managed AKO Install label to new clusters in Avi enabled namespaces
type VKSClusterWebhook struct {
	client kubernetes.Interface
}

func NewVKSClusterWebhook(client kubernetes.Interface) *VKSClusterWebhook {
	return &VKSClusterWebhook{
		client: client,
	}
}

func (w *VKSClusterWebhook) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to read request body: %v", err)
		http.Error(rw, "failed to read request body", http.StatusBadRequest)
		return
	}

	var admissionReview admissionv1.AdmissionReview
	if err := json.Unmarshal(body, &admissionReview); err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to unmarshal admission review: %v", err)
		http.Error(rw, "failed to unmarshal admission review", http.StatusBadRequest)
		return
	}

	if admissionReview.Request == nil {
		utils.AviLog.Errorf("VKS webhook: admission review request is nil")
		http.Error(rw, "admission review request is nil", http.StatusBadRequest)
		return
	}

	admissionResponse := w.ProcessAdmissionRequest(admissionReview.Request)

	responseAdmissionReview := &admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "admission.k8s.io/v1",
			Kind:       "AdmissionReview",
		},
		Response: admissionResponse,
	}
	responseAdmissionReview.Response.UID = admissionReview.Request.UID

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

func (w *VKSClusterWebhook) ProcessAdmissionRequest(req *admissionv1.AdmissionRequest) *admissionv1.AdmissionResponse {
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

	utils.AviLog.Infof("VKS webhook: processing cluster %s/%s", cluster.GetNamespace(), cluster.GetName())

	if !w.shouldManageCluster(cluster) {
		utils.AviLog.Infof("VKS webhook: skipping cluster %s/%s - not eligible for VKS management",
			cluster.GetNamespace(), cluster.GetName())
		return &admissionv1.AdmissionResponse{
			Allowed: true,
			Result:  &metav1.Status{Message: "cluster not eligible for VKS management"},
		}
	}

	// Create patch to add VKS managed AKO Install label
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
	utils.AviLog.Infof("VKS webhook: added label %s=%s to cluster %s/%s",
		VKSManagedLabel, VKSManagedLabelValueTrue,
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
		if value, exists := labels[VKSManagedLabel]; exists && value == VKSManagedLabelValueFalse {
			utils.AviLog.Infof("VKS webhook: cluster %s/%s opted out with label %s=%s",
				cluster.GetNamespace(), cluster.GetName(), VKSManagedLabel, value)
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

	utils.AviLog.Infof("VKS webhook: cluster %s/%s eligible for VKS management",
		cluster.GetNamespace(), cluster.GetName())
	return true
}

// namespaceHasSEG checks if a namespace has service engine group configuration
func (w *VKSClusterWebhook) namespaceHasSEG(namespaceName string) (bool, error) {
	namespace, err := w.client.CoreV1().Namespaces().Get(context.TODO(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to get namespace %s: %w", namespaceName, err)
	}

	if namespace.Annotations != nil {
		if _, exists := namespace.Annotations[internalLib.WCPSEGroup]; exists {
			return true, nil
		}
	}

	return false, nil
}

// createVKSLabelPatch creates a JSON patch to add the VKS managed AKO Install label
func (w *VKSClusterWebhook) createVKSLabelPatch(cluster *unstructured.Unstructured) ([]map[string]interface{}, error) {
	var patches []map[string]interface{}

	labels := cluster.GetLabels()

	if labels == nil {
		patches = append(patches, map[string]interface{}{
			"op":   "add",
			"path": "/metadata/labels",
			"value": map[string]string{
				VKSManagedLabel: VKSManagedLabelValueTrue,
			},
		})
	} else if existingValue, exists := labels[VKSManagedLabel]; exists {
		if existingValue == VKSManagedLabelValueTrue {
			return patches, nil
		} else {
			patches = append(patches, map[string]interface{}{
				"op":    "replace",
				"path":  "/metadata/labels/" + escapeJSONPointer(VKSManagedLabel),
				"value": VKSManagedLabelValueTrue,
			})
		}
	} else {
		patches = append(patches, map[string]interface{}{
			"op":    "add",
			"path":  "/metadata/labels/" + escapeJSONPointer(VKSManagedLabel),
			"value": VKSManagedLabelValueTrue,
		})
	}

	return patches, nil
}

// escapeJSONPointer escapes special characters for JSON Pointer paths
func escapeJSONPointer(s string) string {
	// Replace ~ with ~0 and / with ~1
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

// StartWebhookServer starts the webhook server for VKS cluster admission
func StartWebhookServer(webhook *VKSClusterWebhook, stopCh <-chan struct{}) error {
	port := os.Getenv("VKS_WEBHOOK_PORT")
	if port == "" {
		port = "9443"
	}

	certDir := os.Getenv("VKS_WEBHOOK_CERT_DIR")
	if certDir == "" {
		certDir = "/tmp/k8s-webhook-server/serving-certs"
	}

	certPath := filepath.Join(certDir, "tls.crt")
	keyPath := filepath.Join(certDir, "tls.key")

	mux := http.NewServeMux()
	mux.Handle("/mutate-cluster-x-k8s-io-v1beta1-cluster", webhook)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Channel to receive server startup errors
	serverErrCh := make(chan error, 1)

	go func() {
		utils.AviLog.Infof("VKS webhook: starting server on port %s with certs from %s", port, certDir)
		if err := server.ListenAndServeTLS(certPath, keyPath); err != nil && err != http.ErrServerClosed {
			utils.AviLog.Errorf("VKS webhook server error: %v", err)
			serverErrCh <- err
		}
	}()

	select {
	case err := <-serverErrCh:
		return fmt.Errorf("VKS webhook server failed to start: %w", err)
	case <-stopCh:
		utils.AviLog.Infof("VKS webhook: shutting down server")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		utils.AviLog.Errorf("VKS webhook server shutdown error: %v", err)
		return err
	}

	utils.AviLog.Infof("VKS webhook: server stopped")
	return nil
}

// CreateWebhookConfiguration creates the MutatingWebhookConfiguration
func CreateWebhookConfiguration(kubeClient kubernetes.Interface) error {
	webhookName := "ako-vks-cluster-webhook"
	serviceName := "ako-vks-webhook-service"
	serviceNamespace := utils.GetAKONamespace()
	webhookPath := "/mutate-cluster-x-k8s-io-v1beta1-cluster"
	failurePolicy := admissionregistrationv1.Fail
	sideEffects := admissionregistrationv1.SideEffectClassNone

	_, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(
		context.TODO(), webhookName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Webhook doesn't exist, create it
			webhook := admissionregistrationv1.MutatingWebhook{
				Name: "vks-cluster-labeling.ako.vmware.com",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					Service: &admissionregistrationv1.ServiceReference{
						Name:      serviceName,
						Namespace: serviceNamespace,
						Path:      &webhookPath,
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{{
					Operations: []admissionregistrationv1.OperationType{
						admissionregistrationv1.Create,
					},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"cluster.x-k8s.io"},
						APIVersions: []string{"v1beta1"},
						Resources:   []string{"clusters"},
					},
				}},
				FailurePolicy:           &failurePolicy,
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
			}

			webhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: webhookName,
					Annotations: map[string]string{
						"cert-manager.io/inject-ca-from": fmt.Sprintf("%s/ako-vks-webhook-serving-cert", serviceNamespace),
					},
				},
				Webhooks: []admissionregistrationv1.MutatingWebhook{webhook},
			}

			_, err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(
				context.TODO(), webhookConfig, metav1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("failed to create MutatingWebhookConfiguration: %v", err)
			}
			utils.AviLog.Infof("VKS webhook: created MutatingWebhookConfiguration '%s'", webhookName)
		} else {
			return fmt.Errorf("error checking MutatingWebhookConfiguration: %v", err)
		}
	} else {
		utils.AviLog.Infof("VKS webhook: MutatingWebhookConfiguration '%s' already exists", webhookName)
	}
	return nil
}
