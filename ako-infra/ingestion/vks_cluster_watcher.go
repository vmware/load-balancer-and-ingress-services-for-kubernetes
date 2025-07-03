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
	"fmt"
	"sync"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/client-go/util/workqueue"
)

// VKSClusterWatcher monitors cluster lifecycle events for AKO addon management
type VKSClusterWatcher struct {
	kubeClient      kubernetes.Interface
	dynamicClient   dynamic.Interface
	clusterInformer cache.SharedIndexInformer
	workqueue       workqueue.RateLimitingInterface
	stopCh          <-chan struct{}
	mutex           sync.RWMutex

	// VKS Dependency Manager for cluster-specific resources
	dependencyManager *VKSDependencyManager
}

// NewVKSClusterWatcher creates a new cluster watcher instance
func NewVKSClusterWatcher(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *VKSClusterWatcher {
	workqueue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		VKSClusterWorkQueue,
	)

	clusterInformer := cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				return dynamicClient.Resource(ClusterGVR).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				return dynamicClient.Resource(ClusterGVR).Watch(context.TODO(), options)
			},
		},
		&unstructured.Unstructured{},
		VKSClusterResyncPeriod,
		cache.Indexers{},
	)

	dependencyManager := NewVKSDependencyManager(kubeClient, dynamicClient)

	watcher := &VKSClusterWatcher{
		kubeClient:        kubeClient,
		dynamicClient:     dynamicClient,
		clusterInformer:   clusterInformer,
		workqueue:         workqueue,
		dependencyManager: dependencyManager,
	}

	// Add event handlers
	clusterInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Debugf("Cluster ADD event")
			watcher.enqueueCluster(obj, "ADD")
		},
		UpdateFunc: func(old, new interface{}) {
			utils.AviLog.Debugf("Cluster UPDATE event")
			// Process label changes first
			if oldCluster, ok1 := old.(*unstructured.Unstructured); ok1 {
				if newCluster, ok2 := new.(*unstructured.Unstructured); ok2 {
					watcher.handleLabelChanges(oldCluster, newCluster)
				}
			}
			watcher.enqueueCluster(new, "UPDATE")
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Debugf("Cluster DELETE event")
			watcher.enqueueCluster(obj, "DELETE")
		},
	})

	return watcher
}

// Start begins cluster watcher operation
func (w *VKSClusterWatcher) Start(stopCh <-chan struct{}) error {
	utils.AviLog.Infof("Starting cluster watcher")
	w.stopCh = stopCh

	// Start the informer
	go w.clusterInformer.Run(stopCh)

	// Wait for cache sync
	if !cache.WaitForCacheSync(stopCh, w.clusterInformer.HasSynced) {
		return fmt.Errorf("timed out waiting for cluster cache to sync")
	}
	utils.AviLog.Infof("Cluster cache synced successfully")

	// Start dependency manager reconciler
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-stopCh
		cancel()
	}()

	if err := w.dependencyManager.StartReconciler(ctx); err != nil {
		return fmt.Errorf("failed to start dependency manager: %v", err)
	}
	utils.AviLog.Infof("Dependency manager started")

	utils.AviLog.Infof("Monitoring cluster label changes and lifecycle events")

	// Start worker
	go w.runWorker()

	utils.AviLog.Infof("Cluster watcher started successfully")
	return nil
}

// Stop gracefully shuts down the cluster watcher
func (w *VKSClusterWatcher) Stop() {
	utils.AviLog.Infof("Stopping cluster watcher")

	w.dependencyManager.StopReconciler()
	utils.AviLog.Infof("Dependency manager stopped")

	w.workqueue.ShutDown()
	utils.AviLog.Infof("Cluster watcher stopped")
}

func (w *VKSClusterWatcher) runWorker() {
	for w.processNextWorkItem() {
	}
}

func (w *VKSClusterWatcher) processNextWorkItem() bool {
	obj, shutdown := w.workqueue.Get()
	if shutdown {
		return false
	}

	err := func(obj interface{}) error {
		defer w.workqueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			w.workqueue.Forget(obj)
			utils.AviLog.Errorf("Expected string in workqueue but got %#v", obj)
			return nil
		}
		if err := w.processClusterEvent(key); err != nil {
			w.workqueue.AddRateLimited(key)
			return fmt.Errorf("error processing cluster %s: %s, requeuing", key, err.Error())
		}
		w.workqueue.Forget(obj)
		utils.AviLog.Debugf("Successfully processed cluster: %s", key)
		return nil
	}(obj)

	if err != nil {
		utils.AviLog.Errorf("%v", err)
		return true
	}

	return true
}

func (w *VKSClusterWatcher) enqueueCluster(obj interface{}, eventType string) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utils.AviLog.Errorf("Error getting key for cluster: %v", err)
		return
	}
	utils.AviLog.Debugf("Enqueuing cluster %s for %s", key, eventType)
	w.workqueue.Add(key)
}

func (w *VKSClusterWatcher) handleLabelChanges(oldCluster, newCluster *unstructured.Unstructured) {
	result := w.ProcessClusterLabelChange(oldCluster, newCluster)
	if result.Error != nil {
		utils.AviLog.Errorf("Failed to process label change for cluster %s/%s: %v",
			result.ClusterNamespace, result.ClusterName, result.Error)
	} else if result.Success {
		utils.AviLog.Infof("Successfully processed %s for cluster %s/%s",
			result.Operation, result.ClusterNamespace, result.ClusterName)
	}
}

func (w *VKSClusterWatcher) processClusterEvent(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utils.AviLog.Errorf("Invalid resource key: %s", key)
		return nil
	}

	cluster, err := w.dynamicClient.Resource(ClusterGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if err.Error() == "not found" {
			utils.AviLog.Infof("Cluster deleted: %s", key)
			return w.handleClusterDeletion(namespace, name)
		}
		return fmt.Errorf("failed to get cluster %s: %v", key, err)
	}

	return w.handleClusterAddOrUpdate(cluster)
}

func (w *VKSClusterWatcher) handleClusterAddOrUpdate(cluster *unstructured.Unstructured) error {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	// Check if we should manage this cluster
	if !w.ShouldManageCluster(cluster) {
		utils.AviLog.Debugf("Skipping cluster %s/%s - not managed", clusterNamespace, clusterName)
		return nil
	}

	phase := w.GetClusterPhase(cluster)
	utils.AviLog.Debugf("Processing cluster %s/%s in phase: %s", clusterNamespace, clusterName, phase)

	switch phase {
	case ClusterPhaseProvisioned:
		return w.handleProvisionedCluster(cluster)
	case ClusterPhaseDeleting:
		return w.handleClusterDeletion(clusterNamespace, clusterName)
	default:
		utils.AviLog.Debugf("Cluster %s/%s not in provisioned state, skipping", clusterNamespace, clusterName)
		return nil
	}
}

func (w *VKSClusterWatcher) handleProvisionedCluster(cluster *unstructured.Unstructured) error {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	labels := cluster.GetLabels()
	if labels == nil {
		utils.AviLog.Debugf("Cluster %s/%s has no labels", clusterNamespace, clusterName)
		return nil
	}

	vksManagedValue, exists := labels[VKSManagedLabel]
	if !exists {
		utils.AviLog.Debugf("Cluster %s/%s missing VKS managed label", clusterNamespace, clusterName)
		return nil
	}

	if vksManagedValue != VKSManagedLabelValueTrue {
		utils.AviLog.Debugf("Cluster %s/%s not opted in for VKS management", clusterNamespace, clusterName)
		return nil
	}

	utils.AviLog.Infof("Processing provisioned cluster: %s/%s", clusterNamespace, clusterName)

	// Generate cluster dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.dependencyManager.GenerateClusterDependencies(ctx, cluster); err != nil {
		return fmt.Errorf("failed to generate dependencies for cluster %s/%s: %v", clusterNamespace, clusterName, err)
	}

	utils.AviLog.Infof("Successfully processed cluster: %s/%s", clusterNamespace, clusterName)
	return nil
}

func (w *VKSClusterWatcher) handleClusterDeletion(namespace, name string) error {
	utils.AviLog.Infof("Cleaning up dependencies for deleted cluster: %s/%s", namespace, name)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.dependencyManager.CleanupClusterDependencies(ctx, name, namespace); err != nil {
		utils.AviLog.Errorf("Failed to cleanup dependencies for cluster %s/%s: %v", namespace, name, err)
		return err
	}

	utils.AviLog.Infof("Successfully cleaned up cluster: %s/%s", namespace, name)
	return nil
}

// ProcessClusterLabelChange handles VKS label changes on clusters
func (w *VKSClusterWatcher) ProcessClusterLabelChange(oldCluster, newCluster *unstructured.Unstructured) *LabelingResult {
	clusterName := newCluster.GetName()
	clusterNamespace := newCluster.GetNamespace()

	result := &LabelingResult{
		ClusterName:      clusterName,
		ClusterNamespace: clusterNamespace,
		Operation:        LabelingOperationUpdate,
	}

	oldLabels := oldCluster.GetLabels()
	newLabels := newCluster.GetLabels()

	var oldValue, newValue string
	if oldLabels != nil {
		oldValue = oldLabels[VKSManagedLabel]
	}
	if newLabels != nil {
		newValue = newLabels[VKSManagedLabel]
	}

	result.PreviousValue = oldValue
	result.NewValue = newValue

	// No change in VKS managed label
	if oldValue == newValue {
		result.Skipped = true
		result.SkipReason = "no change in VKS managed label"
		return result
	}

	utils.AviLog.Infof("VKS label change for cluster %s/%s: '%s' -> '%s'",
		clusterNamespace, clusterName, oldValue, newValue)

	// Handle different label transitions
	switch {
	case oldValue == "" && newValue == VKSManagedLabelValueTrue:
		return w.HandleClusterOptIn(newCluster)
	case oldValue == "" && newValue == VKSManagedLabelValueFalse:
		return w.HandleClusterOptOut(newCluster)
	case oldValue == VKSManagedLabelValueTrue && newValue == VKSManagedLabelValueFalse:
		return w.HandleClusterOptOut(newCluster)
	case oldValue == VKSManagedLabelValueFalse && newValue == VKSManagedLabelValueTrue:
		return w.HandleClusterOptIn(newCluster)
	case oldValue == VKSManagedLabelValueTrue && newValue == "":
		return w.HandleClusterOptOut(newCluster)
	default:
		result.Skipped = true
		result.SkipReason = fmt.Sprintf("unhandled label transition: '%s' -> '%s'", oldValue, newValue)
		return result
	}
}

// HandleClusterOptIn processes explicit cluster opt-in
func (w *VKSClusterWatcher) HandleClusterOptIn(cluster *unstructured.Unstructured) *LabelingResult {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	result := &LabelingResult{
		ClusterName:      clusterName,
		ClusterNamespace: clusterNamespace,
		Operation:        LabelingOperationOptIn,
	}

	utils.AviLog.Infof("Handling cluster opt-in: %s/%s", clusterNamespace, clusterName)

	// Check SEG configuration
	hasSEG, err := w.NamespaceHasSEG(clusterNamespace)
	if err != nil {
		result.Error = fmt.Errorf("failed to check SEG configuration: %v", err)
		return result
	}

	if !hasSEG {
		result.Error = fmt.Errorf("cannot opt-in cluster in namespace without SEG configuration")
		return result
	}

	// Check cluster eligibility
	if !w.ShouldManageCluster(cluster) {
		result.Error = fmt.Errorf("cluster is not eligible for VKS management")
		return result
	}

	// Ensure label is set correctly
	labels := cluster.GetLabels()
	if labels == nil || labels[VKSManagedLabel] != VKSManagedLabelValueTrue {
		err = w.setVKSManagedLabel(cluster, VKSManagedLabelValueTrue)
		if err != nil {
			result.Error = err
			return result
		}
	}

	// Generate cluster dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.dependencyManager.GenerateClusterDependencies(ctx, cluster); err != nil {
		utils.AviLog.Errorf("Failed to generate dependencies for opted-in cluster %s/%s: %v", clusterNamespace, clusterName, err)
		result.Error = fmt.Errorf("failed to generate cluster dependencies: %v", err)
		return result
	}
	utils.AviLog.Infof("Successfully generated dependencies for opted-in cluster %s/%s", clusterNamespace, clusterName)

	result.Success = true
	result.NewValue = VKSManagedLabelValueTrue
	return result
}

// HandleClusterOptOut processes explicit cluster opt-out
func (w *VKSClusterWatcher) HandleClusterOptOut(cluster *unstructured.Unstructured) *LabelingResult {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	result := &LabelingResult{
		ClusterName:      clusterName,
		ClusterNamespace: clusterNamespace,
		Operation:        LabelingOperationOptOut,
	}

	utils.AviLog.Infof("Handling cluster opt-out: %s/%s", clusterNamespace, clusterName)

	// Clean up cluster dependencies
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.dependencyManager.CleanupClusterDependencies(ctx, clusterName, clusterNamespace); err != nil {
		utils.AviLog.Errorf("Failed to cleanup dependencies for opted-out cluster %s/%s: %v", clusterNamespace, clusterName, err)
		result.Error = fmt.Errorf("failed to cleanup cluster dependencies: %v", err)
		return result
	}

	utils.AviLog.Infof("Successfully cleaned up dependencies for opted-out cluster %s/%s", clusterNamespace, clusterName)

	result.Success = true
	result.NewValue = VKSManagedLabelValueFalse
	utils.AviLog.Infof("Successfully processed cluster opt-out for %s/%s", clusterNamespace, clusterName)
	return result
}

func (w *VKSClusterWatcher) setVKSManagedLabel(cluster *unstructured.Unstructured, value string) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Get the latest version
		latest, err := w.dynamicClient.Resource(ClusterGVR).
			Namespace(cluster.GetNamespace()).
			Get(context.Background(), cluster.GetName(), metav1.GetOptions{})
		if err != nil {
			return err
		}

		// Update labels
		labels := latest.GetLabels()
		if labels == nil {
			labels = make(map[string]string)
		}
		labels[VKSManagedLabel] = value
		latest.SetLabels(labels)

		// Update the cluster
		_, err = w.dynamicClient.Resource(ClusterGVR).
			Namespace(latest.GetNamespace()).
			Update(context.Background(), latest, metav1.UpdateOptions{})
		return err
	})
}

// NamespaceHasSEG checks if namespace has Service Engine Group configuration
func (w *VKSClusterWatcher) NamespaceHasSEG(namespaceName string) (bool, error) {
	namespace, err := w.kubeClient.CoreV1().Namespaces().Get(context.Background(), namespaceName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	annotations := namespace.GetAnnotations()
	if annotations == nil {
		return false, nil
	}

	_, exists := annotations[ServiceEngineGroupAnnotation]
	return exists, nil
}

// ShouldManageCluster determines if a cluster should be managed by VKS
func (w *VKSClusterWatcher) ShouldManageCluster(cluster *unstructured.Unstructured) bool {
	// Check cluster phase
	phase := w.GetClusterPhase(cluster)
	if phase != ClusterPhaseProvisioned && phase != ClusterPhaseDeleting {
		return false
	}

	// Check if namespace has SEG configuration
	hasSEG, err := w.NamespaceHasSEG(cluster.GetNamespace())
	if err != nil {
		utils.AviLog.Errorf("Failed to check SEG for namespace %s: %v", cluster.GetNamespace(), err)
		return false
	}

	return hasSEG
}

// GetClusterPhase returns the current phase of the cluster
func (w *VKSClusterWatcher) GetClusterPhase(cluster *unstructured.Unstructured) string {
	status, found, err := unstructured.NestedString(cluster.Object, "status", "phase")
	if err != nil || !found {
		return ""
	}
	return status
}
