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
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	types "k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
)

func ParseOptionsFromMetadata(options []UpdateOptions, bulk bool) ([]string, map[string]UpdateOptions) {
	// updateIngressOptions holds as its key ingressNS/ingressName/vsIP and tries aggregating multiple hostnames
	// from various service metadatas. This ensures that a particular ingress hosting a particular IP,
	// can possibly hold multiple hostnames.
	updateIngressOptions := make(map[string]UpdateOptions)

	for _, option := range options {
		if option.ServiceMetadata.InsecureEdgeTermAllow {
			utils.AviLog.Infof("Skipping update of parent VS annotation since the route :%v has InsecureEdgeTerminationAllow set to true", option.ServiceMetadata.IngressName)
			continue
		}
		if len(option.ServiceMetadata.NamespaceIngressName) > 0 {
			// secure VSes, service metadata comes from SNI VS.
			for _, ingressns := range option.ServiceMetadata.NamespaceIngressName {
				ingressArr := strings.Split(ingressns, "/")
				if len(ingressArr) != 2 {
					utils.AviLog.Errorf("key: %s, msg: UpdateIngressStatus IngressNamespace format not correct", option.Key)
					continue
				}
				for _, vip := range option.Vip {
					ingressIPKey := ingressns + "/" + vip
					option.IngSvc = ingressns
					if opt, ok := updateIngressOptions[ingressIPKey]; ok {
						for _, hostname := range option.ServiceMetadata.HostNames {
							if !utils.HasElem(opt.ServiceMetadata.HostNames, hostname) {
								opt.ServiceMetadata.HostNames = append(opt.ServiceMetadata.HostNames, option.ServiceMetadata.HostNames...)
								updateIngressOptions[ingressIPKey] = opt
							}
						}
					} else {
						updateIngressOptions[ingressIPKey] = option
					}
				}
			}
		} else {
			// insecure VSes, servicemetadata comes from Pools.
			ingress := option.ServiceMetadata.Namespace + "/" + option.ServiceMetadata.IngressName
			for _, vip := range option.Vip {
				ingressIPKey := ingress + "/" + vip
				option.IngSvc = ingress
				if opt, ok := updateIngressOptions[ingressIPKey]; ok {
					for _, hostname := range option.ServiceMetadata.HostNames {
						if !utils.HasElem(opt.ServiceMetadata.HostNames, hostname) {
							opt.ServiceMetadata.HostNames = append(opt.ServiceMetadata.HostNames, option.ServiceMetadata.HostNames...)
							updateIngressOptions[ingressIPKey] = opt
						}
					}
				} else {
					updateIngressOptions[ingressIPKey] = option
				}
			}
		}
	}

	ingressesToUpdate := make([]string, len(updateIngressOptions))
	i := 0
	for k := range updateIngressOptions {
		kSlice := strings.Split(k, "/")
		ingressesToUpdate[i] = kSlice[0] + "/" + kSlice[1]
		i++
	}
	return ingressesToUpdate, updateIngressOptions
}

// To Do: Check if it is possible to do update operations under same functions for both
// route and ingress, may be with a single interface with different implementations.
// Currently there are too many api calls, which are different for routes and ingresses,
// to have them under same function.

func (l *leader) UpdateRouteIngressStatus(options []UpdateOptions, bulk bool) {
	if utils.GetInformers().IngressInformer != nil {
		l.UpdateIngressStatus(options, bulk)
	} else if utils.GetInformers().RouteInformer != nil {
		l.UpdateRouteStatus(options, bulk)
	} else {
		utils.AviLog.Errorf("Status update failed, no suitable informers found")
	}
}

func (l *leader) DeleteRouteIngressStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	if utils.GetInformers().IngressInformer != nil {
		return l.DeleteIngressStatus(options, isVSDelete, key)
	} else if utils.GetInformers().RouteInformer != nil {
		return l.DeleteRouteStatus(options, isVSDelete, key)
	} else {
		utils.AviLog.Errorf("key: %s, msg: Status delete failed, no suitable informers found", key)
		return errors.New("Status delete failed, no suitable informers found")
	}
}

func (l *leader) UpdateRouteStatus(options []UpdateOptions, bulk bool) {
	var err error
	routesToUpdate, updateRouteOptions := ParseOptionsFromMetadata(options, bulk)

	// routeMap: {ns/Route: routeObj}
	// this pre-fetches all routes to be candidates for status update
	// after pre-fetching, if a status update comes for that route, then the pre-fetched route would be stale
	// in which case route will be fetched again in updateObject, as part of a retry
	routeMap := getRoutes(routesToUpdate, bulk)
	skipDelete := map[string]bool{}
	for _, option := range updateRouteOptions {
		if route := routeMap[option.IngSvc]; route != nil {
			if err = updateRouteObject(route, option); err != nil {
				utils.AviLog.Errorf("key: %s, msg: updating rorute object failed: %v", option.Key, err)
			}
			skipDelete[option.IngSvc] = true
		}
	}
	if bulk {
		for routeNSName, route := range routeMap {
			if val, ok := skipDelete[routeNSName]; ok && val {
				continue
			}
			l.DeleteRouteStatus([]UpdateOptions{{
				ServiceMetadata: lib.ServiceMetadataObj{
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
		routeList, err := utils.GetInformers().RouteInformer.Lister().List(labels.Set(nil).AsSelector())
		if err != nil {
			utils.AviLog.Warnf("Could not get the route object for UpdateStatus: %s", err)
			// retry get if request timeout or Unauthorized
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) || strings.Contains(err.Error(), utils.K8S_UNAUTHORIZED) {
				return getRoutes(routeNSNames, bulk, retry+1)
			}
		}
		for i := range routeList {
			route := routeList[i]
			if utils.CheckIfNamespaceAccepted(route.Namespace) {
				routeMap[route.Namespace+"/"+route.Name] = route.DeepCopy()
			}
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
			// retry get if request timeout or Unauthorized
			if strings.Contains(err.Error(), utils.K8S_ETIMEDOUT) || strings.Contains(err.Error(), utils.K8S_UNAUTHORIZED) {
				return getRoutes(routeNSNames, bulk, retry+1)
			}
			continue
		}
		routeMap[route.Namespace+"/"+route.Name] = route.DeepCopy()
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

	routeHost := mRoute.Spec.Host
	if routeHost == "" {
		routeHost = lib.GetHostnameforSubdomain(mRoute.Spec.Subdomain)
	}

	rtIngress := routev1.RouteIngress{
		Host:       routeHost,
		RouterName: lib.AKOUser,
		Conditions: []routev1.RouteIngressCondition{
			condition,
		},
	}
	mRoute.Status.Ingress = append(mRoute.Status.Ingress, rtIngress)

	if sameStatus, _, _ := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress); sameStatus {
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

func routeVsUUIDStatus(key, hostname, namespace string, updateOption UpdateOptions) string {
	vsAnnotations := make(map[string]string)
	ctrlAnnotationValStr := avicache.GetControllerClusterUUID()
	for i := 0; i < len(updateOption.ServiceMetadata.HostNames); i++ {
		// only update for given hostname
		if updateOption.ServiceMetadata.HostNames[i] == hostname {
			vsAnnotations[hostname] = updateOption.VirtualServiceUUID
		}
	}
	vsAnnotationsBytes, err := json.Marshal(vsAnnotations)
	if err != nil {
		utils.AviLog.Errorf("error in marshalling vs annotations: %v", err)
		return ""
	}
	vsAnnotationsStrStr := string(vsAnnotationsBytes)
	patchPayload := map[string]string{
		lib.VSAnnotation:         vsAnnotationsStrStr,
		lib.ControllerAnnotation: ctrlAnnotationValStr,
		lib.TenantAnnotation:     updateOption.Tenant,
	}
	return utils.Stringify(patchPayload)
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
	if len(updateOption.Vip) == 0 {
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
		if utils.HasElem(hostnames, mRoute.Status.Ingress[i].Host) && mRoute.Status.Ingress[i].RouterName == lib.AKOUser {
			mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
		}
	}

	// Handle fresh hostname update
	for _, host := range hostnames {
		now := metav1.Now()
		for _, vip := range updateOption.Vip {
			// In 1.12.1, populate both reason and annotation fields.
			//So that during upgrade there will not be any issue of GSLB pools going down.
			reason := routeVsUUIDStatus(key, host, mRoute.Namespace, updateOption)
			condition := routev1.RouteIngressCondition{
				Message:            vip,
				Status:             corev1.ConditionTrue,
				LastTransitionTime: &now,
				Type:               routev1.RouteAdmitted,
				Reason:             reason,
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
	routeHost := mRoute.Spec.Host
	if routeHost == "" {
		routeHost = lib.GetHostnameforSubdomain(mRoute.Spec.Subdomain)
	}
	for i := len(mRoute.Status.Ingress) - 1; i >= 0; i-- {
		if mRoute.Status.Ingress[i].RouterName == lib.AKOUser && routeHost != mRoute.Status.Ingress[i].Host {
			mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
		}
	}

	var updatedRoute *routev1.Route

	sameStatus, beforeHost, afterHost := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress)
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
		} else {
			// The UTs discovered that the route might have gotten deleted just before the PATCH status call to the route.
			// In that case the updatedRoute would turn out to be empty. This else conditional checks saves AKO from
			// the a crash in that case.
			if afterHost != "" {
				lib.AKOControlConfig().EventRecorder().Eventf(updatedRoute, corev1.EventTypeNormal, lib.Synced, "Added virtualservice %s for %s", updateOption.VSName, afterHost)
			} else if beforeHost != "" {
				lib.AKOControlConfig().EventRecorder().Eventf(updatedRoute, corev1.EventTypeNormal, lib.Removed, "Removed virtualservice for %s", beforeHost)
			}
			utils.AviLog.Infof("key: %s, msg: Successfully updated the status of route: %s/%s old: %+v new: %+v",
				key, mRoute.Namespace, mRoute.Name, oldRouteStatus.Ingress, mRoute.Status.Ingress)
		}
	} else {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in route status. old: %+v new: %+v",
			key, oldRouteStatus.Ingress, mRoute.Status.Ingress)
	}
	err = updateRouteAnnotations(updatedRoute, updateOption, mRoute, key, routeHost)

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
	if value, ok := mRoute.Annotations[lib.VSAnnotation]; ok {
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
	// compare the VirtualService annotations for this Route object
	if req := isAnnotationsUpdateRequired(mRoute.Annotations, vsAnnotations, updateOption.Tenant, false); req {
		if err := patchRouteAnnotations(mRoute, vsAnnotations, updateOption.Tenant); err != nil && k8serrors.IsNotFound(err) {
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

func patchRouteAnnotations(mRoute *routev1.Route, vsAnnotations map[string]string, tenant string) error {
	patchPayloadBytes, err := getAnnotationsPayload(vsAnnotations, tenant)
	if err != nil {
		return fmt.Errorf("error in generating payload for vs annotations %v: %v", vsAnnotations, err)
	}
	if _, err = utils.GetInformers().OshiftClient.RouteV1().Routes(mRoute.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayloadBytes, metav1.PatchOptions{}); err != nil {
		return err
	}

	return nil
}

func compareRouteStatus(oldStatus, newStatus []routev1.RouteIngress) (bool, string, string) {
	exists := sets.NewString()
	// Route would essentially consist of single hosts.
	var beforeHost, afterHost string
	var diff *bool

	for _, status := range oldStatus {
		if status.RouterName != lib.AKOUser {
			continue
		}
		if len(status.Conditions) < 1 {
			continue
		}
		// For older created routes, time will be nil
		if status.Conditions[0].LastTransitionTime == nil {
			if diff == nil {
				diff = proto.Bool(false)
			}
			beforeHost = status.Host
			continue
		}
		ip := status.Conditions[0].Message
		reason := status.Conditions[0].Reason
		exists.Insert(ip + ":" + status.Host + ":" + status.RouterName + ":" + reason)
		beforeHost = status.Host
		break
	}

	for _, status := range newStatus {
		if status.RouterName != lib.AKOUser {
			continue
		}
		if len(status.Conditions) < 1 {
			continue
		}
		ip := status.Conditions[0].Message
		reason := status.Conditions[0].Reason
		if !exists.Has(ip + ":" + status.Host + ":" + status.RouterName + ":" + reason) {
			if diff == nil {
				diff = proto.Bool(false)
				afterHost = status.Host
			}
			continue
		}
		afterHost = status.Host
		break
	}

	if len(oldStatus) != len(newStatus) {
		diff = proto.Bool(false)
	}

	if diff == nil {
		diff = proto.Bool(true)
	}
	return *diff, beforeHost, afterHost
}

func (l *leader) DeleteRouteStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	if len(options) == 0 {
		return fmt.Errorf("Length of options is zero")
	}
	var err error
	if len(options[0].ServiceMetadata.NamespaceIngressName) > 0 {
		// This is SNI with hostname sharding.
		for _, ingressns := range options[0].ServiceMetadata.NamespaceIngressName {
			ingressArr := strings.Split(ingressns, "/")
			if len(ingressArr) != 2 {
				utils.AviLog.Errorf("key: %s, msg: DeleteRouteStatus IngressNamespace format not correct", key)
				return errors.New("DeleteRouteStatus IngressNamespace format not correct")
			}
			options[0].ServiceMetadata.Namespace = ingressArr[0]
			options[0].ServiceMetadata.IngressName = ingressArr[1]
			err = deleteRouteObject(options[0], key, isVSDelete)
		}
	} else {
		err = deleteRouteObject(options[0], key, isVSDelete)
	}

	if err != nil {
		return err
	}

	return nil
}

func deleteRouteObject(option UpdateOptions, key string, isVSDelete bool, retryNum ...int) error {
	retry := 0
	if len(retryNum) > 0 {
		utils.AviLog.Infof("key: %s, msg: Retrying to update the route status", key)
		retry = retryNum[0]
		if retry >= 3 {
			utils.AviLog.Errorf("key: %s, msg: DeleteRouteStatus retried 3 times, aborting", key)
			return errors.New("DeleteRouteStatus retried 3 times, aborting")
		}
	}

	mRoute, err := utils.GetInformers().RouteInformer.Lister().Routes(option.ServiceMetadata.Namespace).Get(option.ServiceMetadata.IngressName)
	if err != nil {
		utils.AviLog.Warnf("key: %s, msg: Could not get the Route object for DeleteStatus: %s", key, err)
		return err
	}
	mRoute = mRoute.DeepCopy()

	oldRouteStatus := mRoute.Status.DeepCopy()
	if len(option.ServiceMetadata.HostNames) > 0 {
		// If the route status for the host is already false, then don't delete the status
		if !routeStatusCheck(key, oldRouteStatus.Ingress, option.ServiceMetadata.HostNames[0]) {
			return nil
		}
	}

	utils.AviLog.Infof("key: %s, deleting hostnames %v from Route status %s/%s", key, option.ServiceMetadata.HostNames, option.ServiceMetadata.Namespace, option.ServiceMetadata.IngressName)
	svcMdataHostname := option.ServiceMetadata.HostNames[0]
	routeHost := mRoute.Spec.Host
	if routeHost == "" {
		routeHost = lib.GetHostnameforSubdomain(mRoute.Spec.Subdomain)
	}
	for i := len(mRoute.Status.Ingress) - 1; i >= 0; i-- {
		if mRoute.Status.Ingress[i].Host != svcMdataHostname {
			continue
		}
		// Check if this host is still present in the spec, if so - don't delete it
		// NS migration case: if false -> ns invalid event happened so remove status
		if mRoute.Status.Ingress[i].RouterName == lib.AKOUser && (routeHost != svcMdataHostname || isVSDelete || !utils.CheckIfNamespaceAccepted(option.ServiceMetadata.Namespace)) {
			mRoute.Status.Ingress = append(mRoute.Status.Ingress[:i], mRoute.Status.Ingress[i+1:]...)
		} else {
			utils.AviLog.Debugf("key: %s, msg: skipping status update since host is present in the route: %v", key, svcMdataHostname)
		}
	}

	var updatedRoute *routev1.Route
	sameStatus, _, afterHost := compareRouteStatus(oldRouteStatus.Ingress, mRoute.Status.Ingress)
	if sameStatus {
		utils.AviLog.Debugf("key: %s, msg: No changes detected in Route status. old: %+v new: %+v",
			key, oldRouteStatus.Ingress, mRoute.Status.Ingress)
	} else {
		patchPayload, _ := json.Marshal(map[string]interface{}{
			"status": mRoute.Status,
		})
		if len(mRoute.Status.Ingress) == 0 {
			patchPayload, _ = json.Marshal(map[string]interface{}{
				"status": nil,
			})
		}
		updatedRoute, err = utils.GetInformers().OshiftClient.RouteV1().Routes(option.ServiceMetadata.Namespace).Patch(context.TODO(), mRoute.Name, types.MergePatchType, patchPayload, metav1.PatchOptions{}, "status")
		if err != nil {
			utils.AviLog.Errorf("key: %s, msg: there was an error in deleting the Route status: %v", key, err)
			return deleteRouteObject(option, key, isVSDelete, retry+1)
		} else {
			if afterHost == "" {
				lib.AKOControlConfig().EventRecorder().Eventf(updatedRoute, corev1.EventTypeNormal, lib.Removed, "Removed virtualservice for %s", afterHost)
			}
			utils.AviLog.Infof("key: %s, msg: Successfully deleted status of route: %s/%s old: %+v new: %+v",
				key, mRoute.Namespace, mRoute.Name, oldRouteStatus.Ingress, mRoute.Status.Ingress)
		}
	}

	return deleteRouteAnnotation(updatedRoute, option.ServiceMetadata, isVSDelete, routeHost, key, option.Tenant, mRoute)
}

func deleteRouteAnnotation(routeObj *routev1.Route, svcMeta lib.ServiceMetadataObj, isVSDelete bool,
	routeHost string, key, tenant string, oldRoute *routev1.Route, retryNum ...int) error {
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
	if annotations, exists := routeObj.Annotations[lib.VSAnnotation]; exists {
		if err := json.Unmarshal([]byte(annotations), &existingAnnotations); err != nil {
			return fmt.Errorf("error in unmarshalling annotations %s, %v", annotations, err)
		}
	} else {
		utils.AviLog.Debugf("VS annotations not found for route %s/%s", routeObj.Namespace, routeObj.Name)
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
	if isAnnotationsUpdateRequired(routeObj.Annotations, existingAnnotations, tenant, isVSDelete) {
		if err := patchRouteAnnotations(routeObj, existingAnnotations, tenant); err != nil && k8serrors.IsNotFound(err) {
			utils.AviLog.Errorf("key: %s, msg: error in updating route annotations: %v, will retry", err)
			return deleteRouteAnnotation(routeObj, svcMeta, isVSDelete, routeHost, key, tenant, oldRoute, retry+1)
		}
		utils.AviLog.Debugf("key: %s, msg: annotations updated for route", key)
	}

	return nil
}

func (f *follower) UpdateRouteStatus(options []UpdateOptions, bulk bool) {
	for _, option := range options {
		utils.AviLog.Debugf("key: %s, AKO is not a leader, not updating the Route status", option.Key)
	}
}

func (f *follower) DeleteRouteStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not deleting the Route status", key)
	return nil
}

func (f *follower) UpdateRouteIngressStatus(options []UpdateOptions, bulk bool) {
	for _, option := range options {
		utils.AviLog.Debugf("key: %s, AKO is not a leader, not updating the Route status", option.Key)
	}
}

func (f *follower) DeleteRouteIngressStatus(options []UpdateOptions, isVSDelete bool, key string) error {
	utils.AviLog.Debugf("key: %s, AKO is not a leader, not deleting the Route status", key)
	return nil
}
