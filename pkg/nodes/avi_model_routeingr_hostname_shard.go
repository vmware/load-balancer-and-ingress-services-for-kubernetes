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
	"ako/pkg/lib"
	"ako/pkg/objects"
	"errors"

	"github.com/avinetworks/container-lib/utils"
	routev1 "github.com/openshift/api/route/v1"
	networking "k8s.io/api/networking/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// RouteIngressModel : High Level interfaces that should be implemenetd by
// all l7 route objects, e.g: k8s ingress, openshift route
type RouteIngressModel interface {
	GetName() string
	GetNamespace() string
	GetType() string
	GetSvcLister() *objects.SvcLister
	GetSpec() interface{}
	ParseHostPath() IngressConfig
	// this is required due to different naming convention used in ingress where we dont use service name
	// later if we decide to have common naming for ingress and route, then we can hav a common method
	GetDiffPathSvc(map[string][]string, []IngressHostPathSvc) map[string][]string
}

// OshiftRouteModel : Model for openshift routes with it's own service lister
type OshiftRouteModel struct {
	name      string
	namespace string
	spec      routev1.RouteSpec
}

// K8sIngressModel : Model for openshift routes with default service lister
type K8sIngressModel struct {
	name      string
	namespace string
	spec      networking.IngressSpec
}

func GetOshiftRouteModel(name, namespace string) (*OshiftRouteModel, error, bool) {
	routeModel := OshiftRouteModel{
		name:      name,
		namespace: namespace,
	}
	processObj := true
	routeObj, err := utils.GetInformers().RouteInformer.Lister().Routes(namespace).Get(name)
	if err != nil {
		return &routeModel, err, processObj
	}
	routeModel.spec = routeObj.Spec
	return &routeModel, nil, processObj
}

func (m *OshiftRouteModel) GetName() string {
	return m.name
}

func (m *OshiftRouteModel) GetNamespace() string {
	return m.namespace
}

func (m *OshiftRouteModel) GetType() string {
	return utils.OshiftRoute
}

func (m *OshiftRouteModel) GetSvcLister() *objects.SvcLister {
	return objects.OshiftRouteSvcLister()
}

func (m *OshiftRouteModel) GetSpec() interface{} {
	return m.spec
}

func (or *OshiftRouteModel) ParseHostPath() IngressConfig {
	o := NewNodesValidator()
	return o.ParseHostPathForRoute(or.namespace, or.name, or.spec, "")
}

func (m *OshiftRouteModel) GetDiffPathSvc(storedPathSvc map[string][]string, currentPathSvc []IngressHostPathSvc) map[string][]string {
	currPathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		currPathSvcMap[val.Path] = append(currPathSvcMap[val.Path], val.ServiceName)
	}
	for path, services := range currPathSvcMap {
		storedServices, ok := storedPathSvc[path]
		if ok {
			storedPathSvc[path] = Difference(storedServices, services)
			if len(storedPathSvc[path]) == 0 {
				delete(storedPathSvc, path)
			}
		}
	}
	return storedPathSvc
}

func GetK8sIngressModel(name, namespace string) (*K8sIngressModel, error, bool) {
	ingrModel := K8sIngressModel{
		name:      name,
		namespace: namespace,
	}
	processObj := true
	myIng, err := utils.GetInformers().IngressInformer.Lister().ByNamespace(namespace).Get(name)
	if err != nil {
		return &ingrModel, err, processObj
	}
	ingObj, ok := utils.ToNetworkingIngress(myIng)
	if !ok {
		return &ingrModel, errors.New("Could not convert ingress to net v1beta"), processObj
	}
	processObj = filterIngressOnClass(ingObj)
	ingrModel.spec = ingObj.Spec
	return &ingrModel, nil, processObj
}

func (m *K8sIngressModel) GetName() string {
	return m.name
}

func (m *K8sIngressModel) GetNamespace() string {
	return m.namespace
}

func (m *K8sIngressModel) GetType() string {
	return utils.Ingress
}

func (m *K8sIngressModel) GetSvcLister() *objects.SvcLister {
	return objects.SharedSvcLister()
}

func (m *K8sIngressModel) GetSpec() interface{} {
	return m.spec
}

func (m *K8sIngressModel) ParseHostPath() IngressConfig {
	o := NewNodesValidator()
	return o.ParseHostPathForIngress(m.namespace, m.name, m.spec, "")
}

func (m *K8sIngressModel) GetDiffPathSvc(storedPathSvc map[string][]string, currentPathSvc []IngressHostPathSvc) map[string][]string {
	currPathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		currPathSvcMap[val.Path] = append(currPathSvcMap[val.Path], val.ServiceName)
	}
	for path, _ := range currPathSvcMap {
		_, ok := storedPathSvc[path]
		if ok {
			delete(storedPathSvc, path)
		}
	}
	return storedPathSvc
}

// HostNameShardAndPublishV2 : based on original HostNameShardAndPublish().
// Create model from supported objects - route/ingress, and publish to rest layer
func HostNameShardAndPublishV2(objType, objname, namespace, key string, fullsync bool, sharedQueue *utils.WorkerQueue) {
	utils.AviLog.Infof("key: %s, starting RouteHostNameShardAndPublish", key)
	var routeIgrObj RouteIngressModel
	var err error
	var processObj bool

	switch objType {
	case utils.Ingress:
		if utils.GetInformers().IngressInformer == nil {
			return
		}
		routeIgrObj, err, processObj = GetK8sIngressModel(objname, namespace)
	case utils.OshiftRoute:
		if utils.GetInformers().RouteInformer == nil {
			return
		}
		routeIgrObj, err, processObj = GetOshiftRouteModel(objname, namespace)

	default:
		utils.AviLog.Infof("key: %s, starting unsupported object type: %s", key, objType)
		return
	}

	if err != nil || !processObj {
		utils.AviLog.Infof("key: %s, msg: Error :%v", key, err)
		// Detect a delete condition here.
		if k8serrors.IsNotFound(err) || !processObj {
			utils.AviLog.Infof("key: %s, Deleting Pool for ingress delete", key)
			RouteIngrDeletePoolsByHostname(routeIgrObj, namespace, objname, key, fullsync, sharedQueue)
		}
		return
	}

	utils.AviLog.Infof("key: %s, msg: processed routeIng: %s, type: %s", key, objname, objType)

	var parsedIng IngressConfig
	var modelList []string

	parsedIng = routeIgrObj.ParseHostPath()
	utils.AviLog.Debugf("key: %s, parsed routeIng: %v", key, parsedIng)

	// Check if this ingress and had any previous mappings, if so - delete them first.
	_, Storedhosts := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)

	// Process insecure routes first.
	hostsMap := make(map[string]*objects.RouteIngrhost)
	ProcessInsecureHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap)

	// Process secure routes next.
	ProcessSecureHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap, fullsync, sharedQueue)

	utils.AviLog.Debugf("key: %s, msg: Stored hosts: %s", key, Storedhosts)
	DeleteStaleData(routeIgrObj, key, &modelList, Storedhosts, hostsMap)

	// hostNamePathStore cache operation
	_, oldHostMap := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)
	updateHostPathCacheV2(namespace, objname, oldHostMap, hostsMap)

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

func pathsToEmptySvcMap(paths []string) map[string][]string {
	pathSvcMap := make(map[string][]string)
	for _, path := range paths {
		pathSvcMap[path] = []string{}
	}
	return pathSvcMap
}

func ProcessInsecureHosts(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	for host, pathsvcmap := range parsedIng.IngressHostMap {
		// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
		hostData, found := Storedhosts[host]
		if found && hostData.InsecurePolicy == lib.PolicyAllow {
			// TODO: StoredPaths might be empty if the host was not specified with any paths.
			// Verify the paths and take out the paths that are not need.
			hostData.PathSvc = routeIgrObj.GetDiffPathSvc(hostData.PathSvc, pathsvcmap)
			if len(hostData.PathSvc) == 0 {
				delete(Storedhosts, host)
			}
			utils.AviLog.Infof("hostData.PathSvc: %v", hostData.PathSvc)
		}
		//insecureHostPathMapArr[host] = getPaths(pathsvcmap)
		hostsMap[host] = &objects.RouteIngrhost{
			InsecurePolicy: lib.PolicyAllow,
			SecurePolicy:   lib.PolicyNone,
		}
		//hostsMap[host].Paths = getPaths(pathsvcmap)
		hostsMap[host].PathSvc = getPathSvc(pathsvcmap)

		shardVsName := DeriveHostNameShardVS(host, key)
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return
		}
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Infof("key: %s, msg: model not found, generating new model with name: %s", key, modelName)
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).ConstructAviL7VsNode(shardVsName, key)
		}
		aviModel.(*AviObjectGraph).BuildL7VSGraphHostNameShard(shardVsName, host, routeIgrObj, pathsvcmap, key)
		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}
}

func ProcessSecureHosts(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost,
	hostsMap map[string]*objects.RouteIngrhost, fullsync bool, sharedQueue *utils.WorkerQueue) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before processing securehosts: %v", key, Storedhosts)
	utils.AviLog.Debugf("key: %s, msg: tlscollection: %v", key, parsedIng.TlsCollection)

	// To Do: use service for paths while handling secure routes
	for _, tlssetting := range parsedIng.TlsCollection {
		locSniHostMap := sniNodeHostName(tlssetting, routeIgrObj.GetName(), routeIgrObj.GetNamespace(), key, fullsync, sharedQueue, modelList)
		for host, newPaths := range locSniHostMap {
			// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
			hostData, found := Storedhosts[host]
			if found && hostData.SecurePolicy == lib.PolicyEdgeTerm {
				// TODO: StoredPaths might be empty if the host was not specified with any paths.
				// Verify the paths and take out the paths that are not need.
				pathkeys := []string{}
				for k := range hostData.PathSvc {
					pathkeys = append(pathkeys, k)
				}
				diffStoredPaths := Difference(pathkeys, newPaths)
				if len(diffStoredPaths) == 0 {
					// There's no difference between the paths, we should delete the host entry in the stored Map
					delete(Storedhosts, host)
				} else {
					// These paths are meant for deletion
					Storedhosts[host].PathSvc = pathsToEmptySvcMap(diffStoredPaths)
				}
			}
			hostsMap[host] = &objects.RouteIngrhost{
				InsecurePolicy: lib.PolicyNone,
				SecurePolicy:   lib.PolicyEdgeTerm,
			}
			if routeIgrObj.GetType() == utils.Ingress {
				hostsMap[host].InsecurePolicy = lib.PolicyRedirect
			}
			hostsMap[host].PathSvc = pathsToEmptySvcMap(newPaths)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing securehosts: %v", key, Storedhosts)
}

//DeleteStaleData : delete pool, fqdn and redirect policy which are present in the object store but no longer required.
func DeleteStaleData(routeIgrObj RouteIngressModel, key string, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	for host, hostData := range Storedhosts {
		shardVsName := DeriveHostNameShardVS(host, key)
		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			return
		}
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}
		// By default remove both redirect and fqdn. So if the host isn't transitioning, then we will remove both.
		removeFqdn := true
		removeRedir := true
		currentData, ok := hostsMap[host]
		if ok {
			if currentData.InsecurePolicy == lib.PolicyRedirect {
				removeRedir = false
			}
			utils.AviLog.Infof("key: %s, host: %s, currentData: %v", key, host, currentData)
			removeFqdn = false
		}
		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, host, routeIgrObj, hostData.PathSvc, key, removeFqdn, removeRedir, true)
		} else {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, host, routeIgrObj, hostData.PathSvc, key, removeFqdn, removeRedir, false)

		}
		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
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

	utils.AviLog.Infof("key: %s, msg: hosts to delete are: :%s", key, hostMap)
	for host, hostData := range hostMap {
		shardVsName := DeriveHostNameShardVS(host, key)

		if shardVsName == "" {
			// If we aren't able to derive the ShardVS name, we should return
			utils.AviLog.Infof("key: %s, shard vs ndoe not found for host: %s", host)
			return
		}
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			utils.AviLog.Warnf("key: %s, msg: model not found during delete: %s", key, modelName)
			continue
		}
		// Delete the pool corresponding to this host
		if hostData.SecurePolicy == lib.PolicyEdgeTerm {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, true)
		} else {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, false)
		}
		ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
	// Now remove the secret relationship
	routeIgrObj.GetSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(objname)
	utils.AviLog.Infof("key: %s, removed ingess mapping for: %s", key, objname)

	// Remove the hosts mapping for this ingress
	routeIgrObj.GetSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(objname)

	// remove hostpath mappings
	updateHostPathCacheV2(namespace, objname, hostMap, nil)
}

func updateHostPathCacheV2(ns, ingress string, oldHostMap, newHostMap map[string]*objects.RouteIngrhost) {
	mmapval := ns + "/" + ingress

	// remove from oldHostMap
	for host, oldMap := range oldHostMap {
		for _, path := range oldMap.Paths {
			SharedHostNameLister().RemoveHostPathStore(host+path, mmapval)
		}
	}

	// add from newHostMap
	if newHostMap != nil {
		for host, newMap := range newHostMap {
			for _, path := range newMap.Paths {
				SharedHostNameLister().SaveHostPathStore(host+path, mmapval)
			}
		}
	}
}
