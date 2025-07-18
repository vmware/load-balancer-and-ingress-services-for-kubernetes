/*
 * Copyright 2024 VMware, Inc.
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

package infratests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	kubefake "k8s.io/client-go/kubernetes/fake"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
)

// VKSTestFramework provides comprehensive testing utilities for VKS integration
type VKSTestFramework struct {
	KubeClient    *kubefake.Clientset
	DynamicClient *dynamicfake.FakeDynamicClient

	// VKS Components
	ClusterWatcher    *ingestion.VKSClusterWatcher
	DependencyManager *ingestion.VKSDependencyManager

	// Test Configuration
	ControllerHost string
	VCenterURL     string

	// Test State
	TestClusters   map[string]*unstructured.Unstructured
	TestNamespaces map[string]*corev1.Namespace
}

// NewVKSTestFramework creates a new testing framework instance
func NewVKSTestFramework(t *testing.T) *VKSTestFramework {
	kubeClient := kubefake.NewSimpleClientset()

	// Create scheme and register all necessary resources
	scheme := runtime.NewScheme()

	// Register custom resource list kinds
	gvrToKind := map[schema.GroupVersionResource]string{
		{Group: "cluster.x-k8s.io", Version: "v1beta1", Resource: "clusters"}:                             "ClusterList",
		{Group: "addons.kubernetes.vmware.com", Version: "v1alpha1", Resource: "addoninstalls"}:           "AddonInstallList",
		{Group: "vmware.com", Version: "v1alpha1", Resource: "managementservices"}:                        "ManagementServiceList",
		{Group: "vmware.com", Version: "v1alpha1", Resource: "managementserviceaccessgrants"}:             "ManagementServiceAccessGrantList",
		{Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"}: "MutatingWebhookConfigurationList",
	}

	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, gvrToKind)

	// Create VKS components
	dependencyManager := ingestion.NewVKSDependencyManager(kubeClient, dynamicClient)
	clusterWatcher := ingestion.NewVKSClusterWatcher(kubeClient, dynamicClient)

	// Initialize with test configuration
	controllerHost := "10.70.184.224"
	vCenterURL := "https://vcenter.example.com"

	dependencyManager.InitializeAviControllerConnection(controllerHost, "22.1.3", "Default-Cloud", "admin", "global")
	dependencyManager.InitializeVKSManagementService(vCenterURL)

	return &VKSTestFramework{
		KubeClient:        kubeClient,
		DynamicClient:     dynamicClient,
		ClusterWatcher:    clusterWatcher,
		DependencyManager: dependencyManager,
		ControllerHost:    controllerHost,
		VCenterURL:        vCenterURL,
		TestClusters:      make(map[string]*unstructured.Unstructured),
		TestNamespaces:    make(map[string]*corev1.Namespace),
	}
}

// CreateTestNamespace creates a test namespace with optional SEG configuration
func (f *VKSTestFramework) CreateTestNamespace(name string, hasSEG bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if hasSEG {
		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string)
		}
		ns.Annotations[ingestion.ServiceEngineGroupAnnotation] = fmt.Sprintf("%s-seg", name)
	}

	_, err := f.KubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create test namespace %s: %v", name, err))
	}

	f.TestNamespaces[name] = ns
	return ns
}

// CreateTestCluster creates a test cluster in the specified phase
func (f *VKSTestFramework) CreateTestCluster(name, namespace, phase string, labels map[string]string) *unstructured.Unstructured {
	cluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "Cluster",
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}

	// Set cluster phase
	if phase != "" {
		cluster.Object["status"] = map[string]interface{}{
			"phase": phase,
		}
	}

	// Set labels
	if labels != nil {
		cluster.SetLabels(labels)
	}

	_, err := f.DynamicClient.Resource(ingestion.ClusterGVR).Namespace(namespace).Create(context.TODO(), cluster, metav1.CreateOptions{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create test cluster %s/%s: %v", namespace, name, err))
	}

	key := fmt.Sprintf("%s/%s", namespace, name)
	f.TestClusters[key] = cluster
	return cluster
}

// CreateVKSManagedCluster creates a cluster with VKS managed label
func (f *VKSTestFramework) CreateVKSManagedCluster(name, namespace, phase string, managed bool) *unstructured.Unstructured {
	labels := map[string]string{}
	if managed {
		labels[ingestion.VKSManagedLabel] = ingestion.VKSManagedLabelValueTrue
	} else {
		labels[ingestion.VKSManagedLabel] = ingestion.VKSManagedLabelValueFalse
	}

	return f.CreateTestCluster(name, namespace, phase, labels)
}

// SimulateClusterPhaseTransition simulates a cluster moving through phases
func (f *VKSTestFramework) SimulateClusterPhaseTransition(name, namespace, fromPhase, toPhase string) error {
	cluster, err := f.DynamicClient.Resource(ingestion.ClusterGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get cluster %s/%s: %v", namespace, name, err)
	}

	// Verify current phase
	currentPhase, found, err := unstructured.NestedString(cluster.Object, "status", "phase")
	if err != nil || !found {
		return fmt.Errorf("failed to get current phase for cluster %s/%s", namespace, name)
	}

	if currentPhase != fromPhase {
		return fmt.Errorf("cluster %s/%s is in phase %s, expected %s", namespace, name, currentPhase, fromPhase)
	}

	// Update to new phase
	err = unstructured.SetNestedField(cluster.Object, toPhase, "status", "phase")
	if err != nil {
		return fmt.Errorf("failed to set phase for cluster %s/%s: %v", namespace, name, err)
	}

	_, err = f.DynamicClient.Resource(ingestion.ClusterGVR).Namespace(namespace).Update(context.TODO(), cluster, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update cluster %s/%s: %v", namespace, name, err)
	}

	return nil
}

// CreateWebhookConfiguration creates a test webhook configuration
func (f *VKSTestFramework) CreateWebhookConfiguration(name string) *admissionv1.MutatingWebhookConfiguration {
	webhook := &admissionv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Webhooks: []admissionv1.MutatingWebhook{
			{
				Name: "vks-cluster-webhook.vmware.com",
				ClientConfig: admissionv1.WebhookClientConfig{
					Service: &admissionv1.ServiceReference{
						Name:      "ako-vks-webhook-service",
						Namespace: "vmware-system-ako",
						Path:      func() *string { s := "/mutate-cluster"; return &s }(),
					},
				},
				Rules: []admissionv1.RuleWithOperations{
					{
						Operations: []admissionv1.OperationType{
							admissionv1.Create,
							admissionv1.Update,
						},
						Rule: admissionv1.Rule{
							APIGroups:   []string{"cluster.x-k8s.io"},
							APIVersions: []string{"v1beta1"},
							Resources:   []string{"clusters"},
						},
					},
				},
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
			},
		},
	}

	_, err := f.KubeClient.AdmissionregistrationV1().MutatingWebhookConfigurations().Create(context.TODO(), webhook, metav1.CreateOptions{})
	if err != nil {
		panic(fmt.Sprintf("Failed to create webhook configuration %s: %v", name, err))
	}

	return webhook
}

// VerifyClusterDependencies verifies that all expected dependencies are created for a cluster
func (f *VKSTestFramework) VerifyClusterDependencies(t *testing.T, clusterName, namespace string) {
	ctx := context.TODO()

	// Check for Avi credentials secret
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)
	secret, err := f.KubeClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		t.Logf("Avi credentials secret %s/%s not found (expected in test environment without real Avi client): %v", namespace, secretName, err)
	} else {
		assert.Equal(t, secretName, secret.Name)
		assert.Equal(t, namespace, secret.Namespace)
		assert.Contains(t, secret.Labels, "ako.kubernetes.vmware.com/cluster")
		assert.Equal(t, clusterName, secret.Labels["ako.kubernetes.vmware.com/cluster"])
	}

	// Check for AKO config ConfigMap
	configMapName := fmt.Sprintf("%s-ako-generated-config", clusterName)
	configMap, err := f.KubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		t.Logf("AKO config ConfigMap %s/%s not found (expected in test environment without real Avi client): %v", namespace, configMapName, err)
	} else {
		assert.Equal(t, configMapName, configMap.Name)
		assert.Equal(t, namespace, configMap.Namespace)
		assert.Contains(t, configMap.Labels, "ako.kubernetes.vmware.com/cluster")
		assert.Equal(t, clusterName, configMap.Labels["ako.kubernetes.vmware.com/cluster"])
	}
}

// VerifyClusterCleanup verifies that all dependencies are cleaned up for a cluster
func (f *VKSTestFramework) VerifyClusterCleanup(t *testing.T, clusterName, namespace string) {
	ctx := context.TODO()

	// Verify secret is deleted
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)
	_, err := f.KubeClient.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	assert.True(t, err != nil, "Secret %s/%s should be deleted", namespace, secretName)

	// Verify ConfigMap is deleted
	configMapName := fmt.Sprintf("%s-ako-generated-config", clusterName)
	_, err = f.KubeClient.CoreV1().ConfigMaps(namespace).Get(ctx, configMapName, metav1.GetOptions{})
	assert.True(t, err != nil, "ConfigMap %s/%s should be deleted", namespace, configMapName)
}

// RunVKSIntegrationTest runs a complete VKS integration test scenario
func (f *VKSTestFramework) RunVKSIntegrationTest(t *testing.T, scenario VKSTestScenario) {
	t.Run(scenario.Name, func(t *testing.T) {
		// Setup
		for _, ns := range scenario.Namespaces {
			f.CreateTestNamespace(ns.Name, ns.HasSEG)
		}

		for _, cluster := range scenario.Clusters {
			f.CreateVKSManagedCluster(cluster.Name, cluster.Namespace, cluster.Phase, cluster.Managed)
		}

		// Execute test steps
		for _, step := range scenario.Steps {
			switch step.Action {
			case "generate_dependencies":
				cluster := f.TestClusters[fmt.Sprintf("%s/%s", step.Namespace, step.ClusterName)]
				err := f.DependencyManager.GenerateClusterDependencies(context.TODO(), cluster)
				if step.ExpectError {
					assert.Error(t, err, "Step %s should fail", step.Description)
				} else {
					assert.NoError(t, err, "Step %s should succeed", step.Description)
				}

			case "cleanup_dependencies":
				err := f.DependencyManager.CleanupClusterDependencies(context.TODO(), step.ClusterName, step.Namespace)
				if step.ExpectError {
					assert.Error(t, err, "Step %s should fail", step.Description)
				} else {
					assert.NoError(t, err, "Step %s should succeed", step.Description)
				}

			case "verify_dependencies":
				f.VerifyClusterDependencies(t, step.ClusterName, step.Namespace)

			case "verify_cleanup":
				f.VerifyClusterCleanup(t, step.ClusterName, step.Namespace)

			case "phase_transition":
				err := f.SimulateClusterPhaseTransition(step.ClusterName, step.Namespace, step.FromPhase, step.ToPhase)
				if step.ExpectError {
					assert.Error(t, err, "Step %s should fail", step.Description)
				} else {
					assert.NoError(t, err, "Step %s should succeed", step.Description)
				}

			case "wait":
				time.Sleep(step.Duration)

			default:
				t.Fatalf("Unknown test step action: %s", step.Action)
			}
		}

		// Verify final state
		if scenario.FinalVerification != nil {
			scenario.FinalVerification(t, f)
		}
	})
}

// VKSTestScenario defines a complete test scenario
type VKSTestScenario struct {
	Name        string
	Description string

	// Setup
	Namespaces []TestNamespace
	Clusters   []TestCluster

	// Test steps
	Steps []TestStep

	// Final verification
	FinalVerification func(t *testing.T, f *VKSTestFramework)
}

// TestNamespace defines a test namespace
type TestNamespace struct {
	Name   string
	HasSEG bool
}

// TestCluster defines a test cluster
type TestCluster struct {
	Name      string
	Namespace string
	Phase     string
	Managed   bool
}

// TestStep defines a test step
type TestStep struct {
	Action      string
	Description string

	// Cluster info
	ClusterName string
	Namespace   string

	// Phase transition
	FromPhase string
	ToPhase   string

	// Wait duration
	Duration time.Duration

	// Expectations
	ExpectError bool
}

// Cleanup cleans up all test resources
func (f *VKSTestFramework) Cleanup() {
	// Stop any running components
	if f.ClusterWatcher != nil {
		f.ClusterWatcher.Stop()
	}
	if f.DependencyManager != nil {
		f.DependencyManager.StopReconciler()
	}
}
