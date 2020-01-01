/*
 * [2013] - [2019] Avi Networks Incorporated
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
	"fmt"

	core "k8s.io/api/core/v1"
	extensions "k8s.io/api/extensions/v1beta1"

	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
	avicache "gitlab.eng.vmware.com/orion/akc/pkg/cache"
	"gitlab.eng.vmware.com/orion/akc/pkg/nodes"
	"gitlab.eng.vmware.com/orion/container-lib/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (rest *RestOperations) AviPoolBuild(pool_meta *nodes.AviPoolNode, cache_obj *avicache.AviPoolCache, key string) *utils.RestOp {
	name := pool_meta.Name
	cksum := pool_meta.CloudConfigCksum
	cksumString := fmt.Sprint(cksum)
	tenant := fmt.Sprintf("/api/tenant/?name=%s", pool_meta.Tenant)
	cr := utils.OSHIFT_K8S_CLOUD_CONNECTOR
	svc_mdata_json, _ := json.Marshal(&pool_meta.ServiceMetadata)
	svc_mdata := string(svc_mdata_json)
	cloudRef := "/api/cloud?name=" + utils.CloudName
	pool := avimodels.Pool{Name: &name, CloudConfigCksum: &cksumString,
		CreatedBy: &cr, TenantRef: &tenant, CloudRef: &cloudRef, ServiceMetadata: &svc_mdata}

	for _, server := range pool_meta.Servers {
		sip := server.Ip
		port := pool_meta.Port
		s := avimodels.Server{IP: &sip, Port: &port}
		if server.ServerNode != "" {
			sn := server.ServerNode
			s.ServerNode = &sn
		}
		pool.Servers = append(pool.Servers, &s)
	}

	var hm string
	if pool_meta.Protocol == "udp" {
		hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_UDP_HM)
	} else {
		hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_TCP_HM)
	}
	pool.HealthMonitorRefs = append(pool.HealthMonitorRefs, hm)

	macro := utils.AviRestObjMacro{ModelName: "Pool", Data: pool}

	// TODO Version should be latest from configmap
	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/pool/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pool,
			Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}

	} else {
		path = "/api/macro"
		rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
			Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}
	}

	utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: pool Restop %v K8sAviPoolMeta %v\n", key,
		utils.Stringify(rest_op), *pool_meta))
	return &rest_op
}

func (rest *RestOperations) AviPoolDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/pool/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "Pool", Version: utils.CtrlVersion}
	utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: pool DELETE Restop %v \n", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviPoolCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warning.Printf("key: %s, msg: rest_op has err or no reponse", key)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "pool", key)
	utils.AviLog.Warning.Printf("key: %s, msg: the pool object response %v", key, rest_op.Response)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warning.Printf("key: %s, msg: unable to find pool obj in resp %v", key, rest_op.Response)
		return errors.New("pool not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: Name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: uuid not present in response %v", key, resp)
			continue
		}
		cksum := resp["cloud_config_cksum"].(string)
		var svc_mdata interface{}
		var svc_mdata_map map[string]interface{}
		var svc_mdata_obj nodes.ServiceMetadataObj

		if err := json.Unmarshal([]byte(resp["service_metadata"].(string)),
			&svc_mdata); err == nil {
			svc_mdata_map, ok = svc_mdata.(map[string]interface{})
			if !ok {
				utils.AviLog.Warning.Printf("resp %v svc_mdata %T has invalid service_metadata type", resp, svc_mdata)
			} else {
				SvcMdataMapToObj(&svc_mdata_map, &svc_mdata_obj)
			}
		}
		pool_cache_obj := avicache.AviPoolCache{Name: name, Tenant: rest_op.Tenant,
			Uuid:             uuid,
			CloudConfigCksum: cksum}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.PoolCache.AviCacheAdd(k, &pool_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCache.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				if vs_cache_obj.PoolKeyCollection == nil {
					vs_cache_obj.PoolKeyCollection = []avicache.NamespaceName{k}
				} else {
					if !utils.HasElem(vs_cache_obj.PoolKeyCollection, k) {
						utils.AviLog.Info.Printf("key: %s, msg: Before adding pool collection %v and key :%v", key, vs_cache_obj.PoolKeyCollection, k)
						vs_cache_obj.PoolKeyCollection = append(vs_cache_obj.PoolKeyCollection, k)
					}
				}
				utils.AviLog.Info.Printf("key: %s, msg: modified the VS cache object for Pool Collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
				mClient := utils.GetInformers().ClientSet
				mIngress, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).Get(svc_mdata_obj.IngressName, metav1.GetOptions{})
				// Once the vsvip object is available - we should be able to update the hostname, for now just updating the vip
				lbIngress := core.LoadBalancerIngress{
					IP:       vs_cache_obj.Vip,
					Hostname: "ToBeUpdated",
				}
				mIngress.Status = extensions.IngressStatus{
					LoadBalancer: core.LoadBalancerStatus{
						Ingress: []core.LoadBalancerIngress{lbIngress},
					},
				}
				response, err := mClient.ExtensionsV1beta1().Ingresses(svc_mdata_obj.Namespace).UpdateStatus(mIngress)
				if err != nil {
					utils.AviLog.Error.Printf("key: %s, msg: there was an error in updating the ingress status: %v", key, err)
					return err
				}
				utils.AviLog.Info.Printf("key:%s, msg: Successfully updated the ingress status: %v", key, utils.Stringify(response))
			}
		} else {
			vs_cache_obj := avicache.AviVsCache{Name: vsKey.Name, Tenant: vsKey.Namespace,
				PoolKeyCollection: []avicache.NamespaceName{k}}
			rest.cache.VsCache.AviCacheAdd(vsKey, &vs_cache_obj)
			utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: added VS cache key during pool update %v val %v\n", key, vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Info.Print(spew.Sprintf("key: %s, msg: Added Pool cache k %v val %v\n", key, k,
			pool_cache_obj))
	}

	return nil
}

func SvcMdataMapToObj(svc_mdata_map *map[string]interface{}, svc_mdata *nodes.ServiceMetadataObj) {
	for k, val := range *svc_mdata_map {
		switch k {
		case "ingress_name":
			ingName, ok := val.(string)
			if ok {
				svc_mdata.IngressName = ingName
			} else {
				utils.AviLog.Warning.Printf("Incorrect type %T in svc_mdata_map %v", val, *svc_mdata_map)
			}
		case "namespace":
			namespace, ok := val.(string)
			if ok {
				svc_mdata.Namespace = namespace
			} else {
				utils.AviLog.Warning.Printf("Incorrect type %T in svc_mdata_map %v", val, *svc_mdata_map)
			}
		}
	}
}

func (rest *RestOperations) AviPoolCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	poolKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	utils.AviLog.Info.Printf("key: %s, msg: deleting pool with key :%s", key, poolKey)
	rest.cache.PoolCache.AviCacheDelete(key)
	// Delete the pool from the vs cache as well.
	vs_cache, ok := rest.cache.VsCache.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			utils.AviLog.Info.Printf("key: %s, msg: VS Pool key cache before deletion :%s", key, vs_cache_obj.PoolKeyCollection)
			vs_cache_obj.PoolKeyCollection = Remove(vs_cache_obj.PoolKeyCollection, poolKey)
			utils.AviLog.Info.Printf("key: %s, msg: VS Pool key cache after deletion :%s", key, vs_cache_obj.PoolKeyCollection)
		}
	}

	return nil
}
