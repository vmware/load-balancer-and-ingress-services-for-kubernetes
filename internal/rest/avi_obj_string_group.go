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
	"errors"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/davecgh/go-spew/spew"
	avimodels "github.com/vmware/alb-sdk/go/models"
)

func GetStringGroupMarkers() []*avimodels.RoleFilterMatchLabel {
	stringGroupMarkers := lib.GetMarkers()
	labelKey := "created_by"
	rfml := &avimodels.RoleFilterMatchLabel{
		Key:    &labelKey,
		Values: []string{lib.GetAKOUser()},
	}
	stringGroupMarkers = append(stringGroupMarkers, rfml)
	return stringGroupMarkers
}

func (rest *RestOperations) AviStringGroupBuild(sg_meta *nodes.AviStringGroupNode, cache_obj *avicache.AviStringGroupCache, key string) *utils.RestOp {

	if lib.CheckObjectNameLength(*sg_meta.Name, lib.StringGroup) {
		utils.AviLog.Warnf("key: %s not processing stringgroup object", key)
		return nil
	}
	tenant_ref := "/api/tenant/?name=" + *sg_meta.TenantRef
	stringgroup := avimodels.StringGroup{
		Name:      sg_meta.Name,
		TenantRef: &tenant_ref,
		Kv:        sg_meta.Kv,
		Type:      sg_meta.Type,
	}
	if sg_meta.Description != nil {
		stringgroup.Description = sg_meta.Description
	}
	if sg_meta.LongestMatch != nil {
		stringgroup.LongestMatch = sg_meta.LongestMatch
	}

	stringgroup.Markers = GetStringGroupMarkers()

	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/stringgroup/" + cache_obj.Uuid
		rest_op = utils.RestOp{
			ObjName: *stringgroup.Name,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     stringgroup,
			Tenant:  *sg_meta.TenantRef,
			Model:   "StringGoup",
		}
	} else {
		// Patch an existing stringgroup if it exists in the cache but not associated with this VS.
		stringgroup_key := avicache.NamespaceName{Namespace: *sg_meta.TenantRef, Name: *sg_meta.Name}
		sg_cache, ok := rest.cache.StringGroupCache.AviCacheGet(stringgroup_key)
		if ok {
			sg_cache_obj, _ := sg_cache.(*avicache.AviStringGroupCache)
			path = "/api/stringgroup/" + sg_cache_obj.Uuid
			rest_op = utils.RestOp{
				ObjName: *stringgroup.Name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     stringgroup,
				Tenant:  *sg_meta.TenantRef,
				Model:   "StringGroup",
			}
		} else {
			path = "/api/stringgroup"
			rest_op = utils.RestOp{
				ObjName: *stringgroup.Name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     stringgroup,
				Tenant:  *sg_meta.TenantRef,
				Model:   "StringGroup",
			}
		}
	}

	utils.AviLog.Debugf(spew.Sprintf("key: %s, msg: stringgroup Restop %v StringGroupData %v", key,
		utils.Stringify(rest_op), *sg_meta))
	return &rest_op
}

func (rest *RestOperations) AviStringGroupDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/stringgroup/" + uuid
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "StringGroup",
	}
	utils.AviLog.Infof(spew.Sprintf("key: %s, msg: StringGroup DELETE Restop %v ", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviStringGroupCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for stringgroup err: %v, response: %v", key, rest_op.Err, rest_op.Response)
		return errors.New("errored rest_op")
	}

	resp_elems := rest.restOperator.RestRespArrToObjByType(rest_op, "stringgroup", key)
	utils.AviLog.Debugf("The stringgroup object response %v", rest_op.Response)
	if resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find stringgroup obj in resp %v", key, rest_op.Response)
		return errors.New("stringgroup not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: StringGroup Name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: StringGroup Uuid not present in response %v", key, resp)
			continue
		}

		var keyvalues []*avimodels.KeyValue
		if resp["kv"] != nil {
			kv := resp["kv"].([]interface{})
			keyvalues = make([]*avimodels.KeyValue, len(kv))
			for i, v := range kv {
				if m, ok := v.(map[string]interface{}); ok {
					key := m["key"].(string)
					value, ok := m["value"].(string)
					if ok {
						keyvalues[i] = &avimodels.KeyValue{
							Key:   &key,
							Value: &value,
						}
					} else {
						keyvalues[i] = &avimodels.KeyValue{
							Key:   &key,
							Value: nil,
						}
					}
				} else {
					utils.AviLog.Warnf("key: %s, msg: StringGroup KV value is not a map: %v", key, v)
				}
			}
		}
		stringgroup_cache_obj := avicache.AviStringGroupCache{Name: name, Tenant: rest_op.Tenant,
			Uuid: uuid}

		// StringGroup should not have a checksum
		stringgroup_cache_obj.CloudConfigCksum = lib.StringGroupChecksum(keyvalues, nil, nil, false)

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.StringGroupCache.AviCacheAdd(k, &stringgroup_cache_obj)

		if strings.HasPrefix(name, lib.GetAKOUser()) {
			vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
			if ok {
				vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
				if found {
					vs_cache_obj.AddToStringGroupKeyCollection(k)
					utils.AviLog.Debugf("Modified the VS cache for stringgroup object. The cache now is :%v", utils.Stringify(vs_cache_obj))
				}

			} else {
				vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
				vs_cache_obj.AddToStringGroupKeyCollection(k)
				utils.AviLog.Debug(spew.Sprintf("Added VS cache key %v during string group update with val %v", vsKey,
					vs_cache_obj))
			}
		}

		utils.AviLog.Infof(spew.Sprintf("key: %s, msg: added StringGroup cache k %v val %v", key, k,
			stringgroup_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviStringGroupCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	sgKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	utils.AviLog.Debugf("Deleting StringGroup Cache: %s", sgKey)
	rest.cache.StringGroupCache.AviCacheDelete(sgKey)

	if strings.HasPrefix(rest_op.ObjName, lib.GetAKOUser()) {
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.RemoveFromStringGroupKeyCollection(sgKey)
			}
		}
	}
	return nil
}
