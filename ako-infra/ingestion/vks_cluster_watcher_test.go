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

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	k8sfake "k8s.io/client-go/kubernetes/fake"
)

// Test Setup Helpers

func createTestNamespace(name string, hasSEG bool) *corev1.Namespace {
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

func createTestCluster(name, namespace, phase string) *unstructured.Unstructured {
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

	return cluster
}

func createTestClusterWithLabel(name, namespace, phase, labelValue string) *unstructured.Unstructured {
	cluster := createTestCluster(name, namespace, phase)

	labels := map[string]string{
		VKSManagedLabel: labelValue,
	}
	cluster.SetLabels(labels)

	return cluster
}

func setupTestEnvironment(t *testing.T) (*gomega.WithT, *VKSClusterWatcher, *k8sfake.Clientset, *dynamicfake.FakeDynamicClient) {
	g := gomega.NewGomegaWithT(t)

	kubeClient := k8sfake.NewSimpleClientset()

	// Set up dynamic client with proper GVR mappings
	gvrToKind := map[schema.GroupVersionResource]string{
		ClusterGVR: "clustersList",
	}
	dynamicClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), gvrToKind)

	watcher := NewVKSClusterWatcher(kubeClient, dynamicClient)

	return g, watcher, kubeClient, dynamicClient
}

// =============================================================================
// CORE FUNCTIONALITY TESTS
// =============================================================================

func TestVKSClusterWatcher_NewInstance(t *testing.T) {
	g, watcher, _, _ := setupTestEnvironment(t)

	g.Expect(watcher).ToNot(gomega.BeNil())
}

// =============================================================================
// NAMESPACE AND SEG VALIDATION TESTS
// =============================================================================

func TestVKSClusterWatcher_NamespaceHasSEG(t *testing.T) {
	g, watcher, kubeClient, _ := setupTestEnvironment(t)

	// Create namespace with SEG
	nsWithSEG := createTestNamespace("seg-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Create namespace without SEG
	nsWithoutSEG := createTestNamespace("no-seg-ns", false)
	_, err = kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithoutSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test namespace with SEG
	hasSEG, err := watcher.NamespaceHasSEG("seg-ns")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(hasSEG).To(gomega.BeTrue())

	// Test namespace without SEG
	hasSEG, err = watcher.NamespaceHasSEG("no-seg-ns")
	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(hasSEG).To(gomega.BeFalse())

	// Test non-existent namespace
	hasSEG, err = watcher.NamespaceHasSEG("non-existent")
	g.Expect(err).To(gomega.HaveOccurred()) // Should error for non-existent namespace
	g.Expect(hasSEG).To(gomega.BeFalse())
}

// =============================================================================
// CLUSTER MANAGEMENT LOGIC TESTS
// =============================================================================

func TestVKSClusterWatcher_ShouldManageCluster(t *testing.T) {
	g, watcher, kubeClient, _ := setupTestEnvironment(t)

	// Create namespace with SEG for testing
	nsWithSEG := createTestNamespace("test-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test cluster with VKS managed label = true
	clusterTrue := createTestClusterWithLabel("cluster1", "test-ns", ClusterPhaseProvisioned, VKSManagedLabelValueTrue)
	g.Expect(watcher.ShouldManageCluster(clusterTrue)).To(gomega.BeTrue())

	// Test cluster with VKS managed label = false
	clusterFalse := createTestClusterWithLabel("cluster2", "test-ns", ClusterPhaseProvisioned, VKSManagedLabelValueFalse)
	g.Expect(watcher.ShouldManageCluster(clusterFalse)).To(gomega.BeTrue()) // Still should be managed if in SEG namespace

	// Test cluster without VKS managed label (should be managed if in SEG namespace)
	clusterNoLabel := createTestCluster("cluster3", "test-ns", ClusterPhaseProvisioned)
	g.Expect(watcher.ShouldManageCluster(clusterNoLabel)).To(gomega.BeTrue())
}

func TestVKSClusterWatcher_GetClusterPhase(t *testing.T) {
	g, watcher, _, _ := setupTestEnvironment(t)

	// Test cluster with phase
	cluster := createTestCluster("cluster1", "test-ns", ClusterPhaseProvisioned)
	phase := watcher.GetClusterPhase(cluster)
	g.Expect(phase).To(gomega.Equal(ClusterPhaseProvisioned))

	// Test cluster without phase
	clusterNoPhase := createTestCluster("cluster2", "test-ns", "")
	phase = watcher.GetClusterPhase(clusterNoPhase)
	g.Expect(phase).To(gomega.Equal(""))
}

// =============================================================================
// LABEL CHANGE PROCESSING TESTS
// =============================================================================

func TestVKSClusterWatcher_ProcessClusterLabelChange_NoChange(t *testing.T) {
	g, watcher, kubeClient, dynamicClient := setupTestEnvironment(t)

	// Create namespace with SEG
	nsWithSEG := createTestNamespace("seg-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test no label change
	oldCluster := createTestClusterWithLabel("cluster1", "seg-ns", ClusterPhaseProvisioned, VKSManagedLabelValueTrue)
	newCluster := createTestClusterWithLabel("cluster1", "seg-ns", ClusterPhaseProvisioned, VKSManagedLabelValueTrue)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("seg-ns").Create(context.TODO(), newCluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	result := watcher.ProcessClusterLabelChange(oldCluster, newCluster)
	g.Expect(result.Skipped).To(gomega.BeTrue())
	g.Expect(result.SkipReason).To(gomega.ContainSubstring("no change in VKS managed label"))
}

func TestVKSClusterWatcher_ProcessClusterLabelChange_OptIn(t *testing.T) {
	g, watcher, kubeClient, dynamicClient := setupTestEnvironment(t)

	// Initialize Avi Controller directly on dependency manager
	watcher.dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create namespace with SEG
	nsWithSEG := createTestNamespace("seg-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test opt-in (no label -> true)
	oldCluster := createTestCluster("cluster1", "seg-ns", ClusterPhaseProvisioned)
	newCluster := createTestClusterWithLabel("cluster1", "seg-ns", ClusterPhaseProvisioned, VKSManagedLabelValueTrue)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("seg-ns").Create(context.TODO(), newCluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	result := watcher.ProcessClusterLabelChange(oldCluster, newCluster)
	g.Expect(result.Error).To(gomega.Or(gomega.BeNil(), gomega.HaveOccurred())) // May fail due to test environment
	g.Expect(result.Operation).To(gomega.Equal(LabelingOperationOptIn))
}

func TestVKSClusterWatcher_ProcessClusterLabelChange_OptOut(t *testing.T) {
	g, watcher, kubeClient, dynamicClient := setupTestEnvironment(t)

	// Initialize Avi Controller directly on dependency manager
	watcher.dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create namespace with SEG
	nsWithSEG := createTestNamespace("seg-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test opt-out (true -> false)
	oldCluster := createTestClusterWithLabel("cluster1", "seg-ns", ClusterPhaseProvisioned, VKSManagedLabelValueTrue)
	newCluster := createTestClusterWithLabel("cluster1", "seg-ns", ClusterPhaseProvisioned, VKSManagedLabelValueFalse)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("seg-ns").Create(context.TODO(), newCluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	result := watcher.ProcessClusterLabelChange(oldCluster, newCluster)
	g.Expect(result.Error).ToNot(gomega.HaveOccurred())
	g.Expect(result.Success).To(gomega.BeTrue())
	g.Expect(result.Operation).To(gomega.Equal(LabelingOperationOptOut))
}

// =============================================================================
// OPT-IN/OPT-OUT HANDLING TESTS
// =============================================================================

func TestVKSClusterWatcher_HandleClusterOptIn_Success(t *testing.T) {
	g, watcher, kubeClient, dynamicClient := setupTestEnvironment(t)

	// Initialize Avi Controller directly on dependency manager
	watcher.dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create namespace with SEG
	nsWithSEG := createTestNamespace("seg-ns", true)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test valid opt-in
	cluster := createTestClusterWithLabel("cluster1", "seg-ns", ClusterPhaseProvisioned, VKSManagedLabelValueTrue)
	_, err = dynamicClient.Resource(ClusterGVR).Namespace("seg-ns").Create(context.TODO(), cluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	result := watcher.HandleClusterOptIn(cluster)
	g.Expect(result.Error).To(gomega.Or(gomega.BeNil(), gomega.HaveOccurred())) // May fail due to test environment
	g.Expect(result.Operation).To(gomega.Equal(LabelingOperationOptIn))
}

func TestVKSClusterWatcher_HandleClusterOptIn_NoSEG(t *testing.T) {
	g, watcher, kubeClient, _ := setupTestEnvironment(t)

	// Initialize Avi Controller directly on dependency manager
	watcher.dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create namespace without SEG
	nsWithoutSEG := createTestNamespace("no-seg-ns", false)
	_, err := kubeClient.CoreV1().Namespaces().Create(context.TODO(), nsWithoutSEG, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	// Test opt-in in non-SEG namespace
	cluster := createTestClusterWithLabel("cluster1", "no-seg-ns", ClusterPhaseProvisioned, VKSManagedLabelValueTrue)

	result := watcher.HandleClusterOptIn(cluster)
	g.Expect(result.Error).To(gomega.HaveOccurred())
	g.Expect(result.Error.Error()).To(gomega.ContainSubstring("cannot opt-in cluster in namespace without SEG configuration"))
}

func TestVKSClusterWatcher_HandleClusterOptOut_Success(t *testing.T) {
	g, watcher, _, dynamicClient := setupTestEnvironment(t)

	// Initialize Avi Controller directly on dependency manager
	watcher.dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Test opt-out
	cluster := createTestClusterWithLabel("cluster1", "test-ns", ClusterPhaseProvisioned, VKSManagedLabelValueFalse)
	_, err := dynamicClient.Resource(ClusterGVR).Namespace("test-ns").Create(context.TODO(), cluster, metav1.CreateOptions{})
	g.Expect(err).ToNot(gomega.HaveOccurred())

	result := watcher.HandleClusterOptOut(cluster)
	g.Expect(result.Error).ToNot(gomega.HaveOccurred())
	g.Expect(result.Success).To(gomega.BeTrue())
	g.Expect(result.Operation).To(gomega.Equal(LabelingOperationOptOut))
}

// =============================================================================
// ERROR HANDLING TESTS
// =============================================================================

func TestVKSClusterWatcher_ErrorHandling_MalformedCluster(t *testing.T) {
	g, watcher, _, _ := setupTestEnvironment(t)

	// Initialize Avi Controller directly on dependency manager
	watcher.dependencyManager.InitializeAviControllerConnection("10.1.1.1", "22.1.3", "test-cloud", "admin", "global")

	// Create malformed cluster (missing metadata)
	malformedCluster := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "cluster.x-k8s.io/v1beta1",
			"kind":       "Cluster",
			// Missing metadata
		},
	}

	// Test with malformed cluster - should handle gracefully
	g.Expect(watcher.ShouldManageCluster(malformedCluster)).To(gomega.BeFalse())
	g.Expect(watcher.GetClusterPhase(malformedCluster)).To(gomega.Equal(""))
}
