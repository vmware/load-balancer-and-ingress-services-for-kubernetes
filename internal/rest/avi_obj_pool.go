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
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	avimodels "github.com/avinetworks/sdk/go/models"
	"github.com/davecgh/go-spew/spew"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
)

func (rest *RestOperations) AviPoolBuild(pool_meta *nodes.AviPoolNode, cache_obj *avicache.AviPoolCache, key string) *utils.RestOp {
	name := pool_meta.Name
	cksum := pool_meta.CloudConfigCksum
	cksumString := strconv.Itoa(int(cksum))
	tenant := fmt.Sprintf("/api/tenant/?name=%s", pool_meta.Tenant)
	cr := lib.AKOUser
	svc_mdata_json, _ := json.Marshal(&pool_meta.ServiceMetadata)
	svc_mdata := string(svc_mdata_json)
	cloudRef := "/api/cloud?name=" + utils.CloudName
	vrfContextRef := "/api/vrfcontext?name=" + pool_meta.VrfContext

	placementNetworks := []*avimodels.PlacementNetwork{}

	nodeNetworkMap, _ := lib.GetNodeNetworkMap()

	// set pool placement network if node network details are present and cloud type is CLOUD_VCENTER
	if len(nodeNetworkMap) != 0 && lib.GetCloudType() == lib.CLOUD_VCENTER {
		for network, cidrs := range nodeNetworkMap {
			for _, cidr := range cidrs {
				placementNetwork := avimodels.PlacementNetwork{}
				networkRef := "/api/network/?name=" + network
				placementNetwork.NetworkRef = &networkRef
				_, ipnet, err := net.ParseCIDR(cidr)
				if err != nil {
					utils.AviLog.Warnf("The value of CIDR couldn't be parsed. Failed with error: %v.", err.Error())
					break
				}
				addr := ipnet.IP.String()
				atype := "V4"
				if !utils.IsV4(addr) {
					atype = "V6"
				}

				mask := strings.Split(cidr, "/")[1]
				intCidr, err := strconv.ParseInt(mask, 10, 32)
				if err != nil {
					utils.AviLog.Warnf("The value of CIDR couldn't be converted to int32.")
					break
				}
				int32Cidr := int32(intCidr)

				placementNetwork.Subnet = &avimodels.IPAddrPrefix{IPAddr: &avimodels.IPAddr{Addr: &addr, Type: &atype}, Mask: &int32Cidr}
				placementNetworks = append(placementNetworks, &placementNetwork)
			}

		}
	}

	pool := avimodels.Pool{
		Name:              &name,
		CloudConfigCksum:  &cksumString,
		CreatedBy:         &cr,
		TenantRef:         &tenant,
		CloudRef:          &cloudRef,
		ServiceMetadata:   &svc_mdata,
		VrfRef:            &vrfContextRef,
		SniEnabled:        &pool_meta.SniEnabled,
		SslProfileRef:     &pool_meta.SslProfileRef,
		PlacementNetworks: placementNetworks,
	}

	if pool_meta.PkiProfile != nil {
		pkiProfileName := "/api/pkiprofile?name=" + pool_meta.PkiProfile.Name
		pool.PkiProfileRef = &pkiProfileName
	}

	// there are defaults set by the Avi controller internally
	if pool_meta.LbAlgorithm != "" {
		pool.LbAlgorithm = &pool_meta.LbAlgorithm
	}
	if pool_meta.LbAlgorithmHash != "" {
		pool.LbAlgorithmHash = &pool_meta.LbAlgorithmHash
		if *pool.LbAlgorithmHash == lib.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER {
			pool.LbAlgorithmConsistentHashHdr = &pool_meta.LbAlgoHostHeader
		}
	}

	for i, server := range pool_meta.Servers {
		port := pool_meta.Port
		sip := server.Ip
		if server.Port != 0 {
			port = pool_meta.Servers[i].Port
		}
		s := avimodels.Server{IP: &sip, Port: &port}
		if server.ServerNode != "" {
			sn := server.ServerNode
			s.ServerNode = &sn
		}
		pool.Servers = append(pool.Servers, &s)
	}

	// overwrite with healthmonitors provided by CRD
	if len(pool_meta.HealthMonitors) > 0 {
		pool.HealthMonitorRefs = pool_meta.HealthMonitors
	} else {
		var hm string
		if pool_meta.Protocol == utils.UDP {
			hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_UDP_HM)
		} else {
			hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_TCP_HM)
		}
		pool.HealthMonitorRefs = append(pool.HealthMonitorRefs, hm)
	}

	macro := utils.AviRestObjMacro{ModelName: "Pool", Data: pool}

	// TODO Version should be latest from configmap
	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/pool/" + cache_obj.Uuid
		rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pool,
			Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}
	} else {
		// Patch an existing pool if it exists in the cache but not associated with this VS.
		pool_key := avicache.NamespaceName{Namespace: pool_meta.Tenant, Name: name}
		pool_cache, ok := rest.cache.PoolCache.AviCacheGet(pool_key)
		if ok {
			pool_cache_obj, _ := pool_cache.(*avicache.AviPoolCache)
			path = "/api/pool/" + pool_cache_obj.Uuid
			rest_op = utils.RestOp{Path: path, Method: utils.RestPut, Obj: pool,
				Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}
		} else {
			path = "/api/macro"
			rest_op = utils.RestOp{Path: path, Method: utils.RestPost, Obj: macro,
				Tenant: pool_meta.Tenant, Model: "Pool", Version: utils.CtrlVersion}
		}
	}

	utils.AviLog.Debug(spew.Sprintf("key: %s, msg: pool Restop %v K8sAviPoolMeta %v\n", key,
		utils.Stringify(rest_op), *pool_meta))
	return &rest_op
}

func (rest *RestOperations) AviPoolDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/pool/" + uuid
	rest_op := utils.RestOp{Path: path, Method: "DELETE",
		Tenant: tenant, Model: "Pool", Version: utils.CtrlVersion}
	utils.AviLog.Info(spew.Sprintf("key: %s, msg: pool DELETE Restop %v \n", key,
		utils.Stringify(rest_op)))
	return &rest_op
}

func (rest *RestOperations) AviPoolCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for POOL, err: %s, response: %s", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems, ok := RestRespArrToObjByType(rest_op, "pool", key)
	utils.AviLog.Debugf("key: %s, msg: the pool object response %v", key, rest_op.Response)
	if ok != nil || resp_elems == nil {
		utils.AviLog.Warnf("key: %s, msg: unable to find pool obj in resp %v", key, rest_op.Response)
		return errors.New("pool not found")
	}

	for _, resp := range resp_elems {
		name, ok := resp["name"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: Name not present in response %v", key, resp)
			continue
		}

		uuid, ok := resp["uuid"].(string)
		if !ok {
			utils.AviLog.Warnf("key: %s, msg: uuid not present in response %v", key, resp)
			continue
		}
		cksum := resp["cloud_config_cksum"].(string)

		var svc_mdata_obj avicache.ServiceMetadataObj
		if resp["service_metadata"] != nil {
			if err := json.Unmarshal([]byte(resp["service_metadata"].(string)),
				&svc_mdata_obj); err != nil {
				utils.AviLog.Warnf("Error parsing service metadata :%v", err)
			}
		}

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

		var pkiKey avicache.NamespaceName
		if pkiprof, ok := resp["pki_profile_ref"]; ok && pkiprof != "" {
			pkiUuid := avicache.ExtractUuid(pkiprof.(string), "pkiprofile-.*.#")
			pkiName, foundPki := rest.cache.PKIProfileCache.AviCacheGetNameByUuid(pkiUuid)
			if foundPki {
				pkiKey = avicache.NamespaceName{Namespace: lib.GetTenant(), Name: pkiName.(string)}
			}
		}

		pool_cache_obj := avicache.AviPoolCache{
			Name:                 name,
			Tenant:               rest_op.Tenant,
			Uuid:                 uuid,
			CloudConfigCksum:     cksum,
			ServiceMetadataObj:   svc_mdata_obj,
			PkiProfileCollection: pkiKey,
			LastModified:         lastModifiedStr,
		}
		if lastModifiedStr == "" {
			pool_cache_obj.InvalidData = true
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		rest.cache.PoolCache.AviCacheAdd(k, &pool_cache_obj)
		// Update the VS object
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToPoolKeyCollection(k)
				utils.AviLog.Debugf("key: %s, msg: modified the VS cache object for Pool Collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
				if svc_mdata_obj.Namespace != "" {
					status.UpdateRouteIngressStatus([]status.UpdateStatusOptions{{
						Vip:             vs_cache_obj.Vip,
						ServiceMetadata: svc_mdata_obj,
						Key:             key,
					}}, false)
				}
			}
		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToPoolKeyCollection(k)
			utils.AviLog.Debug(spew.Sprintf("key: %s, msg: added VS cache key during pool update %v val %v\n", key, vsKey,
				vs_cache_obj))
		}
		utils.AviLog.Info(spew.Sprintf("key: %s, msg: Added Pool cache k %v val %v\n", key, k,
			pool_cache_obj))
	}

	return nil
}

func (rest *RestOperations) AviPoolCacheDel(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	// Delete the pool from the vs cache as well.
	poolKey := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: rest_op.ObjName}
	vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
	if ok {
		vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
		if found {
			utils.AviLog.Debugf("key: %s, msg: VsKey: %s, VS Pool key cache before deletion :%s", key, vsKey, vs_cache_obj.PoolKeyCollection)
			vs_cache_obj.RemoveFromPoolKeyCollection(poolKey)
			utils.AviLog.Infof("key: %s, msg: VS Pool key cache after deletion :%s", key, vs_cache_obj.PoolKeyCollection)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: deleting pool with key: %s", key, poolKey)
	// Fetch the pool's cache data and obtain the service metadata
	rest.DeletePoolIngressStatus(poolKey, false, key)
	// Now delete the cache.
	rest.cache.PoolCache.AviCacheDelete(poolKey)

	return nil
}

func (rest *RestOperations) DeletePoolIngressStatus(poolKey avicache.NamespaceName, isVSDelete bool, key string) {
	pool_cache, found := rest.cache.PoolCache.AviCacheGet(poolKey)
	if found {
		pool_cache_obj, success := pool_cache.(*avicache.AviPoolCache)
		if success {
			if pool_cache_obj.ServiceMetadataObj.IngressName != "" {
				// SNI VSes use the VS object metadata, delete ingress status for others
				err := status.DeleteRouteIngressStatus(pool_cache_obj.ServiceMetadataObj, isVSDelete, key)
				if k8serror.IsNotFound(err) {
					// Just log and get away
					utils.AviLog.Infof("key: %s, msg: ingress already deleted, nothing to update in status", key)
				}
			}
		}
	}
}
