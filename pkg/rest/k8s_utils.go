/*
* [2013] - [2019] Avi Networks Incorporated
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

package rest

import (
	core "k8s.io/api/core/v1"

	avicache "ako/pkg/cache"

	"github.com/avinetworks/container-lib/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateIngressStatus(vs_cache_obj *avicache.AviVsCache, svc_mdata_obj avicache.ServiceMetadataObj, key string) error {
	mClient := utils.GetInformers().ClientSet
	mIngress, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).Get(svc_mdata_obj.IngressName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warning.Printf("key :%s, msg: there was a problem in updating the ingress status :%s", key, err)
		return err
	}

	// Handle the hostname --> vip update case
	var updateHost bool
	for _, status := range mIngress.Status.LoadBalancer.Ingress {
		if status.Hostname == svc_mdata_obj.HostName {
			status.IP = vs_cache_obj.Vip
			updateHost = true
		}
	}
	utils.AviLog.Info.Printf("key: %s, msg: status before update: %v", key, mIngress.Status.LoadBalancer.Ingress)
	// Handle fresh hostname update
	if !updateHost {
		lbIngress := core.LoadBalancerIngress{
			IP:       vs_cache_obj.Vip,
			Hostname: svc_mdata_obj.HostName,
		}
		mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress, lbIngress)
	}
	utils.AviLog.Info.Printf("key: %s, msg: status after update: %v", key, mIngress.Status.LoadBalancer.Ingress)

	response, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).UpdateStatus(mIngress)
	if err != nil {
		utils.AviLog.Error.Printf("key: %s, msg: there was an error in updating the ingress status: %v", key, err)
		return err
	}
	utils.AviLog.Info.Printf("key:%s, msg: Successfully updated the ingress status: %v", key, utils.Stringify(response))
	return nil
}

func DeleteIngressStatus(svc_mdata_obj avicache.ServiceMetadataObj, key string) error {
	mClient := utils.GetInformers().ClientSet
	mIngress, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).Get(svc_mdata_obj.IngressName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warning.Printf("key :%s, msg: there was a problem in updating the ingress status :%s", key, err)
		return err
	}
	utils.AviLog.Info.Printf("key: %s, msg: status before update: %v", key, mIngress.Status.LoadBalancer.Ingress)

	for i, status := range mIngress.Status.LoadBalancer.Ingress {
		if status.Hostname == svc_mdata_obj.HostName {
			mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress[:i], mIngress.Status.LoadBalancer.Ingress[i+1:]...)
			break
		}
	}
	utils.AviLog.Info.Printf("key: %s, msg: status after update: %v", key, mIngress.Status.LoadBalancer.Ingress)

	response, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).UpdateStatus(mIngress)
	if err != nil {
		utils.AviLog.Error.Printf("key: %s, msg: there was an error in deleting the ingress status: %v", key, err)
		return err
	}
	utils.AviLog.Info.Printf("key:%s, msg: Successfully deleted the ingress status: %v", key, utils.Stringify(response))
	return nil
}

func UpdateL4LBStatus(vs_cache_obj *avicache.AviVsCache, svc_mdata_obj avicache.LBServiceMetadataObj, key string) error {
	mClient := utils.GetInformers().ClientSet
	mLb, err := mClient.CoreV1().Services(svc_mdata_obj.Namespace).Get(svc_mdata_obj.ServiceName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warning.Printf("key :%s, msg: there was a problem in updating the service status :%s", key, err)
		return err
	}
	// Once the vsvip object is available - we should be able to update the hostname, for now just updating the vip
	lbIngress := core.LoadBalancerIngress{
		IP:       vs_cache_obj.Vip,
		Hostname: "tobeupdated.com",
	}
	mLb.Status = core.ServiceStatus{
		LoadBalancer: core.LoadBalancerStatus{
			Ingress: []core.LoadBalancerIngress{lbIngress},
		},
	}
	response, err := mClient.CoreV1().Services(svc_mdata_obj.Namespace).UpdateStatus(mLb)
	if err != nil {
		utils.AviLog.Error.Printf("key: %s, msg: there was an error in updating the loadbalancer status: %v", key, err)
		return err
	}
	utils.AviLog.Info.Printf("key:%s, msg: Successfully updated the loadbalancer status: %v", key, utils.Stringify(response))
	return nil
}
