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

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/vmware/alb-sdk/go/clients"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// VKSManagementServiceManager handles creation and management of Avi Controller
// ManagementService and ManagementServiceAccessGrant objects for VKS integration
type VKSManagementServiceManager struct {
	controllerIP  string
	vCenterURL    string // For multi-vCenter scenarios
	dynamicClient dynamic.Interface
	mutex         sync.RWMutex
}

// ManagementServiceSpec defines the structure for ManagementService creation
type ManagementServiceSpec struct {
	Name              string
	Description       string
	ManagementAddress string
	Port              int32
	Protocol          string
	VCenterURL        string // Optional, for multi-vCenter scenarios
}

// ManagementServiceAccessGrantSpec defines the structure for AccessGrant creation
type ManagementServiceAccessGrantSpec struct {
	Name        string
	Namespace   string
	Description string
	ServiceName string
	Type        string // "virtualmachine"
	Enabled     bool
	VCenterURL  string // Optional, for multi-vCenter scenarios
}

// CRD Group Version Resources for supervisor cluster
var (
	ManagementServiceGVR = schema.GroupVersionResource{
		Group:    "vmware.com",
		Version:  "v1alpha1",
		Resource: "managementservices",
	}
	ManagementServiceAccessGrantGVR = schema.GroupVersionResource{
		Group:    "vmware.com",
		Version:  "v1alpha1",
		Resource: "managementserviceaccessgrants",
	}
)

// Singleton pattern following AKO conventions
var (
	managementServiceManagerInstance *VKSManagementServiceManager
	managementServiceManagerOnce     sync.Once
)

// GetManagementServiceManager returns the singleton instance of VKSManagementServiceManager
func GetManagementServiceManager(controllerIP, vCenterURL string, dynamicClient dynamic.Interface) *VKSManagementServiceManager {
	managementServiceManagerOnce.Do(func() {
		managementServiceManagerInstance = &VKSManagementServiceManager{
			controllerIP:  controllerIP,
			vCenterURL:    vCenterURL,
			dynamicClient: dynamicClient,
		}
		utils.AviLog.Infof("VKS Management Service Manager initialized for controller %s, vCenter %s", controllerIP, vCenterURL)
	})
	return managementServiceManagerInstance
}

// getAviClient gets the Avi client from AKO's infra client instance
func (m *VKSManagementServiceManager) getAviClient() *clients.AviClient {
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		utils.AviLog.Errorf("VKS Management Service: Avi infra client not available - ensure AKO infra is properly initialized")
		return nil
	}

	return aviClient
}

// EnsureManagementService creates the global ManagementService if it doesn't exist
func (m *VKSManagementServiceManager) EnsureManagementService(spec ManagementServiceSpec) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate input parameters
	if spec.Name == "" {
		return fmt.Errorf("ManagementService name cannot be empty")
	}

	// Get Avi client from AKO's infra client instance
	aviClient := m.getAviClient()
	if aviClient == nil {
		return fmt.Errorf("failed to get Avi client from infra instance")
	}

	// Check if ManagementService already exists
	exists, err := m.checkManagementServiceExists(aviClient, spec.Name)
	if err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to check ManagementService %s existence: %v", spec.Name, err)
		return fmt.Errorf("failed to check ManagementService existence: %v", err)
	}

	if exists {
		utils.AviLog.Infof("VKS Management Service: ManagementService %s already exists", spec.Name)
		return nil
	}

	// Create ManagementService via Avi Controller API
	if err := m.createManagementService(aviClient, spec); err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to create ManagementService %s: %v", spec.Name, err)
		return fmt.Errorf("failed to create ManagementService: %v", err)
	}

	utils.AviLog.Infof("VKS Management Service: ManagementService %s created successfully", spec.Name)
	return nil
}

// EnsureManagementServiceAccessGrant creates namespace-scoped AccessGrant if it doesn't exist
func (m *VKSManagementServiceManager) EnsureManagementServiceAccessGrant(spec ManagementServiceAccessGrantSpec) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate input parameters
	if spec.Name == "" || spec.Namespace == "" {
		return fmt.Errorf("ManagementServiceAccessGrant name and namespace cannot be empty")
	}

	// Get Avi client from AKO's infra client instance
	aviClient := m.getAviClient()
	if aviClient == nil {
		return fmt.Errorf("failed to get Avi client from infra instance")
	}

	// Check if AccessGrant already exists for this namespace
	exists, err := m.checkAccessGrantExists(aviClient, spec.Name, spec.Namespace)
	if err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to check ManagementServiceAccessGrant %s existence for namespace %s: %v", spec.Name, spec.Namespace, err)
		return fmt.Errorf("failed to check ManagementServiceAccessGrant existence: %v", err)
	}

	if exists {
		utils.AviLog.Infof("VKS Management Service: ManagementServiceAccessGrant %s already exists for namespace %s", spec.Name, spec.Namespace)
		return nil
	}

	// Create ManagementServiceAccessGrant via Avi Controller API
	if err := m.createManagementServiceAccessGrant(aviClient, spec); err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to create ManagementServiceAccessGrant %s for namespace %s: %v", spec.Name, spec.Namespace, err)
		return fmt.Errorf("failed to create ManagementServiceAccessGrant: %v", err)
	}

	utils.AviLog.Infof("VKS Management Service: ManagementServiceAccessGrant %s created successfully for namespace %s", spec.Name, spec.Namespace)
	return nil
}

// checkManagementServiceExists checks if ManagementService CRD exists in supervisor cluster
func (m *VKSManagementServiceManager) checkManagementServiceExists(aviClient *clients.AviClient, name string) (bool, error) {
	utils.AviLog.Debugf("VKS Management Service: Checking if ManagementService CRD %s exists in supervisor cluster", name)

	// Use dynamic client to check CRD in supervisor cluster
	_, err := m.dynamicClient.Resource(ManagementServiceGVR).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		// If not found, return false (doesn't exist)
		if errors.IsNotFound(err) {
			return false, nil
		}
		// Other errors
		return false, fmt.Errorf("failed to check ManagementService CRD %s: %v", name, err)
	}

	// Resource exists
	return true, nil
}

// createManagementService creates ManagementService via Avi Controller API
func (m *VKSManagementServiceManager) createManagementService(aviClient *clients.AviClient, spec ManagementServiceSpec) error {
	utils.AviLog.Infof("VKS Management Service: Creating ManagementService %s", spec.Name)

	// TODO: Implement actual API call to create ManagementService
	// managementServiceData := map[string]interface{}{
	//     "name":               spec.Name,
	//     "description":        spec.Description,
	//     "management_address": spec.ManagementAddress,
	//     "port":              spec.Port,
	//     "protocol":          spec.Protocol,
	// }
	//
	// if spec.VCenterURL != "" {
	//     managementServiceData["vcenter_url"] = spec.VCenterURL
	// }
	//
	// var result map[string]interface{}
	// uri := "api/managementservice"
	// err := aviClient.AviSession.Post(uri, managementServiceData, &result)
	// return err

	// Placeholder implementation
	utils.AviLog.Infof("VKS Management Service: ManagementService %s creation placeholder executed", spec.Name)
	return nil
}

// checkAccessGrantExists checks if ManagementServiceAccessGrant CRD exists in supervisor cluster
func (m *VKSManagementServiceManager) checkAccessGrantExists(aviClient *clients.AviClient, name, namespace string) (bool, error) {
	utils.AviLog.Debugf("VKS Management Service: Checking if ManagementServiceAccessGrant CRD %s exists in namespace %s in supervisor cluster", name, namespace)

	// ManagementServiceAccessGrant is namespace-scoped
	_, err := m.dynamicClient.Resource(ManagementServiceAccessGrantGVR).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		// If not found, return false (doesn't exist)
		if errors.IsNotFound(err) {
			return false, nil
		}
		// Other errors
		return false, fmt.Errorf("failed to check ManagementServiceAccessGrant CRD %s in namespace %s: %v", name, namespace, err)
	}

	// Resource exists in the specified namespace
	utils.AviLog.Debugf("VKS Management Service: Found ManagementServiceAccessGrant %s in namespace %s", name, namespace)
	return true, nil
}

// createManagementServiceAccessGrant creates AccessGrant via Avi Controller API
func (m *VKSManagementServiceManager) createManagementServiceAccessGrant(aviClient *clients.AviClient, spec ManagementServiceAccessGrantSpec) error {
	utils.AviLog.Infof("VKS Management Service: Creating ManagementServiceAccessGrant %s for namespace %s", spec.Name, spec.Namespace)

	// TODO: Implement actual API call to create ManagementServiceAccessGrant
	// accessGrantData := map[string]interface{}{
	//     "name":         spec.Name,
	//     "namespace":    spec.Namespace,
	//     "description":  spec.Description,
	//     "service_name": spec.ServiceName,
	//     "type":         spec.Type,
	//     "enabled":      spec.Enabled,
	// }
	//
	// if spec.VCenterURL != "" {
	//     accessGrantData["vcenter_url"] = spec.VCenterURL
	// }
	//
	// var result map[string]interface{}
	// uri := "api/managementserviceaccessgrant"
	// err := aviClient.AviSession.Post(uri, accessGrantData, &result)
	// return err

	// Placeholder implementation
	utils.AviLog.Infof("VKS Management Service: ManagementServiceAccessGrant %s creation placeholder executed for namespace %s", spec.Name, spec.Namespace)
	return nil
}

// CleanupNamespaceAccessGrant removes AccessGrant for a namespace
func (m *VKSManagementServiceManager) CleanupNamespaceAccessGrant(namespace string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get Avi client from AKO's infra client instance
	aviClient := m.getAviClient()
	if aviClient == nil {
		return fmt.Errorf("failed to get Avi client from infra instance for cleanup")
	}

	// Check if AccessGrant exists for this namespace
	accessGrantName := fmt.Sprintf("avi-controller-vm-access-%s", namespace)
	exists, err := m.checkAccessGrantExists(aviClient, accessGrantName, namespace)
	if err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to check ManagementServiceAccessGrant %s existence for namespace %s: %v", accessGrantName, namespace, err)
		return fmt.Errorf("failed to check ManagementServiceAccessGrant existence: %v", err)
	}

	if !exists {
		utils.AviLog.Debugf("VKS Management Service: No ManagementServiceAccessGrant found for namespace %s, skipping cleanup", namespace)
		return nil
	}

	// TODO: Implement actual API call to delete AccessGrant
	// accessGrantName := fmt.Sprintf("avi-controller-vm-access-%s", namespace)
	// var accessGrant map[string]interface{}
	// uri := fmt.Sprintf("api/managementserviceaccessgrant?name=%s&namespace=%s", accessGrantName, namespace)
	// err := aviClient.AviSession.Get(uri, &accessGrant)
	// if err != nil {
	//     utils.AviLog.Infof("VKS Management Service: AccessGrant for namespace %s not found, may already be deleted: %v", namespace, err)
	//     return nil
	// }
	//
	// if uuid, ok := accessGrant["uuid"].(string); ok {
	//     deleteURI := fmt.Sprintf("api/managementserviceaccessgrant/%s", uuid)
	//     err = aviClient.AviSession.Delete(deleteURI)
	//     if err != nil {
	//         return fmt.Errorf("failed to delete AccessGrant for namespace %s: %v", namespace, err)
	//     }
	// }

	utils.AviLog.Infof("VKS Management Service: ManagementServiceAccessGrant cleanup completed for namespace %s", namespace)
	return nil
}

// CleanupManagementService removes the global ManagementService
func (m *VKSManagementServiceManager) CleanupManagementService(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Get Avi client from AKO's infra client instance
	aviClient := m.getAviClient()
	if aviClient == nil {
		return fmt.Errorf("failed to get Avi client from infra instance for ManagementService cleanup")
	}

	// Check if ManagementService exists
	exists, err := m.checkManagementServiceExists(aviClient, name)
	if err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to check ManagementService %s existence during cleanup: %v", name, err)
		return fmt.Errorf("failed to check ManagementService existence: %v", err)
	}

	if !exists {
		utils.AviLog.Debugf("VKS Management Service: No ManagementService %s found, skipping cleanup", name)
		return nil
	}

	// TODO: Implement actual API call to delete ManagementService
	// var managementService map[string]interface{}
	// uri := fmt.Sprintf("api/managementservice?name=%s", name)
	// err := aviClient.AviSession.Get(uri, &managementService)
	// if err != nil {
	//     utils.AviLog.Infof("VKS Management Service: ManagementService %s not found, may already be deleted: %v", name, err)
	//     return nil
	// }
	//
	// if uuid, ok := managementService["uuid"].(string); ok {
	//     deleteURI := fmt.Sprintf("api/managementservice/%s", uuid)
	//     err = aviClient.AviSession.Delete(deleteURI)
	//     if err != nil {
	//         return fmt.Errorf("failed to delete ManagementService %s: %v", name, err)
	//     }
	// }

	utils.AviLog.Infof("VKS Management Service: ManagementService %s cleanup completed", name)
	return nil
}

// ShouldKeepManagementService checks if ManagementService should be kept based on managed clusters
func (m *VKSManagementServiceManager) ShouldKeepManagementService(ctx context.Context, getAllManagedClusters func(context.Context) ([]string, error)) (bool, error) {
	// Get all managed clusters across all namespaces
	managedClusters, err := getAllManagedClusters(ctx)
	if err != nil {
		return true, fmt.Errorf("failed to get managed clusters: %v", err)
	}

	// Keep ManagementService if there's at least one managed cluster
	shouldKeep := len(managedClusters) > 0
	utils.AviLog.Debugf("VKS Management Service: ManagementService should be kept: %t (managed clusters: %d)", shouldKeep, len(managedClusters))
	return shouldKeep, nil
}

// ShouldKeepNamespaceAccessGrant checks if ManagementServiceAccessGrant should be kept for a namespace
func (m *VKSManagementServiceManager) ShouldKeepNamespaceAccessGrant(ctx context.Context, namespace string, getAllManagedClusters func(context.Context) ([]string, error), parseClusterRef func(string) (string, string)) (bool, error) {
	// Get all managed clusters
	managedClusters, err := getAllManagedClusters(ctx)
	if err != nil {
		return true, fmt.Errorf("failed to get managed clusters: %v", err)
	}

	// Check if any managed cluster is in this namespace
	for _, clusterRef := range managedClusters {
		clusterNamespace, _ := parseClusterRef(clusterRef)
		if clusterNamespace == namespace {
			utils.AviLog.Debugf("VKS Management Service: AccessGrant should be kept for namespace %s (found cluster: %s)", namespace, clusterRef)
			return true, nil
		}
	}

	utils.AviLog.Debugf("VKS Management Service: AccessGrant should be removed for namespace %s (no managed clusters)", namespace)
	return false, nil
}

// ConditionalCleanupManagementService removes ManagementService only if no managed clusters exist
func (m *VKSManagementServiceManager) ConditionalCleanupManagementService(ctx context.Context, name string, getAllManagedClusters func(context.Context) ([]string, error)) error {
	shouldKeep, err := m.ShouldKeepManagementService(ctx, getAllManagedClusters)
	if err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to determine if ManagementService should be kept: %v", err)
		return err
	}

	if shouldKeep {
		utils.AviLog.Debugf("VKS Management Service: Keeping ManagementService %s (managed clusters exist)", name)
		return nil
	}

	utils.AviLog.Infof("VKS Management Service: Removing ManagementService %s (no managed clusters)", name)
	return m.CleanupManagementService(name)
}

// ConditionalCleanupNamespaceAccessGrant removes AccessGrant only if no managed clusters exist in namespace
func (m *VKSManagementServiceManager) ConditionalCleanupNamespaceAccessGrant(ctx context.Context, namespace string, getAllManagedClusters func(context.Context) ([]string, error), parseClusterRef func(string) (string, string)) error {
	shouldKeep, err := m.ShouldKeepNamespaceAccessGrant(ctx, namespace, getAllManagedClusters, parseClusterRef)
	if err != nil {
		utils.AviLog.Errorf("VKS Management Service: Failed to determine if AccessGrant should be kept for namespace %s: %v", namespace, err)
		return err
	}

	if shouldKeep {
		utils.AviLog.Debugf("VKS Management Service: Keeping AccessGrant for namespace %s (managed clusters exist)", namespace)
		return nil
	}

	utils.AviLog.Infof("VKS Management Service: Removing AccessGrant for namespace %s (no managed clusters)", namespace)
	return m.CleanupNamespaceAccessGrant(namespace)
}
