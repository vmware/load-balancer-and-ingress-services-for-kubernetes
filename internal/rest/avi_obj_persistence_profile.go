/*
 * Copyright 2024-2025 VMware, Inc.
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

	avimodels "github.com/vmware/alb-sdk/go/models"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/davecgh/go-spew/spew"
)

func (rest *RestOperations) AviPersistenceProfileBuild(appPersProfileNode *nodes.AviApplicationPersistenceProfileNode, cacheObj *avicache.AviPersistenceProfileCache) *utils.RestOp {
	if appPersProfileNode == nil {
		utils.AviLog.Debugf("ApplicationPersistenceProfileNode is nil")
		return nil
	}

	if lib.CheckObjectNameLength(appPersProfileNode.Name, "") {
		utils.AviLog.Warnf("Not processing ApplicationPersistenceProfile object %s due to name length limit", appPersProfileNode.Name)
		return nil
	}

	name := appPersProfileNode.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", lib.GetEscapedValue(appPersProfileNode.Tenant))

	appPersProfile := avimodels.ApplicationPersistenceProfile{
		PersistenceType: &appPersProfileNode.PersistenceType,
		Name:            &name,
		TenantRef:       &tenant,
		Markers:         lib.GetAllMarkers(appPersProfileNode.AviMarkers),
	}
	switch appPersProfileNode.PersistenceType {
	case "PERSISTENCE_TYPE_HTTP_COOKIE":
		appPersProfile.HTTPCookiePersistenceProfile = &avimodels.HTTPCookiePersistenceProfile{
			CookieName:         &appPersProfileNode.HTTPCookiePersistenceProfile.CookieName,
			Timeout:            appPersProfileNode.HTTPCookiePersistenceProfile.Timeout,
			IsPersistentCookie: appPersProfileNode.HTTPCookiePersistenceProfile.IsPersistentCookie,
		}
	default:
		utils.AviLog.Warnf("Unknown persistence type: %s", appPersProfileNode.PersistenceType)
		return nil
	}
	var path string
	var restOp utils.RestOp

	if cacheObj != nil {
		path = "/api/applicationpersistenceprofile/" + cacheObj.Uuid
		restOp = utils.RestOp{
			ObjName: name,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     appPersProfile,
			Tenant:  appPersProfileNode.Tenant,
			Model:   "ApplicationPersistenceProfile",
		}
	} else {
		// Check if it exists in the cache but not associated with this specific context (e.g., pool)
		appPersProfileKey := avicache.NamespaceName{Namespace: appPersProfileNode.Tenant, Name: name}
		existingCache, ok := rest.cache.AppPersProfileCache.AviCacheGet(appPersProfileKey)
		if ok {
			existingCacheObj, _ := existingCache.(*avicache.AviPersistenceProfileCache)
			path = "/api/applicationpersistenceprofile/" + existingCacheObj.Uuid
			restOp = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     appPersProfile,
				Tenant:  appPersProfileNode.Tenant,
				Model:   "ApplicationPersistenceProfile",
			}
		} else {
			path = "/api/applicationpersistenceprofile"
			restOp = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     appPersProfile,
				Tenant:  appPersProfileNode.Tenant,
				Model:   "ApplicationPersistenceProfile",
			}
		}
	}

	utils.AviLog.Debugf(spew.Sprintf("ApplicationPersistenceProfile RestOp: %v, Object: %v", utils.Stringify(restOp), utils.Stringify(appPersProfile)))
	return &restOp
}

func (rest *RestOperations) AviPersistenceProfileDel(uuid string, tenant string) *utils.RestOp {
	path := "/api/applicationpersistenceprofile/" + uuid
	restOp := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "ApplicationPersistenceProfile",
	}
	utils.AviLog.Infof(spew.Sprintf("ApplicationPersistenceProfile DELETE RestOp: %v", utils.Stringify(restOp)))
	return &restOp
}

func (rest *RestOperations) AviPersistenceProfileCacheAdd(restOp *utils.RestOp, poolKey avicache.NamespaceName, key string) error {
	if restOp.Err != nil || restOp.Response == nil {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for ApplicationPersistenceProfile, err: %v, response: %v", key, restOp.Err, restOp.Response)
		return errors.New("errored rest_op")
	}

	respElems := rest.restOperator.RestRespArrToObjByType(restOp, "applicationpersistenceprofile", key)
	if respElems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find ApplicationPersistenceProfile obj in resp %v", key, restOp.Response)
		return errors.New("ApplicationPersistenceProfile not found")
	}

	for _, resp := range respElems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Name not present in response %v for ApplicationPersistenceProfile", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Uuid not present in response %v for ApplicationPersistenceProfile", key, resp)
			continue
		}

		var lastModifiedStr string
		if lastModifiedIntf, ok := resp["_last_modified"]; ok {
			lastModifiedStr, _ = lastModifiedIntf.(string)
		} else {
			utils.AviLog.Warnf("key: %s, msg: _last_modified not present in response %v for ApplicationPersistenceProfile %s", key, resp, name)
		}

		var appPersProfileModel avimodels.ApplicationPersistenceProfile
		switch restOp.Obj.(type) {
		case utils.AviRestObjMacro:
			appPersProfileModel = restOp.Obj.(utils.AviRestObjMacro).Data.(avimodels.ApplicationPersistenceProfile)
		case avimodels.ApplicationPersistenceProfile:
			appPersProfileModel = restOp.Obj.(avimodels.ApplicationPersistenceProfile)
		default:
			utils.AviLog.Warnf("key: %s, msg: Unknown object type for ApplicationPersistenceProfile %v", key, restOp.Obj)
		}

		appPersCacheObj := avicache.AviPersistenceProfileCache{
			Name:         name,
			Tenant:       restOp.Tenant,
			Uuid:         uuid,
			LastModified: lastModifiedStr,
		}
		if appPersProfileModel.PersistenceType != nil {
			appPersCacheObj.Type = *appPersProfileModel.PersistenceType
		}
		emptyIngestionMarkers := utils.AviObjectMarkers{}
		chksum := lib.PersistenceProfileChecksum(name, *appPersProfileModel.PersistenceType, emptyIngestionMarkers, appPersProfileModel.Markers, true)
		if appPersProfileModel.HTTPCookiePersistenceProfile != nil {
			chksum += lib.HTTPCookiePersistenceProfileChecksum(*appPersProfileModel.HTTPCookiePersistenceProfile.CookieName, appPersProfileModel.HTTPCookiePersistenceProfile.Timeout, appPersProfileModel.HTTPCookiePersistenceProfile.IsPersistentCookie)
		}
		appPersCacheObj.CloudConfigCksum = chksum

		if lastModifiedStr == "" {
			appPersCacheObj.InvalidData = true
		}

		k := avicache.NamespaceName{Namespace: restOp.Tenant, Name: name}
		rest.cache.AppPersProfileCache.AviCacheAdd(k, &appPersCacheObj)
		// Update the Pool object
		if poolKey != (avicache.NamespaceName{}) {
			pool_cache, ok := rest.cache.PoolCache.AviCacheGet(poolKey)
			if ok {
				pool_cache_obj, found := pool_cache.(*avicache.AviPoolCache)
				if found {
					utils.AviLog.Debugf("The Pool cache before modification by ApplicationPersistenceProfile is :%v", utils.Stringify(pool_cache_obj))
					pool_cache_obj.PersistenceProfile = k
					utils.AviLog.Infof("Modified the Pool cache object for ApplicationPersistenceProfile Collection. The cache now is :%v", utils.Stringify(pool_cache_obj))
				}
			} else {
				pool_cache_obj := rest.cache.PoolCache.AviCacheAddPool(poolKey)
				pool_cache_obj.PersistenceProfile = k
				utils.AviLog.Infof(spew.Sprintf("Added Pool cache key during ApplicationPersistenceProfile update %v val %v", poolKey,
					pool_cache_obj))
			}
			utils.AviLog.Infof("key: %s, msg: Added ApplicationPersistenceProfile cache k %v val %v", key, k, utils.Stringify(appPersCacheObj))
		}
	}
	return nil
}

func (rest *RestOperations) AviPersistenceProfileCacheDel(restOp *utils.RestOp, poolKey avicache.NamespaceName, key string) error {
	appPersProfileKey := avicache.NamespaceName{Namespace: restOp.Tenant, Name: restOp.ObjName}
	rest.cache.AppPersProfileCache.AviCacheDelete(appPersProfileKey)
	if poolKey != (avicache.NamespaceName{}) {
		poolCache, ok := rest.cache.PoolCache.AviCacheGet(poolKey)
		if ok {
			if poolCacheObj, found := poolCache.(*avicache.AviPoolCache); found {
				poolCacheObj.PersistenceProfile = avicache.NamespaceName{}
			}
		}
	}
	utils.AviLog.Infof("key: %s, msg: Deleted ApplicationPersistenceProfile cache k %v", key, appPersProfileKey)
	return nil
}
