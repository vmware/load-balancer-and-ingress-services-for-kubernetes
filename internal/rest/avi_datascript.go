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

package rest

import (
	"errors"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
)

func (rest *RestOperations) AviDSBuild(ds_meta *nodes.AviHTTPDataScriptNode, cache_obj *avicache.AviDSCache, key string) *utils.RestOp {
	var datascriptlist []*avimodels.VSDataScript
	var poolgroupref []string
	for _, pgname := range ds_meta.PoolGroupRefs {
		// Replace the PoolGroup Ref in the DS.
		pg_ref := "/api/poolgroup/?name=" + pgname
		poolgroupref = append(poolgroupref, pg_ref)
	}
	datascript := avimodels.VSDataScript{Evt: &ds_meta.Evt, Script: &ds_meta.Script}
	datascriptlist = append(datascriptlist, &datascript)
	tenant_ref := "/api/tenant/?name=" + ds_meta.Tenant
	cr := lib.AKOUser
	vsdatascriptset := avimodels.VSDataScriptSet{CreatedBy: &cr, Datascript: datascriptlist, Name: &ds_meta.Name, TenantRef: &tenant_ref, PoolGroupRefs: poolgroupref}
	if len(ds_meta.ProtocolParsers) > 0 {
		vsdatascriptset.ProtocolParserRefs = ds_meta.ProtocolParsers
	}

	var path string
	var rest_op utils.RestOp
	macro := utils.AviRestObjMacro{ModelName: "VSDataScriptSet", Data: vsdatascriptset}
	if cache_obj != nil {
		path = "/api/vsdatascriptset/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: vsdatascriptset,
			Tenant: ds_meta.Tenant, Model: "VSDataScriptSet", Version: utils.CtrlVersion}
	} else {
		// Patch an existing ds if it exists in the cache but not associated with this VS.
		ds_key := avicache.NamespaceName{Namespace: ds_meta.Tenant, Name: ds_meta.Name}
		ds_cache, ok := rest.cache.DSCache.AviCacheGet(ds_key)
		if ok {
			ds_cache_obj, _ := ds_cache.(*avicache.AviDSCache)
			path = "/api/vsdatascriptset/" + ds_cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: vsdatascriptset,
				Tenant: ds_meta.Tenant, Model: "VSDataScriptSet", Version: utils.CtrlVersion}
		} else {
			path = "/api/macro"
			rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
				Tenant: ds_meta.Tenant, Model: "VSDataScriptSet", Version: utils.CtrlVersion}
		}
	}

	utils.AviLog.Debugf(spew.Sprintf("key: %s, msg: ds Restop %v DatascriptData %v\n", key,
		utils.Stringify(rest_op), *ds_meta))
	return &rest_op
}

func (rest *RestOperations) AviDSDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/vsdatascriptset/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "VSDataScriptSet", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("key: %s, msg: DS DELETE Restop %v \n", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviDSCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for datascriptset err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "vsdatascriptset", key)
	utils.AviLog.Debugf("The datascriptset object response %v", rest_op.Response)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find datascriptset obj in resp %v", key, rest_op.Response)
		return errors.New("datascriptset not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: DS Name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: DS Uuid not present in response %v", key, resp)
			continue
		}
		// Datascript should not have a checksum

		var poolgroups []string
		if resp["pool_group_refs"] != nil {
			pgs, _ := resp["pool_group_refs"].([]interface{})
			for _, pg := range pgs {
				pgUuid := avicache.ExtractUuid(pg.(string), "poolgroup-.*.#")
				pgName, found := rest.cache.PgCache.AviCacheGetNameByUuid(pgUuid)
				if found {
					poolgroups = append(poolgroups, pgName.(string))
				}
			}
		}
		ds_cache_obj := avicache.AviDSCache{Name: name, Tenant: rest_op.Tenant,
			Uuid: uuid, PoolGroups: poolgroups}

		ds_cache_obj.CloudConfigCksum = lib.DSChecksum(ds_cache_obj.PoolGroups)

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.DSCache.AviCacheAdd(k, &ds_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToDSKeyCollection(k)
				utils.AviLog.Debugf("key: %s, msg: modified the VS cache object for Datascriptset Collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
			}
		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToDSKeyCollection(k)
			utils.AviLog.Info(spew.Sprintf("key: %s, msg: added VS cache key during datascriptset update %v val %v\n", key, vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Info(spew.Sprintf("key: %s, msg: added Datascriptset cache k %v val %v\n", key, k,
			ds_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviDSCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	dsKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	utils.AviLog.Debugf("Deleting DS: %s", dsKey)
	rest.cache.DSCache.AviCacheDelete(dsKey)
	vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			vs_cache_obj.RemoveFromDSKeyCollection(dsKey)

		}
	}

	return nil
}
