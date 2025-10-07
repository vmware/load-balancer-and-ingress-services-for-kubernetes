/*
 * Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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
	"encoding/json"
	"errors"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/vmware/alb-sdk/go/models"
)

func (rest *RestOperations) AviVrfGet(key, uuid, name string) *avimodels.VrfContext {

	aviRestPoolClient := avicache.SharedAVIClients(lib.GetTenant())
	if aviRestPoolClient == nil {
		utils.AviLog.Warnf("key: %s, msg: aviRestPoolClient not initialized", key)
		return nil
	}
	if len(aviRestPoolClient.AviClient) < 1 {
		utils.AviLog.Warnf("key: %s, msg: client in aviRestPoolClient not initialized", key)
		return nil
	}

	client := aviRestPoolClient.AviClient[0]
	uri := "/api/vrfcontext/" + uuid

	rawData, err := lib.AviGetRaw(client, uri)
	if err != nil {
		utils.AviLog.Warnf("Vrf Get uri %v returned err %v", uri, err)
		return nil
	}
	vrf := avimodels.VrfContext{}
	json.Unmarshal(rawData, &vrf)

	return &vrf
}

func (rest *RestOperations) AviVrfBuild(key string, vrfNode *nodes.AviVrfNode, uuid string) *utils.RestOp {
	vrfCacheObj := rest.getVrfCacheObj(vrfNode.Name)
	if vrfCacheObj == nil {
		return nil
	}
	vrf := rest.AviVrfGet(key, vrfCacheObj.Uuid, vrfCacheObj.Name)
	if vrf == nil {
		return nil
	}
	path := "/api/vrfcontext/" + vrfCacheObj.Uuid
	nodeStaticRoutes := vrfNode.StaticRoutes
	aviStaticRoutes := vrf.StaticRoutes
	mergedStaticRoutes := []*avimodels.StaticRoute{}
	clusterName := lib.GetClusterName()
	utils.AviLog.Infof("key: %s, VRF object in controller %s", key, utils.Stringify(aviStaticRoutes))
	utils.AviLog.Infof("key: %s, VRF object in ako cache %s", key, utils.Stringify(nodeStaticRoutes))
	for _, aviStaticRoute := range aviStaticRoutes {
		if len(aviStaticRoute.Labels) == 0 || (*aviStaticRoute.Labels[0].Key == "clustername" && *aviStaticRoute.Labels[0].Value != clusterName) {
			mergedStaticRoutes = append(mergedStaticRoutes, aviStaticRoute)
		}
	}
	if len(nodeStaticRoutes) != 0 {
		mergedStaticRoutes = append(mergedStaticRoutes, nodeStaticRoutes...)
	}

	vrf.StaticRoutes = []*avimodels.StaticRoute{}
	vrf.StaticRoutes = append(vrf.StaticRoutes, mergedStaticRoutes...)

	opTenant := lib.GetAdminTenant()
	if lib.GetCloudType() == lib.CLOUD_OPENSTACK || lib.GetTenant() != "" {
		//In case of Openstack cloud, use tenant vrf
		opTenant = lib.GetTenant()
	}

	utils.AviLog.Infof("key: %s, VRF object to be sent for update to controller %s", key, utils.Stringify(vrf.StaticRoutes))

	rest_op := utils.RestOp{
		Path:    path,
		Method:  utils.RestPut,
		Obj:     vrf,
		Tenant:  opTenant,
		Model:   "VrfContext",
		ObjName: vrfCacheObj.Name,
	}
	return &rest_op
}

func (rest *RestOperations) getVrfCacheObj(vrfName string) *avicache.AviVrfCache {
	vrfCache, found := rest.cache.VrfCache.AviCacheGet(vrfName)
	if found {
		vrfCacheObj, ok := vrfCache.(*avicache.AviVrfCache)
		if !ok {
			utils.AviLog.Warnf("Vrf object for %s found. Cannot cast. Not doing anything", vrfName)
			return nil
		}
		return vrfCacheObj
	}
	utils.AviLog.Infof("vrf cache object NOT found for vrf name: %s", vrfName)
	return nil
}

func (rest *RestOperations) AviVrfCacheAdd(restOp *utils.RestOp, vrfKey avicache.NamespaceName, key string) error {
	if (restOp.Err != nil) || (restOp.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for vrfcontext, err: %s, response: %v", key, restOp.Err, utils.Stringify(restOp.Response))
		return errors.New("Errored rest_op")
	}
	respElems := rest.restOperator.RestRespArrToObjByType(restOp, "vrfcontext", key)
	if respElems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find vrfcontext obj in resp %v", key, restOp.Response)
		return errors.New("vrfcontext not found")
	}
	vrfName := vrfKey.Name

	var checksum uint32
	var staticRoutes []*avimodels.StaticRoute
	rest.cache.VrfCache.AviCacheGet(vrfName)
	for _, resp := range respElems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: wrong object type %T for name in vrf %s", key, resp["name"], vrfName)
			continue
		}
		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: wrong object type %T for uuid in vrf %s", key, resp["uuid"], vrfName)
			continue
		}
		if resp["static_routes"] == nil {
			staticRoutes = nil
		} else {
			staticRoutesIntf, ok := resp["static_routes"].([]interface{})
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: wrong object type %T for staticroutes in staticroutes %s", key, resp["staticroutes"], vrfName)
				continue
			}
			staticRoutes = lib.StaticRoutesIntfToObj(staticRoutesIntf)
			if len(staticRoutes) == 0 {
				utils.AviLog.Infof("key: %s, no static routes found for vrf %s", key, vrfName)
			}
		}
		checksum = lib.VrfChecksum(name, staticRoutes)
		vrfCacheObj := avicache.AviVrfCache{Name: name, Uuid: uuid, CloudConfigCksum: checksum}
		rest.cache.VrfCache.AviCacheAdd(vrfName, &vrfCacheObj)
	}
	if lib.StaticRouteSyncChan != nil {
		close(lib.StaticRouteSyncChan)
		lib.StaticRouteSyncChan = nil
	}

	return nil
}
