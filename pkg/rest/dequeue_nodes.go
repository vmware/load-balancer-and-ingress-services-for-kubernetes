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
	"fmt"
	"regexp"

	"ako/pkg/objects"

	avicache "ako/pkg/cache"
	"ako/pkg/lib"
	"ako/pkg/nodes"

	"github.com/avinetworks/container-lib/utils"
)

type RestOperations struct {
	cache             *avicache.AviObjCache
	aviRestPoolClient *utils.AviRestClientPool
	// Add locks if some code needs to be protected.
}

func NewRestOperations(cache *avicache.AviObjCache, aviRestPoolClient *utils.AviRestClientPool) RestOperations {
	return RestOperations{cache: cache, aviRestPoolClient: aviRestPoolClient}
}

func (rest *RestOperations) CleanupVS(key string) {
	namespace, name := utils.ExtractNamespaceObjectName(key)
	vsKey := avicache.NamespaceName{Namespace: namespace, Name: name}
	vs_cache_obj := rest.getVsCacheObj(vsKey, key)
	utils.AviLog.Info.Printf("key: %s, msg: cleanup mode, removing all VSes", key)
	rest.deleteVSOper(vsKey, vs_cache_obj, namespace, key)
}

func (rest *RestOperations) DeQueueNodes(key string) {
	utils.AviLog.Info.Printf("key: %s, msg: start rest layer sync.", key)
	// Got the key from the Graph Layer - let's fetch the model
	ok, avimodelIntf := objects.SharedAviGraphLister().Get(key)
	if !ok {
		utils.AviLog.Warning.Printf("key: %s, msg: no model found for the key", key)
		return
	}

	namespace, name := utils.ExtractNamespaceObjectName(key)
	vsKey := avicache.NamespaceName{Namespace: namespace, Name: name}
	vs_cache_obj := rest.getVsCacheObj(vsKey, key)
	if avimodelIntf == nil {
		if vs_cache_obj != nil {
			utils.AviLog.Info.Printf("key: %s, msg: nil model found, this is a vs deletion case", key)
			rest.deleteVSOper(vsKey, vs_cache_obj, namespace, key)
		}
	} else if ok && avimodelIntf != nil {
		avimodel := avimodelIntf.(*nodes.AviObjectGraph)
		if avimodel == nil {
			utils.AviLog.Info.Printf("Enpty Model found, skipping")
			return
		}
		avimodel, ok = avimodel.GetCopy(key)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, failed to get process model", key)
			return
		}
		if avimodel.IsVrf {
			utils.AviLog.Warning.Printf("key: %s, msg: processing vrf object\n", key)
			rest.vrfCU(key, name, avimodel)
			return
		}
		utils.AviLog.Info.Printf("key: %s, msg: VS create/update.", key)
		if len(avimodel.GetAviVS()) != 1 {
			utils.AviLog.Warning.Printf("key: %s, msg: virtualservice in the model is not equal to 1:%v", key, avimodel.GetAviVS())
			return
		}
		rest.RestOperation(name, namespace, avimodel, false, vs_cache_obj, key)
	}

}

func (rest *RestOperations) vrfCU(key, vrfName string, avimodel *nodes.AviObjectGraph) {
	vrfNode := avimodel.GetAviVRF()
	if len(vrfNode) != 1 {
		utils.AviLog.Warning.Printf("key: %s, msg: Number of vrf nodes is not one\n", key)
		return
	}
	aviVrfNode := vrfNode[0]
	vrfCacheObj := rest.getVrfCacheObj(vrfName)
	if vrfCacheObj == nil {
		utils.AviLog.Warning.Printf("key: %s, vrf %s not found in cache, exiting\n", key, vrfName)
		return
	}
	if vrfCacheObj.CloudConfigCksum == aviVrfNode.CloudConfigCksum {
		utils.AviLog.Info.Printf("key: %s, msg: checksum for vrf %s has not changed, skipping\n", key, vrfName)
		return
	}
	var restOps []*utils.RestOp
	restOp := rest.AviVrfBuild(key, aviVrfNode, vrfCacheObj.Uuid)
	if restOp == nil {
		utils.AviLog.Info.Printf("key: %s, no rest operation for vrf %s\n", key, vrfName)
		return
	}
	restOps = append(restOps, restOp)
	vrfKey := avicache.NamespaceName{Namespace: utils.ADMIN_NS, Name: vrfName}
	utils.AviLog.Info.Printf("key: %s, msg: Executing rest for vrf %s\n", key, vrfName)
	utils.AviLog.Trace.Printf("key: %s, msg: restops %v\n", key, *restOp)
	rest.ExecuteRestAndPopulateCache(restOps, vrfKey, avimodel, key)
}

func (rest *RestOperations) RestOperation(vsName string, namespace string, avimodel *nodes.AviObjectGraph, sniNode bool, vs_cache_obj *avicache.AviVsCache, key string) {
	var pools_to_delete []avicache.NamespaceName
	var pgs_to_delete []avicache.NamespaceName
	var ds_to_delete []avicache.NamespaceName
	var vsvip_to_delete []avicache.NamespaceName
	var sni_to_delete []avicache.NamespaceName
	vsKey := avicache.NamespaceName{Namespace: namespace, Name: vsName}
	aviVsNode := avimodel.GetAviVS()[0]
	// Order would be this: 1. Pools 2. PGs  3. DS. 4. SSLKeyCert 5. VS
	if vs_cache_obj != nil {
		var rest_ops []*utils.RestOp
		pools_to_delete, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, vs_cache_obj, namespace, rest_ops, key)
		pgs_to_delete, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, vs_cache_obj, namespace, rest_ops, key)
		ds_to_delete, rest_ops = rest.DatascriptCU(aviVsNode.HTTPDSrefs, vs_cache_obj, namespace, rest_ops, key)
		vsvip_to_delete, rest_ops = rest.VSVipCU(aviVsNode.VSVIPRefs, vs_cache_obj, namespace, rest_ops, key)
		utils.AviLog.Info.Printf("key: %s, msg: stored checksum for VS: %s, model checksum: %s", key, vs_cache_obj.CloudConfigCksum, fmt.Sprint(aviVsNode.GetCheckSum()))
		if vs_cache_obj.CloudConfigCksum == fmt.Sprint(aviVsNode.GetCheckSum()) {
			utils.AviLog.Info.Printf("key: %s, msg: the checksums are same for vs %s, not doing anything", key, vs_cache_obj.Name)
		} else {
			utils.AviLog.Info.Printf("key: %s, msg: the stored checksum for vs is %v, and the obtained checksum for VS is: %v", key, vs_cache_obj.CloudConfigCksum, fmt.Sprint(aviVsNode.GetCheckSum()))
			// The checksums are different, so it should be a PUT call.
			restOp := rest.AviVsBuild(aviVsNode, utils.RestPut, vs_cache_obj, key)
			rest_ops = append(rest_ops, restOp...)

		}
		rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key)
	} else {
		var rest_ops []*utils.RestOp
		_, rest_ops = rest.PoolCU(aviVsNode.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(aviVsNode.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.DatascriptCU(aviVsNode.HTTPDSrefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.VSVipCU(aviVsNode.VSVIPRefs, nil, namespace, rest_ops, key)

		// The cache was not found - it's a POST call.
		restOp := rest.AviVsBuild(aviVsNode, utils.RestPost, nil, key)
		rest_ops = append(rest_ops, restOp...)
		utils.AviLog.Trace.Printf("POST key: %s, vsKey: %s", key, vsKey)
		utils.AviLog.Trace.Printf("POST restops %s", utils.Stringify(rest_ops))
		rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key)
	}
	if vs_cache_obj != nil {
		for _, sni_uuid := range vs_cache_obj.SNIChildCollection {
			sni_vs_key, ok := rest.cache.VsCache.AviCacheGetKeyByUuid(sni_uuid)
			if ok {
				sni_to_delete = append(sni_to_delete, sni_vs_key.(avicache.NamespaceName))
			} else {
				utils.AviLog.Info.Printf("key: %s, msg: Couldn't get SNI key for uuid: %v", key, sni_uuid)
			}
		}
	}
	for _, sni_node := range aviVsNode.SniNodes {
		utils.AviLog.Info.Printf("key: %s, msg: processing sni node: %s", key, sni_node.Name)
		utils.AviLog.Info.Printf("key: %s, msg: probable SNI delete candidates: %s", key, sni_to_delete)
		var rest_ops []*utils.RestOp
		vsKey = avicache.NamespaceName{Namespace: namespace, Name: sni_node.Name}
		if vs_cache_obj != nil {
			sni_to_delete, rest_ops = rest.SNINodeCU(sni_node, vs_cache_obj, namespace, sni_to_delete, rest_ops, key)
		} else {
			_, rest_ops = rest.SNINodeCU(sni_node, nil, namespace, sni_to_delete, rest_ops, key)
		}
		rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key)
	}

	// Let's populate all the DELETE entries
	if len(sni_to_delete) > 0 {
		utils.AviLog.Info.Printf("key: %s, msg: SNI delete candidates are : %s", key, sni_to_delete)
		var rest_ops []*utils.RestOp
		for _, del_sni := range sni_to_delete {
			rest.SNINodeDelete(del_sni, namespace, rest_ops, key)
			rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key)
		}
	} else {
		var rest_ops []*utils.RestOp
		rest_ops = rest.VSVipDelete(vsvip_to_delete, namespace, rest_ops, key)
		rest_ops = rest.DSDelete(ds_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(pgs_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(pools_to_delete, namespace, rest_ops, key)
		rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, avimodel, key)
	}
}

func (rest *RestOperations) getVsCacheObj(vsKey avicache.NamespaceName, key string) *avicache.AviVsCache {
	vs_cache, found := rest.cache.VsCache.AviCacheGet(vsKey)
	if found {
		vs_cache_obj, ok := vs_cache.(*avicache.AviVsCache)
		if !ok {
			utils.AviLog.Warning.Printf("key: %s, msg: invalid vs object found. Cannot cast. Not doing anything", key)
			return nil
		}
		return vs_cache_obj
	}
	utils.AviLog.Info.Printf("key :%s, msg: vs cache object NOT found for vskey: %s", key, vsKey)
	return nil
}

func (rest *RestOperations) deleteVSOper(vsKey avicache.NamespaceName, vs_cache_obj *avicache.AviVsCache, namespace string, key string) bool {
	var rest_ops []*utils.RestOp

	if vs_cache_obj != nil {
		// VS delete should delete everything together.
		for _, sni_uuid := range vs_cache_obj.SNIChildCollection {
			sniVsKey, ok := rest.cache.VsCache.AviCacheGetKeyByUuid(sni_uuid)
			if ok {
				delSNI := sniVsKey.(avicache.NamespaceName)
				rest.SNINodeDelete(delSNI, namespace, rest_ops, key)
			}
		}
		rest_op := rest.AviVSDel(vs_cache_obj.Uuid, namespace, key)
		rest_ops = append(rest_ops, rest_op)
		rest_ops = rest.DataScriptDelete(vs_cache_obj.DSKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.SSLKeyCertDelete(vs_cache_obj.SSLKeyCertCollection, namespace, rest_ops, key)
		rest_ops = rest.HTTPPolicyDelete(vs_cache_obj.HTTPKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(vs_cache_obj.PGKeyCollection, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(vs_cache_obj.PoolKeyCollection, namespace, rest_ops, key)
		rest.ExecuteRestAndPopulateCache(rest_ops, vsKey, nil, key)
		return true
	}
	return false
}

func (rest *RestOperations) ExecuteRestAndPopulateCache(rest_ops []*utils.RestOp, aviObjKey avicache.NamespaceName, avimodel *nodes.AviObjectGraph, key string, sslKey ...utils.NamespaceName) {
	// Choose a avi client based on the model name hash. This would ensure that the same worker queue processes updates for a given VS all the time.
	shardSize := lib.GetshardSize()
	if shardSize != 0 {
		bkt := utils.Bkt(key, shardSize)
		utils.AviLog.Info.Printf("key: %s, msg: processing in rest queue number: %v", key, bkt)
		if len(rest.aviRestPoolClient.AviClient) > 0 && len(rest_ops) > 0 {
			aviclient := rest.aviRestPoolClient.AviClient[bkt]
			err := rest.aviRestPoolClient.AviRestOperate(aviclient, rest_ops)
			if err != nil {
				utils.AviLog.Warning.Printf("key: %s, msg: there was an error sending the macro %v", key, err.Error())
				for i := len(rest_ops) - 1; i >= 0; i-- {
					// Go over each of the failed requests and enqueue them to the worker queue for retry.
					if rest_ops[i].Err != nil {
						// If it's for a SNI child, publish the parent VS's key
						if avimodel != nil {
							publishKey := avimodel.GetAviVS()[0].Name
							if avimodel.GetRetryCounter() != 0 {
								PublishKeyToRetryLayer(publishKey, key, rest_ops[i].Err.Error())
							}
						}
					}
				}
			} else {
				utils.AviLog.Info.Printf("key: %s, msg: rest call executed successfully, will update cache", key)
				// Add to local obj caches
				for _, rest_op := range rest_ops {
					if rest_op.Err == nil && (rest_op.Method == utils.RestPost || rest_op.Method == utils.RestPut) {
						utils.AviLog.Info.Printf("key: %s, msg: creating/updating %s cache", key, rest_op.Model)
						if rest_op.Model == "Pool" {
							rest.AviPoolCacheAdd(rest_op, aviObjKey, key)
						} else if rest_op.Model == "VirtualService" {
							rest.AviVsCacheAdd(rest_op, key)
						} else if rest_op.Model == "PoolGroup" {
							rest.AviPGCacheAdd(rest_op, aviObjKey, key)
						} else if rest_op.Model == "VSDataScriptSet" {
							rest.AviDSCacheAdd(rest_op, aviObjKey, key)
						} else if rest_op.Model == "HTTPPolicySet" {
							rest.AviHTTPPolicyCacheAdd(rest_op, aviObjKey, key)
						} else if rest_op.Model == "SSLKeyAndCertificate" {
							rest.AviSSLKeyCertAdd(rest_op, aviObjKey, key)
						} else if rest_op.Model == "VrfContext" {
							rest.AviVrfCacheAdd(rest_op, aviObjKey, key)
						} else if rest_op.Model == "VsVip" {
							rest.AviVsVipCacheAdd(rest_op, aviObjKey, key)
						}

					} else {
						utils.AviLog.Info.Printf("key: %s, msg: deleting %s cache", rest_op.Model, key)
						if rest_op.Model == "Pool" {
							rest.AviPoolCacheDel(rest_op, aviObjKey, key)
						} else if rest_op.Model == "VirtualService" {
							rest.AviVsCacheDel(aviObjKey, rest_op, key)
						} else if rest_op.Model == "PoolGroup" {
							rest.AviPGCacheDel(rest_op, aviObjKey, key)
						} else if rest_op.Model == "HTTPPolicySet" {
							rest.AviHTTPPolicyCacheDel(rest_op, aviObjKey, key)
						} else if rest_op.Model == "SSLKeyAndCertificate" {
							rest.AviSSLCacheDel(rest_op, aviObjKey, key)
						} else if rest_op.Model == "VsVip" {
							rest.AviVsVipCacheDel(rest_op, aviObjKey, key)
						} else if rest_op.Model == "VSDataScriptSet" {
							rest.AviDSCacheDel(rest_op, aviObjKey, key)
						}
					}
				}

			}
		}
	}
}

func (rest *RestOperations) DataScriptDelete(dsToDelete []avicache.NamespaceName, namespace string, restOps []*utils.RestOp, key string) []*utils.RestOp {
	for _, delDS := range dsToDelete {
		dsKey := avicache.NamespaceName{Namespace: namespace, Name: delDS.Name}
		dsCache, ok := rest.cache.DSCache.AviCacheGet(dsKey)
		if ok {
			dsCacheObj, _ := dsCache.(*avicache.AviDSCache)
			restOp := rest.AviDSDel(dsCacheObj.Uuid, namespace, key)
			restOp.ObjName = delDS.Name
			restOps = append(restOps, restOp)
		}
	}
	return restOps
}

func PublishKeyToRetryLayer(vs_key string, key string, errorStr string) {
	var bkt uint32
	bkt = 0
	// SDK does not give us a error status code, so let's extract it.
	statuscode := ExtractStatusCode(errorStr)
	utils.AviLog.Info.Printf("key: %s, msg: Status code retrieved: %s", key, statuscode)
	if statuscode == "500" || statuscode == "501" || statuscode == "502" || statuscode == "503" {
		slowRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.SLOW_RETRY_LAYER)
		slowRetryQueue.Workqueue[bkt].AddRateLimited(vs_key)
		utils.AviLog.Info.Printf("key: %s, msg: Published key with vs_key to slow path retry queue: %s", key, vs_key)
	} else if statuscode == "404" || statuscode == "400" || statuscode == "409" { // Will account for more error codes.
		fastRetryQueue := utils.SharedWorkQueue().GetQueueByName(lib.FAST_RETRY_LAYER)
		fastRetryQueue.Workqueue[bkt].AddRateLimited(vs_key)
		utils.AviLog.Info.Printf("key: %s, msg: Published key with vs_key to fast path retry queue: %s", key, vs_key)
	}
}

//Candidate for container-lib
func ExtractStatusCode(word string) string {
	r, _ := regexp.Compile("HTTP code: .*.;")
	result := r.FindAllString(word, -1)
	if len(result) == 1 {
		return result[0][len(result[0])-4 : len(result[0])-1]
	}
	return ""
}

func (rest *RestOperations) PoolDelete(pools_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	for _, del_pool := range pools_to_delete {
		// fetch trhe pool uuid from cache
		pool_key := avicache.NamespaceName{Namespace: namespace, Name: del_pool.Name}
		pool_cache, ok := rest.cache.PoolCache.AviCacheGet(pool_key)
		if ok {
			pool_cache_obj, _ := pool_cache.(*avicache.AviPoolCache)
			restOp := rest.AviPoolDel(pool_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_pool.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) VSVipDelete(vsvip_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	for _, del_vsvip := range vsvip_to_delete {
		// fetch trhe pool uuid from cache
		vsvip_key := avicache.NamespaceName{Namespace: namespace, Name: del_vsvip.Name}
		vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvip_key)
		if ok {
			vsvip_cache_obj, _ := vsvip_cache.(*avicache.AviVSVIPCache)
			restOp := rest.AviVsVipDel(vsvip_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_vsvip.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) PoolGroupDelete(pgs_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Info.Printf("key: %s, msg: about to delete the PGs %s", key, pgs_to_delete)
	for _, del_pg := range pgs_to_delete {
		// fetch trhe pool uuid from cache
		pg_key := avicache.NamespaceName{Namespace: namespace, Name: del_pg.Name}
		pg_cache, ok := rest.cache.PgCache.AviCacheGet(pg_key)
		if ok {
			pg_cache_obj, _ := pg_cache.(*avicache.AviPGCache)
			restOp := rest.AviPGDel(pg_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_pg.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) DSDelete(ds_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	utils.AviLog.Info.Printf("key: %s, msg: about to delete the DS %s", key, ds_to_delete)
	for _, del_ds := range ds_to_delete {
		// fetch trhe pool uuid from cache
		ds_key := avicache.NamespaceName{Namespace: namespace, Name: del_ds.Name}
		ds_cache, ok := rest.cache.DSCache.AviCacheGet(ds_key)
		if ok {
			ds_cache_obj, _ := ds_cache.(*avicache.AviDSCache)
			restOp := rest.AviDSDel(ds_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_ds.Name
			rest_ops = append(rest_ops, restOp)
		} else {
			utils.AviLog.Info.Printf("key: %s, msg: ds not found in cache during delete %s", key, ds_to_delete)
		}
	}
	return rest_ops
}

func (rest *RestOperations) PoolCU(pool_nodes []*nodes.AviPoolNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_pool_nodes []avicache.NamespaceName
	if vs_cache_obj != nil {
		cache_pool_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.PoolKeyCollection))
		copy(cache_pool_nodes, vs_cache_obj.PoolKeyCollection)
		utils.AviLog.Info.Printf("key: %s, msg: the cached pools are: %v", key, utils.Stringify(cache_pool_nodes))
		if cache_pool_nodes != nil {
			for _, pool := range pool_nodes {
				// check in the pool cache to see if this pool exists in AVI
				pool_key := avicache.NamespaceName{Namespace: namespace, Name: pool.Name}
				found := utils.HasElem(cache_pool_nodes, pool_key)
				utils.AviLog.Info.Printf("key: %s, msg: processing pool key: %v", key, pool_key)
				if found {
					cache_pool_nodes = Remove(cache_pool_nodes, pool_key)
					utils.AviLog.Info.Printf("key: %s, key: the cache pool nodes are: %v", key, cache_pool_nodes)
					pool_cache, ok := rest.cache.PoolCache.AviCacheGet(pool_key)
					if ok {
						pool_cache_obj, _ := pool_cache.(*avicache.AviPoolCache)
						// Cache found. Let's compare the checksums
						utils.AviLog.Info.Printf("key: %s, msg: poolcache: %v", key, pool_cache_obj)
						if pool_cache_obj.CloudConfigCksum == fmt.Sprint(pool.GetCheckSum()) {
							utils.AviLog.Info.Printf("key: %s, msg: the checksums are same for pool %s, not doing anything", key, pool.Name)
						} else {
							utils.AviLog.Info.Printf("key: %s, msg: the checksums are different for pool %s, operation: PUT", key, pool.Name)
							// The checksums are different, so it should be a PUT call.
							restOp := rest.AviPoolBuild(pool, pool_cache_obj, key)
							rest_ops = append(rest_ops, restOp)
						}
					}
				} else {
					utils.AviLog.Info.Printf("key: %s, msg: pool %s not found in cache, operation: POST", key, pool.Name)
					// Not found - it should be a POST call.
					restOp := rest.AviPoolBuild(pool, nil, key)
					rest_ops = append(rest_ops, restOp)
				}

			}
		}
	} else {
		// Everything is a POST call
		for _, pool := range pool_nodes {
			utils.AviLog.Info.Printf("key: %s, msg: pool cache does not exist %s, operation: POST", key, pool.Name)
			restOp := rest.AviPoolBuild(pool, nil, key)
			rest_ops = append(rest_ops, restOp)
		}

	}
	utils.AviLog.Info.Printf("key: %s, msg: the POOLS rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Info.Printf("key: %s, msg: the POOLs to be deleted are: %s", key, cache_pool_nodes)
	return cache_pool_nodes, rest_ops
}

func (rest *RestOperations) SNINodeDelete(del_sni avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) {
	utils.AviLog.Info.Printf("key: %s, msg: about to delete the SNI child %s", key, del_sni)
	sni_key := avicache.NamespaceName{Namespace: namespace, Name: del_sni.Name}
	sni_cache_obj := rest.getVsCacheObj(sni_key, key)
	if sni_cache_obj != nil {
		utils.AviLog.Info.Printf("key: %s, msg: SNI object before delete %s", key, utils.Stringify(sni_cache_obj))
		rest.deleteVSOper(sni_key, sni_cache_obj, namespace, key)
	}

}

func (rest *RestOperations) SNINodeCU(sni_node *nodes.AviVsNode, vs_cache_obj *avicache.AviVsCache, namespace string, cache_sni_nodes []avicache.NamespaceName, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var sni_pools_to_delete []avicache.NamespaceName
	var sni_pgs_to_delete []avicache.NamespaceName
	var http_policies_to_delete []avicache.NamespaceName
	var sslkey_cert_delete []avicache.NamespaceName
	if vs_cache_obj != nil {
		sni_key := avicache.NamespaceName{Namespace: namespace, Name: sni_node.Name}
		// Search the VS cache and obtain the UUID of this VS. Then see if this UUID is part of the SNIChildCollection or not.
		found := utils.HasElem(cache_sni_nodes, sni_key)
		utils.AviLog.Info.Printf("key: %s, msg: processing node key: %v", key, sni_key)
		if found && cache_sni_nodes != nil {
			cache_sni_nodes = Remove(cache_sni_nodes, sni_key)
			utils.AviLog.Info.Printf("key: %s, msg: the cache sni nodes are: %v", key, cache_sni_nodes)
			sni_cache_obj := rest.getVsCacheObj(sni_key, key)
			if sni_cache_obj != nil {
				// Cache found. Let's compare the checksums
				if sni_cache_obj.CloudConfigCksum == fmt.Sprint(sni_node.GetCheckSum()) {
					utils.AviLog.Info.Printf("key: %s, msg: the checksums are same for sni child %s, not doing anything", key, sni_node.Name)
				} else {
					sni_pools_to_delete, rest_ops = rest.PoolCU(sni_node.PoolRefs, sni_cache_obj, namespace, rest_ops, key)
					sni_pgs_to_delete, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, sni_cache_obj, namespace, rest_ops, key)
					http_policies_to_delete, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, sni_cache_obj, namespace, rest_ops, key)
					sslkey_cert_delete, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, sni_cache_obj, namespace, rest_ops, key)
					utils.AviLog.Info.Printf("key: %s, msg: the checksums are different for sni child %s, operation: PUT", key, sni_node.Name)
					// The checksums are different, so it should be a PUT call.
					restOp := rest.AviVsBuild(sni_node, utils.RestPut, sni_cache_obj, key)
					rest_ops = append(rest_ops, restOp...)
				}
			}
		} else {
			utils.AviLog.Info.Printf("key: %s, msg: sni child %s not found in cache, operation: POST", key, sni_node.Name)
			_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)
			_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)

			// Not found - it should be a POST call.
			restOp := rest.AviVsBuild(sni_node, utils.RestPost, nil, key)
			rest_ops = append(rest_ops, restOp...)
		}
		rest_ops = rest.SSLKeyCertDelete(sslkey_cert_delete, namespace, rest_ops, key)
		rest_ops = rest.HTTPPolicyDelete(http_policies_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolGroupDelete(sni_pgs_to_delete, namespace, rest_ops, key)
		rest_ops = rest.PoolDelete(sni_pools_to_delete, namespace, rest_ops, key)
		utils.AviLog.Info.Printf("key: %s, msg: the SNI VSes to be deleted are: %s", key, cache_sni_nodes)
	} else {
		utils.AviLog.Info.Printf("key: %s, msg: sni child %s not found in cache and SNI parent also does not exist in cache", key, sni_node.Name)
		_, rest_ops = rest.PoolCU(sni_node.PoolRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.PoolGroupCU(sni_node.PoolGroupRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.HTTPPolicyCU(sni_node.HttpPolicyRefs, nil, namespace, rest_ops, key)
		_, rest_ops = rest.SSLKeyCertCU(sni_node.SSLKeyCertRefs, nil, namespace, rest_ops, key)

		// Not found - it should be a POST call.
		restOp := rest.AviVsBuild(sni_node, utils.RestPost, nil, key)
		rest_ops = append(rest_ops, restOp...)
	}
	return cache_sni_nodes, rest_ops
}

func (rest *RestOperations) PoolGroupCU(pg_nodes []*nodes.AviPoolGroupNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_pg_nodes []avicache.NamespaceName
	if vs_cache_obj != nil {
		cache_pg_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.PGKeyCollection))
		copy(cache_pg_nodes, vs_cache_obj.PGKeyCollection)
		utils.AviLog.Info.Printf("key: %s, msg: cached poolgroups before CU :%v", key, cache_pg_nodes)
		// Default is POST
		if cache_pg_nodes != nil {
			for _, pg := range pg_nodes {
				pg_key := avicache.NamespaceName{Namespace: namespace, Name: pg.Name}
				found := utils.HasElem(cache_pg_nodes, pg_key)
				if found {
					cache_pg_nodes = Remove(cache_pg_nodes, pg_key)
					pg_cache, ok := rest.cache.PgCache.AviCacheGet(pg_key)
					if ok {
						pg_cache_obj, _ := pg_cache.(*avicache.AviPGCache)
						// Cache found. Let's compare the checksums
						if pg_cache_obj.CloudConfigCksum == fmt.Sprint(pg.GetCheckSum()) {
							utils.AviLog.Info.Printf("key: %s, msg: the checksums are same for PG %s, not doing anything", key, pg_cache_obj.Name)
						} else {
							// The checksums are different, so it should be a PUT call.
							restOp := rest.AviPoolGroupBuild(pg, pg_cache_obj, key)
							rest_ops = append(rest_ops, restOp)
						}
					}
				} else {
					// Not found - it should be a POST call.
					restOp := rest.AviPoolGroupBuild(pg, nil, key)
					rest_ops = append(rest_ops, restOp)
				}

			}
		}
	} else {
		// Everything is a POST call
		for _, pg := range pg_nodes {
			restOp := rest.AviPoolGroupBuild(pg, nil, key)
			rest_ops = append(rest_ops, restOp)
		}

	}
	utils.AviLog.Info.Printf("key: %s, msg: the PGs rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Info.Printf("key: %s, msg: the PGs to be deleted are: %s", key, cache_pg_nodes)
	return cache_pg_nodes, rest_ops
}

func (rest *RestOperations) DatascriptCU(ds_nodes []*nodes.AviHTTPDataScriptNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_ds_nodes []avicache.NamespaceName

	if vs_cache_obj != nil {
		cache_ds_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.DSKeyCollection))
		copy(cache_ds_nodes, vs_cache_obj.DSKeyCollection)

		// Default is POST
		if cache_ds_nodes != nil {
			for _, ds := range ds_nodes {
				// check in the ds cache to see if this ds exists in AVI
				ds_key := avicache.NamespaceName{Namespace: namespace, Name: ds.Name}
				found := utils.HasElem(cache_ds_nodes, ds_key)
				if found {
					cache_ds_nodes = Remove(cache_ds_nodes, ds_key)
					_, ok := rest.cache.DSCache.AviCacheGet(ds_key)
					if !ok {
						// If the DS Is not found - let's do a POST call. Assume DS will not be a candidate for PUT.
						restOp := rest.AviDSBuild(ds, nil, key)
						rest_ops = append(rest_ops, restOp)
					}
				}
			}
		}
	} else {
		// Everything is a POST call
		for _, ds := range ds_nodes {
			restOp := rest.AviDSBuild(ds, nil, key)
			rest_ops = append(rest_ops, restOp)
		}

	}
	utils.AviLog.Info.Printf("key: %s, msg: the DS rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Info.Printf("key: %s, msg: the DS to be deleted are: %s", key, cache_ds_nodes)
	return cache_ds_nodes, rest_ops
}

func (rest *RestOperations) VSVipCU(vsvip_nodes []*nodes.AviVSVIPNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_vsvip_nodes []avicache.NamespaceName
	if vs_cache_obj != nil {
		cache_vsvip_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.VSVipKeyCollection))
		copy(cache_vsvip_nodes, vs_cache_obj.VSVipKeyCollection)
		// Default is POST
		if cache_vsvip_nodes != nil {
			for _, vsvip := range vsvip_nodes {
				vsvip_key := avicache.NamespaceName{Namespace: namespace, Name: vsvip.Name}
				found := utils.HasElem(cache_vsvip_nodes, vsvip_key)
				if found {
					cache_vsvip_nodes = Remove(cache_vsvip_nodes, vsvip_key)
					vsvip_cache, ok := rest.cache.VSVIPCache.AviCacheGet(vsvip_key)
					if ok {
						vsvip_cache_obj, _ := vsvip_cache.(*avicache.AviVSVIPCache)
						// Cache found. Let's compare the checksums
						utils.AviLog.Info.Printf("key: %s, msg: the model FQDNs: %s, cache_FQDNs: %s", key, vsvip.FQDNs, vsvip_cache_obj.FQDNs)
						if utils.Hash(utils.Stringify(vsvip_cache_obj.FQDNs)) == vsvip.GetCheckSum() {
							utils.AviLog.Info.Printf("key: %s, msg: the checksums are same for VSVIP %s, not doing anything", key, vsvip_cache_obj.Name)
						} else {
							// The checksums are different, so it should be a PUT call.
							restOp, err := rest.AviVsVipBuild(vsvip, vsvip_cache_obj, key)
							if err == nil {
								rest_ops = append(rest_ops, restOp)
							}
						}
					}
				} else {
					// Not found - it should be a POST call.
					restOp, err := rest.AviVsVipBuild(vsvip, nil, key)
					if err == nil {
						rest_ops = append(rest_ops, restOp)
					}
				}

			}
		}
	} else {
		// Everything is a POST call
		for _, vsvip := range vsvip_nodes {
			restOp, err := rest.AviVsVipBuild(vsvip, nil, key)
			if err == nil {
				rest_ops = append(rest_ops, restOp)
			}
		}

	}
	utils.AviLog.Info.Printf("key: %s, msg: the vsvip rest_op is %s", key, utils.Stringify(rest_ops))
	utils.AviLog.Info.Printf("key: %s, msg: the vsvip to be deleted are: %s", key, cache_vsvip_nodes)
	return cache_vsvip_nodes, rest_ops
}

func (rest *RestOperations) HTTPPolicyCU(http_nodes []*nodes.AviHttpPolicySetNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	var cache_http_nodes []avicache.NamespaceName
	// Default is POST
	if vs_cache_obj != nil {
		cache_http_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.HTTPKeyCollection))
		copy(cache_http_nodes, vs_cache_obj.HTTPKeyCollection)
		for _, http := range http_nodes {
			http_key := avicache.NamespaceName{Namespace: namespace, Name: http.Name}
			found := utils.HasElem(cache_http_nodes, http_key)
			if found {
				http_cache, ok := rest.cache.HTTPPolicyCache.AviCacheGet(http_key)
				if ok {
					cache_http_nodes = Remove(cache_http_nodes, http_key)
					http_cache_obj, _ := http_cache.(*avicache.AviHTTPPolicyCache)
					// Cache found. Let's compare the checksums
					if http_cache_obj.CloudConfigCksum == fmt.Sprint(http.GetCheckSum()) {
						utils.AviLog.Info.Printf("The checksums are same for HTTP cache obj %s, not doing anything", http_cache_obj.Name)
					} else {
						// The checksums are different, so it should be a PUT call.
						restOp := rest.AviHttpPSBuild(http, http_cache_obj, key)
						rest_ops = append(rest_ops, restOp)
					}
				}
			} else {
				// Not found - it should be a POST call.
				restOp := rest.AviHttpPSBuild(http, nil, key)
				rest_ops = append(rest_ops, restOp)
			}

		}
	} else {
		// Everything is a POST call
		for _, http := range http_nodes {
			restOp := rest.AviHttpPSBuild(http, nil, key)
			rest_ops = append(rest_ops, restOp)
		}

	}
	utils.AviLog.Info.Printf("The HTTP Policies rest_op is %s", utils.Stringify(rest_ops))
	return cache_http_nodes, rest_ops
}

func (rest *RestOperations) HTTPPolicyDelete(https_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	for _, del_http := range https_to_delete {
		// fetch trhe http policyset uuid from cache
		http_key := avicache.NamespaceName{Namespace: namespace, Name: del_http.Name}
		http_cache, ok := rest.cache.HTTPPolicyCache.AviCacheGet(http_key)
		if ok {
			http_cache_obj, _ := http_cache.(*avicache.AviHTTPPolicyCache)
			restOp := rest.AviHttpPolicyDel(http_cache_obj.Uuid, namespace, key)
			restOp.ObjName = del_http.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func (rest *RestOperations) SSLKeyCertCU(sslkey_nodes []*nodes.AviTLSKeyCertNode, vs_cache_obj *avicache.AviVsCache, namespace string, rest_ops []*utils.RestOp, key string) ([]avicache.NamespaceName, []*utils.RestOp) {
	// Default is POST
	var cache_ssl_nodes []avicache.NamespaceName
	if vs_cache_obj != nil {
		cache_ssl_nodes = make([]avicache.NamespaceName, len(vs_cache_obj.SSLKeyCertCollection))
		copy(cache_ssl_nodes, vs_cache_obj.SSLKeyCertCollection)
		for _, ssl := range sslkey_nodes {
			ssl_key := avicache.NamespaceName{Namespace: namespace, Name: ssl.Name}
			found := utils.HasElem(cache_ssl_nodes, ssl_key)
			if found {
				ssl_cache, ok := rest.cache.SSLKeyCache.AviCacheGet(ssl_key)
				if ok {
					cache_ssl_nodes = Remove(cache_ssl_nodes, ssl_key)
					ssl_cache_obj, _ := ssl_cache.(*avicache.AviSSLCache)

					// The checksums are different, so it should be a PUT call.

					restOp := rest.AviSSLBuild(ssl, ssl_cache_obj)
					rest_ops = append(rest_ops, restOp)

				}
			} else {
				// Not found - it should be a POST call.
				restOp := rest.AviSSLBuild(ssl, nil)
				rest_ops = append(rest_ops, restOp)
			}

		}
	} else {
		// Everything is a POST call
		for _, ssl := range sslkey_nodes {
			restOp := rest.AviSSLBuild(ssl, nil)
			rest_ops = append(rest_ops, restOp)
		}

	}
	return cache_ssl_nodes, rest_ops
}

func (rest *RestOperations) SSLKeyCertDelete(ssl_to_delete []avicache.NamespaceName, namespace string, rest_ops []*utils.RestOp, key string) []*utils.RestOp {
	for _, del_ssl := range ssl_to_delete {
		ssl_key := avicache.NamespaceName{Namespace: namespace, Name: del_ssl.Name}
		ssl_cache, ok := rest.cache.SSLKeyCache.AviCacheGet(ssl_key)
		if ok {
			ssl_cache_obj, _ := ssl_cache.(*avicache.AviSSLCache)
			restOp := rest.AviSSLKeyCertDel(ssl_cache_obj.Uuid, namespace)
			restOp.ObjName = del_ssl.Name
			rest_ops = append(rest_ops, restOp)
		}
	}
	return rest_ops
}

func Remove(s []avicache.NamespaceName, r avicache.NamespaceName) []avicache.NamespaceName {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

func filterKeyFromStringSlice(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
