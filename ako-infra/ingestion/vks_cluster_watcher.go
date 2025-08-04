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

package ingestion

// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/webhook"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

const (
	// VKS cluster monitoring constants
	ClusterPhaseProvisioning = "Provisioning"
	ClusterPhaseProvisioned  = "Provisioned"
	ClusterPhaseDeleting     = "Deleting"
	ClusterPhaseFailed       = "Failed"

	// VKS cluster watcher configuration
	VKSClusterWorkQueue = "vks-cluster-watcher"
)

// AdminCredentials holds the ako-infra admin Avi controller credentials
type AdminCredentials struct {
	Username     string
	Password     string
	AuthToken    string
	ControllerIP string
	CACert       string
}

// VKSClusterWatcher monitors cluster lifecycle events for AKO addon management
type VKSClusterWatcher struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	workqueue     workqueue.RateLimitingInterface
}

// NewVKSClusterWatcher creates a new cluster watcher instance
func NewVKSClusterWatcher(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *VKSClusterWatcher {
	workqueue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(),
		VKSClusterWorkQueue,
	)

	watcher := &VKSClusterWatcher{
		kubeClient:    kubeClient,
		dynamicClient: dynamicClient,
		workqueue:     workqueue,
	}

	return watcher
}

// Start begins cluster watcher operation
func (w *VKSClusterWatcher) Start(stopCh <-chan struct{}) error {
	utils.AviLog.Infof("Starting cluster watcher")
	go w.runWorker()
	utils.AviLog.Infof("Cluster watcher started successfully")
	return nil
}

// Stop gracefully shuts down the cluster watcher
func (w *VKSClusterWatcher) Stop() {
	utils.AviLog.Infof("Stopping cluster watcher")

	w.workqueue.ShutDown()
	utils.AviLog.Infof("Cluster watcher stopped")
}

func (w *VKSClusterWatcher) runWorker() {
	for w.ProcessNextWorkItem() {
	}
}

func (w *VKSClusterWatcher) ProcessNextWorkItem() bool {
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
		if err := w.ProcessClusterEvent(key); err != nil {
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

func (w *VKSClusterWatcher) EnqueueCluster(obj interface{}, eventType string) {
	var key string
	var err error
	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utils.AviLog.Errorf("Error getting key for cluster: %v", err)
		return
	}
	utils.AviLog.Debugf("Enqueuing cluster %s for %s", key, eventType)
	w.workqueue.Add(key)
}

func (w *VKSClusterWatcher) ProcessClusterEvent(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utils.AviLog.Errorf("Invalid resource key: %s", key)
		return nil
	}

	cluster, err := w.dynamicClient.Resource(lib.ClusterGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
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

	phase := w.GetClusterPhase(cluster)
	utils.AviLog.Debugf("Processing cluster %s/%s in phase: %s", clusterNamespace, clusterName, phase)

	switch phase {
	case ClusterPhaseProvisioned:
		return w.HandleProvisionedCluster(cluster)
	case ClusterPhaseDeleting:
		return w.handleClusterDeletion(clusterNamespace, clusterName)
	default:
		utils.AviLog.Debugf("Cluster %s/%s not in provisioned state, skipping", clusterNamespace, clusterName)
		return nil
	}
}

func (w *VKSClusterWatcher) HandleProvisionedCluster(cluster *unstructured.Unstructured) error {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	// Check if cluster should be managed by VKS
	labels := cluster.GetLabels()
	shouldManage := labels != nil && labels[webhook.VKSManagedLabel] == webhook.VKSManagedLabelValueTrue

	// Check if secret already exists
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := w.kubeClient.CoreV1().Secrets(clusterNamespace).Get(ctx, secretName, metav1.GetOptions{})
	secretExists := err == nil

	// Take appropriate action based on desired vs. current state
	switch {
	case shouldManage && !secretExists:
		// Should manage but no secret exists - create it
		utils.AviLog.Infof("Creating VKS dependencies for cluster: %s/%s", clusterNamespace, clusterName)
		if err := w.GenerateClusterSecret(ctx, cluster); err != nil {
			return fmt.Errorf("failed to generate dependencies for cluster %s/%s: %v", clusterNamespace, clusterName, err)
		}
		utils.AviLog.Infof("Successfully created VKS dependencies for cluster: %s/%s", clusterNamespace, clusterName)

	case !shouldManage && secretExists:
		// Should not manage but secret exists - delete it
		utils.AviLog.Infof("Cleaning up VKS dependencies for cluster: %s/%s", clusterNamespace, clusterName)
		if err := w.cleanupClusterSecret(ctx, clusterName, clusterNamespace); err != nil {
			return fmt.Errorf("failed to cleanup dependencies for cluster %s/%s: %v", clusterNamespace, clusterName, err)
		}
		utils.AviLog.Infof("Successfully cleaned up VKS dependencies for cluster: %s/%s", clusterNamespace, clusterName)

	case shouldManage && secretExists:
		// Should manage and secret exists - already in desired state
		utils.AviLog.Debugf("VKS dependencies already exist for cluster: %s/%s", clusterNamespace, clusterName)

	case !shouldManage && !secretExists:
		// Should not manage and no secret exists - already in desired state
		utils.AviLog.Debugf("Cluster %s/%s not managed by VKS and no dependencies exist", clusterNamespace, clusterName)
	}

	return nil
}

func (w *VKSClusterWatcher) handleClusterDeletion(namespace, name string) error {
	utils.AviLog.Infof("Cleaning up dependencies for deleted cluster: %s/%s", namespace, name)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.cleanupClusterSecret(ctx, name, namespace); err != nil {
		utils.AviLog.Errorf("Failed to cleanup dependencies for cluster %s/%s: %v", namespace, name, err)
		return err
	}

	utils.AviLog.Infof("Successfully cleaned up cluster: %s/%s", namespace, name)
	return nil
}

// GetClusterPhase returns the current phase of the cluster
func (w *VKSClusterWatcher) GetClusterPhase(cluster *unstructured.Unstructured) string {
	status, found, err := unstructured.NestedString(cluster.Object, "status", "phase")
	if err != nil || !found {
		return ""
	}
	return status
}

// generateClusterSecret generates all required dependency resources for a cluster
func (w *VKSClusterWatcher) GenerateClusterSecret(ctx context.Context, cluster *unstructured.Unstructured) error {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	utils.AviLog.Infof("Generating VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)

	// Get ako-infra admin credentials for cluster use
	adminCreds, err := w.getAkoInfraAdminCredentials()
	if err != nil {
		return fmt.Errorf("failed to get ako-infra admin credentials: %v", err)
	}

	// Generate Avi credentials secret using admin credentials
	if err := w.createAviCredentialsSecret(ctx, clusterName, clusterNamespace, adminCreds.Username, adminCreds.Password, adminCreds.AuthToken, adminCreds.ControllerIP, adminCreds.CACert); err != nil {
		return fmt.Errorf("failed to generate Avi credentials secret: %v", err)
	}

	utils.AviLog.Infof("Successfully generated VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)
	return nil
}

// cleanupClusterSecret removes dependency resources for a deleted cluster
func (w *VKSClusterWatcher) cleanupClusterSecret(ctx context.Context, clusterName, clusterNamespace string) error {
	utils.AviLog.Infof("Cleaning up VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)

	// Delete Avi credentials secret
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)
	err := w.kubeClient.CoreV1().Secrets(clusterNamespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to delete Avi credentials secret %s/%s: %v", clusterNamespace, secretName, err)
	} else {
		utils.AviLog.Infof("Deleted Avi credentials secret %s/%s", clusterNamespace, secretName)
	}

	utils.AviLog.Infof("Completed cleanup of VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)
	return nil
}

// getAkoInfraAdminCredentials retrieves the ako-infra admin Avi controller credentials
func (w *VKSClusterWatcher) getAkoInfraAdminCredentials() (*AdminCredentials, error) {
	// Get credentials from lib function that reads from avi-secret
	ctrlProps, err := lib.GetControllerPropertiesFromSecret(w.kubeClient)
	if err != nil {
		return nil, fmt.Errorf("failed to get controller properties from secret: %v", err)
	}

	// Get controller IP
	controllerIP := lib.GetControllerIP()
	if controllerIP == "" {
		return nil, fmt.Errorf("controller IP not set")
	}

	creds := &AdminCredentials{
		Username:     ctrlProps[utils.ENV_CTRL_USERNAME],
		Password:     ctrlProps[utils.ENV_CTRL_PASSWORD],
		AuthToken:    ctrlProps[utils.ENV_CTRL_AUTHTOKEN],
		ControllerIP: controllerIP,
		CACert:       ctrlProps[utils.ENV_CTRL_CADATA],
	}

	// Validate required fields
	if creds.Username == "" {
		return nil, fmt.Errorf("username not found in avi-secret")
	}

	// Either password or authtoken should be present
	if creds.Password == "" && creds.AuthToken == "" {
		return nil, fmt.Errorf("neither password nor authtoken found in avi-secret")
	}

	utils.AviLog.Debugf("Retrieved ako-infra admin credentials for user: %s", creds.Username)
	return creds, nil
}

// createAviCredentialsSecret creates the Avi credentials secret for a cluster
func (w *VKSClusterWatcher) createAviCredentialsSecret(ctx context.Context, clusterName, clusterNamespace, username, password, authToken, controllerIP, caCert string) error {
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)

	secretData := map[string][]byte{
		"username":     []byte(base64.StdEncoding.EncodeToString([]byte(username))),
		"controllerIP": []byte(base64.StdEncoding.EncodeToString([]byte(controllerIP))),
	}

	// Add password if available
	if password != "" {
		secretData["password"] = []byte(base64.StdEncoding.EncodeToString([]byte(password)))
	}

	// Add authtoken if available
	if authToken != "" {
		secretData["authtoken"] = []byte(base64.StdEncoding.EncodeToString([]byte(authToken)))
	}

	// Add CA cert if available
	if caCert != "" {
		secretData["certificateAuthorityData"] = []byte(base64.StdEncoding.EncodeToString([]byte(caCert)))
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: clusterNamespace,
			Labels: map[string]string{
				"ako.kubernetes.vmware.com/cluster":    clusterName,
				"ako.kubernetes.vmware.com/managed-by": "ako-infra",
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	_, err := w.kubeClient.CoreV1().Secrets(clusterNamespace).Create(ctx, secret, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create Avi credentials secret %s: %v", secretName, err)
	}

	utils.AviLog.Infof("Created Avi credentials secret %s/%s for cluster %s",
		clusterNamespace, secretName, clusterName)
	return nil
}
