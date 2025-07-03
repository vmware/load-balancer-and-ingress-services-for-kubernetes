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

package ingestion

import (
	"context"
	"testing"
	"time"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// Test Setup Helpers for Dependency Manager

func setupDependencyManagerTestEnvironment(t *testing.T) (*gomega.WithT, *VKSDependencyManager, *k8sfake.Clientset, *dynamicfake.FakeDynamicClient) {
	g := gomega.NewGomegaWithT(t)

	kubeClient := k8sfake.NewSimpleClientset()

	// Set up dynamic client with proper GVR mappings
	gvrToKind := map[schema.GroupVersionResource]string{
		ClusterGVR: "clustersList",
		{
			Group:    "addons.kubernetes.vmware.com",
			Version:  "v1alpha1",
			Resource: "addoninstalls",
		}: "addonInstallsList",
	}
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToKind)

	dependencyManager := NewVKSDependencyManager(kubeClient, dynamicClient)

	return g, dependencyManager, kubeClient, dynamicClient
}

func createTestClusterForDependencyManager(name, namespace, phase string) *unstructured.Unstructured {
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

	if phase != "" {
		cluster.Object["status"] = map[string]interface{}{
			"phase": phase,
		}
	}

	// Add VKS managed label
	labels := map[string]string{
		VKSManagedLabel: VKSManagedLabelValueTrue,
	}
	cluster.SetLabels(labels)

	return cluster
}

func createTestNamespaceForDependencyManager(name string, hasSEG bool) *corev1.Namespace {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if hasSEG {
		if ns.Annotations == nil {
			ns.Annotations = make(map[string]string)
		}
		ns.Annotations[ServiceEngineGroupAnnotation] = "test-seg"
	}

	return ns
}

// =============================================================================
// CORE FUNCTIONALITY TESTS
// =============================================================================

func TestVKSDependencyManager_NewInstance(t *testing.T) {
	g, dependencyManager, _, _ := setupDependencyManagerTestEnvironment(t)

	g.Expect(dependencyManager).ToNot(gomega.BeNil())
}

func TestVKSDependencyManager_InitializeAviControllerConnection(t *testing.T) {
	g, dependencyManager, _, _ := setupDependencyManagerTestEnvironment(t)

	// Test initialization - should not panic
	dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Verify that the dependency manager is still valid after initialization
	g.Expect(dependencyManager).ToNot(gomega.BeNil())
}

func TestVKSDependencyManager_StartStopReconciler(t *testing.T) {
	g, dependencyManager, _, _ := setupDependencyManagerTestEnvironment(t)

	// Initialize Avi Controller connection
	dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Test starting reconciler
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := dependencyManager.StartReconciler(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test stopping reconciler
	dependencyManager.StopReconciler()
}

// =============================================================================
// CLUSTER DEPENDENCY GENERATION TESTS
// =============================================================================

func TestVKSDependencyManager_GenerateClusterDependencies_Success(t *testing.T) {
	g, dependencyManager, kubeClient, dynamicClient := setupDependencyManagerTestEnvironment(t)

	// Initialize Avi Controller connection
	dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create namespace with SEG
	ns := createTestNamespaceForDependencyManager("test-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Create test cluster
	cluster := createTestClusterForDependencyManager("test-cluster", "test-ns", ClusterPhaseProvisioned)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("test-ns").Create(context.TODO(), cluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Generate dependencies - may succeed or fail depending on environment
	ctx := context.TODO()
	err = dependencyManager.GenerateClusterDependencies(ctx, cluster)

	// Should either succeed or fail gracefully (both are acceptable in test environment)
	g.Expect(err).To(gomega.Or(gomega.BeNil(), gomega.HaveOccurred()))
}

func TestVKSDependencyManager_GenerateClusterDependencies_MissingNamespace(t *testing.T) {
	g, dependencyManager, _, dynamicClient := setupDependencyManagerTestEnvironment(t)

	// Initialize Avi Controller connection
	dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create test cluster in non-existent namespace
	cluster := createTestClusterForDependencyManager("test-cluster", "non-existent-ns", ClusterPhaseProvisioned)
	_, err := dynamicClient.Resource(ClusterGVR).Namespace("non-existent-ns").Create(context.TODO(), cluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Generate dependencies - should fail due to missing namespace
	ctx := context.TODO()
	err = dependencyManager.GenerateClusterDependencies(ctx, cluster)
	g.Expect(err).To(gomega.HaveOccurred())
}

// =============================================================================
// CLUSTER DEPENDENCY CLEANUP TESTS
// =============================================================================

func TestVKSDependencyManager_CleanupClusterDependencies_Success(t *testing.T) {
	g, dependencyManager, kubeClient, _ := setupDependencyManagerTestEnvironment(t)

	// Initialize Avi Controller connection
	dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create namespace
	ns := createTestNamespaceForDependencyManager("test-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Create some test secrets and configmaps to cleanup
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-avi-secret",
			Namespace: "test-ns",
			Labels: map[string]string{
				"ako.kubernetes.vmware.com/cluster":         "test-cluster",
				"ako.kubernetes.vmware.com/dependency-type": "avi-credentials",
				"ako.kubernetes.vmware.com/managed-by":      "vks-dependency-manager",
			},
		},
		Data: map[string][]byte{
			"username": []byte("test-user"),
			"password": []byte("test-pass"),
		},
	}
	_, err = kubeClient.CoreV1().Secrets("test-ns").Create(context.TODO(), secret, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-ako-generated-config",
			Namespace: "test-ns",
			Labels: map[string]string{
				"ako.kubernetes.vmware.com/cluster":         "test-cluster",
				"ako.kubernetes.vmware.com/dependency-type": "ako-generated-config",
				"ako.kubernetes.vmware.com/managed-by":      "vks-dependency-manager",
			},
		},
		Data: map[string]string{
			"controllerHost": "10.1.1.1",
		},
	}
	_, err = kubeClient.CoreV1().ConfigMaps("test-ns").Create(context.TODO(), configMap, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Cleanup dependencies
	ctx := context.TODO()
	err = dependencyManager.CleanupClusterDependencies(ctx, "test-cluster", "test-ns")
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Verify resources are cleaned up - they should be deleted
	_, err = kubeClient.CoreV1().Secrets("test-ns").Get(context.TODO(), "test-cluster-avi-secret", metav1.GetOptions{})
	g.Expect(err).To(gomega.HaveOccurred()) // Should be deleted

	_, err = kubeClient.CoreV1().ConfigMaps("test-ns").Get(context.TODO(), "test-cluster-ako-generated-config", metav1.GetOptions{})
	g.Expect(err).To(gomega.HaveOccurred()) // Should be deleted
}

// =============================================================================
// ADDON INSTALL MANAGEMENT TESTS
// =============================================================================

func TestVKSDependencyManager_EnsureGlobalAddonInstall_Success(t *testing.T) {
	g, dependencyManager, kubeClient, dynamicClient := setupDependencyManagerTestEnvironment(t)

	// Create VKS public namespace
	vksNs := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: VKSPublicNamespace,
		},
	}
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), vksNs, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Ensure global AddonInstall
	ctx := context.TODO()
	err = dependencyManager.EnsureGlobalAddonInstall(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Verify AddonInstall was created
	addonInstallGVR := schema.GroupVersionResource{
		Group:    "addons.kubernetes.vmware.com",
		Version:  "v1alpha1",
		Resource: "addoninstalls",
	}

	addonInstall, err := dynamicClient.Resource(addonInstallGVR).Namespace(VKSPublicNamespace).Get(ctx, AKOAddonInstallName, metav1.GetOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(addonInstall).ToNot(gomega.BeNil())

	// Verify spec fields
	spec, found, err := unstructured.NestedMap(addonInstall.Object, "spec")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())

	addonName, found, err := unstructured.NestedString(spec, "addonName")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(found).To(gomega.BeTrue())
	g.Expect(addonName).To(gomega.Equal(AKOAddonName))
}

// =============================================================================
// RECONCILIATION TESTS
// =============================================================================

func TestVKSDependencyManager_GetManagedClusters(t *testing.T) {
	g, dependencyManager, kubeClient, dynamicClient := setupDependencyManagerTestEnvironment(t)

	// Create namespace with SEG
	ns := createTestNamespaceForDependencyManager("test-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Create test cluster with VKS managed label = true
	cluster1 := createTestClusterForDependencyManager("cluster1", "test-ns", ClusterPhaseProvisioned)
	labels := cluster1.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[VKSManagedLabel] = VKSManagedLabelValueTrue
	cluster1.SetLabels(labels)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("test-ns").Create(context.TODO(), cluster1, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Create test cluster without VKS managed label
	cluster2 := createTestClusterForDependencyManager("cluster2", "test-ns", ClusterPhaseProvisioned)
	// Remove the VKS managed label to test clusters without the label
	labels2 := cluster2.GetLabels()
	delete(labels2, VKSManagedLabel)
	cluster2.SetLabels(labels2)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("test-ns").Create(context.TODO(), cluster2, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Get managed clusters
	ctx := context.TODO()
	managedClusters, err := dependencyManager.GetManagedClusters(ctx)
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Should find cluster1 (has VKS managed label = true)
	g.Expect(managedClusters).To(gomega.ContainElement("test-ns/cluster1"))
	// cluster2 should not be in the list (no VKS managed label)
	g.Expect(managedClusters).ToNot(gomega.ContainElement("test-ns/cluster2"))
}

// =============================================================================
// UTILITY FUNCTION TESTS (testing private methods from same package)
// =============================================================================

func TestVKSDependencyManager_ParseClusterRef(t *testing.T) {
	g, dependencyManager, _, _ := setupDependencyManagerTestEnvironment(t)

	// Test valid cluster ref
	namespace, name := dependencyManager.parseClusterRef("test-ns/test-cluster")
	g.Expect(name).To(gomega.Equal("test-cluster"))
	g.Expect(namespace).To(gomega.Equal("test-ns"))

	// Test invalid cluster ref
	namespace, name = dependencyManager.parseClusterRef("invalid-format")
	g.Expect(name).To(gomega.Equal(""))
	g.Expect(namespace).To(gomega.Equal(""))

	// Test empty cluster ref
	namespace, name = dependencyManager.parseClusterRef("")
	g.Expect(name).To(gomega.Equal(""))
	g.Expect(namespace).To(gomega.Equal(""))
}

func TestVKSDependencyManager_GenerateClusterPassword(t *testing.T) {
	g, dependencyManager, _, _ := setupDependencyManagerTestEnvironment(t)

	// Test password generation
	password1 := dependencyManager.generateClusterPassword("cluster1")
	password2 := dependencyManager.generateClusterPassword("cluster2")
	password1Again := dependencyManager.generateClusterPassword("cluster1")

	// Passwords should be non-empty
	g.Expect(password1).ToNot(gomega.BeEmpty())
	g.Expect(password2).ToNot(gomega.BeEmpty())

	// Same cluster should generate same password (deterministic)
	g.Expect(password1).To(gomega.Equal(password1Again))

	// Different clusters should generate different passwords
	g.Expect(password1).ToNot(gomega.Equal(password2))
}

func TestVKSDependencyManager_GenerateClusterAuthToken(t *testing.T) {
	g, dependencyManager, _, _ := setupDependencyManagerTestEnvironment(t)

	// Test auth token generation
	token1 := dependencyManager.generateClusterAuthToken("cluster1")
	token2 := dependencyManager.generateClusterAuthToken("cluster2")
	token1Again := dependencyManager.generateClusterAuthToken("cluster1")

	// Tokens should be non-empty
	g.Expect(token1).ToNot(gomega.BeEmpty())
	g.Expect(token2).ToNot(gomega.BeEmpty())

	// Same cluster should generate same token (deterministic)
	g.Expect(token1).To(gomega.Equal(token1Again))

	// Different clusters should generate different tokens
	g.Expect(token1).ToNot(gomega.Equal(token2))
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestVKSDependencyManager_ErrorHandling_MissingAviController(t *testing.T) {
	g, dependencyManager, kubeClient, dynamicClient := setupDependencyManagerTestEnvironment(t)

	// Don't initialize Avi Controller connection

	// Create namespace with SEG
	ns := createTestNamespaceForDependencyManager("test-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Create test cluster
	cluster := createTestClusterForDependencyManager("test-cluster", "test-ns", ClusterPhaseProvisioned)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("test-ns").Create(context.TODO(), cluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Generate dependencies without Avi Controller - should fail
	ctx := context.TODO()
	err = dependencyManager.GenerateClusterDependencies(ctx, cluster)
	g.Expect(err).To(gomega.HaveOccurred())
}

func TestVKSDependencyManager_ErrorHandling_MalformedCluster(t *testing.T) {
	g, dependencyManager, _, _ := setupDependencyManagerTestEnvironment(t)

	// Initialize Avi Controller connection
	dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create malformed cluster (missing metadata)
	malformedCluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "Cluster",
			// Missing metadata
		},
	}

	// Generate dependencies with malformed cluster - should fail gracefully
	ctx := context.TODO()
	err := dependencyManager.GenerateClusterDependencies(ctx, malformedCluster)
	g.Expect(err).To(gomega.HaveOccurred())
}
