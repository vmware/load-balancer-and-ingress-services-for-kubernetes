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

package rest

import (
	"errors"
	"strings"

	core "k8s.io/api/core/v1"

	avicache "ako/pkg/cache"
	"ako/pkg/lib"

	"github.com/avinetworks/container-lib/utils"
	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type UpdateStatusOptions struct {
	// ingsvc format: namespace/name, not supposed to be provided by the caller
	ingsvc          string
	vip             string
	serviceMetadata avicache.ServiceMetadataObj
	key             string
}

func UpdateIngressStatus(options []UpdateStatusOptions, bulk bool) {
	var err error
	var ingressesToUpdate []string
	var updateIngressOptions []UpdateStatusOptions

	for _, option := range options {
		if len(option.serviceMetadata.NamespaceIngressName) > 0 {
			// This is SNI with hostname sharding.
			for _, ingressns := range option.serviceMetadata.NamespaceIngressName {
				ingressArr := strings.Split(ingressns, "/")
				if len(ingressArr) != 2 {
					utils.AviLog.Errorf("key: %s, msg: UpdateIngressStatus IngressNamespace format not correct", option.key)
					continue
				}

				ingress := ingressArr[0] + "/" + ingressArr[1]
				option.ingsvc = ingress
				ingressesToUpdate = append(ingressesToUpdate, ingress)
				updateIngressOptions = append(updateIngressOptions, option)
			}
		} else {
			ingress := option.serviceMetadata.Namespace + "/" + option.serviceMetadata.IngressName
			option.ingsvc = ingress
			ingressesToUpdate = append(ingressesToUpdate, ingress)
			updateIngressOptions = append(updateIngressOptions, option)
		}
	}

	// ingressMap: {ns/ingress: ingressObj}
	// this pre-fetches all ingresses to be candidates for status update
	// after pre-fetching, if a status update comes for that ingress, then the pre-fetched ingress would be stale
	// in which case ingress will be fetched again in updateObject, as part of a retry
	ingressMap := getIngresses(ingressesToUpdate, bulk)
	for _, option := range updateIngressOptions {
		if ingress := ingressMap[option.ingsvc]; ingress != nil {
			if err = updateObject(ingress, option); err != nil {
				utils.AviLog.Error(err)
			}
		}
	}

	return
}

func updateObject(mIngress *networking.Ingress, updateOption UpdateStatusOptions, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 2 {
			return errors.New("key: %s, msg: UpdateIngressStatus retried 3 times, aborting")
		}
	}

	var err error
	mClient := utils.GetInformers().ClientSet
	hostnames, key := updateOption.serviceMetadata.HostNames, updateOption.key
	oldIngressStatus := mIngress.Status.LoadBalancer.DeepCopy()

	// Clean up all hosts that are not part of the ingress spec.
	var hostListIng []string
	for _, rule := range mIngress.Spec.Rules {
		hostListIng = append(hostListIng, rule.Host)
	}

	// If we find a hostname in the present update, let's first remove it from the existing status.
	for i := len(mIngress.Status.LoadBalancer.Ingress) - 1; i >= 0; i-- {
		if utils.HasElem(hostnames, mIngress.Status.LoadBalancer.Ingress[i].Hostname) {
			mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress[:i], mIngress.Status.LoadBalancer.Ingress[i+1:]...)
		}
	}

	// Handle fresh hostname update
	if updateOption.vip != "" {
		for _, host := range hostnames {
			lbIngress := core.LoadBalancerIngress{
				IP:       updateOption.vip,
				Hostname: host,
			}
			mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress, lbIngress)
		}
	}

	// remove the host from status which is not in spec
	for i := len(mIngress.Status.LoadBalancer.Ingress) - 1; i >= 0; i-- {
		if !utils.HasElem(hostListIng, mIngress.Status.LoadBalancer.Ingress[i].Hostname) {
			mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress[:i], mIngress.Status.LoadBalancer.Ingress[i+1:]...)
		}
	}

	if sameStatus := compareLBStatus(oldIngressStatus, &mIngress.Status.LoadBalancer); sameStatus {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in ingress status. old: %+v new: %+v",
			key, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
		return nil
	}

	if lib.GetIngressApi() == utils.ExtV1IngressInformer {
		mIng, ok := utils.ToExtensionIngress(mIngress)
		if !ok {
			err = errors.New("Unable to convert obj type interface to extensions/v1beta1 ingress")
			utils.AviLog.Error(err)
			return err
		}
		_, err = mClient.ExtensionsV1beta1().Ingresses(mIng.Namespace).UpdateStatus(mIng)
	} else {
		_, err = mClient.NetworkingV1beta1().Ingresses(mIngress.Namespace).UpdateStatus(mIngress)
	}
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the ingress status: %v", key, err)
		// fetch updated ingress and feed for update status
		mIngresses := getIngresses([]string{mIngress.Namespace + "/" + mIngress.Name}, false)
		if len(mIngresses) > 0 {
			return updateObject(mIngresses[mIngress.Namespace+"/"+mIngress.Name], updateOption, retry+1)
		}
	}

	utils.AviLog.Infof("key: %s, msg: Successfully updated the ingress status of ingress: %s/%s old: %+v new: %+v",
		key, mIngress.Namespace, mIngress.Name, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
	return err
}

func DeleteIngressStatus(svc_mdata_obj avicache.ServiceMetadataObj, isVSDelete bool, key string) error {
	var err error
	if len(svc_mdata_obj.NamespaceIngressName) > 0 {
		// This is SNI with hostname sharding.
		for _, ingressns := range svc_mdata_obj.NamespaceIngressName {
			ingressArr := strings.Split(ingressns, "/")
			if len(ingressArr) != 2 {
				return errors.New("key: %s, msg: DeleteIngressStatus IngressNamespace format not correct")
			}
			svc_mdata_obj.Namespace = ingressArr[0]
			svc_mdata_obj.IngressName = ingressArr[1]
			err = deleteObject(svc_mdata_obj, key, isVSDelete)
		}
	} else {
		err = deleteObject(svc_mdata_obj, key, isVSDelete)
	}

	if err != nil {
		utils.AviLog.Warn(err)
	}
	return err
}

func deleteObject(svc_mdata_obj avicache.ServiceMetadataObj, key string, isVSDelete bool, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: Retrying to update the ingress status", key)
		retry = retryNum[0]
		if retry >= 2 {
			return errors.New("key: %s, msg: DeleteIngressStatus retried 3 times, aborting")
		}
	}

	mClient := utils.GetInformers().ClientSet
	var ingObj interface{}
	var err error

	if lib.GetIngressApi() == utils.ExtV1IngressInformer {
		ingObj, err = mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).Get(svc_mdata_obj.IngressName, metav1.GetOptions{})
	} else {
		ingObj, err = mClient.NetworkingV1beta1().Ingresses(svc_mdata_obj.Namespace).Get(svc_mdata_obj.IngressName, metav1.GetOptions{})
	}

	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not get the ingress object for DeleteStatus :%s", key, err)
		return err
	}

	mIngress, ok := utils.ToNetworkingIngress(ingObj)
	if !ok {
		utils.AviLog.Errorf("Unable to convert obj type interface to networking/v1beta1 ingress")
	}

	oldIngressStatus := mIngress.Status.LoadBalancer.DeepCopy()
	var hostListIng []string
	for _, rule := range mIngress.Spec.Rules {
		hostListIng = append(hostListIng, rule.Host)
	}

	for i, status := range mIngress.Status.LoadBalancer.Ingress {
		for _, host := range svc_mdata_obj.HostNames {
			if status.Hostname == host {
				// Check if this host is still present in the spec, if so - don't delete it
				if !utils.HasElem(hostListIng, host) || isVSDelete {
					mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress[:i], mIngress.Status.LoadBalancer.Ingress[i+1:]...)
				} else {
					utils.AviLog.Debugf("key: %s, msg: skipping status update since host is present in the ingress: %v", key, host)
				}
			}
		}
	}

	if sameStatus := compareLBStatus(oldIngressStatus, &mIngress.Status.LoadBalancer); sameStatus {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in ingress status. old: %+v new: %+v",
			key, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
		return nil
	}

	if lib.GetIngressApi() == utils.ExtV1IngressInformer {
		mIng, ok := utils.ToExtensionIngress(mIngress)
		if !ok {
			utils.AviLog.Warn("Unable to convert obj type interface to extensions/v1beta1 ingress")
		}

		_, err = mClient.ExtensionsV1beta1().Ingresses(mIngress.Namespace).UpdateStatus(mIng)
	} else {
		_, err = mClient.NetworkingV1beta1().Ingresses(svc_mdata_obj.Namespace).UpdateStatus(mIngress)
	}
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in deleting the ingress status: %v", key, err)
		return deleteObject(svc_mdata_obj, key, isVSDelete, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully deleted the ingress status of ingress: %s/%s old: %+v new: %+v",
		key, mIngress.Namespace, mIngress.Name, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
	return nil
}

func UpdateL4LBStatus(options []UpdateStatusOptions, bulk bool) {
	mClient := utils.GetInformers().ClientSet
	var servicesToUpdate []string
	var updateServiceOptions []UpdateStatusOptions

	for _, option := range options {
		if len(option.serviceMetadata.HostNames) != 1 {
			utils.AviLog.Error("Hostname length not appropriate for status update, not equals 1")
			continue
		}

		service := option.serviceMetadata.Namespace + "/" + option.serviceMetadata.ServiceName
		option.ingsvc = service
		servicesToUpdate = append(servicesToUpdate, service)
		updateServiceOptions = append(updateServiceOptions, option)
	}

	serviceMap := getServices(servicesToUpdate, bulk)
	for _, option := range updateServiceOptions {
		key, svcMetadata := option.key, option.serviceMetadata
		if service := serviceMap[option.ingsvc]; service != nil {
			oldServiceStatus := service.Status.LoadBalancer.DeepCopy()
			if option.vip == "" {
				// nothing to do here
				continue
			}

			service.Status = core.ServiceStatus{
				LoadBalancer: core.LoadBalancerStatus{
					Ingress: []core.LoadBalancerIngress{core.LoadBalancerIngress{
						IP:       option.vip,
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
				key, option.ingsvc, oldServiceStatus.Ingress, service.Status.LoadBalancer.Ingress)
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
	mLb.Status = core.ServiceStatus{
		LoadBalancer: core.LoadBalancerStatus{
			Ingress: []core.LoadBalancerIngress{},
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

// SyncIngressStatus gets data from L3 cache and does a status update on the ingress objects
// based on the service metadata objects it finds in the cache
// This is executed once AKO is done with populating the L3 cache in reboot scenarios
func (rest *RestOperations) SyncIngressStatus() {
	vsKeys := rest.cache.VsCacheMeta.AviGetAllVSKeys()
	utils.AviLog.Debugf("Ingress status sync for vsKeys %+v", utils.Stringify(vsKeys))

	var allIngressUpdateOptions []UpdateStatusOptions
	var allServiceLBUpdateOptions []UpdateStatusOptions
	for _, vsKey := range vsKeys {
		vsCache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if !ok {
			continue
		}

		vsCacheObj, found := vsCache.(*avicache.AviVsCache)
		if !found {
			continue
		}

		parentVsKey := vsCacheObj.ParentVSRef
		vsSvcMetadataObj := vsCacheObj.ServiceMetadataObj
		if parentVsKey != (avicache.NamespaceName{}) {
			// secure VSes handler
			parentVs, found := rest.cache.VsCacheMeta.AviCacheGet(parentVsKey)
			if !found {
				continue
			}

			parentVsObj, _ := parentVs.(*avicache.AviVsCache)
			if (vsSvcMetadataObj.IngressName != "" || len(vsSvcMetadataObj.NamespaceIngressName) > 0) && vsSvcMetadataObj.Namespace != "" && parentVsObj != nil {
				option := UpdateStatusOptions{vip: parentVsObj.Vip, serviceMetadata: vsSvcMetadataObj, key: "syncstatus"}
				allIngressUpdateOptions = append(allIngressUpdateOptions, option)
			}
		} else if vsSvcMetadataObj.ServiceName != "" && vsSvcMetadataObj.Namespace != "" {
			// serviceLB
			option := UpdateStatusOptions{vip: vsCacheObj.Vip, serviceMetadata: vsSvcMetadataObj, key: "syncstatus"}
			allServiceLBUpdateOptions = append(allServiceLBUpdateOptions, option)
		} else {
			// insecure VSes handler
			for _, poolKey := range vsCacheObj.PoolKeyCollection {
				poolCache, ok := rest.cache.PoolCache.AviCacheGet(poolKey)
				if !ok {
					continue
				}

				poolCacheObj, found := poolCache.(*avicache.AviPoolCache)
				if !found {
					continue
				}

				// insecure pools
				if poolCacheObj.ServiceMetadataObj.Namespace != "" {
					option := UpdateStatusOptions{vip: vsCacheObj.Vip, serviceMetadata: poolCacheObj.ServiceMetadataObj, key: "syncstatus"}
					allIngressUpdateOptions = append(allIngressUpdateOptions, option)
				}
			}
		}
	}

	UpdateIngressStatus(allIngressUpdateOptions, true)
	UpdateL4LBStatus(allServiceLBUpdateOptions, true)
	utils.AviLog.Infof("Status syncing completed")
	return
}

// compareLBStatus returns true if status objects are same, so status update is not required
func compareLBStatus(oldStatus, newStatus *corev1.LoadBalancerStatus) bool {
	if len(oldStatus.Ingress) != len(newStatus.Ingress) {
		return false
	}

	exists := []string{}
	for _, status := range oldStatus.Ingress {
		exists = append(exists, status.IP+":"+status.Hostname)
	}
	for _, status := range newStatus.Ingress {
		if !utils.HasElem(exists, status.IP+":"+status.Hostname) {
			return false
		}
	}

	return true
}

// getIngresses fetches all ingresses and returns a map: {"namespace/name": ingressObj...}
// if bulk is set to true, this fetches all ingresses in a single k8s api-server call
func getIngresses(ingressNSNames []string, bulk bool) map[string]*networking.Ingress {
	mClient := utils.GetInformers().ClientSet
	ingressMap := make(map[string]*networking.Ingress)
	var err error

	if bulk {
		if lib.GetIngressApi() == utils.ExtV1IngressInformer {
			ingressList, err := mClient.ExtensionsV1beta1().Ingresses("").List(metav1.ListOptions{})
			if err != nil {
				utils.AviLog.Warnf("Could not get the ingress object for UpdateStatus :%s", err)
			}
			for _, ing := range ingressList.Items {
				mIngress, ok := utils.ToNetworkingIngress(ing)
				if !ok {
					utils.AviLog.Errorf("Unable to convert obj type interface to networking/v1beta1 ingress %s", ing.Name)
					continue
				}
				ingressMap[mIngress.Namespace+"/"+mIngress.Name] = mIngress
			}
		} else {
			ingressList, err := mClient.NetworkingV1beta1().Ingresses("").List(metav1.ListOptions{})
			if err != nil {
				utils.AviLog.Warnf("Could not get the ingress object for UpdateStatus :%s", err)
			}
			for i := range ingressList.Items {
				ing := ingressList.Items[i]
				ingressMap[ing.Namespace+"/"+ing.Name] = &ing
			}
		}

		return ingressMap
	}

	for _, namespaceName := range ingressNSNames {
		var ingObj interface{}
		nsNameSplit := strings.Split(namespaceName, "/")
		if lib.GetIngressApi() == utils.ExtV1IngressInformer {
			ingObj, err = mClient.ExtensionsV1beta1().Ingresses(nsNameSplit[0]).Get(nsNameSplit[1], metav1.GetOptions{})
		} else {
			ingObj, err = mClient.NetworkingV1beta1().Ingresses(nsNameSplit[0]).Get(nsNameSplit[1], metav1.GetOptions{})
		}
		if err != nil {
			utils.AviLog.Warnf("msg: Could not get the ingress object for UpdateStatus :%s", err)
			continue
		}

		mIngress, ok := utils.ToNetworkingIngress(ingObj)
		if !ok {
			utils.AviLog.Warn("Unable to convert obj type interface to networking/v1beta1 ingress")
			continue
		}
		ingressMap[mIngress.Namespace+"/"+mIngress.Name] = mIngress
	}

	return ingressMap
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
