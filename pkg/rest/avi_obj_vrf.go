/*
 * [2013] - [2018] Avi Networks Incorporated
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
	"errors"

	avimodels "github.com/avinetworks/sdk/go/models"
	avicache "gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/akc/pkg/lib"
	"gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
)

func (rest *RestOperations) AviVrfGet(key, uuid, name string) *avimodels.VrfContext {
	vrf := avimodels.VrfContext{
		Name: &name,
		UUID: &uuid,
	}
	aviRestClientPool := avicache.SharedAVIClients()
	if len(aviRestClientPool.AviClient) == 0 {
		return nil
	}
	client := aviRestClientPool.AviClient[0]
	var restResponse interface{}
	uri := "/api/vrfcontext/" + uuid

	err := client.AviSession.Get(uri, &restResponse)
	if err != nil {
		utils.AviLog.Warning.Printf("Vs Get uri %v returned err %v", uri, err)
		return nil
	}

	resp, ok := restResponse.(map[string]interface{})
	if !ok {
		utils.AviLog.Warning.Printf("Vrfcontext Get uri %v returned %v type %T", uri,
			restResponse, restResponse)
		return nil
	}

	for key, val := range resp {
		switch key {
		case "bgp_profile":
			if bgpProfileIntf, ok := val.(map[string]interface{}); ok {
				vrf.BgpProfile = lib.BgpProfileIntfToObj(bgpProfileIntf)
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for bgpprofile in vrf %s\n", key, val, name)
			}
		case "cloud_ref":
			if cloudref, ok := val.(string); ok {
				vrf.CloudRef = &cloudref
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for cloudref in vrf %s\n", key, val, name)
			}
		case "debugvrfcontest":
			if debugVrfContest, ok := val.(map[string]interface{}); ok {
				vrf.Debugvrfcontext = lib.DebugVrfContestIntfToObj(debugVrfContest)
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for debugvrfcontest in vrf %s\n", key, val, name)
			}
		case "description":
			if description, ok := val.(string); ok {
				vrf.Description = &description
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for description in vrf %s\n", key, val, name)
			}
		case "gateway_mon":
			if gatewaymon, ok := val.([]interface{}); ok {
				vrf.GatewayMon = lib.GatewayMonIntfToObj(gatewaymon)
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for gatewaymon in vrf %s\n", key, val, name)
			}
		case "internal_gateway_monitor":
			if internalGatewayMonitor, ok := val.(map[string]interface{}); ok {
				vrf.InternalGatewayMonitor = lib.InternalGatewayMonIntfToObj(internalGatewayMonitor)
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for internalgatewaymonitor in vrf %s\n", key, val, name)
			}
		case "system_default":
			if systemdefault, ok := val.(bool); ok {
				vrf.SystemDefault = &systemdefault
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for systemdefault in vrf %s\n", key, val, name)
			}
		case "tenant_ref":
			if tenantref, ok := val.(string); ok {
				vrf.TenantRef = &tenantref
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for tenantref in vrf %s\n", key, val, name)
			}
		case "url":
			if url, ok := val.(string); ok {
				vrf.URL = &url
			} else {
				utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for url in vrf %s\n", key, val, name)
			}
		}
	}
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
	vrf.StaticRoutes = vrfNode.StaticRoutes
	restOp := utils.RestOp{Path: path, Method: utils.RestPut, Obj: vrf,
		Tenant: utils.ADMIN_NS, Model: "VrfContext", Version: utils.CtrlVersion}
	return &restOp
}

func (rest *RestOperations) getVrfCacheObj(vrfName string) *avicache.AviVrfCache {
	vrfCache, found := rest.cache.VrfCache.AviCacheGet(vrfName)
	if found {
		vrfCacheObj, ok := vrfCache.(*avicache.AviVrfCache)
		if !ok {
			utils.AviLog.Warning.Printf("Vrf object for %s found. Cannot cast. Not doing anything\n", vrfName)
			return nil
		}
		return vrfCacheObj
	}
	utils.AviLog.Info.Printf("vrf cache object NOT found for vrf name: %s", vrfName)
	return nil
}

func (rest *RestOperations) AviVrfCacheAdd(restOp *utils.RestOp, vrfKey avicache.NamespaceName, key string) error {
	if (restOp.Err != nil) || (restOp.Response == nil) {
		utils.AviLog.Warning.Printf("key: %s, rest_op has err or no reponse for POOL, err: %s, response: %s", key, restOp.Err, restOp.Response)
		return errors.New("Errored rest_op")
	}
	respElems, ok := RestRespArrToObjByType(restOp, "vrfcontext", key)
	if ok != nil || respElems == nil {
		utils.AviLog.Warning.Printf("key: %s, msg: unable to find vrfcontext obj in resp %v", key, restOp.Response)
		return errors.New("vrfcontext not found")
	}
	vrfName := vrfKey.Name
	rest.cache.VrfCache.AviCacheGet(vrfName)
	for _, resp := range respElems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for name in vrf %s\n", key, resp["name"], vrfName)
			continue
		}
		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for uuid in vrf %s\n", key, resp["uuid"], vrfName)
			continue
		}
		staticRoutesIntf, ok := resp["staticroutes"].([]interface{})
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: wrong object type %T for uuid in staticroutes %s\n", key, resp["staticroutes"], vrfName)
			continue
		}
		staticRoutes := lib.StaticRoutesIntfToObj(staticRoutesIntf)
		if len(staticRoutes) == 0 {
			utils.AviLog.Trace.Printf("key: %s, no static routes found for vrf %s\n", key, vrfName)
			continue
		}
		checksum := avicache.VrfChecksum(name, staticRoutes)
		vrfCacheObj := avicache.AviVrfCache{Name: name, Uuid: uuid, CloudConfigCksum: checksum}
		rest.cache.VrfCache.AviCacheAdd(vrfKey, vrfCacheObj)
	}
	return nil
}

func (rest *RestOperations) AviVrfCacheDel(restOp *utils.RestOp, vrfKey avicache.NamespaceName, key string) error {
	if (restOp.Err != nil) || (restOp.Response == nil) {
		utils.AviLog.Warning.Printf("key: %s, rest_op has err or no reponse for POOL, err: %s, response: %s", key, restOp.Err, restOp.Response)
		return errors.New("Errored rest_op")
	}
	respElems, ok := RestRespArrToObjByType(restOp, "vrfcontext", key)
	if ok != nil || respElems == nil {
		utils.AviLog.Warning.Printf("key: %s, msg: unable to find vrfcontext obj in resp %v", key, restOp.Response)
		return errors.New("vrfcontext not found")
	}

	rest.cache.VrfCache.AviCacheGet(vrfKey)
	/*for _, resp := range respElems {
		name, _ := resp["name"].(string)
		staticRoutes, _ := resp["staticroutes"].(string)
		checksum := strconv.Itoa(int(utils.Hash(name) + utils.Hash(utils.Stringify(staticRoutes))))
		vrfCacheObj := avicache.AviVrfCache{Name: name, CloudConfigCksum: checksum}

	}*/
	rest.cache.VrfCache.AviCacheDelete(vrfKey)
	return nil
}
