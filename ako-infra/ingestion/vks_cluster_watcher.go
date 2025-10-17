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

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/alb-sdk/go/clients"
	"github.com/vmware/alb-sdk/go/session"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/webhook"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
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

	VKSReconcileInterval = 6 * time.Hour
)

// VKSClusterConfig holds all the configuration needed for a VKS cluster's AKO deployment
type VKSClusterConfig struct {
	Username                 string
	Password                 string
	ControllerHost           string
	ControllerAddress        string
	ControllerVersion        string
	CertificateAuthorityData string
	VPCMode                  bool
	DedicatedTenantMode      bool
	Managed                  bool

	// Namespace-specific configuration
	ServiceEngineGroup string
	TenantName         string
	NsxtT1LR           string

	CloudName string

	CNIPlugin   string
	ServiceType string
	ClusterName string
}

// VKSSecretInfo holds information about a VKS secret
type VKSSecretInfo struct {
	Name               string
	Namespace          string
	ClusterNameWithUID string
}

// VKSClusterWatcher monitors cluster lifecycle events for AKO addon management
type VKSClusterWatcher struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	workqueue     workqueue.RateLimitingInterface //nolint:staticcheck

	reconcileTicker *time.Ticker
	reconcileStopCh chan struct{}

	// For testing - allows injecting mock behavior
	testMode            bool
	mockCredentialsFunc func(string, string) (*lib.ClusterCredentials, error)
	mockRBACFunc        func(string) error
	mockCleanupFunc     func(string, string) error // func(clusterNameWithUID, tenant) error
}

// getUniqueClusterName generates a unique cluster identifier in format: clustername[:15]-hash
// The hash is computed from clustername, namespace, uid, and supervisorID to ensure uniqueness across supervisors
func (w *VKSClusterWatcher) getUniqueClusterName(cluster *unstructured.Unstructured) string {
	supervisorID := lib.GetClusterID()
	namespace := cluster.GetNamespace()
	name := cluster.GetName()
	uid := cluster.GetUID()

	clusterName := name
	if len(clusterName) > 15 {
		clusterName = clusterName[:15]
	}

	// Create hash from all components to ensure uniqueness across supervisors
	hashInput := fmt.Sprintf("%s-%s-%s-%s", name, namespace, uid, supervisorID)
	hash := utils.Hash(hashInput)

	return fmt.Sprintf("%s-%s", clusterName, fmt.Sprintf("%x", hash))
}

// NewVKSClusterWatcher creates a new cluster watcher instance
func NewVKSClusterWatcher(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *VKSClusterWatcher {
	workqueue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(), //nolint:staticcheck
		VKSClusterWorkQueue,
	)

	watcher := &VKSClusterWatcher{
		kubeClient:      kubeClient,
		dynamicClient:   dynamicClient,
		workqueue:       workqueue,
		reconcileStopCh: make(chan struct{}),
	}

	return watcher
}

func (w *VKSClusterWatcher) EnqueueClusterFromSecret(secretName, secretNamespace string) {
	if !strings.HasSuffix(secretName, "-avi-secret") {
		return
	}

	clusterName := strings.TrimSuffix(secretName, "-avi-secret")
	key := fmt.Sprintf("%s/%s", secretNamespace, clusterName)

	utils.AviLog.Infof("VKS secret watcher: enqueuing cluster %s due to secret change %s/%s", key, secretNamespace, secretName)
	w.workqueue.Add(key)
}

func (w *VKSClusterWatcher) isVKSManagedSecret(secret *corev1.Secret) bool {
	// Only process VKS cluster secrets (pattern: *-avi-secret)
	if !strings.HasSuffix(secret.Name, "-avi-secret") {
		return false
	}

	if secret.Labels == nil {
		return false
	}
	managedBy, exists := secret.Labels["ako.kubernetes.vmware.com/managed-by"]
	return exists && managedBy == "ako-infra"
}

func (w *VKSClusterWatcher) AddVKSSecretEventHandler(stopCh <-chan struct{}) {
	utils.AviLog.Infof("VKS secret watcher: starting secret event handler for VKS cluster secrets")

	vksSecretEventHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, cur interface{}) {
			oldSecret, okOld := old.(*corev1.Secret)
			newSecret, okNew := cur.(*corev1.Secret)

			if !okOld || !okNew {
				return
			}

			if !w.isVKSManagedSecret(oldSecret) {
				return
			}

			// Detect out-of-band changes:
			// 1. Data changed
			// 2. Labels removed (secret no longer VKS-managed)
			dataChanged := !reflect.DeepEqual(newSecret.Data, oldSecret.Data)
			labelsRemoved := !w.isVKSManagedSecret(newSecret)

			if oldSecret.ResourceVersion != newSecret.ResourceVersion && (dataChanged || labelsRemoved) {
				if labelsRemoved {
					utils.AviLog.Warnf("VKS secret watcher: detected label removal from secret %s/%s", newSecret.Namespace, newSecret.Name)
				}
				if dataChanged {
					utils.AviLog.Infof("VKS secret watcher: detected data change to secret %s/%s", newSecret.Namespace, newSecret.Name)
				}
				w.EnqueueClusterFromSecret(newSecret.Name, newSecret.Namespace)
			}
		},
		DeleteFunc: func(obj interface{}) {
			secret, ok := obj.(*corev1.Secret)
			if !ok {
				cacheObj, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("VKS secret watcher: couldn't get object from cache %#v", obj)
					return
				}
				secret, ok = cacheObj.Obj.(*corev1.Secret)
				if !ok {
					utils.AviLog.Errorf("VKS secret watcher: cache contained object that is not a Secret: %#v", obj)
					return
				}
			}

			if !w.isVKSManagedSecret(secret) {
				return
			}

			utils.AviLog.Infof("VKS secret watcher: detected out-of-band deletion of secret %s/%s", secret.Namespace, secret.Name)
			w.EnqueueClusterFromSecret(secret.Name, secret.Namespace)
		},
	}

	if utils.GetInformers().SecretInformer != nil {
		utils.GetInformers().SecretInformer.Informer().AddEventHandler(vksSecretEventHandler)
		utils.AviLog.Infof("VKS secret watcher: added VKS secret event handler to existing secret informer")
	} else {
		utils.AviLog.Warnf("VKS secret watcher: secret informer not available")
	}
}

// Start begins cluster watcher operation
func (w *VKSClusterWatcher) Start() error {
	utils.AviLog.Infof("Starting cluster watcher")
	go w.runWorker()

	w.startPeriodicReconciler()

	utils.AviLog.Infof("Cluster watcher started successfully with periodic reconciliation every %v", VKSReconcileInterval)
	return nil
}

// Stop gracefully shuts down the cluster watcher
func (w *VKSClusterWatcher) Stop() {
	utils.AviLog.Infof("Stopping cluster watcher")

	w.stopPeriodicReconciler()

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
		utils.AviLog.Infof("Successfully processed cluster: %s", key)
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
	utils.AviLog.Infof("Enqueuing cluster %s for %s", key, eventType)
	w.workqueue.Add(key)
}

func (w *VKSClusterWatcher) ProcessClusterEvent(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utils.AviLog.Errorf("Invalid resource key: %s", key)
		return nil
	}

	utils.AviLog.Infof("VKS cluster watcher: processing cluster event for %s", key)

	obj, err := lib.GetDynamicInformers().ClusterInformer.Lister().ByNamespace(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Infof("Cluster deleted: %s", key)
			return w.handleClusterDeletion(namespace, name, "")
		}
		return fmt.Errorf("failed to get cluster %s: %v", key, err)
	}

	cluster, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("unexpected object type for cluster %s: %T", key, obj)
	}

	return w.handleClusterAddOrUpdate(cluster)
}

func (w *VKSClusterWatcher) handleClusterAddOrUpdate(cluster *unstructured.Unstructured) error {
	clusterNameWithUID := w.getUniqueClusterName(cluster)
	clusterNamespace := cluster.GetNamespace()

	phase := w.GetClusterPhase(cluster)
	utils.AviLog.Infof("Processing cluster %s in phase: %s", clusterNameWithUID, phase)

	switch phase {
	case ClusterPhaseProvisioning, ClusterPhaseProvisioned:
		return w.HandleProvisionedCluster(cluster)
	case ClusterPhaseDeleting:
		return w.handleClusterDeletion(clusterNamespace, cluster.GetName(), clusterNameWithUID)
	default:
		utils.AviLog.Infof("Cluster %s/%s (UID: %s) not in provisioning/provisioned state, skipping", clusterNamespace, cluster.GetName(), cluster.GetUID())
		return nil
	}
}

func (w *VKSClusterWatcher) HandleProvisionedCluster(cluster *unstructured.Unstructured) error {
	clusterNameWithUID := w.getUniqueClusterName(cluster)
	clusterNamespace := cluster.GetNamespace()

	// Check if cluster should be managed by VKS
	labels := cluster.GetLabels()
	shouldManage := labels != nil && labels[webhook.VKSManagedLabel] == webhook.VKSManagedLabelValueTrue

	phase := w.GetClusterPhase(cluster)
	utils.AviLog.Infof("Processing cluster %s in phase: %s", clusterNameWithUID, phase)

	secretName := fmt.Sprintf("%s-avi-secret", cluster.GetName())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	secretExists := false
	_, err := utils.GetInformers().SecretInformer.Lister().Secrets(clusterNamespace).Get(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Errorf("Secret %s/%s does not exist", clusterNamespace, secretName)
		} else {
			utils.AviLog.Warnf("Failed to check secret %s/%s: %v", clusterNamespace, secretName, err)
		}
	} else {
		secretExists = true
	}

	switch {
	case shouldManage:
		if err := w.UpsertAviCredentialsSecret(ctx, cluster); err != nil {
			return fmt.Errorf("failed to manage dependencies for cluster %s/%s (UID: %s) in phase %s: %v", clusterNamespace, cluster.GetName(), cluster.GetUID(), phase, err)
		}

	case !shouldManage && secretExists:
		utils.AviLog.Infof("Cluster opted out of VKS management, cleaning up all VKS dependencies for cluster: %s/%s (UID: %s)", clusterNamespace, cluster.GetName(), cluster.GetUID())
		if err := w.cleanupClusterDependencies(cluster.GetName(), clusterNamespace, clusterNameWithUID); err != nil {
			return fmt.Errorf("failed to cleanup dependencies for cluster %s/%s (UID: %s): %v", clusterNamespace, cluster.GetName(), cluster.GetUID(), err)
		}
		utils.AviLog.Infof("Initiated async cleanup for opted-out cluster: %s/%s (UID: %s)", clusterNamespace, cluster.GetName(), cluster.GetUID())

	case !shouldManage && !secretExists:
		// Should not manage and no secret exists - already in desired state
		utils.AviLog.Infof("Cluster %s/%s (UID: %s) not managed by VKS and no dependencies exist", clusterNamespace, cluster.GetName(), cluster.GetUID())
	}

	return nil
}

// handleClusterDeletion handles cleanup for deleted clusters
func (w *VKSClusterWatcher) handleClusterDeletion(namespace, name, clusterNameWithUID string) error {
	utils.AviLog.Infof("Cleaning up dependencies for deleted cluster: %s/%s", namespace, name)

	if err := w.cleanupClusterDependencies(name, namespace, clusterNameWithUID); err != nil {
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

// cleanupClusterDependencies removes all dependency resources for a deleted cluster
func (w *VKSClusterWatcher) cleanupClusterDependencies(clusterName, clusterNamespace, clusterNameWithUID string) error {
	utils.AviLog.Infof("Cleaning up VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)

	if clusterNameWithUID == "" {
		clusterNameWithUID = w.getClusterNameWithUIDFromSecret(clusterName, clusterNamespace)
	}

	if clusterNameWithUID != "" {
		w.cleanupClusterAviObjects(clusterName, clusterNameWithUID, clusterNamespace)
	} else {
		utils.AviLog.Infof("Could not determine cluster identifier for %s/%s - likely already cleaned up", clusterNamespace, clusterName)
	}

	return nil
}

// getClusterNameWithUIDFromSecret gets the cluster name with UID from the secret
// This is used during cluster deletion when the cluster object might be gone
func (w *VKSClusterWatcher) getClusterNameWithUIDFromSecret(clusterName, clusterNamespace string) string {
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)

	secret, err := utils.GetInformers().SecretInformer.Lister().Secrets(clusterNamespace).Get(secretName)
	if err == nil && secret.Data != nil {
		if clusterNameBytes, exists := secret.Data["clusterName"]; exists {
			clusterNameWithUID := string(clusterNameBytes)
			utils.AviLog.Infof("Retrieved cluster name with UID from secret: %s", clusterNameWithUID)
			return clusterNameWithUID
		}
	}

	utils.AviLog.Warnf("Could not get cluster UID for %s/%s from secret", clusterNamespace, clusterName)
	return ""
}

// cleanupClusterAviObjects removes all Avi objects, RBAC, and secret for a VKS cluster
func (w *VKSClusterWatcher) cleanupClusterAviObjects(clusterName, clusterNameWithUID, clusterNamespace string) {
	utils.AviLog.Infof("VKS cleanup: Starting async cleanup for cluster %s in namespace %s", clusterNameWithUID, clusterNamespace)

	go func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()

		utils.AviLog.Infof("VKS cleanup: Executing Avi objects cleanup for cluster %s", clusterNameWithUID)
		if err := w.cleanupAviObjects(cleanupCtx, clusterName, clusterNameWithUID, clusterNamespace); err != nil {
			utils.AviLog.Errorf("VKS cleanup: Avi objects cleanup failed for cluster %s: %v - keeping credentials for retry", clusterNameWithUID, err)
			// Don't delete RBAC or secret if Avi objects cleanup failed
			return
		}
		utils.AviLog.Infof("VKS cleanup: Successfully completed Avi objects cleanup for cluster %s", clusterNameWithUID)

		utils.AviLog.Infof("VKS cleanup: Cleaning up RBAC for cluster %s", clusterNameWithUID)
		if err := w.cleanupClusterSpecificRBAC(clusterNameWithUID); err != nil {
			utils.AviLog.Errorf("VKS cleanup: RBAC cleanup failed for cluster %s: %v - keeping secret for retry", clusterNameWithUID, err)
			// Don't delete secret if RBAC cleanup failed
			return
		}
		utils.AviLog.Infof("VKS cleanup: Successfully cleaned up RBAC for cluster %s", clusterNameWithUID)

		secretName := fmt.Sprintf("%s-avi-secret", clusterName)
		utils.AviLog.Infof("VKS cleanup: Deleting secret %s/%s", clusterNamespace, secretName)
		err := w.kubeClient.CoreV1().Secrets(clusterNamespace).Delete(context.Background(), secretName, metav1.DeleteOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				utils.AviLog.Infof("VKS cleanup: Secret %s/%s already deleted", clusterNamespace, secretName)
			} else {
				utils.AviLog.Warnf("VKS cleanup: Failed to delete secret %s/%s: %v - will retry in next cycle", clusterNamespace, secretName, err)
				return
			}
		}
		utils.AviLog.Infof("VKS cleanup: Successfully deleted secret %s/%s", clusterNamespace, secretName)

		utils.AviLog.Infof("VKS cleanup: Completed all cleanup tasks for cluster %s", clusterNameWithUID)
	}()
}

func (w *VKSClusterWatcher) cleanupAviObjects(ctx context.Context, clusterName, clusterNameWithUID, clusterNamespace string) error {
	// Get cluster-specific credentials and tenant from secret
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)
	secret, err := utils.GetInformers().SecretInformer.Lister().Secrets(clusterNamespace).Get(secretName)
	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Infof("VKS cleanup: Secret %s/%s not found, skipping Avi object cleanup", clusterNamespace, secretName)
			return nil
		}
		return fmt.Errorf("failed to get secret %s/%s: %v", clusterNamespace, secretName, err)
	}

	username, exists := secret.Data["username"]
	if !exists {
		return fmt.Errorf("secret %s/%s missing username field", clusterNamespace, secretName)
	}

	password, exists := secret.Data["password"]
	if !exists {
		return fmt.Errorf("secret %s/%s missing password field", clusterNamespace, secretName)
	}

	tenant, exists := secret.Data["tenantName"]
	if !exists {
		return fmt.Errorf("secret %s/%s missing tenantName field", clusterNamespace, secretName)
	}

	// Use mock cleanup in test mode
	if w.testMode && w.mockCleanupFunc != nil {
		return w.mockCleanupFunc(clusterNameWithUID, string(tenant))
	}

	utils.AviLog.Infof("Cleaning up objects for cluster %s in tenant '%s' using cluster-specific credentials", clusterNameWithUID, string(tenant))

	controllerIP := lib.GetControllerIP()
	ctrlProp := utils.SharedCtrlProp().GetAllCtrlProp()
	caData := ctrlProp[utils.ENV_CTRL_CADATA]

	clusterAviClient, err := w.createCleanupAviClient(controllerIP, string(username), string(password), caData, string(tenant))
	if err != nil {
		return fmt.Errorf("failed to create cleanup AVI client: %v", err)
	}
	defer func() {
		if clusterAviClient != nil && clusterAviClient.AviSession != nil {
			if err := clusterAviClient.AviSession.Logout(); err != nil {
				utils.AviLog.Warnf("VKS cleanup: Failed to logout cleanup Avi client: %v", err)
			}
		}
		clusterAviClient = nil
	}()

	prefixes := []string{lib.AKOPrefix, lib.AKOGWPrefix}
	var errors []string

	for _, prefix := range prefixes {
		createdBy := prefix + clusterNameWithUID
		utils.AviLog.Infof("Cleaning up objects for %s", createdBy)

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := w.deleteAviObjects(clusterAviClient, createdBy); err != nil {
				errorMsg := fmt.Sprintf("cleanup failed for %s: %v", createdBy, err)
				utils.AviLog.Errorf("%s", errorMsg)
				errors = append(errors, errorMsg)
			} else {
				utils.AviLog.Infof("Successfully cleaned up objects for %s", createdBy)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("cleanup failures: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (w *VKSClusterWatcher) createCleanupAviClient(controllerIP, username, password, caData, tenant string) (*clients.AviClient, error) {
	if controllerIP == "" || username == "" || password == "" {
		return nil, fmt.Errorf("missing required AVI controller connection parameters")
	}

	ctrlVersion := lib.GetControllerVersion()
	if ctrlVersion == "" {
		infraClient := avirest.VKSAviClientInstance()
		if infraClient.AviSession != nil {
			version, err := infraClient.AviSession.GetControllerVersion()
			if err != nil {
				return nil, fmt.Errorf("failed to get controller version: %v", err)
			}
			ctrlVersion = version
		} else {
			return nil, fmt.Errorf("controller version not available")
		}
	}

	transport, isSecure := utils.GetHTTPTransportWithCert(caData)
	options := []func(*session.AviSession) error{
		session.SetPassword(password),
		session.DisableControllerStatusCheckOnFailure(true),
		session.SetTransport(transport),
		session.SetTimeout(120 * time.Second),
		session.SetVersion(ctrlVersion),
		session.SetTenant(tenant),
	}

	if !isSecure {
		options = append(options, session.SetInsecure)
	}

	aviClient, err := clients.NewAviClient(controllerIP, username, options...)
	if err != nil {
		return nil, fmt.Errorf("failed to create AVI client: %v", err)
	}

	return aviClient, nil
}

func (w *VKSClusterWatcher) deleteAviObjects(aviClient *clients.AviClient, createdBy string) error {
	utils.AviLog.Infof("Cleaning up VirtualServices for %s", createdBy)
	if err := w.deleteVirtualServices(aviClient, createdBy); err != nil {
		return fmt.Errorf("failed to cleanup virtualservices: %v", err)
	}

	objectTypes := []string{
		"/api/vsvip",
		"/api/vsdatascriptset",
		"/api/httppolicyset",
		"/api/l4policyset",
		"/api/poolgroup",
		"/api/pool",
		"/api/sslkeyandcertificate",
		"/api/stringgroup",
	}

	var errors []string
	for _, apiPath := range objectTypes {
		utils.AviLog.Infof("Cleaning up %s for %s", apiPath, createdBy)
		if err := w.deleteObjectsOfType(aviClient, apiPath, createdBy); err != nil {
			errorMsg := fmt.Sprintf("failed to cleanup %s: %v", apiPath, err)
			utils.AviLog.Warnf("%s", errorMsg)
			errors = append(errors, errorMsg)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("object cleanup failures: %s", strings.Join(errors, "; "))
	}
	return nil
}

// deleteVirtualServices deletes child VS first, then parent VS
func (w *VKSClusterWatcher) deleteVirtualServices(aviClient *clients.AviClient, createdBy string) error {
	apiPath := "/api/virtualservice"
	objects, err := w.fetchAviObjects(aviClient, apiPath, createdBy, "uuid,name,type")
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		utils.AviLog.Infof("No %s found for %s", apiPath, createdBy)
		return nil
	}

	var parentVS []map[string]interface{}
	var childVS []map[string]interface{}
	for _, obj := range objects {
		vsType, _ := obj["type"].(string)
		if vsType == "VS_TYPE_VH_PARENT" {
			parentVS = append(parentVS, obj)
		} else {
			// Child VS or normal VS - delete these first
			childVS = append(childVS, obj)
		}
	}

	utils.AviLog.Infof("Found %d VirtualServices to delete for %s (%d child/normal, %d parent)",
		len(objects), createdBy, len(childVS), len(parentVS))

	var errors []string
	for _, obj := range childVS {
		if err := w.deleteAviObject(aviClient, apiPath, obj); err != nil {
			errors = append(errors, err.Error())
		}
	}
	for _, obj := range parentVS {
		if err := w.deleteAviObject(aviClient, apiPath, obj); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("VS deletion failures: %s", strings.Join(errors, "; "))
	}

	utils.AviLog.Infof("Successfully deleted all VirtualServices for %s", createdBy)
	return nil
}

func (w *VKSClusterWatcher) deleteObjectsOfType(aviClient *clients.AviClient, apiPath, createdBy string) error {
	objects, err := w.fetchAviObjects(aviClient, apiPath, createdBy, "uuid,name")
	if err != nil {
		return err
	}

	if len(objects) == 0 {
		utils.AviLog.Infof("No %s found for %s", apiPath, createdBy)
		return nil
	}

	utils.AviLog.Infof("Found %d %s to delete for %s", len(objects), apiPath, createdBy)

	for _, obj := range objects {
		if err := w.deleteAviObject(aviClient, apiPath, obj); err != nil {
			return err
		}
	}

	return nil
}

func (w *VKSClusterWatcher) fetchAviObjects(aviClient *clients.AviClient, apiPath, createdBy, fields string) ([]map[string]interface{}, error) {
	var allObjects []map[string]interface{}
	baseURI := fmt.Sprintf("%s?created_by=%s&fields=%s&page_size=100", apiPath, createdBy, fields)
	nextURI := ""

	for {
		uri := baseURI
		if nextURI != "" {
			uri = nextURI
		}

		var response map[string]interface{}
		err := aviClient.AviSession.Get(uri, &response)
		if err != nil {
			utils.AviLog.Warnf("Get uri %v returned err %v", uri, err)
			return nil, err
		}

		results, ok := response["results"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("unexpected response format for %s", apiPath)
		}

		for _, result := range results {
			if obj, ok := result.(map[string]interface{}); ok {
				allObjects = append(allObjects, obj)
			}
		}

		if next, exists := response["next"].(string); exists && next != "" {
			utils.AviLog.Infof("Processing next page for %s: %s", apiPath, next)
			nextURI = next
		} else {
			break
		}
	}

	return allObjects, nil
}

func (w *VKSClusterWatcher) deleteAviObject(aviClient *clients.AviClient, apiPath string, obj map[string]interface{}) error {
	uuid, ok := obj["uuid"].(string)
	if !ok {
		return fmt.Errorf("object missing uuid")
	}

	name, _ := obj["name"].(string)
	deleteURI := fmt.Sprintf("%s/%s", apiPath, uuid)

	utils.AviLog.Infof("Deleting %s: %s (UUID: %s)", apiPath, name, uuid)

	if err := aviClient.AviSession.Delete(deleteURI); err != nil {
		utils.AviLog.Warnf("Failed to delete %s %s: %v", apiPath, name, err)
		return fmt.Errorf("failed to delete %s %s: %v", apiPath, name, err)
	}

	w.waitForDeletion(aviClient, deleteURI, 10)
	return nil
}

func (w *VKSClusterWatcher) waitForDeletion(client *clients.AviClient, uri string, maxRetries int) {
	for i := 0; i < maxRetries; i++ {
		var response interface{}
		err := client.AviSession.Get(uri, &response)
		if err != nil {
			if aviError, ok := err.(session.AviError); ok && aviError.HttpStatusCode == 404 {
				return
			}
		}
		time.Sleep(1 * time.Second)
	}
	utils.AviLog.Warnf("Object at %s may not be fully deleted", uri)
}

// cleanupClusterSpecificRBAC removes cluster-specific RBAC from Avi Controller
func (w *VKSClusterWatcher) cleanupClusterSpecificRBAC(clusterName string) error {
	if w.testMode && w.mockRBACFunc != nil {
		return w.mockRBACFunc(clusterName)
	}

	aviClient := avirest.VKSAviClientInstance()

	if err := lib.DeleteClusterUser(aviClient, clusterName); err != nil {
		utils.AviLog.Errorf("Failed to delete VKS cluster user for %s: %v", clusterName, err)
		return fmt.Errorf("failed to delete VKS cluster user for %s: %v", clusterName, err)
	}

	if err := lib.DeleteClusterRoles(aviClient, clusterName); err != nil {
		utils.AviLog.Errorf("Failed to delete VKS cluster roles for %s: %v", clusterName, err)
		return fmt.Errorf("failed to delete VKS cluster roles for %s: %v", clusterName, err)
	}

	utils.AviLog.Infof("Cleaned up VKS cluster RBAC for %s", clusterName)
	return nil
}

// createClusterSpecificCredentials creates cluster-specific RBAC credentials in Avi Controller
func (w *VKSClusterWatcher) createClusterSpecificCredentials(clusterNameWithUID string, operationalTenant string) (*lib.ClusterCredentials, error) {
	// Use mock in test mode
	if w.testMode && w.mockCredentialsFunc != nil {
		return w.mockCredentialsFunc(clusterNameWithUID, operationalTenant)
	}

	aviClient := avirest.VKSAviClientInstance()

	if operationalTenant == "" {
		return nil, fmt.Errorf("no tenant configured for cluster %s: tenant must be provided from namespace annotation %s", clusterNameWithUID, lib.TenantAnnotation)
	}

	utils.AviLog.Infof("Creating VKS cluster RBAC for %s in operational tenant: %s", clusterNameWithUID, operationalTenant)

	roles, err := lib.CreateClusterRoles(aviClient, clusterNameWithUID, operationalTenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create VKS cluster roles: %v", err)
	}

	user, password, err := lib.CreateClusterUserWithRoles(aviClient, clusterNameWithUID, roles, operationalTenant)
	if err != nil {
		lib.DeleteClusterRoles(aviClient, clusterNameWithUID)
		return nil, fmt.Errorf("failed to create VKS cluster user: %v", err)
	}

	utils.AviLog.Infof("Created VKS cluster RBAC for %s: admin-role=%s, tenant-role=%s, all-tenants-role=%s, user=%s",
		clusterNameWithUID, *roles.AdminRole.Name, *roles.TenantRole.Name, *roles.AllTenantsRole.Name, *user.Username)

	return &lib.ClusterCredentials{
		Username: *user.Username,
		Password: password,
	}, nil
}

func (w *VKSClusterWatcher) getCredentialsFromSecret(clusterName, namespace string) (*lib.ClusterCredentials, error) {
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)

	secret, err := utils.GetInformers().SecretInformer.Lister().Secrets(namespace).Get(secretName)
	if err != nil {
		return nil, err
	}

	username, exists := secret.Data["username"]
	if !exists {
		return nil, fmt.Errorf("secret %s/%s missing username field", namespace, secretName)
	}

	password, exists := secret.Data["password"]
	if !exists {
		return nil, fmt.Errorf("secret %s/%s missing password field", namespace, secretName)
	}

	return &lib.ClusterCredentials{
		Username: string(username),
		Password: string(password),
	}, nil
}

// NamespaceConfig holds all required configuration extracted from namespace annotations
type NamespaceConfig struct {
	ServiceEngineGroup string
	Tenant             string
	T1LR               string
}

// getNamespaceConfig fetches namespace and extracts all required configuration in a single call
func (w *VKSClusterWatcher) getNamespaceConfig(clusterNamespace string) (*NamespaceConfig, error) {
	namespace, err := utils.GetInformers().NSInformer.Lister().Get(clusterNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace %s: %v", clusterNamespace, err)
	}

	if namespace.Annotations == nil {
		return nil, fmt.Errorf("namespace %s has no annotations", clusterNamespace)
	}

	seg, exists := namespace.Annotations[lib.WCPSEGroup]
	if !exists || seg == "" {
		return nil, fmt.Errorf("namespace %s does not have annotation %s or it is empty", clusterNamespace, lib.WCPSEGroup)
	}

	tenant, exists := namespace.Annotations[lib.TenantAnnotation]
	if !exists || tenant == "" {
		return nil, fmt.Errorf("namespace %s does not have annotation %s or it is empty", clusterNamespace, lib.TenantAnnotation)
	}

	infraSettingName, exists := namespace.Annotations[lib.InfraSettingNameAnnotation]
	if !exists || infraSettingName == "" {
		return nil, fmt.Errorf("namespace %s does not have annotation %s or it is empty", clusterNamespace, lib.InfraSettingNameAnnotation)
	}

	infraSetting, err := lib.AKOControlConfig().CRDInformers().AviInfraSettingInformer.Lister().Get(infraSettingName)
	if err != nil {
		return nil, fmt.Errorf("failed to get AviInfraSetting %s: %v", infraSettingName, err)
	}

	if infraSetting.Spec.NSXSettings.T1LR == nil || *infraSetting.Spec.NSXSettings.T1LR == "" {
		return nil, fmt.Errorf("AviInfraSetting %s does not have nsxSettings.t1lr configured or it is empty", infraSettingName)
	}
	t1lr := *infraSetting.Spec.NSXSettings.T1LR

	return &NamespaceConfig{
		ServiceEngineGroup: seg,
		Tenant:             tenant,
		T1LR:               t1lr,
	}, nil
}

// buildVKSClusterConfig builds complete configuration for a VKS cluster
func (w *VKSClusterWatcher) buildVKSClusterConfig(cluster *unstructured.Unstructured) (*VKSClusterConfig, error) {
	clusterNameWithUID := w.getUniqueClusterName(cluster)
	clusterNamespace := cluster.GetNamespace()

	nsConfig, err := w.getNamespaceConfig(clusterNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace configuration for cluster %s/%s (UID: %s): %v", clusterNamespace, cluster.GetName(), cluster.GetUID(), err)
	}

	clusterCreds, err := w.getCredentialsFromSecret(cluster.GetName(), clusterNamespace)
	if err != nil {
		if errors.IsNotFound(err) {
			creds, err := w.createClusterSpecificCredentials(clusterNameWithUID, nsConfig.Tenant)
			if err != nil {
				return nil, fmt.Errorf("failed to create cluster-specific credentials: %v", err)
			}
			clusterCreds = creds
		} else {
			return nil, fmt.Errorf("failed to retrieve credentials from existing secret: %v", err)
		}
	}

	// Get common controller properties
	controllerIP := lib.GetControllerIP()
	if controllerIP == "" {
		return nil, fmt.Errorf("controller IP not set")
	}

	ctrlProp := utils.SharedCtrlProp().GetAllCtrlProp()
	certificateAuthorityData := ctrlProp[utils.ENV_CTRL_CADATA]

	cniPlugin, err := w.detectAndValidateCNI(cluster)
	if err != nil {
		return nil, fmt.Errorf("CNI detection failed for cluster %s/%s (UID: %s): %v", clusterNamespace, cluster.GetName(), cluster.GetUID(), err)
	}

	serviceType := "NodePort"
	if cniPlugin == "antrea" {
		serviceType = "NodePortLocal"
	}

	config := &VKSClusterConfig{
		Username:                 clusterCreds.Username,
		Password:                 clusterCreds.Password,
		ControllerHost:           "127.0.0.1",
		ControllerAddress:        controllerIP,
		ControllerVersion:        lib.GetControllerVersion(),
		CertificateAuthorityData: certificateAuthorityData,
		ServiceEngineGroup:       nsConfig.ServiceEngineGroup,
		TenantName:               nsConfig.Tenant,
		NsxtT1LR:                 nsConfig.T1LR,
		CloudName:                utils.CloudName,
		CNIPlugin:                cniPlugin,
		ServiceType:              serviceType,
		ClusterName:              clusterNameWithUID,
		VPCMode:                  true,
		DedicatedTenantMode:      true,
		Managed:                  true,
	}

	if config.ControllerVersion == "" {
		utils.AviLog.Warnf("Controller version not available for cluster %s/%s (UID: %s)", clusterNamespace, cluster.GetName(), cluster.GetUID())
	}

	utils.AviLog.Infof("Built configuration for cluster %s/%s (UID: %s): SEG=%s, Tenant=%s, T1LR=%s, Cloud=%s, CNI=%s, ServiceType=%s",
		clusterNamespace, cluster.GetName(), cluster.GetUID(), config.ServiceEngineGroup, config.TenantName, config.NsxtT1LR, config.CloudName, config.CNIPlugin, config.ServiceType)

	return config, nil
}

// UpsertAviCredentialsSecret creates or updates the Avi credentials secret for a VKS cluster
func (w *VKSClusterWatcher) UpsertAviCredentialsSecret(ctx context.Context, cluster *unstructured.Unstructured) error {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)

	utils.AviLog.Infof("VKS cluster watcher: processing secret upsert for cluster %s/%s", clusterNamespace, clusterName)

	config, err := w.buildVKSClusterConfig(cluster)
	if err != nil {
		return fmt.Errorf("failed to build cluster configuration: %v", err)
	}

	secretData := w.buildSecretData(config)

	existingSecret, err := utils.GetInformers().SecretInformer.Lister().Secrets(clusterNamespace).Get(secretName)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("failed to get existing secret %s: %v", secretName, err)
	}

	if errors.IsNotFound(err) {
		utils.AviLog.Infof("VKS cluster watcher: secret %s/%s not found, creating new secret", clusterNamespace, secretName)
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

		_, err = w.kubeClient.CoreV1().Secrets(clusterNamespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Avi credentials secret %s: %v", secretName, err)
		}

		utils.AviLog.Infof("Created Avi credentials secret %s/%s for cluster %s with complete AKO configuration",
			clusterNamespace, secretName, clusterName)
	} else {
		utils.AviLog.Infof("VKS cluster watcher: secret %s/%s exists, checking if update needed", clusterNamespace, secretName)
		needsUpdate := false

		// Check if data fields changed
		for key, desiredValue := range secretData {
			if existingValue, exists := existingSecret.Data[key]; !exists || string(existingValue) != string(desiredValue) {
				utils.AviLog.Infof("VKS cluster watcher: Secret field %s changed for cluster %s: existing='%s', desired='%s'",
					key, clusterName, string(existingValue), string(desiredValue))
				needsUpdate = true
				break
			}
		}

		// Check for fields that should be removed
		if !needsUpdate {
			for key := range existingSecret.Data {
				if _, shouldExist := secretData[key]; !shouldExist {
					utils.AviLog.Infof("VKS cluster watcher: Secret field %s should be removed for cluster %s", key, clusterName)
					needsUpdate = true
					break
				}
			}
		}

		if !needsUpdate {
			requiredLabels := map[string]string{
				"ako.kubernetes.vmware.com/cluster":    clusterName,
				"ako.kubernetes.vmware.com/managed-by": "ako-infra",
			}
			if existingSecret.Labels == nil {
				utils.AviLog.Infof("VKS cluster watcher: Secret %s/%s is missing labels", clusterNamespace, secretName)
				needsUpdate = true
			} else {
				for key, expectedValue := range requiredLabels {
					if actualValue, exists := existingSecret.Labels[key]; !exists || actualValue != expectedValue {
						utils.AviLog.Infof("VKS cluster watcher: Secret label %s changed for cluster %s: existing='%s', expected='%s'",
							key, clusterName, actualValue, expectedValue)
						needsUpdate = true
						break
					}
				}
			}
		}

		utils.AviLog.Infof("VKS cluster watcher: Secret comparison result for cluster %s: needsUpdate=%t", clusterName, needsUpdate)

		if needsUpdate {
			existingSecret.Data = secretData

			if existingSecret.Labels == nil {
				existingSecret.Labels = make(map[string]string)
			}
			existingSecret.Labels["ako.kubernetes.vmware.com/cluster"] = clusterName
			existingSecret.Labels["ako.kubernetes.vmware.com/managed-by"] = "ako-infra"

			_, err = w.kubeClient.CoreV1().Secrets(clusterNamespace).Update(ctx, existingSecret, metav1.UpdateOptions{})
			if err != nil {
				return fmt.Errorf("failed to update Avi credentials secret %s: %v", secretName, err)
			}

			utils.AviLog.Infof("Updated Avi credentials secret %s/%s for cluster %s with latest configuration",
				clusterNamespace, secretName, clusterName)
		} else {
			utils.AviLog.Infof("VKS cluster watcher: Avi credentials secret %s/%s for cluster %s is up-to-date",
				clusterNamespace, secretName, clusterName)
		}
	}

	return nil
}

// buildSecretData builds the complete secret data for a VKS cluster
func (w *VKSClusterWatcher) buildSecretData(config *VKSClusterConfig) map[string][]byte {
	secretData := make(map[string][]byte)

	secretData["username"] = []byte(config.Username)
	secretData["controllerHost"] = []byte(config.ControllerHost)
	secretData["controllerAddress"] = []byte(config.ControllerAddress)

	secretData["clusterName"] = []byte(config.ClusterName)

	if config.Password != "" {
		secretData["password"] = []byte(config.Password)
	}

	if config.ControllerVersion != "" {
		secretData["controllerVersion"] = []byte(config.ControllerVersion)
	}

	if config.CertificateAuthorityData != "" {
		secretData["certificateAuthorityData"] = []byte(config.CertificateAuthorityData)
	}

	if config.NsxtT1LR != "" {
		secretData["nsxtT1LR"] = []byte(config.NsxtT1LR)
	}

	if config.ServiceEngineGroup != "" {
		secretData["serviceEngineGroupName"] = []byte(config.ServiceEngineGroup)
	}
	if config.TenantName != "" {
		secretData["tenantName"] = []byte(config.TenantName)
	}

	// Add VKS-specific fields from config
	if config.CNIPlugin != "" {
		secretData["cniPlugin"] = []byte(config.CNIPlugin)
	}

	if config.ServiceType != "" {
		secretData["serviceType"] = []byte(config.ServiceType)
	}

	secretData["cloudName"] = []byte(config.CloudName)

	secretData["vpcMode"] = []byte(strconv.FormatBool(config.VPCMode))
	secretData["dedicatedTenantMode"] = []byte(strconv.FormatBool(config.DedicatedTenantMode))
	secretData["managed"] = []byte(strconv.FormatBool(config.Managed))

	return secretData
}

// detectAndValidateCNI detects the CNI plugin used by the VKS cluster and returns the appropriate CNI name
func (w *VKSClusterWatcher) detectAndValidateCNI(cluster *unstructured.Unstructured) (string, error) {
	ctx := context.Background()
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	clusterBootstrapGVR := schema.GroupVersionResource{
		Group:    "run.tanzu.vmware.com",
		Version:  "v1alpha3",
		Resource: "clusterbootstraps",
	}

	clusterBootstrap, err := w.dynamicClient.Resource(clusterBootstrapGVR).Namespace(clusterNamespace).Get(ctx, clusterName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get ClusterBootstrap for cluster %s/%s: %v", clusterNamespace, clusterName, err)
	}

	cniRefName, found, err := unstructured.NestedString(clusterBootstrap.Object, "spec", "cni", "refName")
	if err != nil || !found {
		return "", fmt.Errorf("failed to get CNI refName from ClusterBootstrap %s/%s: %v", clusterNamespace, clusterName, err)
	}

	var detectedCNI string

	cniMapping := map[string]string{
		"antrea":         "antrea",
		"calico":         "calico",
		"canal":          "canal",
		"flannel":        "flannel",
		"cilium":         "cilium",
		"ovn-kubernetes": "ovn-kubernetes",
		"ovn":            "ovn-kubernetes",
		"ncp":            "ncp",
		"nsx":            "ncp",
		"openshift":      "openshift",
	}

	for packagePrefix, akoValue := range cniMapping {
		if strings.Contains(strings.ToLower(cniRefName), packagePrefix) {
			detectedCNI = akoValue
			utils.AviLog.Infof("VKS cluster watcher: detected %s CNI for cluster %s/%s (package: %s)",
				akoValue, clusterNamespace, clusterName, cniRefName)
			break
		}
	}

	if detectedCNI == "" {
		utils.AviLog.Warnf("VKS cluster watcher: unknown CNI package for cluster %s/%s (package: %s), using default CNI configuration",
			clusterNamespace, clusterName, cniRefName)
	}

	return detectedCNI, nil
}

// StartVKSClusterWatcherWithRetry starts cluster watcher with infinite retry
func StartVKSClusterWatcherWithRetry(stopCh <-chan struct{}, dynamicInformers *lib.DynamicInformers) {
	utils.AviLog.Infof("VKS cluster watcher: starting with infinite retry")

	retryInterval := 10 * time.Second

	for {
		if err := StartVKSClusterWatcher(stopCh, dynamicInformers); err != nil {
			utils.AviLog.Warnf("VKS cluster watcher: failed to start, will retry in %v: %v", retryInterval, err)

			// Wait before retry, but also check for shutdown
			select {
			case <-stopCh:
				utils.AviLog.Infof("VKS cluster watcher: shutdown signal received during retry wait")
				return
			case <-time.After(retryInterval):
				// Continue to next retry
				continue
			}
		} else {
			utils.AviLog.Infof("VKS cluster watcher: shutdown gracefully")
			return
		}
	}
}

// StartVKSClusterWatcher starts the VKS cluster watcher - refactored to return error
// It needs dynamic informers to be passed from the controller
func StartVKSClusterWatcher(stopCh <-chan struct{}, dynamicInformers *lib.DynamicInformers) error {
	utils.AviLog.Infof("Starting VKS cluster watcher for cluster lifecycle management")

	kubeClient := utils.GetInformers().ClientSet
	dynamicClient := lib.GetDynamicClientSet()

	if kubeClient == nil || dynamicClient == nil {
		return fmt.Errorf("VKS cluster watcher: missing required clients (kubeClient: %v, dynamicClient: %v)", kubeClient != nil, dynamicClient != nil)
	}

	if dynamicInformers == nil {
		return fmt.Errorf("VKS cluster watcher: dynamic informers not available")
	}

	if dynamicInformers.ClusterInformer == nil {
		return fmt.Errorf("VKS cluster watcher: cluster informer not available")
	}

	clusterWatcher := NewVKSClusterWatcher(kubeClient, dynamicClient)
	if err := clusterWatcher.Start(); err != nil {
		return fmt.Errorf("VKS cluster watcher: failed to start: %v", err)
	}

	clusterWatcher.AddVKSSecretEventHandler(stopCh)

	clusterEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Infof("Cluster ADD event")
			clusterWatcher.EnqueueCluster(obj, "ADD")
		},
		UpdateFunc: func(old, new interface{}) {
			oldCluster, okOld := old.(*unstructured.Unstructured)
			newCluster, okNew := new.(*unstructured.Unstructured)

			if okOld && okNew {
				// Check if this UPDATE actually matters for VKS processing
				needsProcessing := false
				var reasons []string

				// Check VKS management label changes
				oldLabels := oldCluster.GetLabels()
				newLabels := newCluster.GetLabels()
				oldVKSLabel := ""
				newVKSLabel := ""
				if oldLabels != nil {
					oldVKSLabel = oldLabels[webhook.VKSManagedLabel]
				}
				if newLabels != nil {
					newVKSLabel = newLabels[webhook.VKSManagedLabel]
				}
				if oldVKSLabel != newVKSLabel {
					needsProcessing = true
					reasons = append(reasons, fmt.Sprintf("VKS label changed: '%s' -> '%s'", oldVKSLabel, newVKSLabel))
				}

				// Check phase changes
				oldPhase, _, _ := unstructured.NestedString(oldCluster.Object, "status", "phase")
				newPhase, _, _ := unstructured.NestedString(newCluster.Object, "status", "phase")
				if oldPhase != newPhase {
					needsProcessing = true
					reasons = append(reasons, fmt.Sprintf("phase changed: '%s' -> '%s'", oldPhase, newPhase))
				}

				if needsProcessing {
					utils.AviLog.Infof("VKS cluster watcher: UPDATE event for %s/%s requires processing - %s",
						newCluster.GetNamespace(), newCluster.GetName(), strings.Join(reasons, ", "))
					clusterWatcher.EnqueueCluster(new, "UPDATE")
				} else {
					utils.AviLog.Debugf("VKS cluster watcher: UPDATE event for %s/%s skipped",
						newCluster.GetNamespace(), newCluster.GetName())
				}
			} else {
				utils.AviLog.Infof("Cluster UPDATE event")
				clusterWatcher.EnqueueCluster(new, "UPDATE")
			}
		},
		DeleteFunc: func(obj interface{}) {
			cluster, ok := obj.(*unstructured.Unstructured)
			if !ok {
				cacheObj, ok := obj.(cache.DeletedFinalStateUnknown)
				if !ok {
					utils.AviLog.Errorf("VKS cluster watcher: couldn't get object from cache %#v", obj)
					return
				}
				cluster, ok = cacheObj.Obj.(*unstructured.Unstructured)
				if !ok {
					utils.AviLog.Errorf("VKS cluster watcher: cache contained object that is not an Unstructured: %#v", obj)
					return
				}
			}
			utils.AviLog.Infof("Cluster DELETE event for %s/%s", cluster.GetNamespace(), cluster.GetName())
			clusterWatcher.EnqueueCluster(cluster, "DELETE")
		},
	}

	dynamicInformers.ClusterInformer.Informer().AddEventHandler(clusterEventHandler)
	go dynamicInformers.ClusterInformer.Informer().Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, dynamicInformers.ClusterInformer.Informer().HasSynced) {
		return fmt.Errorf("VKS cluster watcher: timed out waiting for cluster caches to sync")
	} else {
		utils.AviLog.Infof("VKS cluster watcher: caches synced for cluster informer")
	}

	// Wait for stop signal and cleanup
	<-stopCh
	clusterWatcher.Stop()
	utils.AviLog.Infof("VKS cluster watcher stopped")
	return nil
}

// SetTestMode enables test mode with mock credentials function
func (w *VKSClusterWatcher) SetTestMode(mockFunc func(string, string) (*lib.ClusterCredentials, error)) {
	w.testMode = true
	w.mockCredentialsFunc = mockFunc
	w.mockRBACFunc = func(clusterName string) error {
		utils.AviLog.Infof("Mock RBAC cleanup for cluster: %s", clusterName)
		return nil
	}
	w.mockCleanupFunc = func(clusterNameWithUID, tenant string) error {
		utils.AviLog.Infof("Mock: Successfully cleaned up Avi objects for cluster %s in tenant %s", clusterNameWithUID, tenant)
		// Simulate some processing time
		time.Sleep(50 * time.Millisecond)
		return nil
	}
}

func (w *VKSClusterWatcher) startPeriodicReconciler() {
	utils.AviLog.Infof("Starting VKS periodic reconciler with interval %v", VKSReconcileInterval)

	ticker := time.NewTicker(VKSReconcileInterval)
	w.reconcileTicker = ticker

	go func() {
		for {
			select {
			case <-ticker.C:
				utils.AviLog.Infof("VKS periodic reconciler: starting reconciliation cycle")
				w.reconcileAllClusters()
			case <-w.reconcileStopCh:
				utils.AviLog.Infof("VKS periodic reconciler: stopping")
				return
			}
		}
	}()
}

func (w *VKSClusterWatcher) stopPeriodicReconciler() {
	// Close the stop channel first to signal the goroutine to exit
	close(w.reconcileStopCh)

	// Then clean up the ticker after the goroutine has exited
	if w.reconcileTicker != nil {
		w.reconcileTicker.Stop()
		w.reconcileTicker = nil
	}

	utils.AviLog.Infof("VKS periodic reconciler stopped")
}

// reconcileAllClusters performs periodic reconciliation of all VKS-managed clusters
func (w *VKSClusterWatcher) reconcileAllClusters() {
	utils.AviLog.Infof("VKS reconciler: starting reconciliation cycle")

	clusterObjects, err := lib.GetDynamicInformers().ClusterInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Errorf("VKS reconciler: failed to list clusters: %v", err)
		return
	}

	activeVKSClusters := make(map[string]bool)

	for _, obj := range clusterObjects {
		cluster, ok := obj.(*unstructured.Unstructured)
		if !ok {
			utils.AviLog.Warnf("VKS reconciler: unexpected object type in cluster list: %T", obj)
			continue
		}

		labels := cluster.GetLabels()
		shouldManage := labels != nil && labels[webhook.VKSManagedLabel] == webhook.VKSManagedLabelValueTrue

		if shouldManage {
			clusterNameWithUID := w.getUniqueClusterName(cluster)
			activeVKSClusters[clusterNameWithUID] = true

			phase := w.GetClusterPhase(cluster)
			if phase == ClusterPhaseProvisioning || phase == ClusterPhaseProvisioned {
				utils.AviLog.Infof("VKS reconciler: reconciling cluster %s/%s (phase: %s)",
					cluster.GetNamespace(), cluster.GetName(), phase)
				w.EnqueueCluster(cluster, "RECONCILE")
			}
		}
	}

	// Clean up orphaned AVI objects for deleted VKS clusters
	orphanedCount := w.cleanupOrphanedAviObjects(activeVKSClusters)

	if orphanedCount > 0 {
		utils.AviLog.Infof("VKS reconciler: completed reconciliation cycle - orphaned cleaned: %d",
			orphanedCount)
	}
}

// cleanupOrphanedAviObjects identifies and cleans up AVI objects belonging to deleted VKS clusters
func (w *VKSClusterWatcher) cleanupOrphanedAviObjects(activeVKSClusters map[string]bool) int {
	utils.AviLog.Infof("VKS reconciler: checking for orphaned AVI objects")

	orphanedSecrets := w.findOrphanedVKSSecrets(activeVKSClusters)
	if len(orphanedSecrets) == 0 {
		return 0
	}

	totalOrphanedCount := 0
	for _, secretInfo := range orphanedSecrets {
		clusterNameWithUID := secretInfo.ClusterNameWithUID
		clusterName := strings.TrimSuffix(secretInfo.Name, "-avi-secret")

		utils.AviLog.Infof("VKS reconciler: processing orphaned cluster %s from secret %s/%s",
			clusterNameWithUID, secretInfo.Namespace, secretInfo.Name)

		utils.AviLog.Infof("VKS reconciler: starting async cleanup for orphaned cluster %s", clusterNameWithUID)
		w.cleanupClusterAviObjects(clusterName, clusterNameWithUID, secretInfo.Namespace)

		totalOrphanedCount++
	}

	if totalOrphanedCount > 0 {
		utils.AviLog.Infof("VKS reconciler: cleaned up %d orphaned clusters", totalOrphanedCount)
	}

	return totalOrphanedCount
}

// findOrphanedVKSSecrets finds VKS secrets for clusters that no longer exist
func (w *VKSClusterWatcher) findOrphanedVKSSecrets(activeVKSClusters map[string]bool) []VKSSecretInfo {
	var orphanedSecrets []VKSSecretInfo

	namespaces, err := utils.GetInformers().NSInformer.Lister().List(labels.Everything())
	if err != nil {
		utils.AviLog.Warnf("VKS reconciler: failed to list namespaces from cache for orphaned secret discovery: %v", err)
		return orphanedSecrets
	}

	for _, namespace := range namespaces {
		secrets, err := utils.GetInformers().SecretInformer.Lister().Secrets(namespace.Name).List(labels.SelectorFromSet(labels.Set{
			"ako.kubernetes.vmware.com/managed-by": "ako-infra",
		}))
		if err != nil {
			utils.AviLog.Infof("VKS reconciler: failed to list secrets from cache in namespace %s: %v", namespace.Name, err)
			continue
		}

		for _, secret := range secrets {
			if clusterNameData, exists := secret.Data["clusterName"]; exists {
				clusterNameWithUID := string(clusterNameData)

				// Check if this cluster is still active
				if !activeVKSClusters[clusterNameWithUID] {
					utils.AviLog.Infof("VKS reconciler: found orphaned secret %s/%s for deleted cluster %s",
						namespace.Name, secret.Name, clusterNameWithUID)

					orphanedSecrets = append(orphanedSecrets, VKSSecretInfo{
						Name:               secret.Name,
						Namespace:          namespace.Name,
						ClusterNameWithUID: clusterNameWithUID,
					})
				}
			}
		}
	}

	if len(orphanedSecrets) > 0 {
		utils.AviLog.Infof("VKS reconciler: found %d orphaned secrets for deleted clusters", len(orphanedSecrets))
	}

	return orphanedSecrets
}
