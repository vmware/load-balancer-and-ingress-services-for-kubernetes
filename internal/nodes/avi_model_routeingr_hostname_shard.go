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
	"errors"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/objects"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	routev1 "github.com/openshift/api/route/v1"
	networkingv1beta1 "k8s.io/api/networking/v1beta1"
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
	key       string
	name      string
	namespace string
	spec      routev1.RouteSpec
}

// K8sIngressModel : Model for openshift routes with default service lister
type K8sIngressModel struct {
	key       string
	name      string
	namespace string
	spec      networkingv1beta1.IngressSpec
}

func GetOshiftRouteModel(name, namespace, key string) (*OshiftRouteModel, error, bool) {
	routeModel := OshiftRouteModel{
		key:       key,
		name:      name,
		namespace: namespace,
	}
	processObj := true
	processObj = utils.CheckIfNamespaceAccepted(namespace, utils.GetGlobalNSFilter(), nil, true)

	routeObj, err := utils.GetInformers().RouteInformer.Lister().Routes(namespace).Get(name)
	if err != nil {
		return &routeModel, err, processObj
	}
	routeModel.spec = routeObj.Spec
	if !lib.HasValidBackends(routeObj.Spec, name, namespace, key) {
		err := errors.New("validation failed for alternate backends for route: " + name)
		return &routeModel, err, false
	}

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
	return o.ParseHostPathForRoute(or.namespace, or.name, or.spec, or.key)
}

func (m *OshiftRouteModel) GetDiffPathSvc(storedPathSvc map[string][]string, currentPathSvc []IngressHostPathSvc) map[string][]string {
	pathSvcCopy := make(map[string][]string)
	for k, v := range storedPathSvc {
		pathSvcCopy[k] = v
	}
	currPathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		currPathSvcMap[val.Path] = append(currPathSvcMap[val.Path], val.ServiceName)
	}
	for path, services := range currPathSvcMap {
		storedServices, ok := pathSvcCopy[path]
		if ok {
			pathSvcCopy[path] = Difference(storedServices, services)
			if len(pathSvcCopy[path]) == 0 {
				delete(pathSvcCopy, path)
			}
		}
	}
	return pathSvcCopy
}

func GetK8sIngressModel(name, namespace, key string) (*K8sIngressModel, error, bool) {
	ingrModel := K8sIngressModel{
		key:       key,
		name:      name,
		namespace: namespace,
	}
	processObj := true
	ingObj, err := utils.GetInformers().IngressInformer.Lister().Ingresses(namespace).Get(name)
	if err != nil {
		return &ingrModel, err, processObj
	}
	processObj = validateIngressForClass(key, ingObj) && utils.CheckIfNamespaceAccepted(namespace, utils.GetGlobalNSFilter(), nil, true)
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
	return o.ParseHostPathForIngress(m.namespace, m.name, m.spec, m.key)
}

func (m *K8sIngressModel) GetDiffPathSvc(storedPathSvc map[string][]string, currentPathSvc []IngressHostPathSvc) map[string][]string {
	pathSvcCopy := make(map[string][]string)
	for k, v := range storedPathSvc {
		pathSvcCopy[k] = v
	}
	currPathSvcMap := make(map[string][]string)
	for _, val := range currentPathSvc {
		currPathSvcMap[val.Path] = append(currPathSvcMap[val.Path], val.ServiceName)
	}
	for path := range currPathSvcMap {
		_, ok := pathSvcCopy[path]
		if ok {
			delete(pathSvcCopy, path)
		}
	}
	return pathSvcCopy
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
		routeIgrObj, err, processObj = GetK8sIngressModel(objname, namespace, key)
	case utils.OshiftRoute:
		if utils.GetInformers().RouteInformer == nil {
			return
		}
		routeIgrObj, err, processObj = GetOshiftRouteModel(objname, namespace, key)

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

	// Check if this ingress and had any previous mappings, if so - delete them first.
	_, Storedhosts := routeIgrObj.GetSvcLister().IngressMappings(namespace).GetRouteIngToHost(objname)

	// Process insecure routes first.
	hostsMap := make(map[string]*objects.RouteIngrhost)
	ProcessInsecureHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap)

	// Process secure routes next.
	ProcessSecureHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap, fullsync, sharedQueue)

	ProcessPassthroughHosts(routeIgrObj, key, parsedIng, &modelList, Storedhosts, hostsMap)

	utils.AviLog.Debugf("key: %s, msg: Stored hosts: %v, hosts map: %v", key, Storedhosts, hostsMap)
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

func ProcessInsecureHosts(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	for host, pathsvcmap := range parsedIng.IngressHostMap {
		// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
		hostData, found := Storedhosts[host]
		if found && hostData.InsecurePolicy != lib.PolicyNone {
			// TODO: StoredPaths might be empty if the host was not specified with any paths.
			// Verify the paths and take out the paths that are not need.
			pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, pathsvcmap)
			if len(pathSvcDiff) == 0 {
				Storedhosts[host].InsecurePolicy = lib.PolicyNone
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
	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing insecurehosts: %s", key, utils.Stringify(Storedhosts))
}

func ProcessSecureHosts(routeIgrObj RouteIngressModel, key string, parsedIng IngressConfig, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost,
	hostsMap map[string]*objects.RouteIngrhost, fullsync bool, sharedQueue *utils.WorkerQueue) {
	utils.AviLog.Debugf("key: %s, msg: Storedhosts before processing securehosts: %v", key, Storedhosts)

	for _, tlssetting := range parsedIng.TlsCollection {
		locSniHostMap := sniNodeHostName(routeIgrObj, tlssetting, routeIgrObj.GetName(), routeIgrObj.GetNamespace(), key, fullsync, sharedQueue, modelList)
		for host, newPathSvc := range locSniHostMap {
			// Remove this entry from storedHosts. First check if the host exists in the stored map or not.
			hostData, found := Storedhosts[host]
			if found && hostData.SecurePolicy == lib.PolicyEdgeTerm {
				// TODO: StoredPaths might be empty if the host was not specified with any paths.
				// Verify the paths and take out the paths that are not need.
				pathSvcDiff := routeIgrObj.GetDiffPathSvc(hostData.PathSvc, newPathSvc)

				// For transtion from insecureEdgeTermination policy Allow -> None in a route
				// pathSvcDiff would be empty, but we still need to delete the pool for insecure route
				// from the shared VS. Hence don't assign the empty value, just update the policy
				if len(pathSvcDiff) == 0 {
					Storedhosts[host].SecurePolicy = lib.PolicyNone
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
			if tlssetting.redirect == true {
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
	for host, pass := range parsedIng.PassthroughCollection {
		hostData, found := Storedhosts[host]
		if found && hostData.SecurePolicy == lib.PolicyPass {
			//
			Storedhosts[host].SecurePolicy = lib.PolicyNone
		}

		if _, ok := hostsMap[host]; !ok {
			hostsMap[host] = &objects.RouteIngrhost{
				InsecurePolicy: lib.PolicyNone,
			}
		}
		hostsMap[host].SecurePolicy = lib.PolicyPass
		redirect := false
		if pass.redirect == true {
			redirect = true
			hostsMap[host].InsecurePolicy = lib.PolicyRedirect
		}

		shardVsName := lib.GetPassthroughShardVSName(host, key)
		modelName := lib.GetModelName(lib.GetTenant(), shardVsName)
		found, aviModel := objects.SharedAviGraphLister().Get(modelName)
		if !found || aviModel == nil {
			aviModel = NewAviObjectGraph()
			aviModel.(*AviObjectGraph).BuildVSForPassthrough(shardVsName, routeIgrObj.GetNamespace(), host, key)
		}
		aviModel.(*AviObjectGraph).BuildGraphForPassthrough(pass.PathSvc, routeIgrObj.GetName(), host, routeIgrObj.GetNamespace(), key, redirect)

		changedModel := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if !utils.HasElem(modelList, modelName) && changedModel {
			*modelList = append(*modelList, modelName)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: Storedhosts after processing passthrough hosts: %s", key, utils.Stringify(Storedhosts))
}

//DeleteStaleData : delete pool, fqdn and redirect policy which are present in the object store but no longer required.
func DeleteStaleData(routeIgrObj RouteIngressModel, key string, modelList *[]string, Storedhosts map[string]*objects.RouteIngrhost, hostsMap map[string]*objects.RouteIngrhost) {
	for host, hostData := range Storedhosts {
		utils.AviLog.Debugf("host to del: %s, data : %s", host, utils.Stringify(hostData))
		shardVsName := DeriveHostNameShardVS(host, key)
		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName = lib.GetPassthroughShardVSName(host, key)
		}
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
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, host, routeIgrObj, hostData.PathSvc, key, removeFqdn, removeRedir, true)
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName, host, routeIgrObj, hostData.PathSvc, key, removeFqdn, removeRedir, true)
		}
		if hostData.InsecurePolicy != lib.PolicyNone {
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

	utils.AviLog.Debugf("key: %s, msg: hosts to delete are :%s", key, utils.Stringify(hostMap))
	for host, hostData := range hostMap {
		shardVsName := DeriveHostNameShardVS(host, key)
		if hostData.SecurePolicy == lib.PolicyPass {
			shardVsName = lib.GetPassthroughShardVSName(host, key)
		}
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
		} else if hostData.SecurePolicy == lib.PolicyPass {
			aviModel.(*AviObjectGraph).DeleteObjectsForPassthroughHost(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, true)
		}
		if hostData.InsecurePolicy == lib.PolicyAllow {
			aviModel.(*AviObjectGraph).DeletePoolForHostname(shardVsName, host, routeIgrObj, hostData.PathSvc, key, true, true, false)
		}
		ok := saveAviModel(modelName, aviModel.(*AviObjectGraph), key)
		if ok && len(aviModel.(*AviObjectGraph).GetOrderedNodes()) != 0 && !fullsync {
			PublishKeyToRestLayer(modelName, key, sharedQueue)
		}
	}
	// Now remove the secret relationship
	routeIgrObj.GetSvcLister().IngressMappings(namespace).RemoveIngressSecretMappings(objname)
	utils.AviLog.Infof("key: %s, removed ingress mapping for: %s", key, objname)

	// Remove the hosts mapping for this ingress
	routeIgrObj.GetSvcLister().IngressMappings(namespace).DeleteIngToHostMapping(objname)

	// remove hostpath mappings
	updateHostPathCacheV2(namespace, objname, hostMap, nil)
}

func updateHostPathCacheV2(ns, ingress string, oldHostMap, newHostMap map[string]*objects.RouteIngrhost) {
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
