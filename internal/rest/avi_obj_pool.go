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

	"google.golang.org/protobuf/proto"

	avicache "github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/cache"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/lib"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/nodes"
	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/internal/status"

	"github.com/vmware/load-balancer-and-ingress-services-for-kubernetes/pkg/utils"

	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/copier"
	avimodels "github.com/vmware/alb-sdk/go/models"
	k8net "k8s.io/utils/net"
)

func (rest *RestOperations) AviPoolBuild(pool_meta *nodes.AviPoolNode, cache_obj *avicache.AviPoolCache, key string) *utils.RestOp {
	if lib.CheckObjectNameLength(pool_meta.Name, lib.Pool) {
		utils.AviLog.Warnf("key: %s not processing pool object", key)
		return nil
	}
	name := pool_meta.Name
	cksum := pool_meta.CloudConfigCksum
	cksumString := strconv.Itoa(int(cksum))
	tenant := fmt.Sprintf("/api/tenant/?name=%s", pool_meta.Tenant)
	cr := lib.AKOUser
	svc_mdata_json, _ := json.Marshal(&pool_meta.ServiceMetadata)
	svc_mdata := string(svc_mdata_json)
	cloudRef := lib.GetCloudRef(lib.GetTenant())
	placementNetworks := []*avimodels.PlacementNetwork{}

	// set pool placement network if node network details are present and cloud type is CLOUD_VCENTER or CLOUD_NSXT (vlan)
	if len(pool_meta.NetworkPlacementSettings) != 0 && lib.IsNodeNetworkAllowedCloud() {
		for network, nwMap := range pool_meta.NetworkPlacementSettings {
			for _, cidr := range nwMap.Cidrs {
				_, ipnet, err := net.ParseCIDR(cidr)
				if err != nil {
					utils.AviLog.Warnf("The value of CIDR couldn't be parsed. Failed with error: %v.", err.Error())
					break
				}
				addr := ipnet.IP.String()
				atype := "V4"
				if k8net.IsIPv6CIDR(ipnet) {
					atype = "V6"
				}

				mask := strings.Split(cidr, "/")[1]
				intCidr, err := strconv.ParseInt(mask, 10, 32)
				if err != nil {
					utils.AviLog.Warnf("The value of CIDR couldn't be converted to int32.")
					break
				}
				int32Cidr := int32(intCidr)
				networkRef := "/api/network/?name=" + network
				if nwMap.NetworkUUID != "" {
					networkRef = "/api/network/" + nwMap.NetworkUUID
				}
				utils.AviLog.Debugf("Pool: %s, Network ref for pool placement setting is: %s", name, networkRef)
				placementNetworks = append(placementNetworks, &avimodels.PlacementNetwork{
					NetworkRef: proto.String(networkRef),
					Subnet: &avimodels.IPAddrPrefix{
						IPAddr: &avimodels.IPAddr{
							Addr: &addr,
							Type: &atype,
						},
						Mask: &int32Cidr,
					},
				})
			}

		}
	}

	pool := avimodels.Pool{
		Name:                    &name,
		CloudConfigCksum:        &cksumString,
		CreatedBy:               &cr,
		TenantRef:               &tenant,
		CloudRef:                &cloudRef,
		ServiceMetadata:         &svc_mdata,
		SniEnabled:              &pool_meta.SniEnabled,
		SslProfileRef:           pool_meta.SslProfileRef,
		SslKeyAndCertificateRef: pool_meta.SslKeyAndCertificateRef,
		PkiProfileRef:           pool_meta.PkiProfileRef,
		PlacementNetworks:       placementNetworks,
	}

	var vrfContextRef string
	if pool_meta.VrfContext != "" {
		vrfContextRef = "/api/vrfcontext?name=" + pool_meta.VrfContext
		pool.VrfRef = &vrfContextRef
	}

	if pool_meta.T1Lr != "" {
		pool.Tier1Lr = &pool_meta.T1Lr
	}

	if !pool_meta.AttachedWithSharedVS {
		pool.Markers = lib.GetAllMarkers(pool_meta.AviMarkers)
	} else {
		pool.Markers = lib.GetMarkers()
	}

	if pool_meta.PkiProfile != nil {
		pkiProfileName := "/api/pkiprofile?name=" + pool_meta.PkiProfile.Name
		pool.PkiProfileRef = &pkiProfileName
	}

	// there are defaults set by the Avi controller internally
	if pool_meta.LbAlgorithm != nil {
		pool.LbAlgorithm = pool_meta.LbAlgorithm
	}
	if pool_meta.LbAlgorithmHash != nil {
		pool.LbAlgorithmHash = pool_meta.LbAlgorithmHash
		if *pool.LbAlgorithmHash == lib.LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER {
			pool.LbAlgorithmConsistentHashHdr = pool_meta.LbAlgorithmConsistentHashHdr
		}
	}

	if pool_meta.ApplicationPersistenceProfileRef != nil {
		pool.ApplicationPersistenceProfileRef = pool_meta.ApplicationPersistenceProfileRef
	}

	if pool_meta.EnableHttp2 != nil {
		pool.EnableHttp2 = pool_meta.EnableHttp2
	}

	for i, server := range pool_meta.Servers {
		port := pool_meta.Port
		sip := server.Ip
		if server.Port != 0 {
			port = pool_meta.Servers[i].Port
		}
		uuid := fmt.Sprintf("%s:%d", *sip.Addr, port)

		s := avimodels.Server{IP: &sip, Port: &port, ExternalUUID: &uuid, Enabled: server.Enabled}
		if server.ServerNode != "" {
			sn := server.ServerNode
			s.ServerNode = &sn
		}
		pool.Servers = append(pool.Servers, &s)
	}

	// overwrite with healthmonitors provided by CRD
	if len(pool_meta.HealthMonitorRefs) > 0 {
		pool.HealthMonitorRefs = pool_meta.HealthMonitorRefs
	} else {
		var hm string
		if pool_meta.Protocol == utils.UDP {
			hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_UDP_HM)
		} else if pool_meta.Protocol == utils.SCTP {
			hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_SCTP_HM)
		} else {
			hm = fmt.Sprintf("/api/healthmonitor/?name=%s", utils.AVI_DEFAULT_TCP_HM)
		}
		pool.HealthMonitorRefs = append(pool.HealthMonitorRefs, hm)
	}

	if err := copier.CopyWithOption(&pool, &pool_meta.AviPoolGeneratedFields, copier.Option{IgnoreEmpty: true}); err != nil {
		utils.AviLog.Warnf("key: %s, msg: unable to set few parameters in the Pool, err: %v", key, err)
	}

	// TODO Version should be latest from configmap
	var path string
	var rest_op utils.RestOp
	if cache_obj != nil {
		path = "/api/pool/" + cache_obj.Uuid
		rest_op = utils.RestOp{
			ObjName: name,
			Path:    path,
			Method:  utils.RestPut,
			Obj:     pool,
			Tenant:  pool_meta.Tenant,
			Model:   "Pool",
		}
	} else {
		// Patch an existing pool if it exists in the cache but not associated with this VS.
		pool_key := avicache.NamespaceName{Namespace: pool_meta.Tenant, Name: name}
		pool_cache, ok := rest.cache.PoolCache.AviCacheGet(pool_key)
		if ok {
			pool_cache_obj, _ := pool_cache.(*avicache.AviPoolCache)
			path = "/api/pool/" + pool_cache_obj.Uuid
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPut,
				Obj:     pool,
				Tenant:  pool_meta.Tenant,
				Model:   "Pool",
			}
		} else {
			path = "/api/pool/"
			rest_op = utils.RestOp{
				ObjName: name,
				Path:    path,
				Method:  utils.RestPost,
				Obj:     pool,
				Tenant:  pool_meta.Tenant,
				Model:   "Pool",
			}
		}
	}

	utils.AviLog.Debug(spew.Sprintf("key: %s, msg: pool Restop %v K8sAviPoolMeta %v", key,
		utils.Stringify(rest_op), *pool_meta))
	return &rest_op
}

func (rest *RestOperations) AviPoolDel(uuid string, tenant string, key string) *utils.RestOp {
	path := "/api/pool/" + uuid
	rest_op := utils.RestOp{
		Path:   path,
		Method: "DELETE",
		Tenant: tenant,
		Model:  "Pool",
	}
	utils.AviLog.Infof("key: %s, msg: pool DELETE Restop %v ", key, utils.Stringify(rest_op))
	return &rest_op
}

func (rest *RestOperations) AviPoolCacheAdd(rest_op *utils.RestOp, vsKey avicache.NamespaceName, key string) error {
	if (rest_op.Err != nil) || (rest_op.Response == nil) {
		utils.AviLog.Warnf("key: %s, rest_op has err or no response for POOL, err: %v, response: %v", key, rest_op.Err, rest_op.Response)
		return errors.New("Errored rest_op")
	}

	resp_elems := rest.restOperator.RestRespArrToObjByType(rest_op, "pool", key)
	utils.AviLog.Debugf("key: %s, msg: the pool object response %v", key, rest_op.Response)
	if resp_elems == nil {
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

		var svc_mdata_obj lib.ServiceMetadataObj
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
			pkiUuid := avicache.ExtractUUID(pkiprof.(string), "pkiprofile-.*.#")
			pkiName, foundPki := rest.cache.PKIProfileCache.AviCacheGetNameByUuid(pkiUuid)
			if foundPki {
				pkiKey = avicache.NamespaceName{Namespace: lib.GetTenant(), Name: pkiName.(string)}
			}
		}

		k := avicache.NamespaceName{Namespace: rest_op.Tenant, Name: name}
		oldCacheServiceMetadataCRD := lib.CRDMetadata{}
		if poolCache, ok := rest.cache.PoolCache.AviCacheGet(k); ok {
			if poolCacheObj, found := poolCache.(*avicache.AviPoolCache); found {
				oldCacheServiceMetadataCRD = poolCacheObj.ServiceMetadataObj.CRDStatus
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

		rest.cache.PoolCache.AviCacheAdd(k, &pool_cache_obj)
		if (oldCacheServiceMetadataCRD != lib.CRDMetadata{}) {
			status.HttpRuleEventBroadcast(k.Name, oldCacheServiceMetadataCRD, svc_mdata_obj.CRDStatus)
		}

		// Update the VS object
		vs_cache, ok := rest.cache.VsCacheMeta.AviCacheGet(vsKey)
		if ok {
			vs_cache_obj, found := vs_cache.(*avicache.AviVsCache)
			if found {
				vs_cache_obj.AddToPoolKeyCollection(k)
				utils.AviLog.Debugf("key: %s, msg: modified the VS cache object for Pool Collection. The cache now is :%v", key, utils.Stringify(vs_cache_obj))
				IPAddrs := rest.GetIPAddrsFromCache(vs_cache_obj)
				if len(IPAddrs) == 0 {
					utils.AviLog.Warnf("key: %s, msg: Unable to find VIP corresponding to Pool %s vsCache %v", key, pool_cache_obj.Name, utils.Stringify(vs_cache_obj))
				} else {
					switch pool_cache_obj.ServiceMetadataObj.ServiceMetadataMapping("Pool") {
					case lib.GatewayPool:
						updateOptions := status.UpdateOptions{
							Vip:                IPAddrs,
							ServiceMetadata:    svc_mdata_obj,
							Key:                key,
							VirtualServiceUUID: vs_cache_obj.Uuid,
							VSName:             vs_cache_obj.Name,
							Tenant:             vs_cache_obj.Tenant,
						}
						statusOption := status.StatusOptions{
							ObjType: utils.L4LBService,
							Op:      lib.UpdateStatus,
							Key:     key,
							Options: &updateOptions,
						}
						utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.NamespaceServiceName[0], utils.Stringify(statusOption))
						status.PublishToStatusQueue(updateOptions.ServiceMetadata.NamespaceServiceName[0], statusOption)
					case lib.SNIInsecureOrEVHPool:
						if vs_cache_obj.Uuid != "" {
							updateOptions := status.UpdateOptions{
								Vip:                IPAddrs,
								ServiceMetadata:    svc_mdata_obj,
								Key:                key,
								VirtualServiceUUID: vs_cache_obj.Uuid,
								VSName:             vs_cache_obj.Name,
								Tenant:             vs_cache_obj.Tenant,
							}
							statusOption := status.StatusOptions{
								ObjType: utils.Ingress,
								Op:      lib.UpdateStatus,
								Options: &updateOptions,
							}
							if utils.GetInformers().RouteInformer != nil {
								statusOption.ObjType = utils.OshiftRoute
							}
							if pool_cache_obj.ServiceMetadataObj.IsMCIIngress {
								statusOption.ObjType = lib.MultiClusterIngress
							}
							utils.AviLog.Debugf("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.HostNames[0], utils.Stringify(statusOption))
							status.PublishToStatusQueue(updateOptions.ServiceMetadata.HostNames[0], statusOption)
						}
					}
				}
			}
		} else {
			vs_cache_obj := rest.cache.VsCacheMeta.AviCacheAddVS(vsKey)
			vs_cache_obj.AddToPoolKeyCollection(k)
			utils.AviLog.Debugf("key: %s, msg: added VS cache key during pool update %v val %v", key, vsKey, utils.Stringify(vs_cache_obj))
		}
		utils.AviLog.Infof("key: %s, msg: Added Pool cache k %v val %v", key, k, utils.Stringify(pool_cache_obj))
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
			vs_cache_obj.RemoveFromPoolKeyCollection(poolKey)
			utils.AviLog.Debugf("key: %s, msg: VS Pool key cache after deletion: %s", key, vs_cache_obj.PoolKeyCollection)
			rest.DeletePoolIngressStatus(poolKey, false, vs_cache_obj.Name, key)
		}
	}
	utils.AviLog.Debugf("key: %s, msg: deleting pool with key: %s", key, poolKey)
	cacheServiceMetadataCRD := lib.CRDMetadata{}
	if poolCache, ok := rest.cache.PoolCache.AviCacheGet(poolKey); ok {
		if poolCacheObj, found := poolCache.(*avicache.AviPoolCache); found {
			cacheServiceMetadataCRD = poolCacheObj.ServiceMetadataObj.CRDStatus
		}
	}
	rest.cache.PoolCache.AviCacheDelete(poolKey)
	if (cacheServiceMetadataCRD != lib.CRDMetadata{}) {
		status.HttpRuleEventBroadcast(poolKey.Name, cacheServiceMetadataCRD, lib.CRDMetadata{})
	}
	return nil
}

func (rest *RestOperations) DeletePoolIngressStatus(poolKey avicache.NamespaceName, isVSDelete bool, vsName, key string) {
	pool_cache, found := rest.cache.PoolCache.AviCacheGet(poolKey)
	if found {
		pool_cache_obj, success := pool_cache.(*avicache.AviPoolCache)
		if success {
			switch pool_cache_obj.ServiceMetadataObj.ServiceMetadataMapping(lib.Pool) {
			case lib.GatewayPool:
				updateOptions := status.UpdateOptions{
					ServiceMetadata: pool_cache_obj.ServiceMetadataObj,
					Key:             key,
					VSName:          vsName,
					Tenant:          pool_cache_obj.Tenant,
				}
				statusOption := status.StatusOptions{
					ObjType: utils.L4LBService,
					Op:      lib.DeleteStatus,
					Key:     key,
					Options: &updateOptions,
				}
				utils.AviLog.Infof("key: %s Publishing to status queue, options: %v", pool_cache_obj.ServiceMetadataObj.NamespaceServiceName[0], utils.Stringify(statusOption))
				status.PublishToStatusQueue(pool_cache_obj.ServiceMetadataObj.NamespaceServiceName[0], statusOption)
			case lib.SNIInsecureOrEVHPool:
				updateOptions := status.UpdateOptions{
					ServiceMetadata: pool_cache_obj.ServiceMetadataObj,
					Key:             key,
					VSName:          vsName,
					Tenant:          pool_cache_obj.Tenant,
				}
				statusOption := status.StatusOptions{
					ObjType: utils.Ingress,
					Op:      lib.DeleteStatus,
					IsVSDel: isVSDelete,
					Key:     key,
					Options: &updateOptions,
				}
				if utils.GetInformers().RouteInformer != nil {
					statusOption.ObjType = utils.OshiftRoute
				}
				if pool_cache_obj.ServiceMetadataObj.IsMCIIngress {
					statusOption.ObjType = lib.MultiClusterIngress
				}
				utils.AviLog.Debugf("key: %s Publishing to status queue, options: %v", updateOptions.ServiceMetadata.HostNames[0], utils.Stringify(statusOption))
				status.PublishToStatusQueue(updateOptions.ServiceMetadata.HostNames[0], statusOption)
			}
		}
	}
}
