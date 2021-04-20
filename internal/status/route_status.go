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

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

func ParseOptionsFromMetadata(options []UpdateOptions, bulk bool) ([]string, []UpdateOptions) {
	var objectsToUpdate []string
	var updateIngressOptions []UpdateOptions

	for _, option := range options {
		if len(option.ServiceMetadata.NamespaceIngressName) > 0 {
			// This is SNI with hostname sharding.
			for _, ingressns := range option.ServiceMetadata.NamespaceIngressName {
				ingressArr := strings.Split(ingressns, "/")
				if len(ingressArr) != 2 {
					utils.AviLog.Errorf("key: %s, msg: UpdateIngressStatus IngressNamespace format not correct", option.Key)
					continue
				}

				ingress := ingressArr[0] + "/" + ingressArr[1]
				option.IngSvc = ingress
				objectsToUpdate = append(objectsToUpdate, ingress)
				updateIngressOptions = append(updateIngressOptions, option)
			}
		} else {
			ingress := option.ServiceMetadata.Namespace + "/" + option.ServiceMetadata.IngressName
			option.IngSvc = ingress
			objectsToUpdate = append(objectsToUpdate, ingress)
			updateIngressOptions = append(updateIngressOptions, option)
		}
	}
	return objectsToUpdate, updateIngressOptions
}

// To Do: Check if it is possible to do update operations under same functions for both
// route and ingress, may be with a single interface with different implementations.
// Currently there are too many api calls, which are different for routes and ingresses,
// to have them under same function.

func UpdateRouteIngressStatus(options []UpdateOptions, bulk bool) {
	if utils.GetInformers().IngressInformer != nil {
		UpdateIngressStatus(options, bulk)
	} else if utils.GetInformers().RouteInformer != nil {
		UpdateRouteStatus(options, bulk)
	} else {
		utils.AviLog.Errorf("Status update failed, no suitable informers found")
	}
}

func DeleteRouteIngressStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	if utils.GetInformers().IngressInformer != nil {
		return DeleteIngressStatus(options, isVSDelete, key)
	} else if utils.GetInformers().RouteInformer != nil {
		return DeleteRouteStatus(options, isVSDelete, key)
	} else {
		utils.AviLog.Errorf("key: %s, msg: Status delete failed, no suitable informers found", key)
		return errors.New("Status delete failed, no suitable informers found")
	}
}

func UpdateRouteStatus(options []UpdateOptions, bulk bool) {
	var err error
	routesToUpdate, updateRouteOptions := ParseOptionsFromMetadata(options, bulk)

	// routeMap: {ns/Route: routeObj}
	// this pre-fetches all routes to be candidates for status update
	// after pre-fetching, if a status update comes for that route, then the pre-fetched route would be stale
	// in which case route will be fetched again in updateObject, as part of a retry
	routeMap := getRoutes(routesToUpdate, bulk)
	for _, option := range updateRouteOptions {
		if route := routeMap[option.IngSvc]; route != nil {
			if err = updateRouteObject(route, option); err != nil {
				utils.AviLog.Errorf("key: %s, msg: updating rorute object failed: %v", option.Key, err)
			}
			delete(routeMap, option.IngSvc)
		}
	}

	if bulk {
		for routeNSName, route := range routeMap {
			DeleteRouteStatus([]UpdateOptions{{
				ServiceMetadata: avicache.ServiceMetadataObj{
					NamespaceIngressName: []string{routeNSName},
					HostNames:            []string{route.Spec.Host},
				},
			}}, true, lib.SyncStatusKey)
		}
	}

	return
}

func getRoutes(routeNSNames []string, bulk bool, retryNum ...int) map[string]*routev1.Route {
	retry := 0
	routeMap := make(map[string]*routev1.Route)
	if len(retryNum) > 0 {
		utils.AviLog.Infof("Retrying to get the routes for status update")
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("getRoutes for status update retried 3 times, aborting")
			return routeMap
		}
	}

	if bulk {
		routeList, err := utils.GetInformers().OshiftClient.RouteV1().Routes(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			utils.AviLog.Warnf("Could not get the route object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getRoutes(routeNSNames, bulk, retry+1)
			}
		}
		for i := range routeList.Items {
			route := routeList.Items[i]
			routeMap[route.Namespace+"/"+route.Name] = &route
		}

		return routeMap
	}

	for _, namespaceName := range routeNSNames {
		nsNameSplit := strings.Split(namespaceName, "/")
		if len(nsNameSplit) != 2 {
			utils.AviLog.Warnf("msg: namespaceName %s has wrong format", namespaceName)
			continue
		}
		route, err := utils.GetInformers().OshiftClient.RouteV1().Routes(nsNameSplit[0]).Get(context.TODO(), nsNameSplit[1], metav1.GetOptions{})
		if err != nil {
			utils.AviLog.Warnf("msg: Could not get the route object for UpdateStatus: %s", err)
			// retry get if request timeout
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) {
				return getRoutes(routeNSNames, bulk, retry+1)
			}
			continue
		}
		routeMap[route.Namespace+"/"+route.Name] = route
	}

	return routeMap
}

func UpdateRouteStatusWithErrMsg(key, routeName, namespace, msg string, retryNum ...int) {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: UpdateRouteStatus retried 3 times, aborting", key)
			return
		}
	}

	mRoutes := getRoutes([]string{namespace + "/" + routeName}, false)
	if len(mRoutes) == 0 {
		return
	}
	mRoute := mRoutes[namespace+"/"+routeName]
	oldRouteStatus := mRoute.Status.DeepCopy()

	mRoute.Status.Ingress = []routev1.RouteIngress{}
	now := metav1.Now()
	condition := routev1.RouteIngressCondition{
		Status:             corev1.ConditionFalse,
		LastTransitionTime: &now,
		Reason:             msg,
		Type:               routev1.RouteAdmitted,
	}

	rtIngress := routev1.RouteIngress{
		Host:       mRoute.Spec.Host,
		RouterName: lib.AKOUser,
		Conditions: []routev1.RouteIngressCondition{
			condition,
		},
	}
	mRoute.Status.Ingress = append(mRoute.Status.Ingress, rtIngress)

	if sameStatus := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress); sameStatus {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in route status. old: %+v new: %+v",
			key, oldRouteStatus.Ingress, mRoute.Status.Ingress)
		return
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": mRoute.Status,
	})
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(mRoute.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in updating the route status: %v", key, err)
		// fetch updated route and feed for update status
		mRoutes := getRoutes([]string{mRoute.Namespace + "/" + mRoute.Name}, false)
		if len(mRoutes) > 0 {
			UpdateRouteStatusWithErrMsg(key, routeName, namespace, msg, retry+1)
		}
	}
	return
}

func routeStatusCheck(key string, oldStatus []routev1.RouteIngress, hostname string) bool {
	for _, status := range oldStatus {
		if len(status.Conditions) < 1 {
			continue
		}
		if status.Host == hostname && status.RouterName == lib.AKOUser {
			if status.Conditions[0].Status == corev1.ConditionFalse {
				utils.AviLog.Infof("key: %s, msg: current status of host %s is False", key, hostname)
				return false
			} else if status.Conditions[0].Status == corev1.ConditionTrue {
				return true
			}
		}
	}
	utils.AviLog.Infof("key: %s, msg: status not found for host %s", key, hostname)

	return false
}

func updateRouteObject(mRoute *routev1.Route, updateOption UpdateOptions, retryNum ...int) error {
	if updateOption.Vip == "" {
		return nil
	}

	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			return errors.New("UpdateRouteStatus retried 3 times, aborting")
		}
	}

	var err error
	hostnames, key := updateOption.ServiceMetadata.HostNames, updateOption.Key
	oldRouteStatus := mRoute.Status.DeepCopy()

	// If we find a hostname in the present update, let's first remove it from the existing status.
	for i := len(mRoute.Status.Ingress) - 1; i >= 0; i-- {
		if utils.HasElem(hostnames, mRoute.Status.Ingress[i].Host) {
			mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
		}
	}

	// Handle fresh hostname update
	if updateOption.Vip != "" {
		for _, host := range hostnames {
			now := metav1.Now()
			condition := routev1.RouteIngressCondition{
				Message:            updateOption.Vip,
				Status:             corev1.ConditionTrue,
				LastTransitionTime: &now,
				Type:               routev1.RouteAdmitted,
			}
			rtIngress := routev1.RouteIngress{
				Host:       host,
				RouterName: lib.AKOUser,
				Conditions: []routev1.RouteIngressCondition{
					condition,
				},
			}
			mRoute.Status.Ingress = append(mRoute.Status.Ingress, rtIngress)
		}
	}

	// remove the host from status which is not in spec
	for i := len(mRoute.Status.Ingress) - 1; i >= 0; i-- {
		if mRoute.Spec.Host != mRoute.Status.Ingress[i].Host {
			mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
		}
	}

	var updatedRoute *routev1.Route

	sameStatus := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress)
	if !sameStatus {
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": mRoute.Status,
		})

		updatedRoute, err = utils.GetInformers().OshiftClient.RouteV1().Routes(mRoute.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
		if err != nil {
			utils.AviLog.Errorf("key: %s, msg: there was an error in updating the route status: %v", key, err)
			// fetch updated route and feed for update status
			mRoutes := getRoutes([]string{mRoute.Namespace + "/" + mRoute.Name}, false)
			if len(mRoutes) > 0 {
				return updateRouteObject(mRoutes[mRoute.Namespace+"/"+mRoute.Name], updateOption, retry+1)
			}
		}

		utils.AviLog.Infof("key: %s, msg: Successfully updated the status of route: %s/%s old: %+v new: %+v",
			key, mRoute.Namespace, mRoute.Name, oldRouteStatus.Ingress, mRoute.Status.Ingress)
	} else {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in route status. old: %+v new: %+v",
			key, oldRouteStatus.Ingress, mRoute.Status.Ingress)
	}
	err = updateRouteAnnotations(updatedRoute, updateOption, mRoute, key, mRoute.Spec.Host)

	return err
}

func updateRouteAnnotations(mRoute *routev1.Route, updateOption UpdateOptions, oldRoute *routev1.Route, key, routeHost string, retryNum ...int) error {
	if mRoute == nil {
		mRoute = oldRoute
	}
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 3 {
			return errors.New("UpdateRouteStatus retried 3 times, aborting")
		}
	}

	vsAnnotations := make(map[string]string)
	if value, ok := mRoute.Annotations[VSAnnotation]; ok {
		if err := json.Unmarshal([]byte(value), &vsAnnotations); err != nil {
			utils.AviLog.Errorf("key: %s, msg: error in unmarshalling route annotations: %v", key, err)
			// nothing else to be done, invalid annotations will be taken care of in the update call
		}
	}

	// update the current hostname's VirtualService UUID
	for i := 0; i < len(updateOption.ServiceMetadata.HostNames); i++ {
		vsAnnotations[updateOption.ServiceMetadata.HostNames[i]] = updateOption.VirtualServiceUUID
	}

	// remove the non-spec hostnames
	for k := range vsAnnotations {
		if routeHost != k {
			delete(vsAnnotations, k)
		}
	}

	// compare the VirtualService annotations for this ingress object
	if req := isAnnotationsUpdateRequired(mRoute.Annotations, vsAnnotations); req {
		if err := patchRouteAnnotations(mRoute, vsAnnotations); err != nil && k8serrors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: error in updating the route annotations: %v", key, err)
			// fetch updated route and retry for updating annotations
			mRoutes := getRoutes([]string{mRoute.Namespace + "/" + mRoute.Name}, false)
			if len(mRoutes) > 0 {
				return updateRouteAnnotations(mRoute, updateOption, oldRoute, key, routeHost, retry+1)
			}
		}
		utils.AviLog.Infof("key: %s, msg: successfully updated route annotations: %s/%s, old: %+v, new: %+v",
			key, mRoute.Namespace, mRoute.Name, oldRoute.Annotations, mRoute.Annotations)
	} else {
		utils.AviLog.Debugf("No annotations update required for this route: %s/%s", mRoute.Namespace,
			mRoute.Name)
	}

	return nil
}

func patchRouteAnnotations(mRoute *routev1.Route, vsAnnotations map[string]string) error {
	patchPayloadBytes, err := getAnnotationsPayload(vsAnnotations, mRoute.GetAnnotations())
	if err != nil {
		return fmt.Errorf("error in generating payload for vs annotations %v: %v", vsAnnotations, err)
	}
	if _, err = utils.GetInformers().OshiftClient.RouteV1().Routes(mRoute.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayloadBytes, metav1.PatchOptions{}); err != nil {
		return err
	}

	return nil
}

func compareRouteStatus(oldStatus, newStatus []routev1.RouteIngress) bool {

	if len(oldStatus) != len(newStatus) {
		return false
	}
	exists := []string{}
	for _, status := range oldStatus {
		if len(status.Conditions) < 1 {
			continue
		}
		// For older created routes, time will be nil
		if status.Conditions[0].LastTransitionTime == nil {
			return false
		}
		ip := status.Conditions[0].Message
		reason := status.Conditions[0].Reason
		exists = append(exists, ip+":"+status.Host+":"+status.RouterName+":"+reason)
	}
	for _, status := range newStatus {
		if len(status.Conditions) < 1 {
			continue
		}
		ip := status.Conditions[0].Message
		reason := status.Conditions[0].Reason
		ipHost := ip + ":" + status.Host + ":" + status.RouterName + ":" + reason

		if !utils.HasElem(exists, ipHost) {
			return false
		}
	}

	return true
}

func DeleteRouteStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	if len(options) == 0 {
		return fmt.Errorf("Length of options is zero")
	}
	svc_mdata_obj := options[0].ServiceMetadata
	var err error
	if len(svc_mdata_obj.NamespaceIngressName) > 0 {
		// This is SNI with hostname sharding.
		for _, ingressns := range svc_mdata_obj.NamespaceIngressName {
			ingressArr := strings.Split(ingressns, "/")
			if len(ingressArr) != 2 {
				utils.AviLog.Errorf("key: %s, msg: DeleteRouteStatus IngressNamespace format not correct", key)
				return errors.New("DeleteRouteStatus IngressNamespace format not correct")
			}
			svc_mdata_obj.Namespace = ingressArr[0]
			svc_mdata_obj.IngressName = ingressArr[1]
			err = deleteRouteObject(svc_mdata_obj, key, isVSDelete)
		}
	} else {
		err = deleteRouteObject(svc_mdata_obj, key, isVSDelete)
	}

	if err != nil {
		return err
	}

	return nil
}

func deleteRouteObject(svc_mdata_obj avicache.ServiceMetadataObj, key string, isVSDelete bool, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: Retrying to update the route status", key)
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: DeleteRouteStatus retried 3 times, aborting", key)
			return errors.New("DeleteRouteStatus retried 3 times, aborting")
		}
	}

	mRoute, err := utils.GetInformers().RouteInformer.Lister().Routes(svc_mdata_obj.Namespace).Get(svc_mdata_obj.IngressName)

	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not get the ingress object for DeleteStatus: %s", key, err)
		return err
	}

	oldRouteStatus := mRoute.Status.DeepCopy()
	if len(svc_mdata_obj.HostNames) > 0 {
		// If the route status for the host is alresay fasle, then don't delete the status
		if !routeStatusCheck(key, oldRouteStatus.Ingress, svc_mdata_obj.HostNames[0]) {
			return nil
		}
	}

	for i := len(mRoute.Status.Ingress) - 1; i >= 0; i-- {
		for _, host := range svc_mdata_obj.HostNames {
			if mRoute.Status.Ingress[i].Host != host {
				continue
			}
			// Check if this host is still present in the spec, if so - don't delete it
			//NS migration case: if false -> ns invalid event happened so remove status
			if mRoute.Spec.Host != host || isVSDelete || !utils.CheckIfNamespaceAccepted(svc_mdata_obj.Namespace) {
				mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
			} else {
				utils.AviLog.Debugf("key: %s, msg: skipping status update since host is present in the route: %v", key, host)
			}
		}
	}

	var updatedRoute *routev1.Route
	sameStatus := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress)
	if sameStatus {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in ingress status. old: %+v new: %+v",
			key, oldRouteStatus.Ingress, mRoute.Status.Ingress)
	} else {
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": mRoute.Status,
		})
		updatedRoute, err = utils.GetInformers().OshiftClient.RouteV1().Routes(svc_mdata_obj.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
		if err != nil {
			utils.AviLog.Errorf("key: %s, msg: there was an error in deleting the ingress status: %v", key, err)
			return deleteObject(svc_mdata_obj, key, isVSDelete, retry+1)
		}

		utils.AviLog.Infof("key: %s, msg: Successfully deleted status of route: %s/%s old: %+v new: %+v",
			key, mRoute.Namespace, mRoute.Name, oldRouteStatus.Ingress, mRoute.Status.Ingress)
	}

	return deleteRouteAnnotation(updatedRoute, svc_mdata_obj, isVSDelete, mRoute.Spec.Host, key, mRoute)
}

func deleteRouteAnnotation(routeObj *routev1.Route, svcMeta avicache.ServiceMetadataObj, isVSDelete bool,
	routeHost string, key string, oldRoute *routev1.Route, retryNum ...int) error {
	if routeObj == nil {
		routeObj = oldRoute
	}
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: retrying to update route annotations", key)
		retry = retryNum[0]
		if retry >= 3 {
			return fmt.Errorf("deleteRouteAnnotation retried 3 times, aborting")
		}
	}
	existingAnnotations := make(map[string]string)
	if annotations, exists := routeObj.Annotations[VSAnnotation]; exists {
		if err := json.Unmarshal([]byte(annotations), &existingAnnotations); err != nil {
			return fmt.Errorf("error in unmarshalling annotations %s, %v", annotations, err)
		}
	} else {
		return fmt.Errorf("error in fetching VS annotations %v", routeObj.Annotations)
	}

	for k := range existingAnnotations {
		for _, host := range svcMeta.HostNames {
			if k == host {
				// Check if:
				// 1. this host is still present in the spec, if so - don't delete it from annotations
				// 2. in case of NS migration, if NS is moved from selected to rejected, this host then
				//    has to be removed from the annotations list.
				nsMigrationFilterFlag := utils.CheckIfNamespaceAccepted(svcMeta.Namespace)
				if routeHost != host || isVSDelete || !nsMigrationFilterFlag {
					delete(existingAnnotations, k)
				} else {
					utils.AviLog.Debugf("key: %s, msg: skipping annotation update since host is present in the route: %v", key, host)
				}
			}
		}
	}

	if isAnnotationsUpdateRequired(routeObj.Annotations, existingAnnotations) {
		if err := patchRouteAnnotations(routeObj, existingAnnotations); err != nil && k8serrors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: error in updating route annotations: %v, will retry", err)
			return deleteRouteAnnotation(routeObj, svcMeta, isVSDelete, routeHost, key, oldRoute, retry+1)
		}
		utils.AviLog.Debugf("key: %s, msg: annotations updated for route", key)
	}

	return nil
}
