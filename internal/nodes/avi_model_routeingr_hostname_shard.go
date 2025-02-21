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

package nodes

import (
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"
	akov1beta1 "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/apis/ako/v1beta1"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// HostNameShardAndPublish: Create model from supported objects - route/ingress, and publish to rest layer
func HostNameShardAndPublish(objType, objname, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	utils.AviLog.Infof("key: %s, starting HostNameShardAndPublish", key)
	var routeIgrObj RouteIngressModel
	var err error
	var processObj bool

	switch objType {
	case utils.Ingress:
		if utils.GetInformers().IngressInformer == nil {
			return
		}
		routeIgrObj, err, processObj = GetK8sIngressModel(objname, namespace, key)
	case utils.OshiftRoute:
		if utils.GetInformers().RouteInformer == nil {
			return
		}
		routeIgrObj, err, processObj = GetOshiftRouteModel(objname, namespace, key)
	case lib.MultiClusterIngress:
		if utils.GetInformers().MultiClusterIngressInformer == nil {
			utils.AviLog.Warnf("key: %s, multi-cluster informer is not initialized for object type: %s", key, objType)
			return
		}
		routeIgrObj, err, processObj = GetMultiClusterIngressModel(objname, namespace, key)
	default:
		utils.AviLog.Infof("key: %s, starting unsupported object type: %s", key, objType)
		return
	}

	defer func(routeIgrObj RouteIngressModel) {
		if aviInfraSetting := routeIgrObj.GetAviInfraSetting(); aviInfraSetting != nil {
			var shardSize string
			if aviInfraSetting.Spec.L7Settings != (akov1beta1.AviInfraL7Settings{}) {
				shardSize = aviInfraSetting.Spec.L7Settings.ShardSize
			}
			objects.InfraSettingL7Lister().UpdateIngRouteInfraSettingMappings(namespace+"/"+objname, aviInfraSetting.Name, shardSize)
		} else {
			objects.InfraSettingL7Lister().RemoveIngRouteInfraSettingMappings(namespace + "/" + objname)
		}
		objects.SharedNamespaceTenantLister().UpdateNamespacedResourceToTenantStore(namespace+"/"+objname, lib.GetTenantInNamespace(namespace))
	}(routeIgrObj)

	// delete old Models in case the modelNames changes because of shardSize updates via AviInfraSetting
	if lib.IsEvhEnabled() {
		DeleteStaleDataForModelChangeForEvh(routeIgrObj, namespace, objname, key, fullsync, sharedQueue)
	} else {
		DeleteStaleDataForModelChange(routeIgrObj, namespace, objname, key, fullsync, sharedQueue)
	}

	if err != nil || !processObj {
		utils.AviLog.Warnf("key: %s, msg: Error %v", key, err)
		// Detect a delete condition here.
		if k8serrors.IsNotFound(err) || !processObj {
			utils.AviLog.Infof("key: %s, Deleting Pool for ingress delete", key)
			if lib.IsEvhEnabled() {
				RouteIngrDeletePoolsByHostnameForEvh(routeIgrObj, namespace, objname, key, fullsync, sharedQueue)
			} else {
				RouteIngrDeletePoolsByHostname(routeIgrObj, namespace, objname, key, fullsync, sharedQueue)
			}
		}
		return
	}

	utils.AviLog.Infof("key: %s, msg: processed routeIng: %s, type: %s", key, objname, objType)

	var parsedIng IngressConfig
	var modelList []string

	parsedIng = routeIgrObj.ParseHostPath()

	// Check if this ingress and had any previous mappings, if so - delete them first.
	_, Storedhosts := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)

	hostsMap := make(map[string]*objects.RouteIngrhost)

	if lib.IsEvhEnabled() {
		// Process insecure hosts
		ProcessInsecureHostsForEVH(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap)
		// process secure hosts
		ProcessSecureHostsForEVH(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap, fullsync, sharedQueue)
		ProcessPassthroughHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap)
		// delete stale data
		DeleteStaleDataForEvh(routeIgrObj, key, &modelList, Storedhosts, hostsMap)
		// hostNamePathStore cache operation
		_, oldHostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
		updateHostPathCache(namespace, objname, oldHostMap, hostsMap)

		routeIgrObj.GetSvcLister().IngressMappings(namespace).UpdateRouteIngToHostMapping(objname, hostsMap)
		// publish to rest layer
		if !fullsync {
			utils.AviLog.Infof("key: %s, msg: List of models to publish: %s", key, modelList)
			for _, modelName := range modelList {
				PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		}
		return
	}

	// TODO: These functions will return true or false. Depeding upon that we should update hostcache to have proper sync
	// Process insecure routes first.
	ProcessInsecureHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap)

	// Process secure routes next.
	ProcessSecureHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap, fullsync, sharedQueue)

	ProcessPassthroughHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap)

	utils.AviLog.Debugf("key: %s, msg: Stored hosts: %v, hosts map: %v", key, Storedhosts, hostsMap)
	DeleteStaleData(routeIgrObj, key, &modelList, Storedhosts, hostsMap)

	// hostNamePathStore cache operation
	_, oldHostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	updateHostPathCache(namespace, objname, oldHostMap, hostsMap)

	routeIgrObj.GetSvcLister().IngressMappings(namespace).UpdateRouteIngToHostMapping(objname, hostsMap)

	if !fullsync {
		utils.AviLog.Infof("key: %s, msg: List of models to publish: %s", key, modelList)
		for _, modelName := range modelList {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
}

func getPathSvc(currentPathSvc []IngressHostPathSvc) map[string][]string {
	pathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		pathSvcMap[val.Path] = append(pathSvcMap[val.Path], val.ServiceName)
	}
	return pathSvcMap
}

func ProcessInsecureHosts(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before processing insecurehosts: %s", key, utils.Stringify(Storedhosts))

	infraSetting := routeIgrObj.GetAviInfraSetting()
	for host, pathsvcmap := range parsedIng.IngressHostMap {
		// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
		hostData, found := Storedhosts[host]
		_, shardVsName := DeriveShardVS(host, key, routeIgrObj)
		if found && hostData.InsecurePolicy != lib.PolicyNone {
			// TODO: StoredPaths might be empty if the host was not specified with any paths.
			// Verify the paths and take out the paths that are not need.
			pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, pathsvcmap.ingressHPSvc, false)
			if len(pathSvcDiff) == 0 {
				Storedhosts[host].InsecurePolicy = lib.PolicyNone
				if shardVsName.Dedicated {
					Storedhosts[host].SecurePolicy = lib.PolicyNone
				}
			} else {
				hostData.PathSvc = pathSvcDiff
			}
			utils.AviLog.Infof("hostData.PathSvc: %v", hostData.PathSvc)
		}
		if _, ok := hostsMap[host]; !ok {
			hostsMap[host] = &objects.RouteIngrhost{
				SecurePolicy: lib.PolicyNone,
			}
		}
		hostsMap[host].InsecurePolicy = lib.PolicyAllow
		hostsMap[host].PathSvc = getPathSvc(pathsvcmap.ingressHPSvc)

		modelName := lib.GetModelName(shardVsName.Tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, modelName)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName.Name, shardVsName.Tenant, key, routeIgrObj, shardVsName.Dedicated, false)
		}

		vsNode := aviModel.(*AviObjectGraph).GetAviVS()

		if !shardVsName.Dedicated {
			aviModel.(*AviObjectGraph).BuildL7VSGraphHostNameShard(shardVsName.Name, host, routeIgrObj, pathsvcmap.ingressHPSvc, pathsvcmap.gslbHostHeader, parsedIng.InsecureEdgeTermAllow, key)
		} else {
			aviModel.(*AviObjectGraph).BuildDedicatedL7VSGraphHostNameShard(shardVsName.Name, host, routeIgrObj, parsedIng.InsecureEdgeTermAllow, pathsvcmap, key)
		}
		if len(vsNode) > 0 && found {
			// if vsNode already exists, check for updates via AviInfraSetting
			if infraSetting != nil {
				buildWithInfraSetting(key, routeIgrObj.GetNamespace(), vsNode[0], vsNode[0].VSVIPRefs[0], infraSetting)
			}
		}
		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing insecurehosts: %s", key, utils.Stringify(Storedhosts))
}

func ProcessSecureHosts(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost,
	hostsMap map[string]*objects.RouteIngrhost, fullsync bool, sharedQueue *utils.WorkerQueue) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before processing securehosts: %v", key, Storedhosts)
	for _, tlssetting := range parsedIng.TlsCollection {
		locSniHostMap, dedicated := sniNodeHostName(routeIgrObj, tlssetting, routeIgrObj.GetName(), routeIgrObj.GetNamespace(), key, fullsync, sharedQueue, modelList)
		for host, newPathSvc := range locSniHostMap {
			// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
			hostData, found := Storedhosts[host]
			if dedicated && found && hostData.InsecurePolicy == lib.PolicyAllow {
				// this is transitioning from insecure to secure host
				Storedhosts[host].InsecurePolicy = lib.PolicyNone
			}
			if found && hostData.SecurePolicy == lib.PolicyEdgeTerm {
				// TODO: StoredPaths might be empty if the host was not specified with any paths.
				// Verify the paths and take out the paths that are not need.
				pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, newPathSvc, false)

				// For transtion from insecureEdgeTermination policy Allow -> None in a route
				// pathSvcDiff would be empty, but we still need to delete the pool for insecure route
				// from the shared VS. Hence don't assign the empty value, just update the policy
				if len(pathSvcDiff) == 0 {
					Storedhosts[host].SecurePolicy = lib.PolicyNone
					if dedicated {
						Storedhosts[host].InsecurePolicy = lib.PolicyNone
					}
				} else {
					hostData.PathSvc = pathSvcDiff
				}
			}
			if _, ok := hostsMap[host]; !ok {
				hostsMap[host] = &objects.RouteIngrhost{
					InsecurePolicy: lib.PolicyNone,
				}
			}
			hostsMap[host].SecurePolicy = lib.PolicyEdgeTerm
			if tlssetting.redirect {
				hostsMap[host].InsecurePolicy = lib.PolicyRedirect
			}
			hostsMap[host].PathSvc = getPathSvc(newPathSvc)
		}
	}

	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing securehosts: %s", key, utils.Stringify(Storedhosts))
}

func ProcessPassthroughHosts(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string,
	Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before processing passthrough hosts: %v", key, Storedhosts)
	infraSetting := routeIgrObj.GetAviInfraSetting()
	tenant := lib.GetTenantInNamespace(routeIgrObj.GetNamespace())
	for host, pass := range parsedIng.PassthroughCollection {
		hostData, found := Storedhosts[host]
		if found && hostData.SecurePolicy == lib.PolicyPass {
			Storedhosts[host].SecurePolicy = lib.PolicyNone
		}

		if _, ok := hostsMap[host]; !ok {
			hostsMap[host] = &objects.RouteIngrhost{
				InsecurePolicy: lib.PolicyNone,
			}
		}
		hostsMap[host].SecurePolicy = lib.PolicyPass
		redirect := false
		if pass.redirect {
			redirect = true
			hostsMap[host].InsecurePolicy = lib.PolicyRedirect
		}
		_, shardVsName := DerivePassthroughVS(host, key, routeIgrObj)
		modelName := lib.GetModelName(tenant, shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).BuildVSForPassthrough(shardVsName, routeIgrObj.GetNamespace(), host, tenant, key, infraSetting)
		}
		vsNode := aviModel.(*AviObjectGraph).GetAviVS()
		if len(vsNode) < 1 {
			return
		}
		if infraSetting != nil {
			buildWithInfraSetting(key, routeIgrObj.GetNamespace(), vsNode[0], vsNode[0].VSVIPRefs[0], infraSetting)
		}
		aviModel.(*AviObjectGraph).BuildGraphForPassthrough(pass.PathSvc, routeIgrObj.GetName(), host, routeIgrObj.GetNamespace(), tenant, key, redirect, infraSetting)

		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing passthrough hosts: %s", key, utils.Stringify(Storedhosts))
}

// DeleteStaleData : delete pool, fqdn and redirect policy which are present in the object store but no longer required.
func DeleteStaleData(routeIgrObj RouteIngressModel, key string, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	var infraSettingName string

	if aviInfraSetting := routeIgrObj.GetAviInfraSetting(); aviInfraSetting != nil {
		if !lib.IsInfraSettingNSScoped(aviInfraSetting.Name, routeIgrObj.GetNamespace()) {
			infraSettingName = aviInfraSetting.Name
		}
	}

	for host, hostData := range Storedhosts {
		utils.AviLog.Debugf("host to del: %s, data : %s", host, utils.Stringify(hostData))
		_, shardVsName := DeriveShardVS(host, key, routeIgrObj)
		if hostData.SecurePolicy == lib.PolicyPass {
			_, shardVsName.Name = DerivePassthroughVS(host, key, routeIgrObj)
		}

		modelName := lib.GetModelName(shardVsName.Tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}
		// By default remove both redirect and fqdn. So if the host isn't transitioning, then we will remove both.
		removeFqdn := true
		removeRedir := true
		currentData, ok := hostsMap[host]
		// if route is transitioning from/to passthrough route, then always remove fqdn
		if ok && hostData.SecurePolicy != lib.PolicyPass && currentData.SecurePolicy != lib.PolicyPass {
			if currentData.InsecurePolicy == lib.PolicyRedirect {
				removeRedir = false
			}
			utils.AviLog.Infof("key: %s, host: %s, currentData: %v", key, host, currentData)
			removeFqdn = false
		}
		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, removeFqdn, removeRedir, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, infraSettingName, key, removeFqdn, removeRedir, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, removeFqdn, removeRedir, false)
		}

		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}
}

func DeleteStaleDataForModelChange(routeIgrObj RouteIngressModel, namespace, objname, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: nothing to delete for route: %s", key, objname)
		return
	}
	var shardVsName lib.VSNameMetadata
	var newShardVsName lib.VSNameMetadata
	for host, hostData := range hostMap {

		shardVsName, newShardVsName = DeriveShardVS(host, key, routeIgrObj)
		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName.Name, newShardVsName.Name = DerivePassthroughVS(host, key, routeIgrObj)
		}
		if shardVsName == newShardVsName {
			continue
		}

		_, infraSettingName := objects.InfraSettingL7Lister().GetIngRouteToInfraSetting(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName())
		if lib.IsInfraSettingNSScoped(infraSettingName, namespace) {
			infraSettingName = ""
		}
		modelName := lib.GetModelName(shardVsName.Tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}

		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, infraSettingName, key, true, true, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, false)
		}

		ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
}

// RouteIngrDeletePoolsByHostname : Based on DeletePoolsByHostname, delete pools and policies that are no longer required
func RouteIngrDeletePoolsByHostname(routeIgrObj RouteIngressModel, namespace, objname, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	ok, hostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	if !ok {
		utils.AviLog.Warnf("key: %s, msg: nothing to delete for route: %s", key, objname)
		return
	}

	_, infraSettingName := objects.InfraSettingL7Lister().GetIngRouteToInfraSetting(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName())
	tenant := objects.SharedNamespaceTenantLister().GetTenantInNamespace(routeIgrObj.GetNamespace() + "/" + routeIgrObj.GetName())
	if tenant == "" {
		tenant = lib.GetTenant()
	}
	if lib.IsInfraSettingNSScoped(infraSettingName, namespace) {
		infraSettingName = ""
	}

	utils.AviLog.Debugf("key: %s, msg: hosts to delete are :%s", key, utils.Stringify(hostMap))
	for host, hostData := range hostMap {
		deleteVS := false
		shardVsName, _ := DeriveShardVS(host, key, routeIgrObj)

		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName.Name, _ = DerivePassthroughVS(host, key, routeIgrObj)
		}

		SharedHostNameLister().DeleteNamespace(host)
		if found, ingressHostMap := SharedHostNameLister().Get(host); found {
			mapkey := namespace + "/" + objname
			delete(ingressHostMap.HostNameMap, mapkey)

		}
		modelName := lib.GetModelName(tenant, shardVsName.Name)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}

		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			deleteVS = aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, infraSettingName, key, true, true, true)
		}
		if hostData.InsecurePolicy == lib.PolicyAllow {
			deleteVS = aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName.Name, host, routeIgrObj, hostData.PathSvc, key, infraSettingName, true, true, false)
		}
		if !deleteVS {
			ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
			if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
				PublishKeyToRestLayer(modelName, key, sharedQueue)
			}
		} else {
			utils.AviLog.Debugf("key: %s, msg: setting up model name :[%v] to nil", key, modelName)
			objects.SharedAviGraphLister().Save(modelName, nil)
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
	// Now remove the secret relationship
	routeIgrObj.GetSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(objname)
	utils.AviLog.Infof("key: %s, removed ingress mapping for: %s", key, objname)

	// Remove the hosts mapping for this ingress
	routeIgrObj.GetSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(objname)

	// remove hostpath mappings
	updateHostPathCache(namespace, objname, hostMap, nil)
}

func updateHostPathCache(ns, ingress string, oldHostMap, newHostMap map[string]*objects.RouteIngrhost) {
	mmapval := ns + "/" + ingress

	// remove from oldHostMap
	for host, oldMap := range oldHostMap {
		for path := range oldMap.PathSvc {
			SharedHostNameLister().RemoveHostPathStore(host, path, mmapval)
		}
	}

	// add from newHostMap
	if newHostMap != nil {
		for host, newMap := range newHostMap {
			for path := range newMap.PathSvc {
				SharedHostNameLister().SaveHostPathStore(host, path, mmapval)
			}
		}
	}
}
