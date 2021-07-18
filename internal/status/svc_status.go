/*
 * Copyright 2019-2020 VMware, Inc.
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

package status

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func UpdateL4LBStatus(options []UpdateOptions, bulk bool) {
	var servicesToUpdate []string
	var updateServiceOptions []UpdateOptions

	for _, option := range options {
		if len(option.ServiceMetadata.HostNames) != 1 && !lib.GetAdvancedL4() {
			utils.AviLog.Warnf("key: %s, msg: Service hostname not found for service %v status update", option.Key, option.ServiceMetadata.NamespaceServiceName)
		}

		for _, svc := range option.ServiceMetadata.NamespaceServiceName {
			option.IngSvc = svc
			servicesToUpdate = append(servicesToUpdate, svc)
			updateServiceOptions = append(updateServiceOptions, option)
		}
	}

	serviceMap := getServices(servicesToUpdate, bulk)
	for _, option := range updateServiceOptions {
		key, svcMetadata := option.Key, option.ServiceMetadata
		if service := serviceMap[option.IngSvc]; service != nil {
			oldServiceStatus := service.Status.LoadBalancer.DeepCopy()
			if option.Vip == "" {
				// nothing to do here
				continue
			}

			var svcHostname string
			if len(svcMetadata.HostNames) > 0 {
				svcHostname = svcMetadata.HostNames[0]
			}
			service.Status = corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{{
						IP:       option.Vip,
						Hostname: svcHostname,
					}}}}

			sameStatus := compareLBStatus(oldServiceStatus, &service.Status.LoadBalancer)
			var updatedSvc *corev1.Service
			var err error
			if !sameStatus {
				patchPayload, _ := json.Marshal(map[string]interface{}{
					"status": service.Status,
				})

				updatedSvc, err = utils.GetInformers().ClientSet.CoreV1().Services(service.Namespace).Patch(context.TODO(), service.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
				if err != nil {
					utils.AviLog.Errorf("key: %s, msg: there was an error in updating the loadbalancer status: %v", key, err)
				} else {
					utils.AviLog.Infof("key: %s, msg: Successfully updated the status of serviceLB: %s old: %+v new %+v",
						key, option.IngSvc, oldServiceStatus.Ingress, service.Status.LoadBalancer.Ingress)
				}
			} else {
				utils.AviLog.Debugf("key: %s, msg: No changes detected in service status. old: %+v new: %+v",
					key, oldServiceStatus.Ingress, service.Status.LoadBalancer.Ingress)
			}

			if err = updateSvcAnnotations(updatedSvc, option, service, svcHostname); err != nil {
				utils.AviLog.Errorf("key: %s, msg: there was an error in updating the service annotations: %v", key, err)
			}
		}
		delete(serviceMap, option.IngSvc)
	}

	if bulk {
		for svcNSName := range serviceMap {
			DeleteL4LBStatus(avicache.ServiceMetadataObj{
				NamespaceServiceName: []string{svcNSName},
			}, lib.SyncStatusKey)
		}
	}
}

func updateSvcAnnotations(svc *corev1.Service, updateOption UpdateOptions, oldSvc *corev1.Service, svcHostname string) error {
	if svcHostname == "" {
		utils.AviLog.Infof("Can't update the service annotations as hostname for this service is empty.")
		return nil
	}
	if svc == nil {
		svc = oldSvc
	}
	vsAnnotations := map[string]string{
		updateOption.ServiceMetadata.HostNames[0]: updateOption.VirtualServiceUUID,
	}

	if !isAnnotationsUpdateRequired(svc.Annotations, vsAnnotations) {
		utils.AviLog.Debugf("No annotations update required for service %s/%s", svc.Namespace, svc.Name)
		return nil
	}

	annotations := svc.Annotations
	vsAnnotationsStr, err := json.Marshal(vsAnnotations)
	if err != nil {
		return fmt.Errorf("error in marshalling the VS annotations for svc %s/%s: %v", svc.Namespace, svc.Name,
			err)
	}
	if len(annotations) == 0 {
		annotations = map[string]string{}
	}
	annotations[VSAnnotation] = string(vsAnnotationsStr)
	annotations[ControllerAnnotation] = avicache.GetControllerClusterUUID()

	patchPayload := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"annotations": annotations,
		},
	}
	patchPayloadBytes, _ := json.Marshal(patchPayload)
	if _, err = utils.GetInformers().ClientSet.CoreV1().Services(svc.Namespace).Patch(context.TODO(), svc.Name,
		types.MergePatchType, patchPayloadBytes, metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("error in patching service %s/%s: %v", svc.Namespace, svc.Name, err)
	}
	return nil
}

func DeleteL4LBStatus(svc_mdata_obj avicache.ServiceMetadataObj, key string) error {
	serviceMap := getServices(svc_mdata_obj.NamespaceServiceName, false)
	for _, service := range svc_mdata_obj.NamespaceServiceName {
		serviceNSName := strings.Split(service, "/")
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": nil,
		})

		if serviceObj := serviceMap[service]; serviceObj != nil && (serviceObj.Status.LoadBalancer.Ingress == nil ||
			(serviceObj.Status.LoadBalancer.Ingress != nil && len(serviceObj.Status.LoadBalancer.Ingress) == 0)) {
			continue
		}

		updatedSvc, err := utils.GetInformers().ClientSet.CoreV1().Services(serviceNSName[0]).Patch(context.TODO(), serviceNSName[1], types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
		if err != nil {
			utils.AviLog.Warnf("key: %s, msg: there was an error in resetting the loadbalancer status: %v", key, err)
			return err
		}
		utils.AviLog.Infof("key: %s, msg: Successfully reset the status of serviceLB: %s", key, svc_mdata_obj.NamespaceServiceName[0])

		err = deleteSvcAnnotation(updatedSvc)
		if err != nil {
			utils.AviLog.Errorf("key: %s, msg: error in deleting service annotation: %v", key, err)
		}
	}
	return nil
}

func deleteSvcAnnotation(svc *corev1.Service) error {
	payloadData := map[string]interface{}{
		"metadata": map[string]map[string]*string{
			"annotations": {
				VSAnnotation:         nil,
				ControllerAnnotation: nil,
			},
		},
	}

	payloadBytes, _ := json.Marshal(payloadData)
	if _, err := utils.GetInformers().ClientSet.CoreV1().Services(svc.Namespace).Patch(context.TODO(), svc.Name,
		types.MergePatchType, payloadBytes, metav1.PatchOptions{}); err != nil {
		return fmt.Errorf("error in updating service: %v", err)
	}

	return nil
}

// getServices fetches all serviceLB and returns a map: {"namespace/name": serviceObj...}
// if bulk is set to true, this fetches all services in a single k8s api-server call
func getServices(serviceNSNames []string, bulk bool, retryNum ...int) map[string]*corev1.Service {
	retry := 0
	mClient := utils.GetInformers().ClientSet
	serviceMap := make(map[string]*corev1.Service)
	if len(retryNum) > 0 {
		utils.AviLog.Infof("Retrying to get the services for status update")
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("getServices for status update retried 3 times, aborting")
			return serviceMap
		}
	}

	if bulk {
		serviceLBList, err := mClient.CoreV1().Services(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the service object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getServices(serviceNSNames, bulk, retry+1)
			}
		}
		for i := range serviceLBList.Items {
			ing := serviceLBList.Items[i]
			serviceMap[ing.Namespace+"/"+ing.Name] = &ing
		}

		return serviceMap
	}

	for _, namespaceName := range serviceNSNames {
		nsNameSplit := strings.Split(namespaceName, "/")
		serviceLB, err := mClient.CoreV1().Services(nsNameSplit[0]).Get(context.TODO(), nsNameSplit[1], metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the service object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getServices(serviceNSNames, bulk, retry+1)
			}
			continue
		}

		serviceMap[serviceLB.Namespace+"/"+serviceLB.Name] = serviceLB
	}

	return serviceMap
}
