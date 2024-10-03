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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
)

type Status struct {
	*gatewayv1.GatewayClassStatus
	*gatewayv1.GatewayStatus
	*gatewayv1.HTTPRouteStatus
}
type UpdateOptions struct {
	// IngSvc format: namespace/name, not supposed to be provided by the caller
	IngSvc             string
	Vip                []string
	ServiceMetadata    lib.ServiceMetadataObj
	Key                string
	VirtualServiceUUID string
	VSName             string
	Message            string
	Tenant             string
	Status             *Status
}

// VSUuidAnnotation is maps a hostname to the UUID of the virtual service where it is placed.
func (l *leader) UpdateIngressStatus(options []UpdateOptions, bulk bool) {
	var err error
	ingressesToUpdate, updateIngressOptions := ParseOptionsFromMetadata(options, bulk)

	// ingressMap: {ns/ingress: ingressObj}
	// this pre-fetches all ingresses to be candidates for status update
	// after pre-fetching, if a status update comes for that ingress, then the pre-fetched ingress would be stale
	// in which case ingress will be fetched again in updateObject, as part of a retry
	ingressMap := getIngresses(ingressesToUpdate, bulk)
	skipDelete := map[string]bool{}
	for _, option := range updateIngressOptions {
		if ingress := ingressMap[option.IngSvc]; ingress != nil {
			if err = updateObject(ingress, option); err != nil {
				utils.AviLog.Errorf("key: %s, msg: updating Ingress object failed: %v", option.Key, err)
			}
			skipDelete[option.IngSvc] = true
		}
	}
	// reset IPAddress and annotations from Ingresses that do not have a corresponding VS in cache
	// this comes in handy when, during bulk ingress removal (lets say IngressClass is removed/deleteConfig etc.),
	// VSes will be deleted from Avi. At that point if the VS is deleted, but the ingress status update
	// is incomplete, this is required to fix it as part of bootup.
	if bulk {
		for ingNSName, ing := range ingressMap {
			if val, ok := skipDelete[ingNSName]; ok && val {
				continue
			}
			var hostnames []string
			for _, rule := range ing.Spec.Rules {
				hostnames = append(hostnames, rule.Host)
			}
			l.DeleteIngressStatus([]UpdateOptions{{
				ServiceMetadata: lib.ServiceMetadataObj{
					NamespaceIngressName: []string{ingNSName},
					HostNames:            hostnames,
				},
			}}, true, lib.SyncStatusKey)
		}
	}
}

func updateObject(mIngress *networkingv1.Ingress, updateOption UpdateOptions, retryNum ...int) error {
	if len(updateOption.Vip) == 0 {
		return nil
	}

	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			return errors.New("UpdateIngressStatus retried 3 times, aborting")
		}
	}

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
	for _, host := range hostnames {
		for _, vip := range updateOption.Vip {
			lbIngress := networkingv1.IngressLoadBalancerIngress{
				IP:       vip,
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

	// Add/Remove Ingress finalizers only when operating in a VCF Cluster.
	if utils.IsVCFCluster() {
		if len(mIngress.Status.LoadBalancer.Ingress) > 0 && mIngress.Status.LoadBalancer.Ingress[0].IP != "" {
			lib.CheckAndSetIngressFinalizer(mIngress)
		} else {
			lib.RemoveIngressFinalizer(mIngress)
		}
	}

	// we need hosts for which the IP is getting removed, and hosts for which it is being added/updated
	sameStatus, hostsBefore, hostsAfter := compareIngressLBStatus(oldIngressStatus, &mIngress.Status.LoadBalancer)
	var updatedIng *networkingv1.Ingress
	var err error
	if !sameStatus {
		ingressNsName := mIngress.Namespace + "/" + mIngress.Name
		//lock here to avoid concurrent updates to same status
		lib.GetLockSet().Lock(ingressNsName)
		latestIngress := getIngresses([]string{ingressNsName}, false)
		if latestIngress[ingressNsName] != nil {
			latestIngressStatus := latestIngress[ingressNsName].Status.LoadBalancer.DeepCopy()
			if latestIngressStatus.String() != oldIngressStatus.String() {
				lib.GetLockSet().Unlock(ingressNsName)
				//unlock and retry if status was changed by concurrent operation with new status
				//retry counter not updated since this is not a failure case
				return updateObject(latestIngress[ingressNsName], updateOption, retry)
			}
		}
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": mIngress.Status,
		})

		updatedIng, err = utils.GetInformers().ClientSet.NetworkingV1().Ingresses(mIngress.Namespace).Patch(context.TODO(), mIngress.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
		if err != nil {
			utils.AviLog.Errorf("key: %s, msg: there was an error in updating the ingress status: %v", key, err)
			// fetch updated ingress and feed for update status
			mIngresses := getIngresses([]string{mIngress.Namespace + "/" + mIngress.Name}, false)
			if len(mIngresses) > 0 {
				lib.GetLockSet().Unlock(ingressNsName)
				return updateObject(mIngresses[mIngress.Namespace+"/"+mIngress.Name], updateOption, retry+1)
			}
		} else {
			for _, hns := range (hostsBefore.Difference(hostsAfter)).UnsortedList() {
				lib.AKOControlConfig().EventRecorder().Eventf(updatedIng, corev1.EventTypeNormal, lib.Removed, "Removed virtualservice for %s", hns)
			}
			for _, hns := range (hostsAfter.Difference(hostsBefore)).UnsortedList() {
				lib.AKOControlConfig().EventRecorder().Eventf(updatedIng, corev1.EventTypeNormal, lib.Synced, "Added virtualservice %s for %s", updateOption.VSName, hns)
			}
			utils.AviLog.Infof("key: %s, msg: Successfully updated the ingress status of ingress: %s/%s old: %+v new: %+v",
				key, mIngress.Namespace, mIngress.Name, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
		}
		lib.GetLockSet().Unlock(ingressNsName)
	} else {
		utils.AviLog.Debugf("key: %s, msg: no changes detected in the ingress %s/%s status", key, mIngress.Namespace, mIngress.Name)
	}

	// update the annotations for this object
	err = updateIngAnnotations(updatedIng, hostnames, updateOption.VirtualServiceUUID, key, updateOption.Tenant, hostListIng, mIngress)
	if err != nil {
		return fmt.Errorf("key: %s, error in updating the Ingress annotations: %v", key, err)
	}
	return nil
}

func updateIngAnnotations(ingObj *networkingv1.Ingress, hostnamesToBeUpdated []string,
	vsUUID, key, tenant string, ingSpecHostnames []string, oldIng *networkingv1.Ingress, retryNum ...int) error {

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

	if value, ok := ingObj.Annotations[lib.VSAnnotation]; ok {
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
	req := isAnnotationsUpdateRequired(ingObj.Annotations, vsAnnotations, tenant, false)
	if !req {
		utils.AviLog.Debugf("annotations update not required for this ingress: %s/%s", ingObj.Namespace, ingObj.Name)
		return nil
	}
	if err = patchIngressAnnotations(ingObj, vsAnnotations, tenant); err != nil && k8serrors.IsNotFound(err) {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the ingress annotations: %v", key, err)
		// fetch updated ingress and feed for update status
		mIngresses := getIngresses([]string{ingObj.Namespace + "/" + ingObj.Name}, false)
		if len(mIngresses) > 0 {
			return updateIngAnnotations(mIngresses[ingObj.Namespace+"/"+ingObj.Name], hostnamesToBeUpdated,
				vsUUID, key, tenant, ingSpecHostnames, oldIng, retry+1)
		}
	}

	return nil
}

func isAnnotationsUpdateRequired(ingAnnotations map[string]string, newVSAnnotations map[string]string, newTenant string, isDelete bool) bool {
	oldVSAnnotationsStr, ok := ingAnnotations[lib.VSAnnotation]
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
	if !isDelete {
		oldTenant, ok := ingAnnotations[lib.TenantAnnotation]
		if !ok {
			return true
		}
		if oldTenant != newTenant {
			return true
		}
	}
	return false
}

func getAnnotationsPayload(vsAnnotations map[string]string, tenant string) ([]byte, error) {
	var vsAnnotationVal, ctrlAnnotationVal, tenantVal *string
	ctrlAnnotationValStr := avicache.GetControllerClusterUUID()
	if len(vsAnnotations) > 0 {
		vsAnnotationsBytes, err := json.Marshal(vsAnnotations)
		if err != nil {
			return nil, fmt.Errorf("error in marshalling vs annotations: %v", err)
		}
		vsAnnotationsStrStr := string(vsAnnotationsBytes)
		vsAnnotationVal = &vsAnnotationsStrStr
		ctrlAnnotationVal = &ctrlAnnotationValStr
		tenantVal = &tenant
	}

	patchPayload := map[string]interface{}{
		"metadata": map[string]map[string]*string{
			"annotations": {
				lib.VSAnnotation:         vsAnnotationVal,
				lib.ControllerAnnotation: ctrlAnnotationVal,
				lib.TenantAnnotation:     tenantVal,
			},
		},
	}
	patchPayloadBytes, err := json.Marshal(patchPayload)
	if err != nil {
		return nil, fmt.Errorf("error in marshalling patch payload %v: %v", patchPayloadBytes, err)
	}
	return patchPayloadBytes, nil
}

func patchIngressAnnotations(ingObj *networkingv1.Ingress, vsAnnotations map[string]string, tenant string) error {
	patchPayloadBytes, err := getAnnotationsPayload(vsAnnotations, tenant)
	if err != nil {
		return fmt.Errorf("error in generating payload for vs annotations %v: %v", vsAnnotations, err)
	}
	if _, err = utils.GetInformers().ClientSet.NetworkingV1().Ingresses(ingObj.Namespace).Patch(context.TODO(), ingObj.Name, types.MergePatchType, patchPayloadBytes, metav1.PatchOptions{}); err != nil {
		return err
	}
	return nil
}

func (l *leader) DeleteIngressStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	if len(options) == 0 {
		return fmt.Errorf("Length of options is zero")
	}
	var err error
	if len(options[0].ServiceMetadata.NamespaceIngressName) > 0 {
		// This is SNI with hostname sharding.
		for _, ingressns := range options[0].ServiceMetadata.NamespaceIngressName {
			ingressArr := strings.Split(ingressns, "/")
			if len(ingressArr) != 2 {
				utils.AviLog.Errorf("key: %s, msg: DeleteIngressStatus IngressNamespace format not correct", key)
				return errors.New("DeleteIngressStatus IngressNamespace format not correct")
			}
			options[0].ServiceMetadata.Namespace = ingressArr[0]
			options[0].ServiceMetadata.IngressName = ingressArr[1]
			err = deleteObject(options[0], key, isVSDelete)
		}
	} else {
		err = deleteObject(options[0], key, isVSDelete)
	}

	if err != nil {
		return err
	}

	return nil
}

func deleteObject(option UpdateOptions, key string, isVSDelete bool, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: Retrying to update the ingress status", key)
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: DeleteIngressStatus retried 3 times, aborting", key)
			return errors.New("DeleteIngressStatus retried 3 times, aborting")
		}
	}

	mIngress, err := utils.GetInformers().IngressInformer.Lister().Ingresses(option.ServiceMetadata.Namespace).Get(option.ServiceMetadata.IngressName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not get the ingress object for DeleteStatus: %s", key, err)
		return err
	}

	oldIngressStatus := mIngress.Status.LoadBalancer.DeepCopy()
	var hostListIng []string
	for _, rule := range mIngress.Spec.Rules {
		hostListIng = append(hostListIng, rule.Host)
	}

	for _, host := range option.ServiceMetadata.HostNames {
		for i := len(mIngress.Status.LoadBalancer.Ingress) - 1; i >= 0; i-- {
			if mIngress.Status.LoadBalancer.Ingress[i].Hostname != host {
				continue
			}
			if !lib.ValidateIngressForClass(key, mIngress) ||
				!utils.CheckIfNamespaceAccepted(option.ServiceMetadata.Namespace) ||
				!utils.HasElem(hostListIng, host) ||
				isVSDelete ||
				mIngress.GetDeletionTimestamp() != nil {
				mIngress.Status.LoadBalancer.Ingress = append(mIngress.Status.LoadBalancer.Ingress[:i], mIngress.Status.LoadBalancer.Ingress[i+1:]...)
			} else {
				utils.AviLog.Debugf("key: %s, msg: skipping status deletion since host is present in the ingress: %v", key, host)
			}
		}
	}

	// Add/Remove Ingress finalizers only when operating in a VCF Cluster.
	if utils.IsVCFCluster() {
		if len(mIngress.Status.LoadBalancer.Ingress) > 0 && mIngress.Status.LoadBalancer.Ingress[0].IP != "" {
			lib.CheckAndSetIngressFinalizer(mIngress)
		} else {
			lib.RemoveIngressFinalizer(mIngress)
		}
	}

	sameStatus, hostsBefore, hostsAfter := compareIngressLBStatus(oldIngressStatus, &mIngress.Status.LoadBalancer)

	var updatedIng *networkingv1.Ingress
	if !sameStatus {
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": mIngress.Status,
		})
		if len(mIngress.Status.LoadBalancer.Ingress) == 0 {
			patchPayload, _ = json.Marshal(map[string]interface{}{
				"status": nil,
			})
		}
		updatedIng, err = utils.GetInformers().ClientSet.NetworkingV1().Ingresses(option.ServiceMetadata.Namespace).Patch(context.TODO(), mIngress.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
		if err != nil {
			utils.AviLog.Errorf("key: %s, msg: there was an error in deleting the ingress status: %v", key, err)
			return deleteObject(option, key, isVSDelete, retry+1)
		} else {
			for _, hns := range (hostsBefore.Difference(hostsAfter)).UnsortedList() {
				lib.AKOControlConfig().EventRecorder().Eventf(updatedIng, corev1.EventTypeNormal, lib.Removed, "Removed virtualservice for %s", hns)
			}
			utils.AviLog.Infof("key: %s, msg: Successfully deleted the ingress status of ingress: %s/%s old: %+v new: %+v",
				key, mIngress.Namespace, mIngress.Name, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
		}
	} else {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in ingress status. old: %+v new: %+v",
			key, oldIngressStatus.Ingress, mIngress.Status.LoadBalancer.Ingress)
	}

	if err = deleteIngressAnnotation(updatedIng, option.ServiceMetadata, isVSDelete, key, option.Tenant, mIngress, hostListIng); err != nil {
		utils.AviLog.Errorf("key: %s, msg: error in deleting ingress annotation: %v", key, err)
	}

	return nil
}

func deleteIngressAnnotation(ingObj *networkingv1.Ingress, svcMeta lib.ServiceMetadataObj, isVSDelete bool,
	key, tenant string, oldIng *networkingv1.Ingress, ingHostList []string, retryNum ...int) error {
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
	if annotations, exists := ingObj.Annotations[lib.VSAnnotation]; exists {
		if err := json.Unmarshal([]byte(annotations), &existingAnnotations); err != nil {
			return fmt.Errorf("error in unmarshalling annotations for ingress: %v", err)
		}
	} else {
		utils.AviLog.Debugf("VS annotations not found for ingress %s/%s", ingObj.Namespace, ingObj.Name)
		return nil
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
					utils.AviLog.Debugf("key: %s, msg: skipping annotation update since host is present in the ing: %v", key, host)
				}
			}
		}
	}
	if isAnnotationsUpdateRequired(ingObj.Annotations, existingAnnotations, tenant, isVSDelete) {
		if err := patchIngressAnnotations(ingObj, existingAnnotations, tenant); err != nil {
			return deleteIngressAnnotation(ingObj, svcMeta, isVSDelete, key, tenant, oldIng, ingHostList, retry+1)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Annotations unchanged for ingress %s/%s", key, ingObj.Namespace, ingObj.Name)

	return nil
}

// compareLBStatus returns true if status objects are same, so status update is not required
func compareIngressLBStatus(oldStatus, newStatus *networkingv1.IngressLoadBalancerStatus) (bool, sets.Set[string], sets.Set[string]) {
	exists := sets.Set[string]{}
	oldHosts := sets.Set[string]{}
	newHosts := sets.Set[string]{}
	var diff *bool
	for _, status := range oldStatus.Ingress {
		exists.Insert(status.IP + ":" + status.Hostname)
		oldHosts.Insert(status.Hostname)
	}

	if len(newStatus.Ingress) != len(oldStatus.Ingress) {
		diff = proto.Bool(false)
	}
	for _, status := range newStatus.Ingress {
		if !exists.Has(status.IP + ":" + status.Hostname) {
			if diff == nil {
				diff = proto.Bool(false)
			}
		}
		newHosts.Insert(status.Hostname)
	}

	if diff == nil {
		diff = proto.Bool(true)
	}
	return *diff, oldHosts, newHosts
}

func compareLBStatus(oldStatus, newStatus *corev1.LoadBalancerStatus) (bool, sets.Set[string], sets.Set[string]) {
	exists := sets.Set[string]{}
	oldHosts := sets.Set[string]{}
	newHosts := sets.Set[string]{}
	var diff *bool
	for _, status := range oldStatus.Ingress {
		exists.Insert(status.IP + ":" + status.Hostname)
		oldHosts.Insert(status.Hostname)
	}

	if len(newStatus.Ingress) != len(oldStatus.Ingress) {
		diff = proto.Bool(false)
	}
	for _, status := range newStatus.Ingress {
		if !exists.Has(status.IP + ":" + status.Hostname) {
			if diff == nil {
				diff = proto.Bool(false)
			}
		}
		newHosts.Insert(status.Hostname)
	}

	if diff == nil {
		diff = proto.Bool(true)
	}
	return *diff, oldHosts, newHosts
}

// getIngresses fetches all ingresses and returns a map: {"namespace/name": ingressObj...}
// if bulk is set to true, this fetches all ingresses in a single k8s api-server call
func getIngresses(ingressNSNames []string, bulk bool, retryNum ...int) map[string]*networkingv1.Ingress {
	retry := 0
	ingressMap := make(map[string]*networkingv1.Ingress)
	if len(retryNum) > 0 {
		utils.AviLog.Infof("Retrying to get the ingress for status update")
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("getIngresses for status update retried 3 times, aborting")
			return ingressMap
		}
	}

	if bulk {
		// Get IngressClasses with Avi set as the controller, get corresponding Ingresses,
		// to return all AKO ingestable Ingresses.
		aviIngClasses := make(map[string]bool)
		ingClassList, err := utils.GetInformers().IngressClassInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("Could not get the IngressClass object for UpdateStatus: %s", err)
			// retry get if request timeout or Unauthorized
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) || strings.Contains(err.Error(), utils.K8S_UNAUTHORIZED) {
				return getIngresses(ingressNSNames, bulk, retry+1)
			}
			return ingressMap
		}

		if len(ingClassList) == 0 {
			return ingressMap
		}

		for i := range ingClassList {
			if ingClassList[i].Spec.Controller == lib.SvcApiAviGatewayController {
				aviIngClasses[ingClassList[i].Name] = true
			}
		}

		ingressList, err := utils.GetInformers().IngressInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("Could not get the ingress object for UpdateStatus: %v", err)
			// retry get if request timeout or Unauthorized
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) || strings.Contains(err.Error(), utils.K8S_UNAUTHORIZED) {
				return getIngresses(ingressNSNames, bulk, retry+1)
			}
		}

		for i := range ingressList {
			var returnIng bool
			if ingressList[i].Spec.IngressClassName != nil {
				if _, ok := aviIngClasses[*ingressList[i].Spec.IngressClassName]; ok {
					returnIng = true
				}
			} else if _, ok := lib.IsAviLBDefaultIngressClassWithClient(); ok {
				returnIng = true
			}

			if returnIng {
				ing := ingressList[i].DeepCopy()
				if utils.CheckIfNamespaceAccepted(ing.Namespace) {
					ingressMap[ing.Namespace+"/"+ing.Name] = ing
				}
			}
		}

		return ingressMap
	}

	for _, namespaceName := range ingressNSNames {
		nsNameSplit := strings.Split(namespaceName, "/")

		mIngress, err := utils.GetInformers().ClientSet.NetworkingV1().Ingresses(nsNameSplit[0]).Get(context.TODO(), nsNameSplit[1], metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the ingress object for UpdateStatus: %v", err)
			// retry get if request timeout or Unauthorized
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) || strings.Contains(err.Error(), utils.K8S_UNAUTHORIZED) {
				return getIngresses(ingressNSNames, bulk, retry+1)
			}
			continue
		}

		ingressMap[mIngress.Namespace+"/"+mIngress.Name] = mIngress.DeepCopy()
	}

	return ingressMap
}

func (f *follower) UpdateIngressStatus(options []UpdateOptions, bulk bool) {
	for _, option := range options {
		utils.AviLog.Debugf("key: %s, AKO is not a leader, not updating the Ingress status", option.Key)
	}
}

func (f *follower) DeleteIngressStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not deleting the Ingress status", key)
	return nil
}
