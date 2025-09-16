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

package proxy

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/ako-infra/avirest"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"

	"gopkg.in/yaml.v2"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	VKSManagementServiceName  = "avi-controller-management-service"
	VKSManagementServiceGrant = "avi-controller-management-service-grant"
	VKSManagementServicePort  = 443

	ManagementServiceRetryInterval = 10 * time.Second

	WorkloadSelectorTypeVirtualMachine = "VIRTUAL_MACHINE"
)

type ManagementServiceController struct {
	supervisorID  string
	serviceName   string
	servicePort   int32
	controllerIPs []string
	cloudUUID     string
	vcenterHost   string
}

func NewManagementServiceController() *ManagementServiceController {
	supervisorID, vcenterHost := getClusterConfigValues()

	if supervisorID == "" || vcenterHost == "" {
		utils.AviLog.Errorf("VKS ManagementService: Cannot create controller without supervisor ID and vCenter host from ConfigMap")
		return nil
	}

	return &ManagementServiceController{
		supervisorID:  supervisorID,
		serviceName:   VKSManagementServiceName,
		servicePort:   VKSManagementServicePort,
		controllerIPs: []string{lib.GetControllerIP()},
		cloudUUID:     utils.CloudUUID,
		vcenterHost:   vcenterHost,
	}
}

// EnsureGlobalManagementService creates the global VKS management service
func EnsureGlobalManagementService() error {
	c := NewManagementServiceController()
	if c == nil {
		return fmt.Errorf("failed to create management service controller")
	}

	existingService, err := c.GetManagementService()
	if err == nil {
		if c.validateManagementServiceConfig(existingService) {
			utils.AviLog.Infof("VKS Management Service %s already exists with correct configuration", c.serviceName)
			return nil
		}
		utils.AviLog.Infof("VKS Management Service %s exists but has outdated configuration, updating via AVI Controller", c.serviceName)
	} else {
		utils.AviLog.Infof("VKS Management Service %s not found, creating new service", c.serviceName)
	}

	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return fmt.Errorf("avi Controller client not available")
	}

	payload := map[string]interface{}{
		"cloud_uuid":    c.cloudUUID,
		"supervisor_id": c.supervisorID,
		"vcenter_host":  c.vcenterHost,
		"management_service": map[string]interface{}{
			"management_service": c.serviceName,
			"ports": []map[string]interface{}{
				{
					"port": c.servicePort,
					"tls_configuration": map[string]interface{}{
						"certificate_authority_chain": utils.SharedCtrlProp().GetAllCtrlProp()[utils.ENV_CTRL_CADATA],
						"hostname":                    lib.GetControllerIP(),
					},
				},
			},
		},
	}
	var response interface{}
	err = aviClient.AviSession.Post(
		"api/vimgrvcenterruntime/initiate/managementservice",
		payload,
		&response,
	)
	if err != nil {
		return fmt.Errorf("management service creation API call failed: %v", err)
	}
	utils.AviLog.Infof("VKS Management Service %s created successfully. Response: %v", c.serviceName, response)
	return nil
}

func (c *ManagementServiceController) GetManagementService() (map[string]interface{}, error) {
	dynamicClient := lib.GetDynamicClientSet()
	if dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not available")
	}

	gvr := schema.GroupVersionResource{
		Group:    "netoperator.vmware.com",
		Version:  "v1alpha1",
		Resource: "managementservices",
	}

	resource, err := dynamicClient.Resource(gvr).Get(context.TODO(), c.serviceName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("management service not found")
		}
		return nil, fmt.Errorf("failed to get management service: %v", err)
	}

	return resource.Object, nil
}

func (c *ManagementServiceController) validateManagementServiceConfig(serviceObj map[string]interface{}) bool {
	spec, ok := serviceObj["spec"].(map[string]interface{})
	if !ok {
		utils.AviLog.Infof("ManagementService spec not found or invalid format")
		return false
	}

	managementAddresses, ok := spec["managementAddresses"].([]interface{})
	if !ok || len(managementAddresses) == 0 {
		utils.AviLog.Infof("ManagementService managementAddresses not found or empty")
		return false
	}

	expectedAddress := lib.GetControllerIP()
	addressFound := false
	for _, addr := range managementAddresses {
		if addrStr, ok := addr.(string); ok && addrStr == expectedAddress {
			addressFound = true
			break
		}
	}
	if !addressFound {
		utils.AviLog.Infof("ManagementService expected address %s not found in managementAddresses", expectedAddress)
		return false
	}

	ports, ok := spec["ports"].([]interface{})
	if !ok || len(ports) == 0 {
		utils.AviLog.Infof("ManagementService ports not found or empty")
		return false
	}

	portFound := false
	for _, portInterface := range ports {
		if port, ok := portInterface.(map[string]interface{}); ok {
			if value, ok := port["value"].(float64); ok && int32(value) == c.servicePort {
				if tls, ok := port["tls"].(map[string]interface{}); ok {
					if hostname, ok := tls["hostname"].(string); ok && hostname == expectedAddress {
						if caCert, ok := tls["certificateAuthorityChain"].(string); ok && len(caCert) > 0 {
							expectedCaCert := utils.SharedCtrlProp().GetAllCtrlProp()[utils.ENV_CTRL_CADATA]
							if caCert == expectedCaCert {
								portFound = true
								break
							} else {
								utils.AviLog.Infof("ManagementService CA certificate mismatch: expected length %d, got length %d", len(expectedCaCert), len(caCert))
							}
						}
					}
				}
			}
		}
	}

	if !portFound {
		utils.AviLog.Infof("ManagementService expected port configuration not found")
		return false
	}

	utils.AviLog.Infof("ManagementService configuration validation passed: address=%s, port=%d, hostname=%s, ca_cert_length=%d",
		expectedAddress, c.servicePort, expectedAddress, len(utils.SharedCtrlProp().GetAllCtrlProp()[utils.ENV_CTRL_CADATA]))
	return true
}

func CleanupGlobalManagementService() error {
	c := NewManagementServiceController()
	if c == nil {
		return fmt.Errorf("failed to cleanup management service controller")
	}
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return fmt.Errorf("avi Controller client not available")
	}
	payload := map[string]interface{}{
		"cloud_uuid":            c.cloudUUID,
		"supervisor_id":         c.supervisorID,
		"management_service_id": c.serviceName,
		"vcenter_host":          c.vcenterHost,
	}
	var response interface{}
	err := aviClient.AviSession.Post(
		"api/vimgrvcenterruntime/delete/managementservice",
		payload,
		&response,
	)
	if err != nil {
		return fmt.Errorf("management service delete API call failed: %v", err)
	}

	utils.AviLog.Infof("VKS Management Service %s deleted successfully. Response: %v", c.serviceName, response)
	return nil
}

func EnsureGlobalManagementServiceWithRetry(stopCh <-chan struct{}) {
	utils.AviLog.Infof("VKS Management Service: starting with infinite retry")

	for {
		if err := EnsureGlobalManagementService(); err != nil {
			utils.AviLog.Warnf("VKS Management Service: failed to ensure, will retry in %v: %v",
				ManagementServiceRetryInterval, err)

			// Wait before retry, but also check for shutdown
			select {
			case <-stopCh:
				utils.AviLog.Infof("VKS Management Service: shutdown signal received during retry wait")
				return
			case <-time.After(ManagementServiceRetryInterval):
				// Continue to next retry
				continue
			}
		} else {
			utils.AviLog.Infof("VKS Management Service: ensured successfully")
			return
		}
	}
}

type WCPClusterConfig struct {
	SupervisorID string `yaml:"supervisor_id"`
	VCPnid       string `yaml:"vc_pnid"`
}

func getClusterConfigValues() (string, string) {
	clientset := utils.GetInformers().ClientSet
	if clientset == nil {
		return "", ""
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	configMap, err := clientset.CoreV1().ConfigMaps("kube-system").Get(ctx, "wcp-cluster-config", metav1.GetOptions{})
	if err != nil {
		return "", ""
	}

	configYAML, exists := configMap.Data["wcp-cluster-config.yaml"]
	if !exists {
		return "", ""
	}

	var config WCPClusterConfig
	if err := yaml.Unmarshal([]byte(configYAML), &config); err != nil {
		return "", ""
	}

	if config.SupervisorID == "" {
		utils.AviLog.Errorf("VKS ManagementService: supervisor_id not found in wcp-cluster-config")
		return "", ""
	}

	if config.VCPnid == "" {
		utils.AviLog.Errorf("VKS ManagementService: vc_pnid not found in wcp-cluster-config")
		return "", ""
	}

	utils.AviLog.Infof("VKS ManagementService: Using supervisor ID: %s, vCenter host: %s", config.SupervisorID, config.VCPnid)
	return config.SupervisorID, config.VCPnid
}

func (c *ManagementServiceController) CreateManagementServiceGrant(namespace string) error {
	grantName := fmt.Sprintf("%s-%s", namespace, VKSManagementServiceGrant)
	existingGrant, err := c.GetManagementServiceGrant(namespace)
	if err == nil {
		if c.validateManagementServiceGrantConfig(existingGrant) {
			utils.AviLog.Infof("VKS ManagementServiceGrant %s in namespace %s already exists with correct configuration", grantName, namespace)
			return nil
		}
		utils.AviLog.Infof("VKS ManagementServiceGrant %s in namespace %s exists but has outdated configuration, updating via AVI Controller", grantName, namespace)
	} else {
		utils.AviLog.Infof("VKS ManagementServiceGrant %s in namespace %s not found, creating new grant with workload selector %s", grantName, namespace, WorkloadSelectorTypeVirtualMachine)
	}

	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return fmt.Errorf("avi Controller client not available")
	}

	payload := map[string]interface{}{
		"cloud_uuid":    c.cloudUUID,
		"supervisor_id": c.supervisorID,
		"vcenter_host":  c.vcenterHost,
		"namespace":     namespace,
		"management_service_access_grant": map[string]interface{}{
			"access_grant":       grantName,
			"management_service": c.serviceName,
			"workload_selector":  WorkloadSelectorTypeVirtualMachine,
		},
	}

	var response interface{}
	err = aviClient.AviSession.Post(
		"api/vimgrvcenterruntime/initiate/managementserviceaccessgrant",
		payload,
		&response,
	)
	if err != nil {
		return fmt.Errorf("management service grant creation API call failed: %v", err)
	}

	utils.AviLog.Infof("VKS ManagementServiceGrant %s in namespace %s created successfully. Response: %v", grantName, namespace, response)
	return nil
}

func (c *ManagementServiceController) GetManagementServiceGrant(namespace string) (map[string]interface{}, error) {
	grantName := fmt.Sprintf("%s-%s", namespace, VKSManagementServiceGrant)

	dynamicClient := lib.GetDynamicClientSet()
	if dynamicClient == nil {
		return nil, fmt.Errorf("dynamic client not available")
	}

	gvr := schema.GroupVersionResource{
		Group:    "netoperator.vmware.com",
		Version:  "v1alpha1",
		Resource: "managementserviceaccessgrants",
	}

	resource, err := dynamicClient.Resource(gvr).Namespace(namespace).Get(context.TODO(), grantName, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return nil, fmt.Errorf("management service grant not found")
		}
		return nil, fmt.Errorf("failed to get management service grant: %v", err)
	}

	return resource.Object, nil
}

func (c *ManagementServiceController) validateManagementServiceGrantConfig(grantObj map[string]interface{}) bool {
	spec, ok := grantObj["spec"].(map[string]interface{})
	if !ok {
		utils.AviLog.Infof("ManagementServiceAccessGrant spec not found or invalid format")
		return false
	}

	managementServiceRef, ok := spec["managementServiceRef"].(string)
	if !ok || managementServiceRef != c.serviceName {
		utils.AviLog.Infof("ManagementServiceAccessGrant managementServiceRef mismatch: expected %s, got %s", c.serviceName, managementServiceRef)
		return false
	}

	utils.AviLog.Infof("ManagementServiceAccessGrant configuration validation passed: managementServiceRef=%s",
		managementServiceRef)
	return true
}

func (c *ManagementServiceController) DeleteManagementServiceGrant(namespace string) error {
	grantName := fmt.Sprintf("%s-%s", namespace, VKSManagementServiceGrant)
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return fmt.Errorf("avi Controller client not available")
	}

	payload := map[string]interface{}{
		"cloud_uuid":                         c.cloudUUID,
		"supervisor_id":                      c.supervisorID,
		"vcenter_host":                       c.vcenterHost,
		"namespace":                          namespace,
		"management_service_access_grant_id": grantName,
	}

	var response interface{}
	err := aviClient.AviSession.Post(
		"api/vimgrvcenterruntime/delete/managementserviceaccessgrant",
		payload,
		&response,
	)
	if err != nil {
		return fmt.Errorf("management service grant delete API call failed: %v", err)
	}

	utils.AviLog.Infof("VKS ManagementServiceGrant %s in namespace %s deleted successfully. Response: %v", grantName, namespace, response)
	return nil
}

// HandleNamespaceGrantAdd creates a ManagementServiceGrant when a namespace with SEG annotation is added
func HandleNamespaceGrantAdd(obj interface{}) {
	if !lib.GetVPCMode() || !lib.IsVKSCapabilityActivated() {
		return
	}
	namespace, ok := obj.(*corev1.Namespace)
	if !ok {
		utils.AviLog.Warnf("VKS ManagementServiceGrant: expected namespace object, got %T", obj)
		return
	}

	if !namespaceHasSEG(namespace) {
		utils.AviLog.Debugf("VKS ManagementServiceGrant: namespace %s does not have SEG annotation, skipping", namespace.Name)
		return
	}

	controller := NewManagementServiceController()
	if controller == nil {
		utils.AviLog.Errorf("VKS ManagementServiceGrant: failed to create controller for namespace %s", namespace.Name)
		return
	}

	utils.AviLog.Infof("VKS ManagementServiceGrant: namespace %s added with SEG annotation, creating grant", namespace.Name)
	if err := controller.CreateManagementServiceGrant(namespace.Name); err != nil {
		utils.AviLog.Errorf("VKS ManagementServiceGrant: failed to create grant for namespace %s: %v", namespace.Name, err)
	}
}

// HandleNamespaceGrantUpdate manages ManagementServiceGrant when namespace SEG annotation changes
func HandleNamespaceGrantUpdate(oldObj, newObj interface{}) {
	if !lib.GetVPCMode() || !lib.IsVKSCapabilityActivated() {
		return
	}
	oldNamespace, ok := oldObj.(*corev1.Namespace)
	if !ok {
		utils.AviLog.Warnf("VKS ManagementServiceGrant: expected namespace object, got %T", oldObj)
		return
	}

	newNamespace, ok := newObj.(*corev1.Namespace)
	if !ok {
		utils.AviLog.Warnf("VKS ManagementServiceGrant: expected namespace object, got %T", newObj)
		return
	}

	oldHasSEG := namespaceHasSEG(oldNamespace)
	newHasSEG := namespaceHasSEG(newNamespace)

	if oldHasSEG == newHasSEG {
		return
	}

	controller := NewManagementServiceController()
	if controller == nil {
		utils.AviLog.Errorf("VKS ManagementServiceGrant: failed to create controller for namespace %s", newNamespace.Name)
		return
	}

	if !oldHasSEG && newHasSEG {
		// SEG annotation was added
		utils.AviLog.Infof("VKS ManagementServiceGrant: namespace %s now has SEG annotation, creating grant", newNamespace.Name)
		if err := controller.CreateManagementServiceGrant(newNamespace.Name); err != nil {
			utils.AviLog.Errorf("VKS ManagementServiceGrant: failed to create grant for namespace %s: %v", newNamespace.Name, err)
		}
	} else if oldHasSEG && !newHasSEG {
		// SEG annotation was removed
		utils.AviLog.Infof("VKS ManagementServiceGrant: namespace %s no longer has SEG annotation, deleting grant", newNamespace.Name)
		if err := controller.DeleteManagementServiceGrant(newNamespace.Name); err != nil {
			utils.AviLog.Errorf("VKS ManagementServiceGrant: failed to delete grant for namespace %s: %v", newNamespace.Name, err)
		}
	}
}

// HandleNamespaceGrantDelete removes a ManagementServiceGrant when a namespace is deleted
func HandleNamespaceGrantDelete(obj interface{}) {
	if !lib.GetVPCMode() || !lib.IsVKSCapabilityActivated() {
		return
	}
	namespace, ok := obj.(*corev1.Namespace)
	if !ok {
		utils.AviLog.Warnf("VKS ManagementServiceGrant: expected namespace object, got %T", obj)
		return
	}

	if !namespaceHasSEG(namespace) {
		utils.AviLog.Infof("VKS ManagementServiceGrant: namespace %s did not have SEG annotation, skipping", namespace.Name)
		return
	}

	controller := NewManagementServiceController()
	if controller == nil {
		utils.AviLog.Errorf("VKS ManagementServiceGrant: failed to create controller for namespace %s", namespace.Name)
		return
	}

	utils.AviLog.Infof("VKS ManagementServiceGrant: namespace %s deleted, removing grant", namespace.Name)
	if err := controller.DeleteManagementServiceGrant(namespace.Name); err != nil {
		utils.AviLog.Errorf("VKS ManagementServiceGrant: failed to delete grant for namespace %s: %v", namespace.Name, err)
	}
}

// ReconcileManagementServiceGrants ensures ManagementServiceGrants exist for all namespaces with SEG annotations
func ReconcileManagementServiceGrants() {
	utils.AviLog.Infof("VKS reconciler: reconciling ManagementServiceGrants")

	controller := NewManagementServiceController()
	if controller == nil {
		utils.AviLog.Errorf("VKS reconciler: failed to create ManagementServiceController")
		return
	}

	// Get all namespaces
	informers := utils.GetInformers()
	if informers == nil || informers.NSInformer == nil {
		utils.AviLog.Infof("VKS reconciler: namespace informer not initialized yet, skipping reconciliation")
		return
	}

	lister := informers.NSInformer.Lister()
	if lister == nil {
		utils.AviLog.Infof("VKS reconciler: namespace lister not available yet, skipping reconciliation")
		return
	}

	namespaces, err := lister.List(labels.Everything())
	if err != nil {
		utils.AviLog.Errorf("VKS reconciler: failed to list namespaces: %v", err)
		return
	}

	grantCount := 0
	for _, namespace := range namespaces {
		if namespaceHasSEG(namespace) {
			if err := controller.CreateManagementServiceGrant(namespace.Name); err != nil {
				utils.AviLog.Errorf("VKS reconciler: failed to ensure grant for namespace %s: %v", namespace.Name, err)
			} else {
				grantCount++
			}
		}
	}

	utils.AviLog.Infof("VKS reconciler: reconciled %d ManagementServiceGrants", grantCount)
}

func namespaceHasSEG(namespace *corev1.Namespace) bool {
	if namespace.Annotations != nil {
		if _, exists := namespace.Annotations[lib.WCPSEGroup]; exists {
			return true
		}
	}
	return false
}
