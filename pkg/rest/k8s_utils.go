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
	extensions "k8s.io/api/extensions/v1beta1"

	avicache "gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func UpdateIngressStatus(vs_cache_obj *avicache.AviVsCache, svc_mdata_obj avicache.ServiceMetadataObj, key string) error {
	mClient := utils.GetInformers().ClientSet
	mIngress, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).Get(svc_mdata_obj.IngressName, metav1.GetOptions{})
	// Once the vsvip object is available - we should be able to update the hostname, for now just updating the vip
	lbIngress := core.LoadBalancerIngress{
		IP:       vs_cache_obj.Vip,
		Hostname: "tobeupdated.com",
	}
	mIngress.Status = extensions.IngressStatus{
		LoadBalancer: core.LoadBalancerStatus{
			Ingress: []core.LoadBalancerIngress{lbIngress},
		},
	}
	response, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).UpdateStatus(mIngress)
	if err != nil {
		utils.AviLog.Error.Printf("key: %s, msg: there was an error in updating the ingress status: %v", key, err)
		return err
	}
	utils.AviLog.Info.Printf("key:%s, msg: Successfully updated the ingress status: %v", key, utils.Stringify(response))
	return nil
}

func UpdateL4LBStatus(vs_cache_obj *avicache.AviVsCache, svc_mdata_obj avicache.LBServiceMetadataObj, key string) error {
	mClient := utils.GetInformers().ClientSet
	mLb, err := mClient.CoreV1().Services(svc_mdata_obj.Namespace).Get(svc_mdata_obj.ServiceName, metav1.GetOptions{})
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
