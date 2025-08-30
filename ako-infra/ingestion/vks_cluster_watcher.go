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
	"strconv"
	"strings"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/webhook"
	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/k8s"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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
)

// VKSClusterConfig holds all the configuration needed for a VKS cluster's AKO deployment
type VKSClusterConfig struct {
	Username            string
	Password            string
	ControllerIP        string
	ControllerVersion   string
	VPCMode             bool
	DedicatedTenantMode bool
	Managed             bool

	// Namespace-specific configuration
	ServiceEngineGroup string
	TenantName         string
	NsxtT1LR           string

	CloudName string

	CNIPlugin   string
	ServiceType string
	ClusterName string
}

// VKSClusterWatcher monitors cluster lifecycle events for AKO addon management
type VKSClusterWatcher struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface
	workqueue     workqueue.RateLimitingInterface //nolint:staticcheck

	// For testing - allows injecting mock behavior
	testMode            bool
	mockCredentialsFunc func(string, string) (*lib.ClusterCredentials, error)
}

// getUniqueClusterName generates a unique cluster identifier using namespace, name and UID
func (w *VKSClusterWatcher) getUniqueClusterName(cluster *unstructured.Unstructured) string {
	namespace := cluster.GetNamespace()
	name := cluster.GetName()
	uid := cluster.GetUID()

	maxNamespaceLen := 15
	maxNameLen := 25
	maxUIDLen := 8

	if len(namespace) > maxNamespaceLen {
		namespace = namespace[:maxNamespaceLen]
	}
	if len(name) > maxNameLen {
		name = name[:maxNameLen]
	}
	if len(uid) > maxUIDLen {
		uid = uid[:maxUIDLen]
	}

	return fmt.Sprintf("%s-%s-%s", namespace, name, uid)
}

// NewVKSClusterWatcher creates a new cluster watcher instance
func NewVKSClusterWatcher(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *VKSClusterWatcher {
	workqueue := workqueue.NewNamedRateLimitingQueue(
		workqueue.DefaultControllerRateLimiter(), //nolint:staticcheck
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
			return w.handleClusterDeletion(namespace, name, "")
		}
		return fmt.Errorf("failed to get cluster %s: %v", key, err)
	}

	return w.handleClusterAddOrUpdate(cluster)
}

func (w *VKSClusterWatcher) handleClusterAddOrUpdate(cluster *unstructured.Unstructured) error {
	clusterNameWithUID := w.getUniqueClusterName(cluster)
	clusterNamespace := cluster.GetNamespace()

	phase := w.GetClusterPhase(cluster)
	utils.AviLog.Debugf("Processing cluster %s in phase: %s", clusterNameWithUID, phase)

	switch phase {
	case ClusterPhaseProvisioning, ClusterPhaseProvisioned:
		return w.HandleProvisionedCluster(cluster)
	case ClusterPhaseDeleting:
		return w.handleClusterDeletion(clusterNamespace, cluster.GetName(), clusterNameWithUID)
	default:
		utils.AviLog.Debugf("Cluster %s/%s (UID: %s) not in provisioning/provisioned state, skipping", clusterNamespace, cluster.GetName(), cluster.GetUID())
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
	utils.AviLog.Debugf("Processing cluster %s in phase: %s", clusterNameWithUID, phase)

	secretName := fmt.Sprintf("%s-avi-secret", cluster.GetName())
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := utils.GetInformers().SecretInformer.Lister().Secrets(clusterNamespace).Get(secretName)
	secretExists := err == nil

	switch {
	case shouldManage:
		if err := w.UpsertAviCredentialsSecret(ctx, cluster); err != nil {
			return fmt.Errorf("failed to manage dependencies for cluster %s/%s (UID: %s) in phase %s: %v", clusterNamespace, cluster.GetName(), cluster.GetUID(), phase, err)
		}

	case !shouldManage && secretExists:
		utils.AviLog.Infof("Cluster opted out of VKS management, cleaning up all VKS dependencies for cluster: %s/%s (UID: %s)", clusterNamespace, cluster.GetName(), cluster.GetUID())
		if err := w.cleanupClusterDependencies(ctx, cluster.GetName(), clusterNamespace, clusterNameWithUID); err != nil {
			return fmt.Errorf("failed to cleanup dependencies for cluster %s/%s (UID: %s): %v", clusterNamespace, cluster.GetName(), cluster.GetUID(), err)
		}
		utils.AviLog.Infof("Successfully cleaned up all VKS dependencies for opted-out cluster: %s/%s (UID: %s)", clusterNamespace, cluster.GetName(), cluster.GetUID())

	case !shouldManage && !secretExists:
		// Should not manage and no secret exists - already in desired state
		utils.AviLog.Debugf("Cluster %s/%s (UID: %s) not managed by VKS and no dependencies exist", clusterNamespace, cluster.GetName(), cluster.GetUID())
	}

	return nil
}

// handleClusterDeletion handles cleanup for deleted clusters
func (w *VKSClusterWatcher) handleClusterDeletion(namespace, name, clusterNameWithUID string) error {
	utils.AviLog.Infof("Cleaning up dependencies for deleted cluster: %s/%s", namespace, name)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := w.cleanupClusterDependencies(ctx, name, namespace, clusterNameWithUID); err != nil {
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
func (w *VKSClusterWatcher) cleanupClusterDependencies(ctx context.Context, clusterName, clusterNamespace, clusterNameWithUID string) error {
	utils.AviLog.Infof("Cleaning up VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)

	// If clusterNameWithUID is not provided, try to get it from the secret
	if clusterNameWithUID == "" {
		clusterNameWithUID = w.getClusterNameWithUIDFromSecret(clusterName, clusterNamespace)
	}

	if clusterNameWithUID != "" {
		if err := w.cleanupAviObjects(clusterNameWithUID); err != nil {
			utils.AviLog.Errorf("AVI objects cleanup failed for cluster %s/%s: %v", clusterNamespace, clusterName, err)
		}

		utils.AviLog.Infof("Cleaning up RBAC for cluster: %s", clusterNameWithUID)
		w.cleanupClusterSpecificRBAC(clusterNameWithUID)
		utils.AviLog.Infof("Successfully cleaned up RBAC for cluster: %s", clusterNameWithUID)
	} else {
		utils.AviLog.Infof("Could not determine cluster identifier for %s/%s - skipping AVI objects and RBAC cleanup", clusterNamespace, clusterName)
	}

	secretName := fmt.Sprintf("%s-avi-secret", clusterName)
	err := w.kubeClient.CoreV1().Secrets(clusterNamespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			utils.AviLog.Debugf("Avi credentials secret %s/%s already deleted", clusterNamespace, secretName)
		} else {
			utils.AviLog.Warnf("Failed to delete Avi credentials secret %s/%s: %v", clusterNamespace, secretName, err)
		}
	} else {
		utils.AviLog.Infof("Deleted Avi credentials secret %s/%s", clusterNamespace, secretName)
	}

	utils.AviLog.Infof("Completed cleanup of VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)
	return nil
}

// getClusterNameWithUIDFromSecret gets the cluster name with UID from the secret
// This is used during cluster deletion when the cluster object might be gone
func (w *VKSClusterWatcher) getClusterNameWithUIDFromSecret(clusterName, clusterNamespace string) string {
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)

	secret, err := w.kubeClient.CoreV1().Secrets(clusterNamespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err == nil && secret.Data != nil {
		if clusterNameBytes, exists := secret.Data["clusterName"]; exists {
			clusterNameWithUID := string(clusterNameBytes)
			utils.AviLog.Debugf("Retrieved cluster name with UID from secret: %s", clusterNameWithUID)
			return clusterNameWithUID
		}
	}

	utils.AviLog.Warnf("Could not get cluster UID for %s/%s from secret", clusterNamespace, clusterName)
	return ""
}

// cleanupClusterSpecificRBAC removes cluster-specific RBAC from Avi Controller
func (w *VKSClusterWatcher) cleanupClusterSpecificRBAC(clusterName string) {
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		utils.AviLog.Warnf("Avi Controller client not available for RBAC cleanup of cluster %s", clusterName)
		return
	}

	err := lib.DeleteClusterUser(aviClient, clusterName)
	if err != nil {
		utils.AviLog.Errorf("Failed to delete VKS cluster user for %s: %v", clusterName, err)
	}

	err = lib.DeleteClusterRoles(aviClient, clusterName)
	if err != nil {
		utils.AviLog.Errorf("Failed to delete VKS cluster roles for %s: %v", clusterName, err)
	}

	utils.AviLog.Infof("Cleaned up VKS cluster RBAC for %s", clusterName)
}

// createClusterSpecificCredentials creates cluster-specific RBAC credentials in Avi Controller
func (w *VKSClusterWatcher) createClusterSpecificCredentials(clusterNameWithUID string, operationalTenant string) (*lib.ClusterCredentials, error) {
	// Use mock in test mode
	if w.testMode && w.mockCredentialsFunc != nil {
		return w.mockCredentialsFunc(clusterNameWithUID, operationalTenant)
	}

	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return nil, fmt.Errorf("avi Controller client not available - ensure AKO infra is properly initialized")
	}

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
		return nil, fmt.Errorf("secret %s/%s not found in cache: %v", namespace, secretName, err)
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
	namespace, err := w.kubeClient.CoreV1().Namespaces().Get(context.Background(), clusterNamespace, metav1.GetOptions{})
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
		// Secret doesn't exist, create new credentials
		creds, err := w.createClusterSpecificCredentials(clusterNameWithUID, nsConfig.Tenant)
		if err != nil {
			return nil, fmt.Errorf("failed to create cluster-specific credentials: %v", err)
		}
		clusterCreds = creds
	}

	// Get common controller properties
	controllerIP := lib.GetControllerIP()
	if controllerIP == "" {
		return nil, fmt.Errorf("controller IP not set")
	}

	cniPlugin, err := w.detectAndValidateCNI(cluster)
	if err != nil {
		return nil, fmt.Errorf("CNI detection failed for cluster %s/%s (UID: %s): %v", clusterNamespace, cluster.GetName(), cluster.GetUID(), err)
	}

	serviceType := "NodePort"
	if cniPlugin == "antrea" {
		serviceType = "NodePortLocal"
	}

	config := &VKSClusterConfig{
		Username:            clusterCreds.Username,
		Password:            clusterCreds.Password,
		ControllerIP:        controllerIP,
		ControllerVersion:   lib.GetControllerVersion(),
		ServiceEngineGroup:  nsConfig.ServiceEngineGroup,
		TenantName:          nsConfig.Tenant,
		NsxtT1LR:            nsConfig.T1LR,
		CloudName:           utils.CloudName,
		CNIPlugin:           cniPlugin,
		ServiceType:         serviceType,
		ClusterName:         clusterNameWithUID,
		VPCMode:             true,
		DedicatedTenantMode: true,
		Managed:             true,
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

		utils.AviLog.Infof("Created Avi credentials secret %s/%s for cluster %s with complete AKO configuration",
			clusterNamespace, secretName, clusterName)
	} else {
		needsUpdate := false

		for key, desiredValue := range secretData {
			if existingValue, exists := existingSecret.Data[key]; !exists || string(existingValue) != string(desiredValue) {
				utils.AviLog.Debugf("Secret field %s changed for cluster %s: existing='%s', desired='%s'",
					key, clusterName, string(existingValue), string(desiredValue))
				needsUpdate = true
				break
			}
		}

		// Check for fields that should be removed
		if !needsUpdate {
			for key := range existingSecret.Data {
				if _, shouldExist := secretData[key]; !shouldExist {
					utils.AviLog.Debugf("Secret field %s should be removed for cluster %s", key, clusterName)
					needsUpdate = true
					break
				}
			}
		}

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
			utils.AviLog.Debugf("Avi credentials secret %s/%s for cluster %s is up-to-date",
				clusterNamespace, secretName, clusterName)
		}
	}

	return nil
}

// buildSecretData builds the complete secret data for a VKS cluster
func (w *VKSClusterWatcher) buildSecretData(config *VKSClusterConfig) map[string][]byte {
	secretData := make(map[string][]byte)

	secretData["username"] = []byte(config.Username)
	secretData["controllerIP"] = []byte(config.ControllerIP)

	secretData["clusterName"] = []byte(config.ClusterName)

	if config.Password != "" {
		secretData["password"] = []byte(config.Password)
	}

	if config.ControllerVersion != "" {
		secretData["controllerVersion"] = []byte(config.ControllerVersion)
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
		detectedCNI = ""
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
				lib.CleanupSharedRoles(avirest.InfraAviClientInstance())
				return
			case <-time.After(retryInterval):
				// Continue to next retry
				continue
			}
		} else {
			utils.AviLog.Infof("VKS cluster watcher: shutdown gracefully")
			lib.CleanupSharedRoles(avirest.InfraAviClientInstance())
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
	if err := clusterWatcher.Start(stopCh); err != nil {
		return fmt.Errorf("VKS cluster watcher: failed to start: %v", err)
	}

	clusterEventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			utils.AviLog.Debugf("Cluster ADD event")
			clusterWatcher.EnqueueCluster(obj, "ADD")
		},
		UpdateFunc: func(old, new interface{}) {
			utils.AviLog.Debugf("Cluster UPDATE event")
			clusterWatcher.EnqueueCluster(new, "UPDATE")
		},
		DeleteFunc: func(obj interface{}) {
			utils.AviLog.Debugf("Cluster DELETE event")
			clusterWatcher.EnqueueCluster(obj, "DELETE")
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

// cleanupAviObjects removes all AVI objects created by the cluster
func (w *VKSClusterWatcher) cleanupAviObjects(clusterNameWithUID string) error {
	utils.AviLog.Infof("Cleaning up AVI objects for cluster: %s", clusterNameWithUID)

	aviObjCache := avicache.SharedAviObjCache()
	if aviObjCache == nil {
		return fmt.Errorf("AVI object cache not available")
	}

	// Get all VS cache keys and filter for this cluster
	var parentVSKeys []avicache.NamespaceName
	vsCacheKeys := aviObjCache.VsCacheMeta.AviGetAllKeys()

	for _, vsKey := range vsCacheKeys {
		if strings.Contains(vsKey.Name, clusterNameWithUID) {
			parentVSKeys = append(parentVSKeys, vsKey)
		}
	}

	if len(parentVSKeys) == 0 {
		utils.AviLog.Infof("No parent VS keys found for cluster %s", clusterNameWithUID)
		return nil
	}

	utils.AviLog.Infof("Found %d parent VS keys for cluster %s", len(parentVSKeys), clusterNameWithUID)

	k8s.DeleteAviObjects(parentVSKeys, aviObjCache)

	utils.AviLog.Infof("Completed AVI object deletion for VKS cluster: %s", clusterNameWithUID)
	return nil
}

// SetTestMode enables test mode with mock credentials function
func (w *VKSClusterWatcher) SetTestMode(mockFunc func(string, string) (*lib.ClusterCredentials, error)) {
	w.testMode = true
	w.mockCredentialsFunc = mockFunc
}
