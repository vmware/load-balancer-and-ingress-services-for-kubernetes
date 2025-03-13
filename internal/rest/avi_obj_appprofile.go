/*
 * Copyright 2020-2021 VMware, Inc.
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

	"github.com/davecgh/go-spew/spew"
	avimodels "github.com/vmware/alb-sdk/go/models"
)

func (rest *RestOperations) AviAppProfileBuild(node *nodes.AviVSVIPNode, cacheObj *avicache.AviAppProfileCache, key string) (*utils.RestOp, error) {
	if lib.CheckObjectNameLength(node.Name, lib.ApplicationProfile) {
		utils.AviLog.Warnf("Not processing Application Profile object")
		return nil
	}
	name := node.Name
	tenant := fmt.Sprintf("/api/tenant/?name=%s", lib.GetEscapedValue(node.Tenant))
	appType := node.Type
	createBy := lib.GetAKOUser()
	cksum := node.CloudConfigCksum
	cksumstr := strconv.Itoa(int(cksum))
	tcpAppProfile := avimodels.TCPApplicationProfile{
		ProxyProtocolEnabled: node.EnableProxyProtocol,
	}

	appProfile := avimodels.ApplicationProfile{
		Name:             &name,
		TenantRef:        &tenant,
		Type:             &appType,
		CreatedBy:        &createBy,
		TcpAppProfile:    &tcpAppProfile,
		CloudConfigCksum: &cksumstr,
	}

	var rest_op utils.RestOp
	var path string

	if cacheObj != nil {
		path = "/api/applicationprofile/" + cacheObj.Uuid
		rest_op = utils.RestOp{
			ObjName: name,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     appProfile,
			Tenant:  node.Tenant,
			Model:   "ApplicationProfile",
		}
	} else {
		key := avicache.NamespaceName{Namespace: node.Tenant, Name: name}
		cache, ok := rest.cache.AppProfileCache.AviCacheGet(key)
		if ok {
			appCacheObj, _ := cache.(*avicache.AviAppProfileCache)
			path = "/api/sslkeyandcertificate/" + appCacheObj.Uuid
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     appProfile,
				Tenant:  node.Tenant,
				Model:   "ApplicationProfile",
			}
		} else {
			path = "/api/applicationprofile"
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     appProfile,
				Tenant:  node.Tenant,
				Model:   "ApplicationProfile",
			}
		}
	}
	utils.AviLog.Debugf(spew.Sprintf("key: %s, msg: app profile Restop %v ApplicationProfileData %v", key,
		utils.Stringify(rest_op), *node))
	return &rest_op, nil
}

func (rest *RestOperations) AviAppProfileDel(uuid, tenant, key string) *utils.RestOp {
	path := "/api/applicationprofile/" + uuid
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "ApplicationProfile",
	}
	utils.AviLog.Infof(spew.Sprintf("key: %s, msg: AppProfile DELETE Restop %v ", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviAppProfileCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for app profile err: %v, response: %v", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored app profile rest_op")
	}

	resp_elems := rest.restOperator.RestRespArrToObjByType(rest_op, "applicationprofile", key)
	if resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find app profile obj in resp %v", key, rest_op.Response)
		return errors.New("app profile not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: app profile name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: app profile Uuid not present in response %v", key, resp)
			continue
		}

		appType, ok := resp["type"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: app profile type not present in response %v", key, resp)
			continue
		}

		tap, ok := resp["tcp_app_profile"].(map[string]interface{})
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: tcp app profile not present in response %v", key, resp)
			continue
		}
		ppe := tap["proxy_protocol_enabled"].(bool)
		cksum := resp["cloud_config_cksum"].(string)
		appProfileCache := avicache.AviAppProfileCache{
			Uuid:                uuid,
			Name:                name,
			Tenant:              rest_op.Tenant,
			Type:                appType,
			EnableProxyProtocol: ppe,
			CloudConfigCksum:    cksum,
		}
		appKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.AppProfileCache.AviCacheAdd(appKey, &appProfileCache)
		vsCache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vsCacheObj, found := vsCache.(*avicache.AviVsCache)
			if found {
				vsCacheObj.AddToHTTPKeyCollection(appKey)
				utils.AviLog.Debugf("Modified the VS cache for app profile object. The cache now is :%v",
					utils.Stringify(vsCacheObj))
			}
		} else {
			vsCacheObj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vsCacheObj.AddToHTTPKeyCollection(appKey)
			utils.AviLog.Debug(spew.Sprintf("Added VS cache key during app profile update %v val %v", vsKey,
				vsCacheObj))
		}
		utils.AviLog.Debug(spew.Sprintf("Added App Profile cache k %v val %v", appKey, appProfileCache))
	}
	return nil
}

func (rest *RestOperations) AviAppProfileCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string, del bool) error {
	appKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	if del {
		rest.cache.AppProfileCache.AviCacheDelete(appKey)
	}
	vsCache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if ok {
		vsCacheObj, found := vsCache.(*avicache.AviVsCache)
		if found {
			vsCacheObj.RemoveFromAppProfileCollection(appKey)
		}
	}
	return nil
}
