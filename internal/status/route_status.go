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
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
)

func ParseOptionsFromMetadata(options []UpdateStatusOptions, bulk bool) ([]string, []UpdateStatusOptions) {
	var objectsToUpdate []string
	var updateIngressOptions []UpdateStatusOptions

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

func UpdateRouteIngressStatus(options []UpdateStatusOptions, bulk bool) {
	if utils.GetInformers().IngressInformer != nil {
		UpdateIngressStatus(options, bulk)
	} else if utils.GetInformers().RouteInformer != nil {
		UpdateRouteStatus(options, bulk)
	} else {
		utils.AviLog.Errorf("Status update failed, no suitable informers found")
	}
}

func DeleteRouteIngressStatus(svc_mdata_obj avicache.ServiceMetadataObj, isVSDelete bool, key string) error {
	if utils.GetInformers().IngressInformer != nil {
		return DeleteIngressStatus(svc_mdata_obj, isVSDelete, key)
	} else if utils.GetInformers().RouteInformer != nil {
		return DeleteRouteStatus(svc_mdata_obj, isVSDelete, key)
	} else {
		utils.AviLog.Errorf("key: %s, msg: Status delete failed, no suitable informers found", key)
		return errors.New("Status delete failed, no suitable informers found")
	}
}

func UpdateRouteStatus(options []UpdateStatusOptions, bulk bool) {
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
				utils.AviLog.Error(err)
			}
		}
	}

	return
}

func getRoutes(routeNSNames []string, bulk bool, retryNum ...int) map[string]*routev1.Route {
	retry := 0
	routeMap := make(map[string]*routev1.Route)
	if len(retryNum) > 0 {
		utils.AviLog.Infof("msg: Retrying to get the routes for status update")
		retry = retryNum[0]
		if retry >= 2 {
			utils.AviLog.Errorf("msg: getRoutes for status update retried 3 times, aborting")
			return routeMap
		}
	}

	if bulk {
		routeList, err := utils.GetInformers().OshiftClient.RouteV1().Routes("").List(context.TODO(), metav1.ListOptions{})
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
	utils.AviLog.Infof("routeNSNames: %v", routeNSNames)
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

func UpdateRouteStatusWithErrMsg(routeName, namespace, msg string, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 2 {
			return errors.New("msg: UpdateRouteStatus retried 3 times, aborting")
		}
	}

	mRoutes := getRoutes([]string{namespace + "/" + routeName}, false)
	if len(mRoutes) == 0 {
		return nil
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
		utils.AviLog.Debugf("msg: No changes detected in route status. old: %+v new: %+v",
			oldRouteStatus.Ingress, mRoute.Status.Ingress)
		return nil
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": mRoute.Status,
	})
	_, err := utils.GetInformers().OshiftClient.RouteV1().Routes(mRoute.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("msg: there was an error in updating the route status: %v", err)
		// fetch updated route and feed for update status
		mRoutes := getRoutes([]string{mRoute.Namespace + "/" + mRoute.Name}, false)
		if len(mRoutes) > 0 {
			return UpdateRouteStatusWithErrMsg(routeName, namespace, msg, retry+1)
		}
	}
	return err
}

func routeStatusCheck(oldStatus []routev1.RouteIngress, hostname string) bool {
	for _, status := range oldStatus {
		if len(status.Conditions) < 1 {
			continue
		}
		if status.Host == hostname && status.RouterName == lib.AKOUser {
			if status.Conditions[0].Status == corev1.ConditionFalse {
				utils.AviLog.Infof("current status of host %s is False", hostname)
				return false
			} else if status.Conditions[0].Status == corev1.ConditionTrue {
				return true
			}
		}
	}
	utils.AviLog.Infof("status not found for host %s", hostname)

	return false
}

func updateRouteObject(mRoute *routev1.Route, updateOption UpdateStatusOptions, retryNum ...int) error {
	if updateOption.Vip == "" {
		return nil
	}
	retry := 0
	if len(retryNum) > 0 {
		retry = retryNum[0]
		if retry >= 2 {
			return errors.New("key: %s, msg: UpdateRouteStatus retried 3 times, aborting")
		}
	}

	var err error
	utils.AviLog.Infof("updateOption: %v", updateOption)
	hostnames, key := updateOption.ServiceMetadata.HostNames, updateOption.Key
	oldRouteStatus := mRoute.Status.DeepCopy()

	// Clean up all hosts that are not part of the route spec.
	var hostListIng []string
	hostListIng = append(hostListIng, mRoute.Spec.Host)

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
		if !utils.HasElem(hostListIng, mRoute.Status.Ingress[i].Host) {
			mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
		}
	}

	if sameStatus := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress); sameStatus {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in route status. old: %+v new: %+v",
			key, oldRouteStatus.Ingress, mRoute.Status.Ingress)
		return nil
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": mRoute.Status,
	})
	_, err = utils.GetInformers().OshiftClient.RouteV1().Routes(mRoute.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
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
	return err
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

func DeleteRouteStatus(svc_mdata_obj avicache.ServiceMetadataObj, isVSDelete bool, key string) error {
	var err error
	if len(svc_mdata_obj.NamespaceIngressName) > 0 {
		// This is SNI with hostname sharding.
		for _, ingressns := range svc_mdata_obj.NamespaceIngressName {
			ingressArr := strings.Split(ingressns, "/")
			if len(ingressArr) != 2 {
				return errors.New("key: %s, msg: DeleteRouteStatus IngressNamespace format not correct")
			}
			svc_mdata_obj.Namespace = ingressArr[0]
			svc_mdata_obj.IngressName = ingressArr[1]
			err = deleteRouteObject(svc_mdata_obj, key, isVSDelete)
		}
	} else {
		err = deleteRouteObject(svc_mdata_obj, key, isVSDelete)
	}

	if err != nil {
		utils.AviLog.Warn(err)
	}

	return nil
}

func deleteRouteObject(svc_mdata_obj avicache.ServiceMetadataObj, key string, isVSDelete bool, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: Retrying to update the route status", key)
		retry = retryNum[0]
		if retry >= 2 {
			return errors.New("key: %s, msg: DeleteRouteStatus retried 3 times, aborting")
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
		if !routeStatusCheck(oldRouteStatus.Ingress, svc_mdata_obj.HostNames[0]) {
			return nil
		}
	}
	var hostListIng []string
	hostListIng = append(hostListIng, mRoute.Spec.Host)

	for i := len(mRoute.Status.Ingress) - 1; i >= 0; i-- {
		for _, host := range svc_mdata_obj.HostNames {
			if mRoute.Status.Ingress[i].Host == host {
				// Check if this host is still present in the spec, if so - don't delete it
				//NS migration case: if false -> ns invalid event happend so remove status
				nsMigrationFilterFlag := utils.CheckIfNamespaceAccepted(svc_mdata_obj.Namespace, utils.GetGlobalNSFilter(), nil, true)
				if !utils.HasElem(hostListIng, host) || isVSDelete || !nsMigrationFilterFlag {
					mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
				} else {
					utils.AviLog.Debugf("key: %s, msg: skipping status update since host is present in the route: %v", key, host)
				}
			}
		}
	}

	if sameStatus := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress); sameStatus {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in ingress status. old: %+v new: %+v",
			key, oldRouteStatus.Ingress, mRoute.Status.Ingress)
		return nil
	}

	patchPayload, _ := json.Marshal(map[string]interface{}{
		"status": mRoute.Status,
	})
	_, err = utils.GetInformers().OshiftClient.RouteV1().Routes(svc_mdata_obj.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
	if err != nil {
		utils.AviLog.Errorf("key: %s, msg: there was an error in deleting the ingress status: %v", key, err)
		return deleteObject(svc_mdata_obj, key, isVSDelete, retry+1)
	}

	utils.AviLog.Infof("key: %s, msg: Successfully deleted status of route: %s/%s old: %+v new: %+v",
		key, mRoute.Namespace, mRoute.Name, oldRouteStatus.Ingress, mRoute.Status.Ingress)
	return nil
}
