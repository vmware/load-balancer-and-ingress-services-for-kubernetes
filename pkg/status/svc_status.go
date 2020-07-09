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
	avicache "ako/pkg/cache"
	"strings"

	"github.com/avinetworks/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateL4LBStatus(options []UpdateStatusOptions, bulk bool) {
	mClient := utils.GetInformers().ClientSet
	var servicesToUpdate []string
	var updateServiceOptions []UpdateStatusOptions

	for _, option := range options {
		if len(option.ServiceMetadata.HostNames) != 1 {
			utils.AviLog.Error("Hostname length not appropriate for status update, not equals 1")
			continue
		}

		service := option.ServiceMetadata.Namespace + "/" + option.ServiceMetadata.ServiceName
		option.IngSvc = service
		servicesToUpdate = append(servicesToUpdate, service)
		updateServiceOptions = append(updateServiceOptions, option)
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

			service.Status = corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{corev1.LoadBalancerIngress{
						IP:       option.Vip,
						Hostname: svcMetadata.HostNames[0],
					}}}}

			if sameStatus := compareLBStatus(oldServiceStatus, &service.Status.LoadBalancer); sameStatus {
				utils.AviLog.Debugf("key: %s, msg: No changes detected in service status. old: %+v new: %+v",
					key, oldServiceStatus.Ingress, service.Status.LoadBalancer.Ingress)
				continue
			}

			_, err := mClient.CoreV1().Services(svcMetadata.Namespace).UpdateStatus(service)
			if err != nil {
				utils.AviLog.Errorf("key: %s, msg: there was an error in updating the loadbalancer status: %v", key, err)
				continue
			}
			utils.AviLog.Infof("key: %s, msg: Successfully updated the status of serviceLB: %s old: %+v new %+v",
				key, option.IngSvc, oldServiceStatus.Ingress, service.Status.LoadBalancer.Ingress)
		}
	}

	return
}

func DeleteL4LBStatus(svc_mdata_obj avicache.ServiceMetadataObj, key string) error {
	mClient := utils.GetInformers().ClientSet
	mLb, err := mClient.CoreV1().Services(svc_mdata_obj.Namespace).Get(svc_mdata_obj.ServiceName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: there was a problem in resetting the service status :%s", key, err)
		return err
	}
	mLb.Status = corev1.ServiceStatus{
		LoadBalancer: corev1.LoadBalancerStatus{
			Ingress: []corev1.LoadBalancerIngress{},
		},
	}
	_, err = mClient.CoreV1().Services(svc_mdata_obj.Namespace).UpdateStatus(mLb)
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in resetting the loadbalancer status: %v", key, err)
		return err
	}
	utils.AviLog.Infof("key: %s, msg: Successfully reset the status of serviceLB: %s/%s",
		key, svc_mdata_obj.Namespace, svc_mdata_obj.ServiceName)
	return nil
}

// getServices fetches all serviceLB and returns a map: {"namespace/name": serviceObj...}
// if bulk is set to true, this fetches all services in a single k8s api-server call
func getServices(serviceNSNames []string, bulk bool) map[string]*corev1.Service {
	mClient := utils.GetInformers().ClientSet
	serviceMap := make(map[string]*corev1.Service)

	if bulk {
		serviceLBList, err := mClient.CoreV1().Services("").List(metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the ingress object for UpdateStatus :%s", err)
		}
		for i := range serviceLBList.Items {
			ing := serviceLBList.Items[i]
			serviceMap[ing.Namespace+"/"+ing.Name] = &ing
		}

		return serviceMap
	}

	for _, namespaceName := range serviceNSNames {
		nsNameSplit := strings.Split(namespaceName, "/")
		serviceLB, err := mClient.CoreV1().Services(nsNameSplit[0]).Get(nsNameSplit[1], metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("msg: Could not get the service object for UpdateStatus :%s", err)
			continue
		}

		serviceMap[serviceLB.Namespace+"/"+serviceLB.Name] = serviceLB
	}

	return serviceMap
}
