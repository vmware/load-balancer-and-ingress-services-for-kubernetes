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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/ingestion"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TestVKSCompleteLifecycle tests the complete VKS cluster lifecycle
func TestVKSCompleteLifecycle(t *testing.T) {
	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	scenario := VKSTestScenario{
		Name:        "VKS Complete Cluster Lifecycle",
		Description: "Tests cluster creation, dependency generation, phase transitions, and cleanup",

		Namespaces: []TestNamespace{
			{Name: "test-ns", HasSEG: true},
		},

		Clusters: []TestCluster{
			{Name: "test-cluster", Namespace: "test-ns", Phase: ingestion.ClusterPhaseProvisioning, Managed: true},
		},

		Steps: []TestStep{
			{
				Action:      "generate_dependencies",
				Description: "Generate dependencies for provisioning cluster",
				ClusterName: "test-cluster",
				Namespace:   "test-ns",
				ExpectError: true, // Expected to fail without real Avi client
			},
			{
				Action:      "phase_transition",
				Description: "Transition cluster to provisioned state",
				ClusterName: "test-cluster",
				Namespace:   "test-ns",
				FromPhase:   ingestion.ClusterPhaseProvisioning,
				ToPhase:     ingestion.ClusterPhaseProvisioned,
				ExpectError: false,
			},
			{
				Action:      "generate_dependencies",
				Description: "Generate dependencies for provisioned cluster",
				ClusterName: "test-cluster",
				Namespace:   "test-ns",
				ExpectError: true, // Expected to fail without real Avi client
			},
			{
				Action:      "cleanup_dependencies",
				Description: "Cleanup cluster dependencies",
				ClusterName: "test-cluster",
				Namespace:   "test-ns",
				ExpectError: false,
			},
			{
				Action:      "verify_cleanup",
				Description: "Verify all dependencies are cleaned up",
				ClusterName: "test-cluster",
				Namespace:   "test-ns",
			},
		},

		FinalVerification: func(t *testing.T, f *VKSTestFramework) {
			// Verify cluster still exists but dependencies are cleaned up
			cluster, err := f.DynamicClient.Resource(ingestion.ClusterGVR).Namespace("test-ns").Get(context.TODO(), "test-cluster", metav1.GetOptions{})
			assert.NoError(t, err, "Cluster should still exist")
			assert.NotNil(t, cluster, "Cluster should not be nil")
		},
	}

	framework.RunVKSIntegrationTest(t, scenario)
}

// TestVKSNamespaceWithoutSEG tests VKS behavior with namespaces that don't have SEG
func TestVKSNamespaceWithoutSEG(t *testing.T) {
	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	scenario := VKSTestScenario{
		Name:        "VKS Namespace Without SEG",
		Description: "Tests that VKS skips clusters in namespaces without Service Engine Group",

		Namespaces: []TestNamespace{
			{Name: "no-seg-ns", HasSEG: false},
		},

		Clusters: []TestCluster{
			{Name: "test-cluster", Namespace: "no-seg-ns", Phase: ingestion.ClusterPhaseProvisioned, Managed: true},
		},

		Steps: []TestStep{
			{
				Action:      "generate_dependencies",
				Description: "Attempt to generate dependencies for cluster in namespace without SEG",
				ClusterName: "test-cluster",
				Namespace:   "no-seg-ns",
				ExpectError: false, // Should succeed but skip VKS management
			},
			{
				Action:      "verify_dependencies",
				Description: "Verify no dependencies were created",
				ClusterName: "test-cluster",
				Namespace:   "no-seg-ns",
			},
		},

		FinalVerification: func(t *testing.T, f *VKSTestFramework) {
			// Verify no secrets or configmaps were created
			secrets, err := f.KubeClient.CoreV1().Secrets("no-seg-ns").List(context.TODO(), metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Empty(t, secrets.Items, "No secrets should be created for namespace without SEG")

			configMaps, err := f.KubeClient.CoreV1().ConfigMaps("no-seg-ns").List(context.TODO(), metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Empty(t, configMaps.Items, "No configmaps should be created for namespace without SEG")
		},
	}

	framework.RunVKSIntegrationTest(t, scenario)
}

// TestVKSClusterWatcherIntegration tests the cluster watcher functionality
func TestVKSClusterWatcherIntegration(t *testing.T) {
	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	// Create namespace with SEG
	framework.CreateTestNamespace("watcher-test", true)

	// Test cluster labeling operations
	t.Run("Cluster Opt-In", func(t *testing.T) {
		cluster := framework.CreateVKSManagedCluster("opt-in-cluster", "watcher-test", ingestion.ClusterPhaseProvisioned, false)

		// Test opt-in
		result := framework.ClusterWatcher.HandleClusterOptIn(cluster)
		assert.NotNil(t, result, "Opt-in result should not be nil")
		assert.Equal(t, "opt-in-cluster", result.ClusterName)
		assert.Equal(t, "watcher-test", result.ClusterNamespace)
		assert.Equal(t, ingestion.LabelingOperationOptIn, result.Operation)

		// Should fail due to missing Avi client in test environment
		assert.Error(t, result.Error, "Opt-in should fail without real Avi client")
	})

	t.Run("Cluster Opt-Out", func(t *testing.T) {
		cluster := framework.CreateVKSManagedCluster("opt-out-cluster", "watcher-test", ingestion.ClusterPhaseProvisioned, true)

		// Test opt-out
		result := framework.ClusterWatcher.HandleClusterOptOut(cluster)
		assert.NotNil(t, result, "Opt-out result should not be nil")
		assert.Equal(t, "opt-out-cluster", result.ClusterName)
		assert.Equal(t, "watcher-test", result.ClusterNamespace)
		assert.Equal(t, ingestion.LabelingOperationOptOut, result.Operation)

		// Should succeed in test environment
		assert.NoError(t, result.Error, "Opt-out should succeed")
		assert.True(t, result.Success, "Opt-out should be successful")
	})

	t.Run("Cluster Eligibility", func(t *testing.T) {
		// Test provisioned cluster (should be eligible)
		provisionedCluster := framework.CreateVKSManagedCluster("provisioned-cluster", "watcher-test", ingestion.ClusterPhaseProvisioned, true)
		assert.True(t, framework.ClusterWatcher.ShouldManageCluster(provisionedCluster), "Provisioned cluster should be eligible")

		// Test provisioning cluster (should not be eligible)
		provisioningCluster := framework.CreateVKSManagedCluster("provisioning-cluster", "watcher-test", ingestion.ClusterPhaseProvisioning, true)
		assert.False(t, framework.ClusterWatcher.ShouldManageCluster(provisioningCluster), "Provisioning cluster should not be eligible")

		// Test failed cluster (should not be eligible)
		failedCluster := framework.CreateVKSManagedCluster("failed-cluster", "watcher-test", ingestion.ClusterPhaseFailed, true)
		assert.False(t, framework.ClusterWatcher.ShouldManageCluster(failedCluster), "Failed cluster should not be eligible")
	})
}

// TestVKSDependencyManagerReconciliation tests the dependency manager reconciliation
func TestVKSDependencyManagerReconciliation(t *testing.T) {
	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	// Create test environment
	framework.CreateTestNamespace("reconcile-test", true)
	framework.CreateVKSManagedCluster("reconcile-cluster", "reconcile-test", ingestion.ClusterPhaseProvisioned, true)

	// Test reconciler start/stop
	t.Run("Reconciler Lifecycle", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Start reconciler
		err := framework.DependencyManager.StartReconciler(ctx)
		assert.NoError(t, err, "Reconciler should start successfully")

		// Give it a moment to run
		time.Sleep(100 * time.Millisecond)

		// Stop reconciler
		framework.DependencyManager.StopReconciler()
	})

	t.Run("Managed Clusters Discovery", func(t *testing.T) {
		ctx := context.TODO()

		// Get managed clusters
		managedClusters, err := framework.DependencyManager.GetManagedClusters(ctx)
		assert.NoError(t, err, "Should be able to get managed clusters")

		// Should find our test cluster
		expectedClusterRef := "reconcile-test/reconcile-cluster"
		assert.Contains(t, managedClusters, expectedClusterRef, "Should find the test cluster")
	})
}

// TestVKSWebhookIntegration tests webhook-related functionality
func TestVKSWebhookIntegration(t *testing.T) {
	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	t.Run("Webhook Configuration Management", func(t *testing.T) {
		// Create webhook configuration
		webhook := framework.CreateWebhookConfiguration("ako-vks-cluster-webhook")
		assert.NotNil(t, webhook, "Webhook configuration should be created")
		assert.Equal(t, "ako-vks-cluster-webhook", webhook.Name)
		assert.Len(t, webhook.Webhooks, 1, "Should have one webhook")

		// Verify webhook configuration
		createdWebhook := webhook.Webhooks[0]
		assert.Equal(t, "vks-cluster-webhook.vmware.com", createdWebhook.Name)
		assert.NotNil(t, createdWebhook.ClientConfig.Service, "Should have service configuration")
		assert.Equal(t, "ako-vks-webhook-service", createdWebhook.ClientConfig.Service.Name)
		assert.Equal(t, "vmware-system-ako", createdWebhook.ClientConfig.Service.Namespace)
	})
}

// TestVKSMultiClusterScenario tests VKS with multiple clusters and namespaces
func TestVKSMultiClusterScenario(t *testing.T) {
	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	scenario := VKSTestScenario{
		Name:        "VKS Multi-Cluster Scenario",
		Description: "Tests VKS with multiple clusters across different namespaces",

		Namespaces: []TestNamespace{
			{Name: "team-a", HasSEG: true},
			{Name: "team-b", HasSEG: true},
			{Name: "team-c", HasSEG: false}, // No SEG
		},

		Clusters: []TestCluster{
			{Name: "cluster-a1", Namespace: "team-a", Phase: ingestion.ClusterPhaseProvisioned, Managed: true},
			{Name: "cluster-a2", Namespace: "team-a", Phase: ingestion.ClusterPhaseProvisioned, Managed: true},
			{Name: "cluster-b1", Namespace: "team-b", Phase: ingestion.ClusterPhaseProvisioned, Managed: true},
			{Name: "cluster-c1", Namespace: "team-c", Phase: ingestion.ClusterPhaseProvisioned, Managed: true}, // Should be skipped
		},

		Steps: []TestStep{
			{
				Action:      "generate_dependencies",
				Description: "Generate dependencies for cluster-a1",
				ClusterName: "cluster-a1",
				Namespace:   "team-a",
				ExpectError: true, // Expected to fail without real Avi client
			},
			{
				Action:      "generate_dependencies",
				Description: "Generate dependencies for cluster-a2",
				ClusterName: "cluster-a2",
				Namespace:   "team-a",
				ExpectError: true, // Expected to fail without real Avi client
			},
			{
				Action:      "generate_dependencies",
				Description: "Generate dependencies for cluster-b1",
				ClusterName: "cluster-b1",
				Namespace:   "team-b",
				ExpectError: true, // Expected to fail without real Avi client
			},
			{
				Action:      "generate_dependencies",
				Description: "Attempt to generate dependencies for cluster-c1 (no SEG)",
				ClusterName: "cluster-c1",
				Namespace:   "team-c",
				ExpectError: false, // Should succeed but skip VKS management
			},
		},

		FinalVerification: func(t *testing.T, f *VKSTestFramework) {
			ctx := context.TODO()

			// Get all managed clusters
			managedClusters, err := f.DependencyManager.GetManagedClusters(ctx)
			assert.NoError(t, err, "Should be able to get managed clusters")

			// Should find clusters from namespaces with SEG
			expectedClusters := []string{
				"team-a/cluster-a1",
				"team-a/cluster-a2",
				"team-b/cluster-b1",
				"team-c/cluster-c1", // This will be found but dependency generation will be skipped
			}

			for _, expected := range expectedClusters {
				assert.Contains(t, managedClusters, expected, "Should find cluster %s", expected)
			}

			// Verify that team-c namespace has no resources (due to no SEG)
			secrets, err := f.KubeClient.CoreV1().Secrets("team-c").List(ctx, metav1.ListOptions{})
			assert.NoError(t, err)
			assert.Empty(t, secrets.Items, "team-c should have no secrets due to missing SEG")
		},
	}

	framework.RunVKSIntegrationTest(t, scenario)
}

// TestVKSErrorHandling tests various error scenarios
func TestVKSErrorHandling(t *testing.T) {
	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	t.Run("Malformed Cluster Handling", func(t *testing.T) {
		// Create malformed cluster (missing metadata)
		malformedCluster := framework.CreateTestCluster("malformed", "default", "", nil)

		// Remove metadata to make it malformed
		delete(malformedCluster.Object, "metadata")

		// Test that watcher handles malformed cluster gracefully
		assert.False(t, framework.ClusterWatcher.ShouldManageCluster(malformedCluster), "Should reject malformed cluster")
		assert.Equal(t, "", framework.ClusterWatcher.GetClusterPhase(malformedCluster), "Should return empty phase for malformed cluster")
	})

	t.Run("Missing Namespace Handling", func(t *testing.T) {
		// Try to generate dependencies for cluster in non-existent namespace
		cluster := framework.CreateTestCluster("orphan-cluster", "non-existent-ns", ingestion.ClusterPhaseProvisioned, map[string]string{
			ingestion.VKSManagedLabel: ingestion.VKSManagedLabelValueTrue,
		})

		err := framework.DependencyManager.GenerateClusterDependencies(context.TODO(), cluster)
		assert.Error(t, err, "Should fail for cluster in non-existent namespace")
		assert.Contains(t, err.Error(), "failed to get namespace", "Error should mention missing namespace")
	})
}

// TestVKSPerformanceScenario tests VKS performance with many clusters
func TestVKSPerformanceScenario(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	framework := NewVKSTestFramework(t)
	defer framework.Cleanup()

	// Create multiple namespaces and clusters
	numNamespaces := 5
	clustersPerNamespace := 10

	start := time.Now()

	// Setup phase
	for i := 0; i < numNamespaces; i++ {
		nsName := fmt.Sprintf("perf-ns-%d", i)
		framework.CreateTestNamespace(nsName, true)

		for j := 0; j < clustersPerNamespace; j++ {
			clusterName := fmt.Sprintf("cluster-%d", j)
			framework.CreateVKSManagedCluster(clusterName, nsName, ingestion.ClusterPhaseProvisioned, true)
		}
	}

	setupTime := time.Since(start)
	t.Logf("Setup time for %d namespaces with %d clusters each: %v", numNamespaces, clustersPerNamespace, setupTime)

	// Test getting managed clusters
	start = time.Now()
	managedClusters, err := framework.DependencyManager.GetManagedClusters(context.TODO())
	queryTime := time.Since(start)

	assert.NoError(t, err, "Should be able to get managed clusters")
	expectedCount := numNamespaces * clustersPerNamespace
	assert.Len(t, managedClusters, expectedCount, "Should find all %d clusters", expectedCount)

	t.Logf("Query time for %d clusters: %v", expectedCount, queryTime)

	// Performance assertions
	assert.Less(t, setupTime, 5*time.Second, "Setup should complete within 5 seconds")
	assert.Less(t, queryTime, 1*time.Second, "Query should complete within 1 second")
}
