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
	"fmt"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	corev1 "k8s.io/api/core/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type UpdateOptions struct {
	// IngSvc format: namespace/name, not supposed to be provided by the caller
	IngSvc             string
	Vip                string
	ServiceMetadata    avicache.ServiceMetadataObj
	Key                string
	VirtualServiceUUID string
}

const (
	VSAnnotation         = "ako.vmware.com/host-fqdn-vs-uuid-map"
	ControllerAnnotation = "ako.vmware.com/controller-cluster-uuid"
)

// VSUuidAnnotation is maps a hostname to the UUID of the virtual service where it is placed.

func UpdateIngressStatus(options []UpdateOptions, bulk bool) {
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

func updateObject(mIngress *networkingv1beta1.Ingress, updateOption UpdateOptions, retryNum ...int) error {
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

	sameStatus := compareLBStatus(oldIngressStatus, &mIngress.Status.LoadBalancer)

	var updatedIng *networkingv1beta1.Ingress
	var err error
	if !sameStatus {
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": mIngress.Status,
		})
		updatedIng, err = mClient.NetworkingV1beta1().Ingresses(mIngress.Namespace).Patch(context.TODO(), mIngress.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
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
	} else {
		utils.AviLog.Debugf("key: %s, msg: no changes detected in the ingress %s/%s status", key, mIngress.Namespace, mIngress.Name)
	}

	// update the annotations for this object
	err = updateIngAnnotations(mClient, updatedIng, hostnames, updateOption.VirtualServiceUUID, key, hostListIng, mIngress)
	if err != nil {
		return fmt.Errorf("key: %s, error in updating the Ingress annotations: %v", key, err)
	}
	return nil
}

func updateIngAnnotations(mClient kubernetes.Interface, ingObj *networkingv1beta1.Ingress, hostnamesToBeUpdated []string,
	vsUUID, key string, ingSpecHostnames []string, oldIng *networkingv1beta1.Ingress, retryNum ...int) error {

	if ingObj == nil {
		ingObj = oldIng
	}
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			return fmt.Errorf("retried 3 times to update ingress annotations, aborting")
		}
	}
	var err error
	vsAnnotations := make(map[string]string)

	if value, ok := ingObj.Annotations[VSAnnotation]; ok {
		if err := json.Unmarshal([]byte(value), &vsAnnotations); err != nil {
			// just print an error and continue, this will be taken care of during the update
			utils.AviLog.Errorf("key: %s, error in unmarshalling Ingress %s/%s annotations for VS: %v",
				key, ingObj.Namespace, ingObj.Name, err)
		}
	}

	// update the existing hostname vs uuid for the current update
	for i := 0; i < len(hostnamesToBeUpdated); i++ {
		vsAnnotations[hostnamesToBeUpdated[i]] = vsUUID
	}

	// remove the hostname from annotations which is not part of the spec
	for k := range vsAnnotations {
		if !utils.HasElem(ingSpecHostnames, k) {
			delete(vsAnnotations, k)
		}
	}

	// compare the vs annotations for this ingress object
	req := isAnnotationsUpdateRequired(ingObj.Annotations, vsAnnotations)
	if !req {
		utils.AviLog.Debugf("annotations update not required for this ingress: %s/%s", ingObj.Namespace, ingObj.Name)
		return nil
	}
	if err = patchIngressAnnotations(ingObj, vsAnnotations, mClient); err != nil && k8serrors.IsNotFound(err) {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the ingress annotations: %v", key, err)
		// fetch updated ingress and feed for update status
		mIngresses := getIngresses([]string{ingObj.Namespace + "/" + ingObj.Name}, false)
		if len(mIngresses) > 0 {
			return updateIngAnnotations(mClient, mIngresses[ingObj.Namespace+"/"+ingObj.Name], hostnamesToBeUpdated,
				vsUUID, key, ingSpecHostnames, oldIng, retry+1)
		}
	}

	return nil
}

func isAnnotationsUpdateRequired(ingAnnotations map[string]string, newVSAnnotations map[string]string) bool {
	oldVSAnnotationsStr, ok := ingAnnotations[VSAnnotation]
	if !ok {
		if len(newVSAnnotations) > 0 {
			return true
		}
		return false
	}

	var oldVSAnnotations map[string]string
	if err := json.Unmarshal([]byte(oldVSAnnotationsStr), &oldVSAnnotations); err != nil {
		utils.AviLog.Errorf("error in unmarshalling old vs annotations %s: %v", oldVSAnnotationsStr, err)
		return true
	}

	if len(oldVSAnnotations) != len(newVSAnnotations) {
		return true
	}
	for oldHost, oldVS := range oldVSAnnotations {
		newVS, exists := newVSAnnotations[oldHost]
		if !exists || (newVS != oldVS) {
			return true
		}
	}
	return false
}

func getAnnotationsPayload(vsAnnotations map[string]string, existingAnnotations map[string]string) ([]byte, error) {
	vsAnnotationsStr, err := json.Marshal(vsAnnotations)
	if err != nil {
		return nil, fmt.Errorf("error in marshalling vs annotations: %v", err)
	}
	if len(existingAnnotations) == 0 {
		existingAnnotations = make(map[string]string)
	}
	existingAnnotations[VSAnnotation] = string(vsAnnotationsStr)
	existingAnnotations[ControllerAnnotation] = avicache.GetControllerClusterUUID()
	patchPayload := map[string]interface{}{
		"metadata": map[string]map[string]string{
			"annotations": existingAnnotations,
		},
	}
	patchPayloadBytes, err := json.Marshal(patchPayload)
	if err != nil {
		return nil, fmt.Errorf("error in marshalling patch payload %v: %v", patchPayloadBytes, err)
	}
	return patchPayloadBytes, nil
}

func patchIngressAnnotations(ingObj *networkingv1beta1.Ingress, vsAnnotations map[string]string, mClient kubernetes.Interface) error {
	annotations := ingObj.GetAnnotations()
	patchPayloadBytes, err := getAnnotationsPayload(vsAnnotations, annotations)
	if err != nil {
		return fmt.Errorf("error in generating payload for vs annotations %v: %v", vsAnnotations, err)
	}

	_, err = mClient.NetworkingV1beta1().Ingresses(ingObj.Namespace).Patch(context.TODO(), ingObj.Name, types.MergePatchType, patchPayloadBytes, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error in updating ingress: %v", err)
	}
	return nil
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

	for i, status := range mIngress.Status.LoadBalancer.Ingress {
		for _, host := range svc_mdata_obj.HostNames {
			if status.Hostname == host {
				mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress[:i], mIngress.Status.LoadBalancer.Ingress[i+1:]...)
			}
		}
	}

	sameStatus := compareLBStatus(oldIngressStatus, &mIngress.Status.LoadBalancer)

	var updatedIng *networkingv1beta1.Ingress
	if !sameStatus {
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": mIngress.Status,
		})
		if len(mIngress.Status.LoadBalancer.Ingress) == 0 {
			patchPayload, _ = json.Marshal(map[string]interface{}{
				"status": nil,
			})
		}
		updatedIng, err = mClient.NetworkingV1beta1().Ingresses(svc_mdata_obj.Namespace).Patch(context.TODO(), mIngress.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
		if err != nil {
			utils.AviLog.Errorf("key: %s, msg: there was an error in deleting the ingress status: %v", key, err)
			return deleteObject(svc_mdata_obj, key, isVSDelete, retry+1)
		}

		utils.AviLog.Infof("key: %s, msg: Successfully deleted the ingress status of ingress: %s/%s old: %+v new: %+v",
			key, mIngress.Namespace, mIngress.Name, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)

	} else {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in ingress status. old: %+v new: %+v",
			key, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
	}

	if err = deleteIngressAnnotation(updatedIng, svc_mdata_obj, isVSDelete, hostListIng, key, mClient, mIngress); err != nil {
		utils.AviLog.Errorf("key: %s, msg: error in deleting ingress annotation: %v", key, err)
	}

	return nil
}

func deleteIngressAnnotation(ingObj *networkingv1beta1.Ingress, svcMeta avicache.ServiceMetadataObj, isVSDelete bool,
	ingHostList []string, key string, mClient kubernetes.Interface, oldIng *networkingv1beta1.Ingress,
	retryNum ...int) error {
	if ingObj == nil {
		ingObj = oldIng
	}
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: Retrying to update the ingress status", key)
		retry = retryNum[0]
		if retry >= 3 {
			return fmt.Errorf("retried 3 times to delete the ingress annotations, aborting")
		}
	}
	existingAnnotations := make(map[string]string)
	if annotations, exists := ingObj.Annotations[VSAnnotation]; exists {
		if err := json.Unmarshal([]byte(annotations), &existingAnnotations); err != nil {
			return fmt.Errorf("error in unmarshalling annotations for ingress: %v", err)
		}
	} else {
		return fmt.Errorf("error in fetching annotations for ingress %s/%s", ingObj.Namespace, ingObj.Name)
	}

	for k := range existingAnnotations {
		for _, host := range svcMeta.HostNames {
			if k == host {
				// Check if:
				// 1. this host is still present in the spec, if so - don't delete it from annotations
				// 2. in case of NS migration, if NS is moved from selected to rejected, this host then
				//    has to be removed from the annotations list.
				nsMigrationFilterFlag := utils.CheckIfNamespaceAccepted(svcMeta.Namespace)
				if !utils.HasElem(ingHostList, host) || isVSDelete || !nsMigrationFilterFlag {
					delete(existingAnnotations, k)
				} else {
					utils.AviLog.Debugf("key: %s, msg: skipping annotation update since host is present in the ingress: %v", key, host)
				}
			}
		}
	}

	if isAnnotationsUpdateRequired(ingObj.Annotations, existingAnnotations) {
		if err := patchIngressAnnotations(ingObj, existingAnnotations, mClient); err != nil {
			return deleteIngressAnnotation(ingObj, svcMeta, isVSDelete, ingHostList, key, mClient, oldIng, retry+1)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Annotations unchanged for ingress %s/%s", key, ingObj.Namespace, ingObj.Name)

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
