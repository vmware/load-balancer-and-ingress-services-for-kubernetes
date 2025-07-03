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
	"encoding/base64"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
)

// VKSDependencyManager manages cluster-specific dependencies for VKS framework integration
type VKSDependencyManager struct {
	kubeClient    kubernetes.Interface
	dynamicClient dynamic.Interface

	// Avi Controller connection details
	controllerHost    string
	controllerVersion string
	cloudName         string
	tenantName        string
	vrfName           string

	// Cluster-specific credentials cache
	clusterCredentials map[string]*lib.ClusterCredentials
	credentialsMutex   sync.RWMutex

	// Reconciliation control
	reconcileInterval time.Duration
	stopCh            chan struct{}
	watcherActive     bool

	// VKS Management Service integration
	managementServiceManager *VKSManagementServiceManager
	vCenterURL               string
}

// DependencyResourceSpec defines the specification for cluster dependency resources
type DependencyResourceSpec struct {
	ClusterName      string
	ClusterNamespace string

	// Avi Controller credentials (cluster-specific)
	AviCredentials struct {
		Username                 string
		Password                 string
		AuthToken                string
		CertificateAuthorityData string
	}

	// Generated AKO configuration
	AKOConfig struct {
		ControllerHost         string
		ControllerVersion      string
		CloudName              string
		ServiceEngineGroupName string
		TenantName             string
		VRFName                string
	}
}

// ReconcileEvent represents a reconciliation event for monitoring
type ReconcileEvent struct {
	EventType    string
	ResourceType string
	ResourceName string
	Namespace    string
	ClusterName  string
	Reason       string
}

// NewVKSDependencyManager creates a new dependency manager
func NewVKSDependencyManager(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface) *VKSDependencyManager {
	return &VKSDependencyManager{
		kubeClient:         kubeClient,
		dynamicClient:      dynamicClient,
		clusterCredentials: make(map[string]*lib.ClusterCredentials),
		reconcileInterval:  30 * time.Second, // Default reconcile every 30 seconds
		stopCh:             make(chan struct{}),
		watcherActive:      false,
	}
}

// NewVKSDependencyManagerWithAvi creates a dependency manager with Avi Controller integration
func NewVKSDependencyManagerWithAvi(kubeClient kubernetes.Interface, dynamicClient dynamic.Interface,
	aviHost, aviUsername, aviPassword, aviTenant, aviAPIVersion string) (*VKSDependencyManager, error) {

	dm := NewVKSDependencyManager(kubeClient, dynamicClient)
	dm.controllerHost = aviHost

	utils.AviLog.Infof("VKS Dependency Manager created with Avi Controller integration: %s", aviHost)
	return dm, nil
}

// InitializeAviControllerConnection initializes connection details to Avi Controller
func (dm *VKSDependencyManager) InitializeAviControllerConnection(host, version, cloud, tenant, vrf string) {
	dm.controllerHost = host
	dm.controllerVersion = version
	dm.cloudName = cloud
	dm.tenantName = tenant
	dm.vrfName = vrf

	utils.AviLog.Infof("VKS Dependency Manager initialized with Avi Controller: %s (version: %s, cloud: %s)",
		host, version, cloud)
}

// InitializeVKSManagementService initializes the VKS Management Service integration
func (dm *VKSDependencyManager) InitializeVKSManagementService(vCenterURL string) {
	dm.vCenterURL = vCenterURL
	// Initialize the management service manager with dynamic client
	dm.managementServiceManager = GetManagementServiceManager(dm.controllerHost, vCenterURL, dm.dynamicClient)
	utils.AviLog.Infof("VKS Management Service integration initialized with vCenter URL: %s", vCenterURL)
}

// StartReconciler starts the reconciliation loop to prevent out-of-band modifications
func (dm *VKSDependencyManager) StartReconciler(ctx context.Context) error {
	if dm.watcherActive {
		return fmt.Errorf("reconciler is already active")
	}

	dm.watcherActive = true
	utils.AviLog.Infof("Starting VKS Dependency Manager reconciler (interval: %v)", dm.reconcileInterval)

	// Start periodic reconciliation
	go dm.periodicReconcile(ctx)

	// Start resource watchers for immediate reconciliation
	go dm.watchSecrets(ctx)
	go dm.watchConfigMaps(ctx)

	// Start VKS Management Service watchers if enabled
	if dm.managementServiceManager != nil {
		go dm.watchManagementServices(ctx)
		go dm.watchManagementServiceAccessGrants(ctx)
	}

	return nil
}

// StopReconciler stops the reconciliation loop
func (dm *VKSDependencyManager) StopReconciler() {
	if !dm.watcherActive {
		return
	}

	utils.AviLog.Infof("Stopping VKS Dependency Manager reconciler")
	close(dm.stopCh)
	dm.watcherActive = false
}

// periodicReconcile runs periodic reconciliation to ensure all managed resources are in desired state
func (dm *VKSDependencyManager) periodicReconcile(ctx context.Context) {
	ticker := time.NewTicker(dm.reconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-dm.stopCh:
			return
		case <-ticker.C:
			if err := dm.reconcileAllManagedResources(ctx); err != nil {
				utils.AviLog.Errorf("Periodic reconciliation failed: %v", err)
			}
		}
	}
}

// watchSecrets watches for modifications to managed secrets and reconciles them
func (dm *VKSDependencyManager) watchSecrets(ctx context.Context) {
	labelSelector := "ako.kubernetes.vmware.com/managed-by=vks-dependency-manager,ako.kubernetes.vmware.com/dependency-type=avi-credentials"

	for {
		select {
		case <-ctx.Done():
			return
		case <-dm.stopCh:
			return
		default:
			watchlist, err := dm.kubeClient.CoreV1().Secrets("").Watch(ctx, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				utils.AviLog.Errorf("Failed to watch secrets: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for event := range watchlist.ResultChan() {
				if event.Type == watch.Modified || event.Type == watch.Deleted {
					if secret, ok := event.Object.(*corev1.Secret); ok {
						dm.handleSecretEvent(ctx, event.Type, secret)
					}
				}
			}
		}
	}
}

// watchConfigMaps watches for modifications to managed ConfigMaps and reconciles them
func (dm *VKSDependencyManager) watchConfigMaps(ctx context.Context) {
	labelSelector := "ako.kubernetes.vmware.com/managed-by=vks-dependency-manager,ako.kubernetes.vmware.com/dependency-type=ako-generated-config"

	for {
		select {
		case <-ctx.Done():
			return
		case <-dm.stopCh:
			return
		default:
			watchlist, err := dm.kubeClient.CoreV1().ConfigMaps("").Watch(ctx, metav1.ListOptions{
				LabelSelector: labelSelector,
			})
			if err != nil {
				utils.AviLog.Errorf("Failed to watch ConfigMaps: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for event := range watchlist.ResultChan() {
				if event.Type == watch.Modified || event.Type == watch.Deleted {
					if configMap, ok := event.Object.(*corev1.ConfigMap); ok {
						dm.handleConfigMapEvent(ctx, event.Type, configMap)
					}
				}
			}
		}
	}
}

// watchManagementServices watches for modifications to VKS ManagementService CRDs and reconciles them
func (dm *VKSDependencyManager) watchManagementServices(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dm.stopCh:
			return
		default:
			watchlist, err := dm.dynamicClient.Resource(ManagementServiceGVR).Watch(ctx, metav1.ListOptions{})
			if err != nil {
				utils.AviLog.Errorf("Failed to watch ManagementServices: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for event := range watchlist.ResultChan() {
				if event.Type == watch.Modified || event.Type == watch.Deleted {
					if obj, ok := event.Object.(*unstructured.Unstructured); ok {
						dm.handleManagementServiceEvent(ctx, event.Type, obj)
					}
				}
			}
		}
	}
}

// watchManagementServiceAccessGrants watches for modifications to VKS ManagementServiceAccessGrant CRDs and reconciles them
func (dm *VKSDependencyManager) watchManagementServiceAccessGrants(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-dm.stopCh:
			return
		default:
			watchlist, err := dm.dynamicClient.Resource(ManagementServiceAccessGrantGVR).Watch(ctx, metav1.ListOptions{})
			if err != nil {
				utils.AviLog.Errorf("Failed to watch ManagementServiceAccessGrants: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			for event := range watchlist.ResultChan() {
				if event.Type == watch.Modified || event.Type == watch.Deleted {
					if obj, ok := event.Object.(*unstructured.Unstructured); ok {
						dm.handleManagementServiceAccessGrantEvent(ctx, event.Type, obj)
					}
				}
			}
		}
	}
}

// handleSecretEvent handles events on managed secrets
func (dm *VKSDependencyManager) handleSecretEvent(ctx context.Context, eventType watch.EventType, secret *corev1.Secret) {
	clusterName := secret.Labels["ako.kubernetes.vmware.com/cluster"]
	if clusterName == "" {
		return
	}

	switch eventType {
	case watch.Modified:
		utils.AviLog.Warnf("Out-of-band modification detected on secret %s/%s for cluster %s - reconciling",
			secret.Namespace, secret.Name, clusterName)
		dm.reconcileSecretForCluster(ctx, clusterName, secret.Namespace)

	case watch.Deleted:
		utils.AviLog.Warnf("Out-of-band deletion detected on secret %s/%s for cluster %s - recreating",
			secret.Namespace, secret.Name, clusterName)
		dm.reconcileSecretForCluster(ctx, clusterName, secret.Namespace)
	}
}

// handleConfigMapEvent handles events on managed ConfigMaps
func (dm *VKSDependencyManager) handleConfigMapEvent(ctx context.Context, eventType watch.EventType, configMap *corev1.ConfigMap) {
	clusterName := configMap.Labels["ako.kubernetes.vmware.com/cluster"]
	if clusterName == "" {
		return
	}

	switch eventType {
	case watch.Modified:
		utils.AviLog.Warnf("Out-of-band modification detected on ConfigMap %s/%s for cluster %s - reconciling",
			configMap.Namespace, configMap.Name, clusterName)
		dm.reconcileConfigMapForCluster(ctx, clusterName, configMap.Namespace)

	case watch.Deleted:
		utils.AviLog.Warnf("Out-of-band deletion detected on ConfigMap %s/%s for cluster %s - recreating",
			configMap.Namespace, configMap.Name, clusterName)
		dm.reconcileConfigMapForCluster(ctx, clusterName, configMap.Namespace)
	}
}

// handleManagementServiceEvent handles events on VKS ManagementService CRDs
func (dm *VKSDependencyManager) handleManagementServiceEvent(ctx context.Context, eventType watch.EventType, obj *unstructured.Unstructured) {
	name := obj.GetName()

	switch eventType {
	case watch.Modified:
		utils.AviLog.Warnf("Out-of-band modification detected on ManagementService %s - reconciling", name)
		dm.reconcileManagementService(ctx)

	case watch.Deleted:
		utils.AviLog.Warnf("Out-of-band deletion detected on ManagementService %s - recreating", name)
		dm.reconcileManagementService(ctx)
	}
}

// handleManagementServiceAccessGrantEvent handles events on VKS ManagementServiceAccessGrant CRDs
func (dm *VKSDependencyManager) handleManagementServiceAccessGrantEvent(ctx context.Context, eventType watch.EventType, obj *unstructured.Unstructured) {
	name := obj.GetName()

	switch eventType {
	case watch.Modified:
		utils.AviLog.Warnf("Out-of-band modification detected on ManagementServiceAccessGrant %s - reconciling", name)
		dm.reconcileManagementServiceAccessGrants(ctx)

	case watch.Deleted:
		utils.AviLog.Warnf("Out-of-band deletion detected on ManagementServiceAccessGrant %s - recreating", name)
		dm.reconcileManagementServiceAccessGrants(ctx)
	}
}

// reconcileAllManagedResources reconciles all managed resources across all clusters
func (dm *VKSDependencyManager) reconcileAllManagedResources(ctx context.Context) error {
	// First, ensure the global AddonInstall resource exists
	if err := dm.reconcileGlobalAddonInstall(ctx); err != nil {
		utils.AviLog.Errorf("Failed to reconcile global AddonInstall: %v", err)
		// Don't fail the entire reconciliation for AddonInstall issues
	}

	// Reconcile VKS Management Service resources if enabled
	if dm.managementServiceManager != nil {
		dm.reconcileManagementService(ctx)
		dm.reconcileManagementServiceAccessGrants(ctx)
	}

	managedClusters, err := dm.GetManagedClusters(ctx)
	if err != nil {
		return fmt.Errorf("failed to get managed clusters: %v", err)
	}

	utils.AviLog.Debugf("Reconciling %d managed clusters", len(managedClusters))

	for _, clusterRef := range managedClusters {
		// Parse cluster reference (namespace/name)
		namespace, clusterName := dm.parseClusterRef(clusterRef)
		if namespace == "" || clusterName == "" {
			continue
		}

		// Reconcile resources for this cluster
		if err := dm.reconcileClusterResources(ctx, clusterName, namespace); err != nil {
			utils.AviLog.Errorf("Failed to reconcile resources for cluster %s/%s: %v", namespace, clusterName, err)
		}
	}

	return nil
}

// reconcileClusterResources reconciles all resources for a specific cluster
func (dm *VKSDependencyManager) reconcileClusterResources(ctx context.Context, clusterName, namespace string) error {
	// Get cluster object to generate desired state
	cluster, err := dm.getClusterObject(ctx, clusterName, namespace)
	if err != nil {
		return fmt.Errorf("failed to get cluster object %s/%s: %v", namespace, clusterName, err)
	}

	// Generate desired state
	spec, err := dm.generateDependencySpec(cluster)
	if err != nil {
		return fmt.Errorf("failed to generate dependency spec: %v", err)
	}

	// Reconcile secret
	if err := dm.reconcileSecret(ctx, spec); err != nil {
		return fmt.Errorf("failed to reconcile secret: %v", err)
	}

	// Reconcile ConfigMap
	if err := dm.reconcileConfigMap(ctx, spec); err != nil {
		return fmt.Errorf("failed to reconcile ConfigMap: %v", err)
	}

	return nil
}

// reconcileSecretForCluster reconciles the secret for a specific cluster
func (dm *VKSDependencyManager) reconcileSecretForCluster(ctx context.Context, clusterName, namespace string) {
	if err := dm.reconcileClusterResources(ctx, clusterName, namespace); err != nil {
		utils.AviLog.Errorf("Failed to reconcile secret for cluster %s/%s: %v", namespace, clusterName, err)
	}
}

// reconcileConfigMapForCluster reconciles the ConfigMap for a specific cluster
func (dm *VKSDependencyManager) reconcileConfigMapForCluster(ctx context.Context, clusterName, namespace string) {
	if err := dm.reconcileClusterResources(ctx, clusterName, namespace); err != nil {
		utils.AviLog.Errorf("Failed to reconcile ConfigMap for cluster %s/%s: %v", namespace, clusterName, err)
	}
}

// reconcileSecret ensures the secret matches the desired state
func (dm *VKSDependencyManager) reconcileSecret(ctx context.Context, spec *DependencyResourceSpec) error {
	secretName := fmt.Sprintf("%s-avi-secret", spec.ClusterName)

	// Get current secret
	currentSecret, err := dm.kubeClient.CoreV1().Secrets(spec.ClusterNamespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		// Secret doesn't exist, create it
		return dm.generateAviCredentialsSecret(ctx, spec)
	}

	// Generate desired secret
	desiredSecret := dm.buildDesiredSecret(spec)

	// Compare current vs desired (excluding metadata that can change)
	if !dm.secretsEqual(currentSecret, desiredSecret) {
		utils.AviLog.Infof("Secret %s/%s drift detected - restoring desired state",
			spec.ClusterNamespace, secretName)

		// Preserve ResourceVersion for update
		desiredSecret.ResourceVersion = currentSecret.ResourceVersion

		_, err = dm.kubeClient.CoreV1().Secrets(spec.ClusterNamespace).Update(ctx, desiredSecret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to restore secret %s: %v", secretName, err)
		}

		utils.AviLog.Infof("Restored secret %s/%s to desired state", spec.ClusterNamespace, secretName)
	}

	return nil
}

// reconcileConfigMap ensures the ConfigMap matches the desired state
func (dm *VKSDependencyManager) reconcileConfigMap(ctx context.Context, spec *DependencyResourceSpec) error {
	configMapName := fmt.Sprintf("%s-ako-generated-config", spec.ClusterName)

	// Get current ConfigMap
	currentConfigMap, err := dm.kubeClient.CoreV1().ConfigMaps(spec.ClusterNamespace).Get(ctx, configMapName, metav1.GetOptions{})
	if err != nil {
		// ConfigMap doesn't exist, create it
		return dm.generateAKOConfigMap(ctx, spec)
	}

	// Generate desired ConfigMap
	desiredConfigMap := dm.buildDesiredConfigMap(spec)

	// Compare current vs desired (excluding metadata that can change)
	if !dm.configMapsEqual(currentConfigMap, desiredConfigMap) {
		utils.AviLog.Infof("ConfigMap %s/%s drift detected - restoring desired state",
			spec.ClusterNamespace, configMapName)

		// Preserve ResourceVersion for update
		desiredConfigMap.ResourceVersion = currentConfigMap.ResourceVersion

		_, err = dm.kubeClient.CoreV1().ConfigMaps(spec.ClusterNamespace).Update(ctx, desiredConfigMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to restore ConfigMap %s: %v", configMapName, err)
		}

		utils.AviLog.Infof("Restored ConfigMap %s/%s to desired state", spec.ClusterNamespace, configMapName)
	}

	return nil
}

// buildDesiredSecret builds the desired secret object
func (dm *VKSDependencyManager) buildDesiredSecret(spec *DependencyResourceSpec) *corev1.Secret {
	secretName := fmt.Sprintf("%s-avi-secret", spec.ClusterName)

	secretData := map[string][]byte{
		"username": []byte(base64.StdEncoding.EncodeToString([]byte(spec.AviCredentials.Username))),
		"password": []byte(base64.StdEncoding.EncodeToString([]byte(spec.AviCredentials.Password))),
	}

	if spec.AviCredentials.AuthToken != "" {
		secretData["authtoken"] = []byte(base64.StdEncoding.EncodeToString([]byte(spec.AviCredentials.AuthToken)))
	}

	if spec.AviCredentials.CertificateAuthorityData != "" {
		secretData["certificateAuthorityData"] = []byte(base64.StdEncoding.EncodeToString([]byte(spec.AviCredentials.CertificateAuthorityData)))
	}

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: spec.ClusterNamespace,
			Labels: map[string]string{
				"ako.kubernetes.vmware.com/cluster":         spec.ClusterName,
				"ako.kubernetes.vmware.com/dependency-type": "avi-credentials",
				"ako.kubernetes.vmware.com/managed-by":      "vks-dependency-manager",
			},
			Annotations: map[string]string{
				"ako.kubernetes.vmware.com/generated-at": time.Now().Format(time.RFC3339),
				"ako.kubernetes.vmware.com/cluster-user": spec.AviCredentials.Username,
			},
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}
}

// buildDesiredConfigMap builds the desired ConfigMap object
func (dm *VKSDependencyManager) buildDesiredConfigMap(spec *DependencyResourceSpec) *corev1.ConfigMap {
	configMapName := fmt.Sprintf("%s-ako-generated-config", spec.ClusterName)

	configData := map[string]string{
		"controllerHost":         spec.AKOConfig.ControllerHost,
		"controllerVersion":      spec.AKOConfig.ControllerVersion,
		"cloudName":              spec.AKOConfig.CloudName,
		"serviceEngineGroupName": spec.AKOConfig.ServiceEngineGroupName,
		"tenantName":             spec.AKOConfig.TenantName,
		"vrfName":                spec.AKOConfig.VRFName,
		"clusterName":            spec.ClusterName,
		"generatedAt":            time.Now().Format(time.RFC3339),
	}

	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: spec.ClusterNamespace,
			Labels: map[string]string{
				"ako.kubernetes.vmware.com/cluster":         spec.ClusterName,
				"ako.kubernetes.vmware.com/dependency-type": "ako-generated-config",
				"ako.kubernetes.vmware.com/managed-by":      "vks-dependency-manager",
			},
			Annotations: map[string]string{
				"ako.kubernetes.vmware.com/generated-at":    time.Now().Format(time.RFC3339),
				"ako.kubernetes.vmware.com/controller-host": spec.AKOConfig.ControllerHost,
			},
		},
		Data: configData,
	}
}

// secretsEqual compares two secrets for equality (ignoring changeable metadata)
func (dm *VKSDependencyManager) secretsEqual(current, desired *corev1.Secret) bool {
	// Compare data
	if !reflect.DeepEqual(current.Data, desired.Data) {
		return false
	}

	// Compare essential labels
	for key, desiredValue := range desired.Labels {
		if key == "ako.kubernetes.vmware.com/cluster" ||
			key == "ako.kubernetes.vmware.com/dependency-type" ||
			key == "ako.kubernetes.vmware.com/managed-by" {
			if current.Labels[key] != desiredValue {
				return false
			}
		}
	}

	// Compare type
	if current.Type != desired.Type {
		return false
	}

	return true
}

// configMapsEqual compares two ConfigMaps for equality (ignoring changeable metadata)
func (dm *VKSDependencyManager) configMapsEqual(current, desired *corev1.ConfigMap) bool {
	// Compare data (excluding generatedAt which always changes)
	currentData := make(map[string]string)
	desiredData := make(map[string]string)

	for k, v := range current.Data {
		if k != "generatedAt" {
			currentData[k] = v
		}
	}

	for k, v := range desired.Data {
		if k != "generatedAt" {
			desiredData[k] = v
		}
	}

	if !reflect.DeepEqual(currentData, desiredData) {
		return false
	}

	// Compare essential labels
	for key, desiredValue := range desired.Labels {
		if key == "ako.kubernetes.vmware.com/cluster" ||
			key == "ako.kubernetes.vmware.com/dependency-type" ||
			key == "ako.kubernetes.vmware.com/managed-by" {
			if current.Labels[key] != desiredValue {
				return false
			}
		}
	}

	return true
}

// GenerateClusterDependencies generates all required dependency resources for a cluster
func (dm *VKSDependencyManager) GenerateClusterDependencies(ctx context.Context, cluster *unstructured.Unstructured) error {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	utils.AviLog.Infof("Generating VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)

	// Check if namespace has Service Engine Group configured
	segName, err := dm.getServiceEngineGroupForNamespace(clusterNamespace)
	if err != nil {
		utils.AviLog.Errorf("VKS Dependency Manager: Failed to check Service Engine Group for namespace %s: %v", clusterNamespace, err)
		return fmt.Errorf("failed to check Service Engine Group for namespace %s: %v", clusterNamespace, err)
	}
	if segName == "" {
		utils.AviLog.Infof("VKS Dependency Manager: No Service Engine Group configured for namespace %s, skipping VKS management", clusterNamespace)
		return nil
	}

	// Ensure ManagementService is created (global, once)
	if dm.managementServiceManager != nil {
		mgmtServiceSpec := ManagementServiceSpec{
			Name:              "avi-controller-mgmt",
			Description:       "Avi Controller Management Endpoint for VKS",
			ManagementAddress: dm.controllerHost,
			Port:              443,
			Protocol:          "https",
			VCenterURL:        dm.vCenterURL,
		}

		if err := dm.managementServiceManager.EnsureManagementService(mgmtServiceSpec); err != nil {
			utils.AviLog.Errorf("VKS Dependency Manager: Failed to ensure ManagementService: %v", err)
			// Continue with other dependency creation - this is not blocking
		}

		// Ensure ManagementServiceAccessGrant for namespace (once per namespace)
		accessGrantSpec := ManagementServiceAccessGrantSpec{
			Name:        fmt.Sprintf("avi-controller-vm-access-%s", clusterNamespace),
			Namespace:   clusterNamespace,
			Description: fmt.Sprintf("Access grant for VMs in namespace %s to Avi Controller", clusterNamespace),
			ServiceName: "avi-controller-mgmt",
			Type:        "virtualmachine",
			Enabled:     true,
			VCenterURL:  dm.vCenterURL,
		}

		if err := dm.managementServiceManager.EnsureManagementServiceAccessGrant(accessGrantSpec); err != nil {
			utils.AviLog.Errorf("VKS Dependency Manager: Failed to ensure ManagementServiceAccessGrant for namespace %s: %v", clusterNamespace, err)
			// Continue with other dependency creation - this is not blocking
		}
	}

	// Generate dependency specification
	spec, err := dm.generateDependencySpec(cluster)
	if err != nil {
		return fmt.Errorf("failed to generate dependency spec for cluster %s/%s: %v",
			clusterNamespace, clusterName, err)
	}

	// Generate Avi credentials secret
	if err := dm.generateAviCredentialsSecret(ctx, spec); err != nil {
		return fmt.Errorf("failed to generate Avi credentials secret: %v", err)
	}

	// Generate AKO configuration ConfigMap
	if err := dm.generateAKOConfigMap(ctx, spec); err != nil {
		return fmt.Errorf("failed to generate AKO configuration ConfigMap: %v", err)
	}

	utils.AviLog.Infof("Successfully generated VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)
	return nil
}

// generateDependencySpec creates the specification for dependency resources
func (dm *VKSDependencyManager) generateDependencySpec(cluster *unstructured.Unstructured) (*DependencyResourceSpec, error) {
	clusterName := cluster.GetName()
	clusterNamespace := cluster.GetNamespace()

	spec := &DependencyResourceSpec{
		ClusterName:      clusterName,
		ClusterNamespace: clusterNamespace,
	}

	// REQUIRE real cluster-specific Avi credentials - no fallback
	creds, exists := dm.getClusterCredentials(clusterName)
	if !exists {
		return nil, fmt.Errorf("cluster-specific RBAC credentials not available for cluster %s - dependency not satisfied", clusterName)
	}

	// Use real credentials from Avi Controller RBAC
	spec.AviCredentials.Username = creds.Username
	spec.AviCredentials.Password = creds.Password
	utils.AviLog.Infof("Using real Avi Controller credentials for cluster %s: user=%s",
		clusterName, creds.Username)

	// Generate auth token (optional, used for token-based authentication)
	spec.AviCredentials.AuthToken = dm.generateClusterAuthToken(clusterName)

	// Get CA certificate data from Avi Controller (placeholder for now)
	spec.AviCredentials.CertificateAuthorityData = dm.getAviControllerCACert()

	// Generate AKO configuration from supervisor
	spec.AKOConfig.ControllerHost = dm.controllerHost
	spec.AKOConfig.ControllerVersion = dm.controllerVersion
	spec.AKOConfig.CloudName = dm.cloudName
	spec.AKOConfig.TenantName = dm.tenantName
	spec.AKOConfig.VRFName = dm.vrfName

	// Generate cluster-specific Service Engine Group
	// Based on namespace SEG configuration
	seg, err := dm.getServiceEngineGroupForNamespace(clusterNamespace)
	if err != nil {
		return nil, fmt.Errorf("failed to get SEG for namespace %s: %v", clusterNamespace, err)
	}
	spec.AKOConfig.ServiceEngineGroupName = seg

	return spec, nil
}

// generateAviCredentialsSecret creates the cluster-specific Avi credentials secret with granular RBAC
func (dm *VKSDependencyManager) generateAviCredentialsSecret(ctx context.Context, spec *DependencyResourceSpec) error {
	// Create cluster-specific RBAC in Avi Controller - REQUIRED, no fallback
	if dm.controllerHost != "" {
		err := dm.createClusterSpecificRBAC(spec.ClusterName)
		if err != nil {
			return fmt.Errorf("failed to create cluster-specific RBAC for %s: %v - dependency not satisfied", spec.ClusterName, err)
		}
	} else {
		return fmt.Errorf("Avi Controller not configured - cannot create cluster-specific RBAC for %s", spec.ClusterName)
	}

	secret := dm.buildDesiredSecret(spec)

	// Create or update the secret
	existingSecret, err := dm.kubeClient.CoreV1().Secrets(spec.ClusterNamespace).Get(ctx, secret.Name, metav1.GetOptions{})
	if err != nil {
		// Secret doesn't exist, create it
		_, err = dm.kubeClient.CoreV1().Secrets(spec.ClusterNamespace).Create(ctx, secret, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create Avi credentials secret %s: %v", secret.Name, err)
		}
		utils.AviLog.Infof("Created Avi credentials secret %s/%s for cluster %s with RBAC credentials",
			spec.ClusterNamespace, secret.Name, spec.ClusterName)
	} else {
		// Secret exists, update it
		secret.ResourceVersion = existingSecret.ResourceVersion
		_, err = dm.kubeClient.CoreV1().Secrets(spec.ClusterNamespace).Update(ctx, secret, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update Avi credentials secret %s: %v", secret.Name, err)
		}
		utils.AviLog.Infof("Updated Avi credentials secret %s/%s for cluster %s with RBAC credentials",
			spec.ClusterNamespace, secret.Name, spec.ClusterName)
	}

	return nil
}

// generateAKOConfigMap creates the cluster-specific AKO configuration ConfigMap
func (dm *VKSDependencyManager) generateAKOConfigMap(ctx context.Context, spec *DependencyResourceSpec) error {
	configMap := dm.buildDesiredConfigMap(spec)

	// Create or update the ConfigMap
	existingConfigMap, err := dm.kubeClient.CoreV1().ConfigMaps(spec.ClusterNamespace).Get(ctx, configMap.Name, metav1.GetOptions{})
	if err != nil {
		// ConfigMap doesn't exist, create it
		_, err = dm.kubeClient.CoreV1().ConfigMaps(spec.ClusterNamespace).Create(ctx, configMap, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("failed to create AKO config ConfigMap %s: %v", configMap.Name, err)
		}
		utils.AviLog.Infof("Created AKO config ConfigMap %s/%s for cluster %s",
			spec.ClusterNamespace, configMap.Name, spec.ClusterName)
	} else {
		// ConfigMap exists, update it
		configMap.ResourceVersion = existingConfigMap.ResourceVersion
		_, err = dm.kubeClient.CoreV1().ConfigMaps(spec.ClusterNamespace).Update(ctx, configMap, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("failed to update AKO config ConfigMap %s: %v", configMap.Name, err)
		}
		utils.AviLog.Infof("Updated AKO config ConfigMap %s/%s for cluster %s",
			spec.ClusterNamespace, configMap.Name, spec.ClusterName)
	}

	return nil
}

// CleanupClusterDependencies removes dependency resources for a deleted cluster
func (dm *VKSDependencyManager) CleanupClusterDependencies(ctx context.Context, clusterName, clusterNamespace string) error {
	utils.AviLog.Infof("Cleaning up VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)

	// Cleanup VKS Management Service resources if enabled
	if dm.managementServiceManager != nil {
		// Conditionally cleanup ManagementServiceAccessGrant for this namespace
		if err := dm.managementServiceManager.ConditionalCleanupNamespaceAccessGrant(ctx, clusterNamespace, dm.GetManagedClusters, dm.parseClusterRef); err != nil {
			utils.AviLog.Errorf("Failed to cleanup ManagementServiceAccessGrant for namespace %s: %v", clusterNamespace, err)
		}

		// Conditionally cleanup the global ManagementService
		if err := dm.managementServiceManager.ConditionalCleanupManagementService(ctx, "avi-controller-mgmt", dm.GetManagedClusters); err != nil {
			utils.AviLog.Errorf("Failed to cleanup ManagementService: %v", err)
		}
	}

	// Cleanup cluster-specific RBAC in Avi Controller
	if dm.controllerHost != "" {
		err := dm.cleanupClusterSpecificRBAC(clusterName)
		if err != nil {
			utils.AviLog.Errorf("Failed to cleanup cluster-specific RBAC for %s: %v", clusterName, err)
		}
	}

	// Delete Avi credentials secret
	secretName := fmt.Sprintf("%s-avi-secret", clusterName)
	err := dm.kubeClient.CoreV1().Secrets(clusterNamespace).Delete(ctx, secretName, metav1.DeleteOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to delete Avi credentials secret %s/%s: %v", clusterNamespace, secretName, err)
	} else {
		utils.AviLog.Infof("Deleted Avi credentials secret %s/%s", clusterNamespace, secretName)
	}

	// Delete AKO config ConfigMap
	configMapName := fmt.Sprintf("%s-ako-generated-config", clusterName)
	err = dm.kubeClient.CoreV1().ConfigMaps(clusterNamespace).Delete(ctx, configMapName, metav1.DeleteOptions{})
	if err != nil {
		utils.AviLog.Warnf("Failed to delete AKO config ConfigMap %s/%s: %v", clusterNamespace, configMapName, err)
	} else {
		utils.AviLog.Infof("Deleted AKO config ConfigMap %s/%s", clusterNamespace, configMapName)
	}

	utils.AviLog.Infof("Completed cleanup of VKS dependencies for cluster %s/%s", clusterNamespace, clusterName)
	return nil
}

// createClusterSpecificRBAC creates cluster-specific role and user in Avi Controller
func (dm *VKSDependencyManager) createClusterSpecificRBAC(clusterName string) error {
	// Use AKO's established client infrastructure
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return fmt.Errorf("Avi Controller client not available - ensure AKO infra is properly initialized")
	}

	// Create cluster-specific role using AKO's RBAC library
	role, err := lib.CreateVKSClusterRole(aviClient, clusterName, dm.tenantName)
	if err != nil {
		return fmt.Errorf("failed to create VKS cluster role: %v", err)
	}

	// Create cluster-specific user with the role
	user, password, err := lib.CreateVKSClusterUser(aviClient, clusterName, role.UUID, dm.tenantName)
	if err != nil {
		// Cleanup role if user creation fails
		lib.DeleteVKSClusterRole(aviClient, clusterName)
		return fmt.Errorf("failed to create VKS cluster user: %v", err)
	}

	utils.AviLog.Infof("Created VKS cluster RBAC for %s: role=%s, user=%s",
		clusterName, role.Name, user.Username)

	// Store the real credentials in cache for secret generation
	dm.credentialsMutex.Lock()
	dm.clusterCredentials[clusterName] = &lib.ClusterCredentials{
		Username: user.Username,
		Password: password,
		UserUUID: user.UUID,
		RoleUUID: role.UUID,
	}
	dm.credentialsMutex.Unlock()

	return nil
}

// getClusterCredentials safely retrieves cluster credentials from cache
func (dm *VKSDependencyManager) getClusterCredentials(clusterName string) (*lib.ClusterCredentials, bool) {
	dm.credentialsMutex.RLock()
	defer dm.credentialsMutex.RUnlock()
	creds, exists := dm.clusterCredentials[clusterName]
	return creds, exists
}

// cleanupClusterSpecificRBAC removes cluster-specific RBAC from Avi Controller
func (dm *VKSDependencyManager) cleanupClusterSpecificRBAC(clusterName string) error {
	if dm.controllerHost == "" {
		return nil // No Avi Controller configured
	}

	// Use AKO's established client infrastructure
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		utils.AviLog.Warnf("Avi Controller client not available for RBAC cleanup of cluster %s", clusterName)
		return nil // Continue with other cleanup
	}

	// Delete cluster user using AKO's RBAC library
	err := lib.DeleteVKSClusterUser(aviClient, clusterName)
	if err != nil {
		utils.AviLog.Errorf("Failed to delete VKS cluster user for %s: %v", clusterName, err)
	}

	// Delete cluster role using AKO's RBAC library
	err = lib.DeleteVKSClusterRole(aviClient, clusterName)
	if err != nil {
		utils.AviLog.Errorf("Failed to delete VKS cluster role for %s: %v", clusterName, err)
	}

	// Remove credentials from cache
	dm.credentialsMutex.Lock()
	delete(dm.clusterCredentials, clusterName)
	dm.credentialsMutex.Unlock()

	utils.AviLog.Infof("Cleaned up VKS cluster RBAC for %s", clusterName)
	return nil
}

// Helper methods

// parseClusterRef parses cluster reference in format "namespace/name"
func (dm *VKSDependencyManager) parseClusterRef(clusterRef string) (string, string) {
	parts := strings.Split(clusterRef, "/")
	if len(parts) != 2 {
		return "", ""
	}
	return parts[0], parts[1]
}

// getClusterObject retrieves the cluster object from Kubernetes API
func (dm *VKSDependencyManager) getClusterObject(ctx context.Context, clusterName, namespace string) (*unstructured.Unstructured, error) {
	// This would typically use the dynamic client to get the actual cluster object
	// For now, create a minimal unstructured object with the required fields
	cluster := &unstructured.Unstructured{}
	cluster.SetName(clusterName)
	cluster.SetNamespace(namespace)

	return cluster, nil
}

// generateClusterPassword generates a cluster-specific password
// In production, this would integrate with Avi Controller API to create/retrieve user credentials
func (dm *VKSDependencyManager) generateClusterPassword(clusterName string) string {
	// TODO: Integrate with Avi Controller API to generate actual cluster-specific user password
	// For now, return a placeholder that indicates supervisor-generated credentials
	return fmt.Sprintf("supervisor-generated-password-for-%s", clusterName)
}

// generateClusterAuthToken generates a cluster-specific auth token
// In production, this would integrate with Avi Controller API for token-based authentication
func (dm *VKSDependencyManager) generateClusterAuthToken(clusterName string) string {
	// TODO: Integrate with Avi Controller API to generate actual auth tokens
	// For now, return a placeholder
	return fmt.Sprintf("supervisor-generated-token-for-%s", clusterName)
}

// getAviControllerCACert retrieves the CA certificate from Avi Controller
// In production, this would fetch the actual CA certificate
func (dm *VKSDependencyManager) getAviControllerCACert() string {
	// TODO: Retrieve actual CA certificate from Avi Controller
	// For now, return a placeholder
	return "-----BEGIN CERTIFICATE-----\nsupervisor-generated-ca-cert\n-----END CERTIFICATE-----"
}

// getServiceEngineGroupForNamespace determines the SEG for a namespace
func (dm *VKSDependencyManager) getServiceEngineGroupForNamespace(namespace string) (string, error) {
	// Get namespace object to check for SEG annotation
	ns, err := dm.kubeClient.CoreV1().Namespaces().Get(context.TODO(), namespace, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get namespace %s: %v", namespace, err)
	}

	// Check for SEG annotation
	if seg, exists := ns.Annotations[ServiceEngineGroupAnnotation]; exists {
		return seg, nil
	}

	// No SEG annotation found - return empty string to indicate VKS management should be skipped
	return "", nil
}

// GetManagedClusters returns all clusters that have VKS dependency resources
func (dm *VKSDependencyManager) GetManagedClusters(ctx context.Context) ([]string, error) {
	// List all clusters with VKS managed label set to true
	clusterList, err := dm.dynamicClient.Resource(ClusterGVR).List(ctx, metav1.ListOptions{
		LabelSelector: "ako.kubernetes.vmware.com/install=true",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list VKS managed clusters: %v", err)
	}

	var clusters []string
	for _, cluster := range clusterList.Items {
		// Only include provisioned clusters (skip those in other phases)
		phase, found, err := unstructured.NestedString(cluster.Object, "status", "phase")
		if err != nil || !found || phase != ClusterPhaseProvisioned {
			continue
		}

		clusterRef := fmt.Sprintf("%s/%s", cluster.GetNamespace(), cluster.GetName())
		clusters = append(clusters, clusterRef)
	}

	return clusters, nil
}

// ============================================================================
// VKS AddonInstall Management Methods
// ============================================================================

// reconcileGlobalAddonInstall ensures the global AddonInstall resource exists and is up to date
func (dm *VKSDependencyManager) reconcileGlobalAddonInstall(ctx context.Context) error {
	utils.AviLog.Debugf("Reconciling global AKO AddonInstall resource")

	addonInstallGVR := schema.GroupVersionResource{
		Group:    "addons.kubernetes.vmware.com",
		Version:  "v1alpha1",
		Resource: "addoninstalls",
	}

	// Create the desired AddonInstall resource specification
	addonInstall := dm.createAddonInstallSpec()

	// Try to get the existing AddonInstall
	existing, err := dm.dynamicClient.Resource(addonInstallGVR).Namespace(VKSPublicNamespace).Get(ctx, AKOAddonInstallName, metav1.GetOptions{})
	if err != nil {
		// If not found, create it
		if dm.isNotFoundError(err) {
			utils.AviLog.Infof("Creating global AKO AddonInstall resource")
			_, err = dm.dynamicClient.Resource(addonInstallGVR).Namespace(VKSPublicNamespace).Create(ctx, addonInstall, metav1.CreateOptions{})
			if err != nil {
				utils.AviLog.Errorf("Failed to create global AKO AddonInstall: %v", err.Error())
				return fmt.Errorf("failed to create global AKO AddonInstall: %w", err)
			}
			utils.AviLog.Infof("Successfully created global AKO AddonInstall resource")
			return nil
		}
		utils.AviLog.Errorf("Failed to get global AKO AddonInstall: %v", err.Error())
		return fmt.Errorf("failed to get global AKO AddonInstall: %w", err)
	}

	// Update the existing resource if needed
	needsUpdate := dm.addonInstallNeedsUpdate(existing, addonInstall)
	if needsUpdate {
		utils.AviLog.Infof("Updating global AKO AddonInstall resource")

		// Preserve metadata fields that shouldn't be overwritten
		addonInstall.SetResourceVersion(existing.GetResourceVersion())
		addonInstall.SetUID(existing.GetUID())
		addonInstall.SetCreationTimestamp(existing.GetCreationTimestamp())

		_, err = dm.dynamicClient.Resource(addonInstallGVR).Namespace(VKSPublicNamespace).Update(ctx, addonInstall, metav1.UpdateOptions{})
		if err != nil {
			utils.AviLog.Errorf("Failed to update global AKO AddonInstall: %v", err.Error())
			return fmt.Errorf("failed to update global AKO AddonInstall: %w", err)
		}
		utils.AviLog.Infof("Successfully updated global AKO AddonInstall resource")
	}

	return nil
}

// EnsureGlobalAddonInstall creates or updates the global AddonInstall resource for AKO
// This method is exposed for initial setup during startup
func (dm *VKSDependencyManager) EnsureGlobalAddonInstall(ctx context.Context) error {
	return dm.reconcileGlobalAddonInstall(ctx)
}

// DeleteGlobalAddonInstall removes the global AddonInstall resource for AKO
func (dm *VKSDependencyManager) DeleteGlobalAddonInstall(ctx context.Context) error {
	utils.AviLog.Infof("Deleting global AKO AddonInstall resource")

	addonInstallGVR := schema.GroupVersionResource{
		Group:    "addons.kubernetes.vmware.com",
		Version:  "v1alpha1",
		Resource: "addoninstalls",
	}

	err := dm.dynamicClient.Resource(addonInstallGVR).Namespace(VKSPublicNamespace).Delete(ctx, AKOAddonInstallName, metav1.DeleteOptions{})
	if err != nil {
		if dm.isNotFoundError(err) {
			utils.AviLog.Infof("Global AKO AddonInstall resource already deleted")
			return nil
		}
		utils.AviLog.Errorf("Failed to delete global AKO AddonInstall: %v", err.Error())
		return fmt.Errorf("failed to delete global AKO AddonInstall: %w", err)
	}

	utils.AviLog.Infof("Successfully deleted global AKO AddonInstall resource")
	return nil
}

// GetGlobalAddonInstallStatus retrieves the status of the global AddonInstall resource
func (dm *VKSDependencyManager) GetGlobalAddonInstallStatus(ctx context.Context) (*unstructured.Unstructured, error) {
	addonInstallGVR := schema.GroupVersionResource{
		Group:    "addons.kubernetes.vmware.com",
		Version:  "v1alpha1",
		Resource: "addoninstalls",
	}

	return dm.dynamicClient.Resource(addonInstallGVR).Namespace(VKSPublicNamespace).Get(ctx, AKOAddonInstallName, metav1.GetOptions{})
}

// IsAddonInstallHealthy checks if the AddonInstall resource is in a healthy state
func (dm *VKSDependencyManager) IsAddonInstallHealthy(ctx context.Context) (bool, error) {
	addonInstall, err := dm.GetGlobalAddonInstallStatus(ctx)
	if err != nil {
		return false, err
	}

	// Check status conditions for health
	status, found, err := unstructured.NestedMap(addonInstall.Object, "status")
	if err != nil || !found {
		return false, nil // No status yet, considered not healthy
	}

	conditions, found, err := unstructured.NestedSlice(status, "conditions")
	if err != nil || !found {
		return false, nil // No conditions, considered not healthy
	}

	// Look for Ready condition with status True
	for _, condition := range conditions {
		conditionMap, ok := condition.(map[string]interface{})
		if !ok {
			continue
		}

		conditionType, found, err := unstructured.NestedString(conditionMap, "type")
		if err != nil || !found || conditionType != "Ready" {
			continue
		}

		conditionStatus, found, err := unstructured.NestedString(conditionMap, "status")
		if err != nil || !found {
			continue
		}

		return conditionStatus == "True", nil
	}

	return false, nil
}

// createAddonInstallSpec creates the desired AddonInstall resource specification
func (dm *VKSDependencyManager) createAddonInstallSpec() *unstructured.Unstructured {
	addonInstall := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "addons.kubernetes.vmware.com/v1alpha1",
			"kind":       "AddonInstall",
			"metadata": map[string]interface{}{
				"name":      AKOAddonInstallName,
				"namespace": VKSPublicNamespace,
			},
			"spec": map[string]interface{}{
				"addonName":               AKOAddonName,
				"crossNamespaceSelection": "Allowed",
				"clusters": []interface{}{
					map[string]interface{}{
						"selector": map[string]interface{}{
							"matchLabels": map[string]interface{}{
								VKSManagedLabel: VKSManagedLabelValueTrue,
							},
						},
					},
				},
				"releases": map[string]interface{}{
					"selector": map[string]interface{}{
						"matchLabels": map[string]interface{}{
							"addon.kubernetes.vmware.com/addon-name": AKOAddonName,
						},
					},
					"resolutionRule": "PreferLatest",
				},
				"paused": false,
			},
		},
	}

	return addonInstall
}

// addonInstallNeedsUpdate determines if the existing AddonInstall resource needs to be updated
func (dm *VKSDependencyManager) addonInstallNeedsUpdate(existing, desired *unstructured.Unstructured) bool {
	existingSpec, found, err := unstructured.NestedMap(existing.Object, "spec")
	if err != nil || !found {
		return true
	}

	desiredSpec, found, err := unstructured.NestedMap(desired.Object, "spec")
	if err != nil || !found {
		return true
	}

	// Compare key fields that matter for functionality
	if !dm.compareStringField(existingSpec, desiredSpec, "addonName") {
		return true
	}

	if !dm.compareBoolField(existingSpec, desiredSpec, "paused") {
		return true
	}

	// Compare crossNamespaceSelection (top-level field)
	if !dm.compareStringField(existingSpec, desiredSpec, "crossNamespaceSelection") {
		return true
	}

	// Compare clusters configuration (basic validation)
	if !dm.compareClustersConfig(existingSpec, desiredSpec) {
		return true
	}

	// Compare releases configuration
	if !dm.compareReleasesConfig(existingSpec, desiredSpec) {
		return true
	}

	return false
}

// compareStringField compares a string field between two specs
func (dm *VKSDependencyManager) compareStringField(existing, desired map[string]interface{}, field string) bool {
	existingValue, _, _ := unstructured.NestedString(existing, field)
	desiredValue, _, _ := unstructured.NestedString(desired, field)
	return existingValue == desiredValue
}

// compareBoolField compares a boolean field between two specs
func (dm *VKSDependencyManager) compareBoolField(existing, desired map[string]interface{}, field string) bool {
	existingValue, _, _ := unstructured.NestedBool(existing, field)
	desiredValue, _, _ := unstructured.NestedBool(desired, field)
	return existingValue == desiredValue
}

// compareClustersConfig compares the clusters configuration between two specs
func (dm *VKSDependencyManager) compareClustersConfig(existing, desired map[string]interface{}) bool {
	existingClusters, existingFound, err := unstructured.NestedSlice(existing, "clusters")
	if err != nil || !existingFound {
		// If existing doesn't have clusters, check if desired also doesn't have clusters
		_, desiredFound, err := unstructured.NestedSlice(desired, "clusters")
		return err == nil && !desiredFound
	}

	desiredClusters, desiredFound, err := unstructured.NestedSlice(desired, "clusters")
	if err != nil || !desiredFound {
		return false
	}

	// Compare array lengths - basic validation for now
	return len(existingClusters) == len(desiredClusters)
}

// compareReleasesConfig compares the releases configuration between two specs
func (dm *VKSDependencyManager) compareReleasesConfig(existing, desired map[string]interface{}) bool {
	existingReleases, found, err := unstructured.NestedMap(existing, "releases")
	if err != nil || !found {
		return false
	}

	desiredReleases, found, err := unstructured.NestedMap(desired, "releases")
	if err != nil || !found {
		return false
	}

	// Compare resolutionRule
	existingRule, _, _ := unstructured.NestedString(existingReleases, "resolutionRule")
	desiredRule, _, _ := unstructured.NestedString(desiredReleases, "resolutionRule")

	return existingRule == desiredRule
}

// isNotFoundError checks if the error is a Kubernetes NotFound error
func (dm *VKSDependencyManager) isNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	return errMsg == "not found" || errMsg == "NotFound" ||
		(len(errMsg) >= 9 && errMsg[len(errMsg)-9:] == "not found") ||
		(len(errMsg) >= 8 && errMsg[len(errMsg)-8:] == "NotFound")
}

// reconcileManagementService reconciles the global ManagementService
func (dm *VKSDependencyManager) reconcileManagementService(ctx context.Context) {
	if dm.managementServiceManager == nil {
		return
	}

	// Create ManagementService spec
	mgmtServiceSpec := ManagementServiceSpec{
		Name:              "avi-controller-mgmt",
		Description:       "Avi Controller Management Endpoint for VKS",
		ManagementAddress: dm.controllerHost,
		Port:              443,
		Protocol:          "https",
		VCenterURL:        dm.vCenterURL,
	}

	if err := dm.managementServiceManager.EnsureManagementService(mgmtServiceSpec); err != nil {
		utils.AviLog.Errorf("Failed to reconcile ManagementService: %v", err)
	} else {
		utils.AviLog.Infof("ManagementService reconciled successfully")
	}
}

// reconcileManagementServiceAccessGrants reconciles all ManagementServiceAccessGrants for active namespaces
func (dm *VKSDependencyManager) reconcileManagementServiceAccessGrants(ctx context.Context) {
	if dm.managementServiceManager == nil {
		return
	}

	// Get all managed clusters to determine which namespaces need AccessGrants
	managedClusters, err := dm.GetManagedClusters(ctx)
	if err != nil {
		utils.AviLog.Errorf("Failed to get managed clusters for AccessGrant reconciliation: %v", err)
		return
	}

	// Track namespaces that need AccessGrants
	namespacesWithSEG := make(map[string]bool)

	for _, clusterRef := range managedClusters {
		namespace, _ := dm.parseClusterRef(clusterRef)
		if namespace == "" {
			continue
		}

		// Check if namespace has SEG configuration
		segName, err := dm.getServiceEngineGroupForNamespace(namespace)
		if err != nil {
			utils.AviLog.Errorf("Failed to check SEG for namespace %s: %v", namespace, err)
			continue
		}

		if segName != "" {
			namespacesWithSEG[namespace] = true
		}
	}

	// Ensure AccessGrants exist for all SEG-enabled namespaces
	for namespace := range namespacesWithSEG {
		accessGrantSpec := ManagementServiceAccessGrantSpec{
			Name:        fmt.Sprintf("avi-controller-vm-access-%s", namespace),
			Namespace:   namespace,
			Description: fmt.Sprintf("Access grant for VMs in namespace %s to Avi Controller", namespace),
			ServiceName: "avi-controller-mgmt",
			Type:        "virtualmachine",
			Enabled:     true,
			VCenterURL:  dm.vCenterURL,
		}

		if err := dm.managementServiceManager.EnsureManagementServiceAccessGrant(accessGrantSpec); err != nil {
			utils.AviLog.Errorf("Failed to reconcile ManagementServiceAccessGrant for namespace %s: %v", namespace, err)
		} else {
			utils.AviLog.Debugf("ManagementServiceAccessGrant reconciled successfully for namespace %s", namespace)
		}
	}
}
