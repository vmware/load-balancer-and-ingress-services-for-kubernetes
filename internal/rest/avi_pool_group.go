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
	"fmt"
	"strconv"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
)

func (rest *RestOperations) AviPoolGroupBuild(pg_meta *nodes.AviPoolGroupNode, cache_obj *avicache.AviPGCache, key string) *utils.RestOp {
	name := pg_meta.Name
	cksum := pg_meta.CloudConfigCksum
	cksumString := strconv.Itoa(int(cksum))
	tenant := fmt.Sprintf("/api/tenant/?name=%s", pg_meta.Tenant)
	members := rest.SanitizePGMembers(pg_meta.Members, key)
	cr := lib.AKOUser
	cloudRef := "/api/cloud?name=" + utils.CloudName

	pg := avimodels.PoolGroup{Name: &name, CloudConfigCksum: &cksumString,
		CreatedBy: &cr, TenantRef: &tenant, Members: members, CloudRef: &cloudRef, ImplicitPriorityLabels: &pg_meta.ImplicitPriorityLabel}

	macro := utils.AviRestObjMacro{ModelName: "PoolGroup", Data: pg}

	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/poolgroup/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pg,
			Tenant: pg_meta.Tenant, Model: "PoolGroup", Version: utils.CtrlVersion}
	} else {
		// Patch an existing pg if it exists in the cache but not associated with this VS.
		pg_key := avicache.NamespaceName{Namespace: pg_meta.Tenant, Name: name}
		pg_cache, ok := rest.cache.PgCache.AviCacheGet(pg_key)
		if ok {
			pg_cache_obj, _ := pg_cache.(*avicache.AviPGCache)
			path = "/api/poolgroup/" + pg_cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pg,
				Tenant: pg_meta.Tenant, Model: "PoolGroup", Version: utils.CtrlVersion}
		} else {
			path = "/api/macro"
			rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
				Tenant: pg_meta.Tenant, Model: "PoolGroup", Version: utils.CtrlVersion}
		}
	}

	return &rest_op
}

func (rest *RestOperations) SanitizePGMembers(Members []*avimodels.PoolGroupMember, key string) []*avimodels.PoolGroupMember {
	// This method iterates over the pg members and removes any duplicate.
	var newList []string
	var pgmemberscopy []*avimodels.PoolGroupMember
	pgmemberscopy = make([]*avimodels.PoolGroupMember, len(Members))
	copy(pgmemberscopy, Members)
	for i, member := range Members {
		if utils.HasElem(newList, member.PoolRef) {
			// Duplicate detected, remove it from the copy
			pgmemberscopy = append(pgmemberscopy[:i], pgmemberscopy[i+1:]...)
			utils.AviLog.Warnf("key: %s, msg: detected duplicate poolref :%s", key, member.PoolRef)
		} else {
			newList = append(newList, *member.PoolRef)
		}
	}
	return pgmemberscopy
}

func (rest *RestOperations) AviPGDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/poolgroup/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "PoolGroup", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("key: %s, msg: PG DELETE Restop %v \n", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviPGCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for PG err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "poolgroup", key)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find pool group obj in resp %v", key, rest_op.Response)
		return errors.New("poolgroup not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Uuid not present in response %v", key, resp)
			continue
		}

		cksum := resp["cloud_config_cksum"].(string)

		var lastModifiedStr string
		lastModifiedIntf, ok := resp["_last_modified"]
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: last_modified not present in response %v", key, resp)
		} else {
			lastModifiedStr, ok = lastModifiedIntf.(string)
			if !ok {
				utils.AviLog.Warnf("key: %s, msg: last_modified is not of type string", key)
			}
		}
		var poolMembers []string
		if resp["members"] != nil {
			pools, poolsOk := resp["members"].([]interface{})
			if poolsOk {
				for _, poolIntf := range pools {
					poolmap, _ := poolIntf.(map[string]interface{})
					poolUuid := avicache.ExtractUuid(poolmap["pool_ref"].(string), "pool-.*.#")
					// Search the poolName using this Uuid in the poolcache.
					poolName, found := rest.cache.PoolCache.AviCacheGetNameByUuid(poolUuid)
					if found {
						poolMembers = append(poolMembers, poolName.(string))
					}
				}
			}
		}
		pg_cache_obj := avicache.AviPGCache{Name: name, Tenant: rest_op.Tenant,
			Uuid:             uuid,
			CloudConfigCksum: cksum,
			LastModified:     lastModifiedStr,
			Members:          poolMembers,
		}
		if lastModifiedStr == "" {
			pg_cache_obj.InvalidData = true
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.PgCache.AviCacheAdd(k, &pg_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToPGKeyCollection(k)
				utils.AviLog.Infof("key: %s, msg: modified the VS cache object for PG collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
			}

		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToPGKeyCollection(k)
			utils.AviLog.Info(spew.Sprintf("key: %s, msg: added VS cache key during poolgroup update %v val %v\n", key, vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Info(spew.Sprintf("key: %s, msg: added PG cache k %v val %v\n", key, k,
			pg_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviPGCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	pgKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	rest.cache.PgCache.AviCacheDelete(pgKey)
	vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			vs_cache_obj.RemoveFromPGKeyCollection(pgKey)
		}
	}

	return nil

}
