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
	"errors"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

type UpdateStatusOptions struct {
	// IngSvc format: namespace/name, not supposed to be provided by the caller
	IngSvc          string
	Vip             string
	ServiceMetadata avicache.ServiceMetadataObj
	Key             string
}

func UpdateIngressStatus(options []UpdateStatusOptions, bulk bool) {
	var err error
	ingressesToUpdate, updateIngressOptions := ParseOptionsFromMetadata(options, bulk)

	// ingressMap: {ns/ingress: ingressObj}
	// this pre-fetches all ingresses to be candidates for status update
	// after pre-fetching, if a status update comes for that ingress, then the pre-fetched ingress would be stale
	// in which case ingress will be fetched again in updateObject, as part of a retry
	ingressMap := getIngresses(ingressesToUpdate, bulk)
	for _, option := range updateIngressOptions {
		if ingress := ingressMap[option.IngSvc]; ingress != nil {
			if err = updateObject(ingress, option); err != nil {
				utils.AviLog.Error("key: %s, msg: updating Ingress object failed: %v", option.Key, err)
			}
		}
	}

	return
}

func updateObject(mIngress *networkingv1beta1.Ingress, updateOption UpdateStatusOptions, retryNum ...int) error {
	if updateOption.Vip == "" {
		return nil
	}

	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			return errors.New("UpdateIngressStatus retried 3 times, aborting")
		}
	}

	mClient := utils.GetInformers().ClientSet
	hostnames, key := updateOption.ServiceMetadata.HostNames, updateOption.Key
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
	if updateOption.Vip != "" {
		for _, host := range hostnames {
			lbIngress := corev1.LoadBalancerIngress{
				IP:       updateOption.Vip,
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

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": mIngress.Status,
	})
	_, err := mClient.NetworkingV1beta1().Ingresses(mIngress.Namespace).Patch(context.TODO(), mIngress.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
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
				utils.AviLog.Errorf("key: %s, msg: DeleteIngressStatus IngressNamespace format not correct", key)
				return errors.New("DeleteIngressStatus IngressNamespace format not correct")
			}
			svc_mdata_obj.Namespace = ingressArr[0]
			svc_mdata_obj.IngressName = ingressArr[1]
			err = deleteObject(svc_mdata_obj, key, isVSDelete)
		}
	} else {
		err = deleteObject(svc_mdata_obj, key, isVSDelete)
	}

	if err != nil {
		return err
	}

	return nil
}

func deleteObject(svc_mdata_obj avicache.ServiceMetadataObj, key string, isVSDelete bool, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: Retrying to update the ingress status", key)
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: DeleteIngressStatus retried 3 times, aborting", key)
			return errors.New("DeleteIngressStatus retried 3 times, aborting")
		}
	}

	mClient := utils.GetInformers().ClientSet
	mIngress, err := mClient.NetworkingV1beta1().Ingresses(svc_mdata_obj.Namespace).Get(context.TODO(), svc_mdata_obj.IngressName, metav1.GetOptions{})
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not get the ingress object for DeleteStatus: %s", key, err)
		return err
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
				//NS migration case: if false -> ns invalid event happened so remove status
				nsMigrationFilterFlag := utils.CheckIfNamespaceAccepted(svc_mdata_obj.Namespace, utils.GetGlobalNSFilter(), nil, true)
				if !utils.HasElem(hostListIng, host) || isVSDelete || !nsMigrationFilterFlag {
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

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": mIngress.Status,
	})
	if len(mIngress.Status.LoadBalancer.Ingress) == 0 {
		patchPayload, _ = json.Marshal(map[string]interface{}{
			"status": nil,
		})
	}
	_, err = mClient.NetworkingV1beta1().Ingresses(svc_mdata_obj.Namespace).Patch(context.TODO(), mIngress.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in deleting the ingress status: %v", key, err)
		return deleteObject(svc_mdata_obj, key, isVSDelete, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully deleted the ingress status of ingress: %s/%s old: %+v new: %+v",
		key, mIngress.Namespace, mIngress.Name, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
	return nil
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
func getIngresses(ingressNSNames []string, bulk bool, retryNum ...int) map[string]*networkingv1beta1.Ingress {
	retry := 0
	mClient := utils.GetInformers().ClientSet
	ingressMap := make(map[string]*networkingv1beta1.Ingress)
	if len(retryNum) > 0 {
		utils.AviLog.Infof("Retrying to get the ingress for status update")
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("getIngresses for status update retried 3 times, aborting")
			return ingressMap
		}
	}

	if bulk {
		ingressList, err := mClient.NetworkingV1beta1().Ingresses("").List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the ingress object for UpdateStatus: %v", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getIngresses(ingressNSNames, bulk, retry+1)
			}
		}
		for i := range ingressList.Items {
			ing := ingressList.Items[i]
			ingressMap[ing.Namespace+"/"+ing.Name] = &ing
		}

		return ingressMap
	}

	for _, namespaceName := range ingressNSNames {
		nsNameSplit := strings.Split(namespaceName, "/")

		mIngress, err := mClient.NetworkingV1beta1().Ingresses(nsNameSplit[0]).Get(context.TODO(), nsNameSplit[1], metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the ingress object for UpdateStatus: %v", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getIngresses(ingressNSNames, bulk, retry+1)
			}
			continue
		}

		ingressMap[mIngress.Namespace+"/"+mIngress.Name] = mIngress
	}

	return ingressMap
}
