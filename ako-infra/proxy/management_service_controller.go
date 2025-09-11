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

	"gopkg.in/yaml.v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	VKSManagementServiceNamePrefix  = "avi-management-service"
	VKSManagementServiceGrantPrefix = "avi-management-service-grant"
	VKSManagementServicePort        = 443

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
		serviceName:   fmt.Sprintf("%s-%s", VKSManagementServiceNamePrefix, supervisorID),
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

	_, err := c.GetManagementService()
	if err == nil {
		utils.AviLog.Infof("VKS Management Service %s already exists", c.serviceName)
		return nil
	}

	utils.AviLog.Infof("VKS Management Service %s not found, creating new service", c.serviceName)

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
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return nil, fmt.Errorf("avi Controller client not available")
	}

	payload := map[string]interface{}{
		"cloud_uuid":            c.cloudUUID,
		"supervisor_id":         c.supervisorID,
		"management_service_id": c.serviceName,
		"vcenter_host":          c.vcenterHost,
	}
	var response interface{}
	err := aviClient.AviSession.Post(
		"api/vimgrvcenterruntime/retrieve/managementservice",
		payload,
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("management service get API call failed: %v", err)
	}
	if respMap, ok := response.(map[string]interface{}); ok {
		return respMap, nil
	}

	return map[string]interface{}{"response": response}, nil
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
	grantName := fmt.Sprintf("%s-%s-%s", VKSManagementServiceGrantPrefix, namespace, c.supervisorID)
	_, err := c.GetManagementServiceGrant(namespace)
	if err == nil {
		utils.AviLog.Infof("VKS ManagementServiceGrant %s in namespace %s already exists", grantName, namespace)
		return nil
	}

	utils.AviLog.Infof("VKS ManagementServiceGrant %s in namespace %s not found, creating new grant with workload selector %s", grantName, namespace, WorkloadSelectorTypeVirtualMachine)

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
	grantName := fmt.Sprintf("%s-%s-%s", VKSManagementServiceGrantPrefix, namespace, c.supervisorID)
	aviClient := avirest.InfraAviClientInstance()
	if aviClient == nil {
		return nil, fmt.Errorf("avi Controller client not available")
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
		"api/vimgrvcenterruntime/retrieve/managementserviceaccessgrant",
		payload,
		&response,
	)
	if err != nil {
		return nil, fmt.Errorf("management service grant get API call failed: %v", err)
	}

	if respMap, ok := response.(map[string]interface{}); ok {
		return respMap, nil
	}

	return map[string]interface{}{"response": response}, nil
}

func (c *ManagementServiceController) DeleteManagementServiceGrant(namespace string) error {
	grantName := fmt.Sprintf("%s-%s-%s", VKSManagementServiceGrantPrefix, namespace, c.supervisorID)
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
