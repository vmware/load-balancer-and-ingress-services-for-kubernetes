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
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=cert-manager.io,resources=certificates;issuers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.kubernetes.vmware.com,resources=addoninstalls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=iaas.vmware.com,resources=capabilities,verbs=get;list;watch
// +kubebuilder:rbac:groups=run.tanzu.vmware.com,resources=clusterbootstraps,verbs=get;list;watch

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
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	internalLib "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"
)

const (
	VKSManagedLabel           = "ako.kubernetes.vmware.com/install"
	VKSManagedLabelValueTrue  = "true"
	VKSManagedLabelValueFalse = "false"

	WebhookName            = "ako-vks-cluster-webhook"
	WebhookServiceName     = "vmware-system-ako-ako-vks-webhook-service"
	WebhookPath            = "/ako-vks-mutate-cluster-x-k8s-io"
	WebhookIssuerName      = "ako-vks-webhook-selfsigned-issuer"
	WebhookCertificateName = "ako-vks-webhook-serving-cert"
	WebhookCertSecretName  = "ako-vks-webhook-serving-cert"
)

var vksWebhookOnce sync.Once

func StartVKSWebhook(kubeClient kubernetes.Interface, stopCh <-chan struct{}) {
	vksWebhookOnce.Do(func() {
		utils.AviLog.Infof("VKS webhook: capability activated, starting webhook with infinite retry")

		retryInterval := 10 * time.Second

		for {
			if err := setupAndStartWebhook(kubeClient, stopCh); err != nil {
				utils.AviLog.Warnf("VKS webhook: setup failed, will retry in %v: %v", retryInterval, err)

				// Wait before retry, but also check for shutdown
				select {
				case <-stopCh:
					utils.AviLog.Infof("VKS webhook: shutdown signal received during retry wait")
					// Clean up webhook resources on shutdown
					if err := CleanupAllWebhookResources(kubeClient); err != nil {
						utils.AviLog.Warnf("VKS webhook: failed to cleanup resources during shutdown: %v", err)
					}
					return
				case <-time.After(retryInterval):
					// Continue to next retry
					continue
				}
			} else {
				// Server returned without error - this means graceful shutdown
				utils.AviLog.Infof("VKS webhook: server stopped gracefully")
				// Clean up webhook resources on graceful shutdown
				if err := CleanupAllWebhookResources(kubeClient); err != nil {
					utils.AviLog.Warnf("VKS webhook: failed to cleanup resources during graceful shutdown: %v", err)
				}
				return
			}
		}
	})
}

// setupAndStartWebhook attempts to setup certificates, configuration, and start the webhook server
// Returns error if any step fails, allowing the caller to retry
func setupAndStartWebhook(kubeClient kubernetes.Interface, stopCh <-chan struct{}) error {
	if err := ensureWebhookCertificates(kubeClient); err != nil {
		return fmt.Errorf("failed to ensure certificates: %v", err)
	}

	if err := CreateWebhookConfiguration(kubeClient); err != nil {
		return fmt.Errorf("failed to create webhook configuration: %v", err)
	}

	vksWebhook := NewVKSClusterWebhook(kubeClient)
	if err := StartWebhookServer(vksWebhook, stopCh); err != nil {
		return fmt.Errorf("webhook server failed: %v", err)
	}

	return nil
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
		utils.AviLog.Infof("VKS webhook: namespace %s does not have SEG configuration", cluster.GetNamespace())
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
			"value": map[string]interface{}{
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

// Certificate Management Functions

// isCertManagerAvailable checks if cert-manager CRDs are available
func isCertManagerAvailable(kubeClient kubernetes.Interface) bool {
	discoveryClient := kubeClient.Discovery()

	_, err := discoveryClient.ServerResourcesForGroupVersion("cert-manager.io/v1")
	if err != nil {
		utils.AviLog.Infof("cert-manager.io/v1 not available: %v", err)
		return false
	}

	return true
}

// createCertManagerResources creates Issuer and Certificate resources for the webhook
func createCertManagerResources(kubeClient kubernetes.Interface, namespace string) error {
	if !isCertManagerAvailable(kubeClient) {
		return fmt.Errorf("cert-manager CRDs not available")
	}

	dynamicClient := internalLib.GetDynamicClientSet()
	if dynamicClient == nil {
		return fmt.Errorf("dynamic client not available - AKO-infra not properly initialized")
	}

	if err := createWebhookIssuer(dynamicClient, namespace); err != nil {
		return fmt.Errorf("failed to create issuer: %v", err)
	}

	if err := createWebhookCertificate(dynamicClient, namespace); err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	return nil
}

// createWebhookIssuer creates a self-signed issuer for webhook certificates
func createWebhookIssuer(dynamicClient dynamic.Interface, namespace string) error {
	issuerGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "issuers",
	}

	issuerName := WebhookIssuerName

	// Check if issuer already exists
	_, err := dynamicClient.Resource(issuerGVR).Namespace(namespace).Get(
		context.TODO(), issuerName, metav1.GetOptions{})
	if err == nil {
		utils.AviLog.Infof("Issuer %s already exists", issuerName)
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("error checking for existing issuer: %v", err)
	}

	// Create issuer
	issuer := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Issuer",
			"metadata": map[string]interface{}{
				"name":      issuerName,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"selfSigned": map[string]interface{}{},
			},
		},
	}

	_, err = dynamicClient.Resource(issuerGVR).Namespace(namespace).Create(
		context.TODO(), issuer, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create issuer: %v", err)
	}

	utils.AviLog.Infof("Created webhook issuer: %s", issuerName)
	return nil
}

// createWebhookCertificate creates a certificate for the webhook service
func createWebhookCertificate(dynamicClient dynamic.Interface, namespace string) error {
	certificateGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	certificateName := WebhookCertificateName
	serviceName := WebhookServiceName

	_, err := dynamicClient.Resource(certificateGVR).Namespace(namespace).Get(
		context.TODO(), certificateName, metav1.GetOptions{})
	if err == nil {
		utils.AviLog.Infof("Certificate %s already exists", certificateName)
		return nil
	}

	if !errors.IsNotFound(err) {
		return fmt.Errorf("error checking for existing certificate: %v", err)
	}

	certificate := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]interface{}{
				"name":      certificateName,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"dnsNames": []interface{}{
					fmt.Sprintf("%s.%s.svc", serviceName, namespace),
					fmt.Sprintf("%s.%s.svc.cluster.local", serviceName, namespace),
				},
				"issuerRef": map[string]interface{}{
					"kind": "Issuer",
					"name": WebhookIssuerName,
				},
				"secretName":  certificateName,
				"duration":    "2160h", // 90 days
				"renewBefore": "720h",  // 30 days
			},
		},
	}

	_, err = dynamicClient.Resource(certificateGVR).Namespace(namespace).Create(
		context.TODO(), certificate, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	utils.AviLog.Infof("Created webhook certificate: %s", certificateName)
	return nil
}

func waitForWebhookCertificates(kubeClient kubernetes.Interface, namespace string) error {
	secretName := os.Getenv("VKS_WEBHOOK_CERT_SECRET")
	if secretName == "" {
		secretName = WebhookCertSecretName
	}

	utils.AviLog.Infof("Checking webhook certificate secret '%s'...", secretName)

	secret, err := kubeClient.CoreV1().Secrets(namespace).Get(
		context.TODO(), secretName, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("certificate secret '%s' not found", secretName)
		}
		return fmt.Errorf("error checking certificate secret: %v", err)
	}

	// Verify secret has required certificate data
	if _, hasCert := secret.Data["tls.crt"]; !hasCert {
		return fmt.Errorf("certificate secret missing 'tls.crt' key")
	}
	if _, hasKey := secret.Data["tls.key"]; !hasKey {
		return fmt.Errorf("certificate secret missing 'tls.key' key")
	}
	if _, hasCA := secret.Data["ca.crt"]; !hasCA {
		return fmt.Errorf("certificate secret missing 'ca.crt' key")
	}

	utils.AviLog.Infof("Webhook certificates are ready!")
	return nil
}

func ensureWebhookCertificates(kubeClient kubernetes.Interface) error {
	namespace := utils.GetAKONamespace()

	if err := createCertManagerResources(kubeClient, namespace); err != nil {
		return fmt.Errorf("cert-manager not ready: %v", err)
	}

	if err := waitForWebhookCertificates(kubeClient, namespace); err != nil {
		return fmt.Errorf("certificate not ready: %v", err)
	}

	utils.AviLog.Infof("Webhook certificates are ready")
	return nil
}

// StartWebhookServer starts the webhook server for VKS cluster admission
func StartWebhookServer(webhook *VKSClusterWebhook, stopCh <-chan struct{}) error {
	port := os.Getenv("VKS_WEBHOOK_PORT")
	if port == "" {
		port = "9998"
	}

	certDir := os.Getenv("VKS_WEBHOOK_CERT_DIR")
	if certDir == "" {
		certDir = "/tmp/k8s-webhook-server/serving-certs"
	}

	certPath := filepath.Join(certDir, "tls.crt")
	keyPath := filepath.Join(certDir, "tls.key")

	mux := http.NewServeMux()
	mux.Handle(WebhookPath, webhook)
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
	webhookName := WebhookName
	serviceName := WebhookServiceName
	serviceNamespace := utils.GetAKONamespace()
	webhookPath := WebhookPath
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
						APIVersions: []string{"v1beta2"},
						Resources:   []string{"clusters"},
					},
				}},
				FailurePolicy:           &failurePolicy,
				SideEffects:             &sideEffects,
				AdmissionReviewVersions: []string{"v1", "v1beta2"},
			}

			webhookConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
				ObjectMeta: metav1.ObjectMeta{
					Name: webhookName,
					Annotations: map[string]string{
						"cert-manager.io/inject-ca-from": fmt.Sprintf("%s/%s", serviceNamespace, WebhookCertSecretName),
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

// waitForCertificates waits for the TLS certificates to be available
func waitForCertificates(certDir string, timeout time.Duration) error {
	certPath := filepath.Join(certDir, "tls.crt")
	keyPath := filepath.Join(certDir, "tls.key")

	utils.AviLog.Infof("VKS webhook: waiting for certificates at %s and %s", certPath, keyPath)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	timeoutCh := time.After(timeout)

	for {
		select {
		case <-ticker.C:
			if filesExist(certPath, keyPath) {
				utils.AviLog.Infof("VKS webhook: certificates ready and valid")
				return nil
			}
			utils.AviLog.Infof("VKS webhook: certificates not ready yet, continuing to wait...")
		case <-timeoutCh:
			return fmt.Errorf("timeout waiting for certificates after %v", timeout)
		}
	}
}

// filesExist checks if certificate files exist
func filesExist(certPath, keyPath string) bool {
	if _, err := os.Stat(certPath); err != nil {
		if os.IsNotExist(err) {
			utils.AviLog.Errorf("VKS webhook: certificate file does not exist: %s", certPath)
		} else {
			utils.AviLog.Warnf("VKS webhook: error checking certificate file %s: %v", certPath, err)
		}
		return false
	}
	if _, err := os.Stat(keyPath); err != nil {
		if os.IsNotExist(err) {
			utils.AviLog.Errorf("VKS webhook: key file does not exist: %s", keyPath)
		} else {
			utils.AviLog.Warnf("VKS webhook: error checking key file %s: %v", keyPath, err)
		}
		return false
	}

	utils.AviLog.Infof("VKS webhook: certificate files exist")
	return true
}

// CleanupWebhookConfiguration deletes the MutatingWebhookConfiguration
func CleanupWebhookConfiguration(kubeClient kubernetes.Interface) error {
	webhookName := WebhookName

	utils.AviLog.Infof("VKS webhook: deleting MutatingWebhookConfiguration '%s'", webhookName)

	err := kubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Delete(
		context.TODO(), webhookName, metav1.DeleteOptions{})

	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Infof("VKS webhook: MutatingWebhookConfiguration '%s' already deleted", webhookName)
			return nil
		}
		return fmt.Errorf("failed to delete MutatingWebhookConfiguration '%s': %v", webhookName, err)
	}

	utils.AviLog.Infof("VKS webhook: successfully deleted MutatingWebhookConfiguration '%s'", webhookName)
	return nil
}

// CleanupCertManagerResources deletes the cert-manager Issuer and Certificate resources
func CleanupCertManagerResources(kubeClient kubernetes.Interface) error {
	namespace := utils.GetAKONamespace()

	if !isCertManagerAvailable(kubeClient) {
		utils.AviLog.Infof("VKS webhook: cert-manager CRDs not available, skipping cert-manager resource cleanup")
		return nil
	}

	dynamicClient := internalLib.GetDynamicClientSet()
	if dynamicClient == nil {
		utils.AviLog.Warnf("VKS webhook: dynamic client not available, skipping cert-manager resource cleanup")
		return nil
	}

	issuerGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "issuers",
	}

	certificateGVR := schema.GroupVersionResource{
		Group:    "cert-manager.io",
		Version:  "v1",
		Resource: "certificates",
	}

	certificateName := WebhookCertificateName
	utils.AviLog.Infof("VKS webhook: deleting Certificate '%s'", certificateName)

	err := dynamicClient.Resource(certificateGVR).Namespace(namespace).Delete(
		context.TODO(), certificateName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Infof("VKS webhook: Certificate '%s' already deleted", certificateName)
		} else {
			utils.AviLog.Warnf("VKS webhook: failed to delete Certificate '%s': %v", certificateName, err)
		}
	} else {
		utils.AviLog.Infof("VKS webhook: successfully deleted Certificate '%s'", certificateName)
	}

	issuerName := WebhookIssuerName
	utils.AviLog.Infof("VKS webhook: deleting Issuer '%s'", issuerName)

	err = dynamicClient.Resource(issuerGVR).Namespace(namespace).Delete(
		context.TODO(), issuerName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Infof("VKS webhook: Issuer '%s' already deleted", issuerName)
		} else {
			utils.AviLog.Warnf("VKS webhook: failed to delete Issuer '%s': %v", issuerName, err)
		}
	} else {
		utils.AviLog.Infof("VKS webhook: successfully deleted Issuer '%s'", issuerName)
	}

	return nil
}

// CleanupAllWebhookResources cleans up all webhook-related resources
func CleanupAllWebhookResources(kubeClient kubernetes.Interface) error {
	utils.AviLog.Infof("VKS webhook: starting cleanup of all webhook resources")

	if err := CleanupWebhookConfiguration(kubeClient); err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to cleanup webhook configuration: %v", err)
	}

	if err := CleanupCertManagerResources(kubeClient); err != nil {
		utils.AviLog.Errorf("VKS webhook: failed to cleanup cert-manager resources: %v", err)
	}

	utils.AviLog.Infof("VKS webhook: completed cleanup of all webhook resources")
	return nil
}
